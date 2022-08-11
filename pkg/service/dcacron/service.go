package dca

import (
	"context"
	"fmt"
	"math"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/alert"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/eventbus"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/wallet"
	"github.com/asaskevich/EventBus"
	"github.com/dcaf-labs/solana-go-clients/pkg/drip"
	"github.com/dcaf-labs/solana-go-clients/pkg/whirlpool"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type DCACronService struct {
	DCACrons       cmap.ConcurrentMap
	solClient      *rpc.Client
	walletProvider *wallet.WalletProvider
	alertService   alert.Service
	env            configs.Environment
}

type DCACron struct {
	Cron   *cron.Cron
	Config configs.DripConfig
}

func NewDCACron(
	lc fx.Lifecycle,
	config *configs.Config,
	eventBus EventBus.Bus,
	solClient *rpc.Client,
	walletProvider *wallet.WalletProvider,
	alertService alert.Service,
) (*DCACronService, error) {
	logrus.Info("initializing dca cron service")
	dcaCronService := DCACronService{
		DCACrons:       cmap.New(),
		solClient:      solClient,
		walletProvider: walletProvider,
		alertService:   alertService,
		env:            config.Environment,
	}
	// Start this before lifecycle to ensure it is subscribed as soon as invoke is called
	if err := eventBus.Subscribe(string(eventbus.VaultConfigTopic), dcaCronService.createCron); err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			for !dcaCronService.DCACrons.IsEmpty() {
				if config.ShouldDiscoverNewConfigs {
					// Don't return err if this fails
					// we need to stop the cronJobs
					if err := eventBus.Unsubscribe(string(eventbus.VaultConfigTopic), dcaCronService.createCron); err != nil {
						logrus.WithError(err).WithField("bus", eventbus.VaultConfigTopic).Error("failed to unsubscribe to event bus")
					}
				}
				for _, key := range dcaCronService.DCACrons.Keys() {
					v, ok := dcaCronService.DCACrons.Pop(key)
					if !ok {
						continue
					}
					dcaCron := v.(*DCACron)
					if err := dcaCronService.stopCron(ctx, dcaCron.Cron); err != nil {
						logrus.WithError(err).WithField("vault", dcaCron.Config.Vault).Error("failed to stop dca cron job")
					}
				}
			}
			return nil
		},
	})
	return nil, nil
}

// TODO(Mocha): We can cache the vault proto configs
func (dca *DCACronService) createCron(newConfig configs.DripConfig) (*DCACron, error) {
	logrus.WithField("vault", newConfig.Vault).Info("received vault newConfig")
	if v, ok := dca.DCACrons.Get(newConfig.Vault); ok {
		dcaCron := v.(*DCACron)
		// If there is a new whirlpool config, and it's different from what we have, set it
		// If there is a new splTokenSwap config, and it's different from what we have, set it
		if (newConfig.OrcaWhirlpoolConfig.Whirlpool != "" && dcaCron.Config.OrcaWhirlpoolConfig.Whirlpool != newConfig.OrcaWhirlpoolConfig.Whirlpool) ||
			(newConfig.SPLTokenSwapConfig.Swap != "" && dcaCron.Config.SPLTokenSwapConfig.Swap != newConfig.SPLTokenSwapConfig.Swap) {
			logrus.
				WithField("vault", newConfig.Vault).
				WithField("oldSwap", dcaCron.Config.SPLTokenSwapConfig.Swap).
				WithField("newSwap", newConfig.SPLTokenSwapConfig.Swap).
				WithField("oldSwap", dcaCron.Config.SPLTokenSwapConfig.Swap).
				WithField("newSwap", newConfig.SPLTokenSwapConfig.Swap).
				Info("vault already registered, overriding swap")
			dcaCron.Config = newConfig
			dca.DCACrons.Set(newConfig.Vault, dcaCron)
			return dcaCron, nil
		}
		logrus.WithField("vault", newConfig.Vault).Info("vault already registered, skipping cron creation")
		return nil, nil
	}
	logrus.WithField("vault", newConfig.Vault).Info("creating cron")
	var vaultProtoConfigData drip.VaultProtoConfig
	vaultProtoConfigPubKey, err := solana.PublicKeyFromBase58(newConfig.VaultProtoConfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := dca.solClient.GetAccountDataInto(ctx, vaultProtoConfigPubKey, &vaultProtoConfigData); err != nil {
		return nil, err
	}
	logrus.WithField("vaultProtoConfig", fmt.Sprintf("%+v", vaultProtoConfigData)).Info("fetched vault proto newConfig")

	cronJob := cron.New()
	runWithConfig := func() {
		dca.runWithRetry(newConfig.Vault, 0, 5, 1)
	}
	if _, err := cronJob.AddFunc(fmt.Sprintf("@every %ds", vaultProtoConfigData.Granularity), runWithConfig); err != nil {
		return nil, err
	}
	dcaCron := DCACron{
		Config: newConfig,
		Cron:   cronJob,
	}
	dca.DCACrons.Set(newConfig.Vault, &dcaCron)
	// Run the first trigger dca right now in case we created this cron past the lastDCAActivation timestamp
	go runWithConfig()
	dcaCron.Cron.Start()
	return &dcaCron, nil
}

func (dca *DCACronService) stopCron(
	ctx context.Context, cron *cron.Cron,
) error {
	// Stop cron and wait for it to finish or timeout
	timeout := time.Minute
	ticker := time.NewTicker(timeout)
	cronStop := cron.Stop().Done()
	ctxDone := ctx.Done()
	select {
	case <-ticker.C:
		return fmt.Errorf("failed to stop dca cron in %s", timeout.String())
	case <-cronStop:
	case <-ctxDone:
		return nil
	}
	return nil
}

func (dca *DCACronService) runWithRetry(vault string, try, maxTry int, timeout int64) {
	v, ok := dca.DCACrons.Get(vault)
	if !ok {
		logrus.
			WithField("try", try).
			WithField("maxTry", maxTry).
			WithField("timeout", timeout).
			WithField("vault", vault).
			Error("failed to get dcaCron from DCACrons")
		return
	}
	dcaCron := v.(*DCACron)
	config := dcaCron.Config

	defer func() {
		if r := recover(); r != nil {
			_ = dca.alertService.SendError(context.Background(), fmt.Errorf("panic in runWithRetry"))
			logrus.
				WithField("r", r).
				WithField("stackTrace", string(debug.Stack())).
				WithField("config", config).
				WithField("try", try).
				WithField("maxTry", maxTry).
				WithField("timeOut", timeout).
				Errorf("panic in dca cron")
		}
	}()
	if err := dca.run(config); err != nil {
		if try >= maxTry {
			logrus.WithField("try", try).WithField("maxTry", maxTry).WithField("timeout", timeout).Info("failed to DCA with retry")
			if alertErr := dca.alertService.SendError(context.Background(), fmt.Errorf("err in runWithRetry, try %d, maxTry %d, err %w", try, maxTry, err)); alertErr != nil {
				logrus.WithError(err).Errorf("failed to send error alert, alertErr: %s", alertErr)
			}
			return
		}
		logrus.WithError(err).WithField("timeout", timeout).WithField("try", try).Info("waiting before retrying DCA")
		time.Sleep(time.Duration(timeout) * time.Second)
		dca.runWithRetry(config.Vault, try+1, maxTry, timeout*timeout)
	}
}

//nolint:funlen
func (dca *DCACronService) run(dripConfig configs.DripConfig) error {
	logrus.WithField("vault", dripConfig.Vault).Info("preparing trigger dca")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	var vaultData drip.Vault
	vaultPubKey := solana.MustPublicKeyFromBase58(dripConfig.Vault)
	// Use GetAccountInfoWithOpts so we can pass in a commitment level
	resp, err := dca.solClient.GetAccountInfoWithOpts(ctx, vaultPubKey, &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: "confirmed",
		DataSlice:  nil,
	})
	if err != nil {
		return err
	}

	if err := bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&vaultData); err != nil {
		return err
	}

	// Check if Vault can Drip
	if vaultData.DripAmount == 0 {
		logrus.
			WithField("dripAmount", vaultData.DripAmount).
			WithField("vault", dripConfig.Vault).
			Info("exiting, drip amount is 0")
		return nil
	}
	if vaultData.DripActivationTimestamp > time.Now().Unix() {
		logrus.
			WithField("dripActivationTimestamp", time.Unix(vaultData.DripActivationTimestamp, 0).String()).
			WithField("vault", dripConfig.Vault).
			Info("exiting, dca already triggered")
		return nil
	}
	balance, err := dca.solClient.GetTokenAccountBalance(ctx, solana.MustPublicKeyFromBase58(dripConfig.VaultTokenAAccount), rpc.CommitmentConfirmed)
	if err != nil || balance == nil || balance.Value == nil {
		logrus.
			WithError(err).
			WithField("vault", dripConfig.Vault).
			Errorf("failed to fetch vault balance")
		return err
	}
	vaultTokenABalance, err := strconv.ParseUint(balance.Value.Amount, 10, 64)
	if err != nil {
		logrus.
			WithError(err).
			WithField("vault", dripConfig.Vault).
			Errorf("failed to parse vault balance")
		return err
	}
	if vaultTokenABalance == 0 || vaultTokenABalance < vaultData.DripAmount {
		logrus.
			WithField("tokenABalance", balance.Value.Amount).
			WithField("dripAmount", vaultData.DripAmount).
			WithField("vault", dripConfig.Vault).
			Errorf("exiting, token balance is too low")
		return nil
	}

	var instructions []solana.Instruction
	lastVaultPeriod := int64(vaultData.LastDripPeriod)
	vaultPeriodI, instruction, err := dca.fetchVaultPeriod(ctx, vaultPubKey, vaultData.ProtoConfig, vaultData.TokenAMint, vaultData.TokenBMint, lastVaultPeriod)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get vaultPeriodI %d PDA", lastVaultPeriod)
		return err
	}
	if instruction != nil {
		instructions = append(instructions, instruction)
	}
	logrus.WithField("publicKey", vaultPeriodI.String()).Infof("fetched vaultPeriod %d PDA", lastVaultPeriod)

	currentVaultPeriod := lastVaultPeriod + 1
	vaultPeriodJ, instruction, err := dca.fetchVaultPeriod(ctx, vaultPubKey, vaultData.ProtoConfig, vaultData.TokenAMint, vaultData.TokenBMint, currentVaultPeriod)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get vaultPeriodJ %d PDA", currentVaultPeriod)
		return err
	}
	if instruction != nil {
		instructions = append(instructions, instruction)
	}
	logrus.WithField("publicKey", vaultPeriodJ.String()).Infof("fetched vaultPeriod %d PDA", currentVaultPeriod)

	botTokenAFeeAccount, instruction, err := dca.fetchBotTokenAFeeAccount(ctx, vaultData)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get botTokenAFeeAccount")
		return err
	}
	if instruction != nil {
		instructions = append(instructions, instruction)
	}
	logrus.WithField("publicKey", botTokenAFeeAccount.String()).Infof("fetched botTokenAFeeAccount")

	switch {
	case dripConfig.OrcaWhirlpoolConfig.Whirlpool != "":
		newInstructions, err := dca.dripOrcaWhirlpool(ctx, dripConfig, vaultData, vaultPeriodI, vaultPeriodJ, botTokenAFeeAccount)
		if err != nil {
			return err
		}
		instructions = append(instructions, newInstructions...)
	case dripConfig.SPLTokenSwapConfig.Swap != "":
		newInstructions, err := dca.dripSplTokenSwap(ctx, dripConfig, vaultData, vaultPeriodI, vaultPeriodJ, botTokenAFeeAccount)
		if err != nil {
			return err
		}
		instructions = append(instructions, newInstructions...)
	default:
		logrus.WithField("vault", dripConfig.Vault).Infof("missing drip config")
	}
	if err := dca.walletProvider.Send(ctx, instructions...); err != nil {
		logrus.
			WithField("vault", dripConfig.Vault).
			WithField("numInstructions", len(instructions)).
			WithError(err).
			Errorf("failed to trigger dca")
		return err
	}
	logrus.
		WithFields(logrus.Fields{"vault": dripConfig.Vault}).
		Info("processed drip")
	return nil
}

func (dca *DCACronService) dripOrcaWhirlpool(
	ctx context.Context,
	dripConfig configs.DripConfig,
	vaultData drip.Vault,
	vaultPeriodI solana.PublicKey,
	vaultPeriodJ solana.PublicKey,
	botTokenAFeeAccount solana.PublicKey,
) ([]solana.Instruction, error) {
	var instructions []solana.Instruction
	// Get WhirlpoolsConfig
	whirlpoolPubkey := solana.MustPublicKeyFromBase58(dripConfig.OrcaWhirlpoolConfig.Whirlpool)
	resp, err := dca.solClient.GetAccountInfoWithOpts(ctx, whirlpoolPubkey, &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: "confirmed",
		DataSlice:  nil,
	})
	if err != nil {
		return []solana.Instruction{}, err
	}
	var whirlpoolData whirlpool.Whirlpool
	if err := bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&whirlpoolData); err != nil {
		return []solana.Instruction{}, err
	}

	if err := dca.ensureTickArrays(ctx, dripConfig, vaultData, whirlpoolData); err != nil {
		return []solana.Instruction{}, err
	}
	quoteEstimate, err := wallet.GetOrcaWhirlpoolQuoteEstimate(
		whirlpoolData.WhirlpoolsConfig.String(),
		whirlpoolData.TokenMintA.String(),
		whirlpoolData.TokenMintB.String(),
		vaultData.TokenAMint.String(),
		whirlpoolData.TickSpacing,
		dca.env,
	)
	if err != nil {
		return []solana.Instruction{}, err
	}
	logrus.WithFields(logrus.Fields{
		"vault":               dripConfig.Vault,
		"tokenAMint":          vaultData.TokenAMint.String(),
		"tokenBMint":          vaultData.TokenBMint.String(),
		"swapTokenAAcount":    dripConfig.OrcaWhirlpoolConfig.SwapTokenAAccount,
		"swapTokenBAccount":   dripConfig.OrcaWhirlpoolConfig.SwapTokenBAccount,
		"i":                   vaultData.LastDripPeriod,
		"j":                   vaultData.LastDripPeriod + 1,
		"vaultPeriodI":        vaultPeriodI.String(),
		"vaultPeriodJ":        vaultPeriodJ.String(),
		"botTokenAFeeAccount": botTokenAFeeAccount.String(),
	}).Info("running drip")

	instruction, err := dca.walletProvider.DripOrcaWhirlpool(ctx,
		wallet.DripOrcaWhirlpoolParams{
			VaultData:           vaultData,
			Vault:               solana.MustPublicKeyFromBase58(dripConfig.Vault),
			VaultPeriodI:        vaultPeriodI,
			VaultPeriodJ:        vaultPeriodJ,
			BotTokenAFeeAccount: botTokenAFeeAccount,
			WhirlpoolData:       whirlpoolData,
			Whirlpool:           whirlpoolPubkey,
			TickArray0:          solana.MustPublicKeyFromBase58(quoteEstimate.TickArray0),
			TickArray1:          solana.MustPublicKeyFromBase58(quoteEstimate.TickArray1),
			TickArray2:          solana.MustPublicKeyFromBase58(quoteEstimate.TickArray2),
			Oracle:              solana.MustPublicKeyFromBase58(dripConfig.OrcaWhirlpoolConfig.Oracle),
		},
	)
	if err != nil {
		logrus.
			WithError(err).
			WithField("dcaProgram", drip.ProgramID.String()).
			Errorf("failed to create DripSPLTokenSwap instruction")
		return []solana.Instruction{}, err
	}
	instructions = append(instructions, instruction)

	return instructions, nil
}

func (dca *DCACronService) ensureTickArrays(
	ctx context.Context,
	dripConfig configs.DripConfig,
	vault drip.Vault,
	whirlpoolData whirlpool.Whirlpool,
) error {
	whirlpoolPubkey := solana.MustPublicKeyFromBase58(dripConfig.OrcaWhirlpoolConfig.Whirlpool)
	var instructions []solana.Instruction
	realIndex := math.Floor(float64(whirlpoolData.TickCurrentIndex) / float64(whirlpoolData.TickSpacing) / 88.0)
	startTickIndex := int32(realIndex) * int32(whirlpoolData.TickSpacing) * 88

	aToB := vault.TokenAMint.String() == whirlpoolData.TokenMintA.String()
	var tickArrayIndexs []int32
	if aToB {
		tickArrayIndexs = []int32{
			startTickIndex,
			startTickIndex - int32(whirlpoolData.TickSpacing*88)*1,
			startTickIndex - int32(whirlpoolData.TickSpacing*88)*2,
		}
	} else {
		tickArrayIndexs = []int32{
			startTickIndex,
			startTickIndex + int32(whirlpoolData.TickSpacing*88)*1,
			startTickIndex + int32(whirlpoolData.TickSpacing*88)*2,
		}
	}
	for _, tickArrayIndex := range tickArrayIndexs {
		tickArrayPubkey, _, _ := solana.FindProgramAddress([][]byte{
			[]byte("tick_array"),
			whirlpoolPubkey[:],
			[]byte(strconv.FormatInt(int64(tickArrayIndex), 10)),
		}, whirlpool.ProgramID)
		// Use GetAccountInfoWithOpts so we can pass in a commitment level
		if _, err := dca.solClient.GetAccountInfoWithOpts(ctx, tickArrayPubkey, &rpc.GetAccountInfoOpts{
			Encoding:   solana.EncodingBase64,
			Commitment: "confirmed",
			DataSlice:  nil,
		}); err != nil && err.Error() == "not found" {
			initTickArrayInstruction, err := dca.walletProvider.InitializeTickArray(ctx,
				wallet.InitializeTickArrayParams{
					Whirlpool:  whirlpoolPubkey,
					StartIndex: tickArrayIndex,
					TickArray:  tickArrayPubkey,
				})
			if err != nil {
				logrus.
					WithError(err).
					Errorf("failed to create InitializeTickArrayParams instruction")
				return err
			}
			instructions = append(instructions, initTickArrayInstruction)
		}
	}
	if err := dca.walletProvider.Send(ctx, instructions...); err != nil {
		logrus.
			WithError(err).
			WithField("whirlpool", dripConfig.OrcaWhirlpoolConfig.Whirlpool).
			WithField("numInstructions", len(instructions)).
			Errorf("failed to initialize tick arrays")
		return err
	}
	logrus.
		WithField("whirlpool", dripConfig.OrcaWhirlpoolConfig.Whirlpool).
		WithField("numInstructions", len(instructions)).
		Info("initialized tick arrays")
	return nil
}

func (dca *DCACronService) dripSplTokenSwap(
	ctx context.Context,
	dripConfig configs.DripConfig,
	vaultData drip.Vault,
	vaultPeriodI solana.PublicKey,
	vaultPeriodJ solana.PublicKey,
	botTokenAFeeAccount solana.PublicKey,
) ([]solana.Instruction, error) {
	var instructions []solana.Instruction
	swapTokenAAccount, swapTokenBAccount, err := dca.fetchSplTokenSwapTokenAccounts(ctx, dripConfig)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get swap token accounts")
		return []solana.Instruction{}, err
	}
	dripConfig.SPLTokenSwapConfig.SwapTokenAAccount = swapTokenAAccount
	dripConfig.SPLTokenSwapConfig.SwapTokenBAccount = swapTokenBAccount

	logrus.WithFields(logrus.Fields{
		"vault":               dripConfig.Vault,
		"tokenAMint":          vaultData.TokenAMint.String(),
		"tokenBMint":          vaultData.TokenBMint.String(),
		"swapTokenAAcount":    dripConfig.SPLTokenSwapConfig.SwapTokenAAccount,
		"swapTokenBAccount":   dripConfig.SPLTokenSwapConfig.SwapTokenBAccount,
		"i":                   vaultData.LastDripPeriod,
		"j":                   vaultData.LastDripPeriod + 1,
		"vaultPeriodI":        vaultPeriodI.String(),
		"vaultPeriodJ":        vaultPeriodJ.String(),
		"botTokenAFeeAccount": botTokenAFeeAccount.String(),
	}).Info("running drip")

	instruction, err := dca.walletProvider.DripSPLTokenSwap(ctx, dripConfig, vaultPeriodI, vaultPeriodJ, botTokenAFeeAccount)
	if err != nil {
		logrus.
			WithError(err).
			WithField("dcaProgram", drip.ProgramID.String()).
			Errorf("failed to create DripSPLTokenSwap instruction")
		return []solana.Instruction{}, err
	}
	instructions = append(instructions, instruction)

	return instructions, nil
}

func (dca *DCACronService) fetchVaultPeriod(
	ctx context.Context,
	vault, vaultProtoConfig, tokenAMint, tokenBMint solana.PublicKey,
	vaultPeriodID int64,
) (solana.PublicKey, solana.Instruction, error) {
	vaultPeriod, _, err := solana.FindProgramAddress([][]byte{
		[]byte("vault_period"),
		vault[:],
		[]byte(strconv.FormatInt(vaultPeriodID, 10)),
	}, drip.ProgramID)
	if err != nil {
		logrus.
			WithError(err).
			WithField("dcaProgram", drip.ProgramID.String()).
			WithField("vaultPeriodID", vaultPeriodID).
			Errorf("failed to get vaultPeriodI PDA")
		return solana.PublicKey{}, nil, err
	}
	var instruction solana.Instruction
	// Use GetAccountInfoWithOpts so we can pass in a commitment level
	if resp, err := dca.solClient.GetAccountInfoWithOpts(ctx, vaultPeriod, &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: "confirmed",
		DataSlice:  nil,
	}); err != nil {
		// Failure is likely because the vault period is not initialized
		instruction, err = dca.walletProvider.InitVaultPeriod(ctx, vault.String(), vaultProtoConfig.String(), vaultPeriod.String(), tokenAMint.String(), tokenBMint.String(), vaultPeriodID)
		if err != nil {
			logrus.
				WithError(err).
				WithField("dcaProgram", drip.ProgramID.String()).
				WithField("vaultPeriodID", vaultPeriodID).
				Errorf("failed to create InitVaultPeriod instruction")
			return solana.PublicKey{}, nil, err
		}
	} else {
		var vaultPeriodData drip.VaultPeriod
		if err := bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&vaultPeriodData); err != nil {
			return solana.PublicKey{}, nil, err
		}
		logrus.
			WithField("vaultPeriodID", vaultPeriodID).
			WithField("dar", vaultPeriodData.Dar).
			WithField("twap", vaultPeriodData.Twap).
			Infof("fetched vault period")
	}
	return vaultPeriod, instruction, nil
}

func (dca *DCACronService) fetchBotTokenAFeeAccount(
	ctx context.Context, vault drip.Vault,
) (solana.PublicKey, solana.Instruction, error) {
	botTokenAAccount, _, err := solana.FindAssociatedTokenAddress(
		dca.walletProvider.FeeWalletPubkey,
		vault.TokenAMint,
	)
	if err != nil {
		logrus.
			WithError(err).
			WithField("dcaProgram", drip.ProgramID.String()).
			WithField("feeWallet", dca.walletProvider.FeeWalletPubkey.String()).
			WithField("mint", vault.TokenAMint.String()).
			Errorf("failed to get botTokenAAccount")
		return solana.PublicKey{}, nil, err
	}
	var instruction solana.Instruction

	if resp, err := dca.solClient.GetTokenAccountBalance(ctx, botTokenAAccount, "confirmed"); err != nil {
		instruction, err = dca.walletProvider.CreateTokenAccount(ctx, vault.TokenAMint.String())
		if err != nil {
			logrus.
				WithError(err).
				WithField("dcaProgram", drip.ProgramID.String()).
				WithField("mint", vault.TokenAMint.String()).
				Errorf("failed to create createTokenAccount instruction")
			return solana.PublicKey{}, nil, err
		}
	} else {
		logrus.
			WithField("botTokenAAccount", botTokenAAccount.String()).
			WithField("balance", resp.Value.Amount).
			WithField("decimals", resp.Value.Decimals).
			Info("fetched vault period")
	}
	return botTokenAAccount, instruction, nil
}

func (dca *DCACronService) fetchSplTokenSwapTokenAccounts(
	ctx context.Context,
	config configs.DripConfig,
) (string, string, error) {
	// Fetch Token A
	resp, err := dca.solClient.GetAccountInfoWithOpts(ctx, solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapTokenAAccount), &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: "confirmed",
		DataSlice:  nil,
	})
	if err != nil {
		return "", "", err
	}
	var swapTokenAAccount token.Account
	if err := bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&swapTokenAAccount); err != nil {
		return "", "", err
	}

	// Fetch token B
	resp, err = dca.solClient.GetAccountInfoWithOpts(ctx, solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapTokenBAccount), &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: "confirmed",
		DataSlice:  nil,
	})
	if err != nil {
		return "", "", err
	}
	var swapTokenBAccount token.Account
	if err := bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&swapTokenBAccount); err != nil {
		return "", "", err
	}

	if swapTokenAAccount.Mint.String() == config.SPLTokenSwapConfig.TokenAMint && swapTokenBAccount.Mint.String() == config.SPLTokenSwapConfig.TokenBMint {
		// Normal A -> b
		return config.SPLTokenSwapConfig.SwapTokenAAccount, config.SPLTokenSwapConfig.SwapTokenBAccount, nil
	} else if swapTokenAAccount.Mint.String() == config.SPLTokenSwapConfig.TokenBMint && swapTokenBAccount.Mint.String() == config.SPLTokenSwapConfig.TokenAMint {
		// Need to swap token accounts for inverse
		return config.SPLTokenSwapConfig.SwapTokenBAccount, config.SPLTokenSwapConfig.SwapTokenAAccount, nil
	}
	err = fmt.Errorf("token swap token accounts do not match config mints, or the inverse of the config mints")
	logrus.
		WithField("swapTokenAAccount", config.SPLTokenSwapConfig.SwapTokenAAccount).
		WithField("swapTokenBAccount", config.SPLTokenSwapConfig.SwapTokenBAccount).
		WithField("swapTokenAMint", swapTokenAAccount.Mint.String()).
		WithField("swapTokenBMint", swapTokenBAccount.Mint.String()).
		WithField("configTokenAMint", config.SPLTokenSwapConfig.TokenAMint).
		WithField("configTokenBMint", config.SPLTokenSwapConfig.TokenBMint).
		WithField("vault", config.Vault).
		WithError(err).
		Error("failed to get swap token accounts")
	return "", "", err
}

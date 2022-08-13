package dca

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/alert"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/eventbus"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/wallet"
	"github.com/asaskevich/EventBus"
	"github.com/dcaf-labs/solana-go-clients/pkg/drip"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type DripScheduler struct {
	At    time.Time
	Every time.Duration
}

func (s *DripScheduler) Next(t time.Time) time.Time {
	if t.After(s.At) {
		return t.Add(s.Every)
	}
	return s.At
}

func newScheduler(granularity uint64) DripScheduler {
	return DripScheduler{time.Now().Add(-1 * time.Duration(time.Now().Unix()%int64(granularity))), time.Second * time.Duration(granularity)}
}

const ErrDripAmount0 = "drip amount is 0"
const ErrDripAlreadyTriggered = "drip already triggered"

type DCACronService struct {
	DripConfigs    cmap.ConcurrentMap
	solClient      *rpc.Client
	walletProvider *wallet.WalletProvider
	alertService   alert.Service
	env            configs.Environment
}

type DripConfig struct {
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
		DripConfigs:    cmap.New(),
		solClient:      solClient,
		walletProvider: walletProvider,
		alertService:   alertService,
		env:            config.Environment,
	}
	// Start this before lifecycle to ensure it is subscribed as soon as invoke is called
	if err := eventBus.Subscribe(string(eventbus.VaultConfigTopic), dcaCronService.registerDripConfig); err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			for !dcaCronService.DripConfigs.IsEmpty() {
				if config.ShouldDiscoverNewConfigs {
					// Don't return err if this fails
					// we need to stop the cronJobs
					if err := eventBus.Unsubscribe(string(eventbus.VaultConfigTopic), dcaCronService.registerDripConfig); err != nil {
						logrus.WithError(err).WithField("bus", eventbus.VaultConfigTopic).Error("failed to unsubscribe to event bus")
					}
				}
				for _, key := range dcaCronService.DripConfigs.Keys() {
					v, ok := dcaCronService.DripConfigs.Pop(key)
					if !ok {
						continue
					}
					dcaCron := v.(*DripConfig)
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

func (dca *DCACronService) registerDripConfig(newConfig configs.DripConfig) (*DripConfig, error) {
	logrus.WithField("vault", newConfig.Vault).Info("received new config")

	if v, ok := dca.DripConfigs.Get(newConfig.Vault); ok {
		dripConfig := v.(*DripConfig)
		// If there is a new whirlpool config, and it's different from what we have, set it
		// If there is a new splTokenSwap config, and it's different from what we have, set it
		if (newConfig.OrcaWhirlpoolConfig.Whirlpool != "" && dripConfig.Config.OrcaWhirlpoolConfig.Whirlpool != newConfig.OrcaWhirlpoolConfig.Whirlpool) ||
			(newConfig.SPLTokenSwapConfig.Swap != "" && dripConfig.Config.SPLTokenSwapConfig.Swap != newConfig.SPLTokenSwapConfig.Swap) {
			logrus.
				WithField("vault", newConfig.Vault).
				WithField("oldSwap", dripConfig.Config.SPLTokenSwapConfig.Swap).
				WithField("newSwap", newConfig.SPLTokenSwapConfig.Swap).
				WithField("oldSwap", dripConfig.Config.SPLTokenSwapConfig.Swap).
				WithField("newSwap", newConfig.SPLTokenSwapConfig.Swap).
				Info("vault already registered, overriding swap")
			dripConfig.Config = newConfig
			dca.DripConfigs.Set(newConfig.Vault, dripConfig)
			return dripConfig, nil
		}
		logrus.WithField("vault", newConfig.Vault).Info("vault already registered, skipping cron creation")
		return nil, nil
	}
	logrus.WithField("vault", newConfig.Vault).Info("creating cron")
	return dca.scheduleDrip(newConfig)
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
	v, ok := dca.DripConfigs.Get(vault)
	if !ok {
		logrus.
			WithField("vault", vault).
			Error("failed to get dcaCron from DripConfigs")
		return
	}
	dripConfig := v.(*DripConfig)
	config := dripConfig.Config

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
				Errorf("panic in doDrip")
		}
	}()
	if err := dca.run(config); err != nil {
		if try >= maxTry {
			if err.Error() != ErrDripAmount0 && err.Error() != ErrDripAlreadyTriggered {
				logrus.WithField("try", try).WithField("maxTry", maxTry).WithField("timeout", timeout).Info("failed to drip")
				if alertErr := dca.alertService.SendError(context.Background(), fmt.Errorf("err in runWithRetry, try %d, maxTry %d, err %w", try, maxTry, err)); alertErr != nil {
					logrus.WithError(err).Errorf("failed to send error alert, alertErr: %s", alertErr)
				}
			}
			// first stop the current cron to avoid mem leaks
			if err := dca.stopCron(context.Background(), dripConfig.Cron); err != nil {
				logrus.WithError(err).Error("failed to stop cron job while trying to reschedule")
			}
			// create new drip handler
			if _, err := dca.scheduleDrip(config); err != nil {
				logrus.WithError(err).WithField("vault", config.Vault).Errorf("failed to reschedule drip")
			}
			return
		}
		logrus.WithError(err).WithField("timeout", timeout).WithField("try", try).Info("waiting before retrying DCA")
		time.Sleep(time.Duration(timeout) * time.Second)
		dca.runWithRetry(config.Vault, try+1, maxTry, timeout*timeout)
	}
}

func (dca *DCACronService) getSchedulerFromProtoConfig(address string) (DripScheduler, uint64, error) {
	protoConfigPubkey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return DripScheduler{}, 0, err
	}
	var vaultProtoConfigData drip.VaultProtoConfig
	if err := dca.solClient.GetAccountDataInto(context.Background(), protoConfigPubkey, &vaultProtoConfigData); err != nil {
		return DripScheduler{}, 0, err
	}
	return newScheduler(vaultProtoConfigData.Granularity), vaultProtoConfigData.Granularity, nil
}

func (dca *DCACronService) getDripFunc(vault string) func() {
	return func() {
		dca.runWithRetry(vault, 0, 5, 2)
	}
}

func (dca *DCACronService) scheduleDrip(config configs.DripConfig) (*DripConfig, error) {
	schedule, granularity, err := dca.getSchedulerFromProtoConfig(config.VaultProtoConfig)
	if err != nil {
		logrus.WithError(err).Errorf("failed to getSchedulerFromProtoConfig")
		return nil, err
	}
	cronJob := cron.New()
	doDrip := dca.getDripFunc(config.Vault)
	if _, err := cronJob.AddFunc(fmt.Sprintf("@every %ds", granularity), doDrip); err != nil {
		logrus.WithError(err).Errorf("failed to addFunc to cronJob while trying to reschedule")
		return nil, err
	}
	cronJob.Schedule(&schedule, cronJob)
	newDripConfig := DripConfig{
		Config: config,
		Cron:   cronJob,
	}
	dca.DripConfigs.Set(config.Vault, &newDripConfig)
	// Run the first trigger dca right now in case we created this cron past the lastDCAActivation timestamp
	go doDrip()
	newDripConfig.Cron.Start()
	logrus.WithField("vault", config.Vault).Info("scheduled doDrip")
	return &newDripConfig, nil
}

// TODO(Mocha): this function shouldn't be this long
// it should just verify the vault can drip
// then orchestrate instructions, then send them
//nolint:funlen
func (dca *DCACronService) run(dripConfig configs.DripConfig) error {
	logrus.WithField("vault", dripConfig.Vault).Info("preparing trigger dca")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()
	// TODO(Mocha): consider moving this vault verification logic somewhere else
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
		return errors.New(ErrDripAmount0)
	}
	if vaultData.DripActivationTimestamp > time.Now().Unix() {
		logrus.
			WithField("dripActivationTimestamp", time.Unix(vaultData.DripActivationTimestamp, 0).String()).
			WithField("vault", dripConfig.Vault).
			Info("exiting, drip already triggered")
		return errors.New(ErrDripAlreadyTriggered)
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
		return fmt.Errorf("token balance is too low")
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

// TODO(Mocha): consider moving this fn somewhere else
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

// TODO(Mocha): consider moving this fn somewhere else
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
			Errorf("failed to get botTokenAAccount address")
		return solana.PublicKey{}, nil, err
	}
	var instruction solana.Instruction

	// Use GetAccountInfoWithOpts so we can pass in a commitment level
	if _, err := dca.solClient.GetAccountInfoWithOpts(ctx, botTokenAAccount, &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: "confirmed",
		DataSlice:  nil,
	}); err != nil && err.Error() == "not found" {
		instruction, err = dca.walletProvider.CreateTokenAccount(ctx, dca.walletProvider.FeeWalletPubkey, vault.TokenAMint)
		if err != nil {
			return solana.PublicKey{}, nil, err
		}
	}
	return botTokenAAccount, instruction, nil
}

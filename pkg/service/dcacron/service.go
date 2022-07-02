package dca

import (
	"context"
	"fmt"
	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/Dcaf-Protocol/drip-keeper/generated/drip"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/eventbus"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/wallet"
	"github.com/asaskevich/EventBus"
	"github.com/gagliardetto/solana-go"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/robfig/cron/v3"
	"runtime/debug"
	"strconv"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type DCACronService struct {
	DCACrons       cmap.ConcurrentMap
	solClient      *rpc.Client
	walletProvider *wallet.WalletProvider
}

type DCACron struct {
	Cron   *cron.Cron
	Config configs.TriggerDCAConfig
}

func NewDCACron(
	lc fx.Lifecycle,
	config *configs.Config,
	eventBus EventBus.Bus,
	solClient *rpc.Client,
	walletProvider *wallet.WalletProvider,
) (*DCACronService, error) {
	logrus.Info("initializing dca cron service")
	dcaCronService := DCACronService{
		DCACrons:       cmap.New(),
		solClient:      solClient,
		walletProvider: walletProvider,
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
func (dca *DCACronService) createCron(config configs.TriggerDCAConfig) (*DCACron, error) {
	logrus.WithField("vault", config.Vault).Info("received vault config")
	if _, ok := dca.DCACrons.Get(config.Vault); ok {
		logrus.WithField("vault", config.Vault).Info("vault already registered, skipping cron creation")
		return nil, nil
	}
	logrus.WithField("vault", config.Vault).Info("creating cron")
	var vaultProtoConfigData drip.VaultProtoConfig
	vaultProtoConfigPubKey, err := solana.PublicKeyFromBase58(config.VaultProtoConfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := dca.solClient.GetAccountDataInto(ctx, vaultProtoConfigPubKey, &vaultProtoConfigData); err != nil {
		return nil, err
	}
	logrus.WithField("vaultProtoConfig", fmt.Sprintf("%+v", vaultProtoConfigData)).Info("fetched vault proto config")

	cronJob := cron.New()
	runWithConfig := func() {
		dca.runWithRetry(config, 0, 5, 1)
	}
	// Run the first trigger dca right now and schedule the rest in the future
	runWithConfig()
	if _, err := cronJob.AddFunc(fmt.Sprintf("@every %ds", vaultProtoConfigData.Granularity), runWithConfig); err != nil {
		return nil, err
	}
	dcaCron := DCACron{
		Config: config,
		Cron:   cronJob,
	}
	dca.DCACrons.Set(config.Vault, &dcaCron)
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

func (dca *DCACronService) runWithRetry(config configs.TriggerDCAConfig, try, maxTry int, timeout int64) {
	defer func() {
		if r := recover(); r != nil {
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
			return
		}
		logrus.WithError(err).WithField("timeout", timeout).WithField("try", try).Info("waiting before retrying DCA")
		time.Sleep(time.Duration(timeout) * time.Second)
		dca.runWithRetry(config, try+1, maxTry, timeout*timeout)
	}
}

func (dca *DCACronService) run(config configs.TriggerDCAConfig) error {
	logrus.WithField("vault", config.Vault).Info("preparing trigger dca")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var vaultData drip.Vault
	vaultPubKey := solana.MustPublicKeyFromBase58(config.Vault)
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
	if vaultData.DripAmount == 0 {
		logrus.
			WithField("dripAmount", vaultData.DripAmount).
			WithField("vault", config.Vault).
			Info("exiting, drip amount is 0")
		return nil
	}
	if vaultData.DcaActivationTimestamp > time.Now().Unix() {
		logrus.
			WithField("DcaActivationTimestamp", time.Unix(vaultData.DcaActivationTimestamp, 0).String()).
			WithField("vault", config.Vault).
			Info("exiting, dca already triggered")
		return nil
	}

	balance, err := dca.solClient.GetTokenAccountBalance(ctx, solana.MustPublicKeyFromBase58(config.VaultTokenAAccount), rpc.CommitmentConfirmed)
	if err != nil || balance == nil || balance.Value == nil {
		logrus.
			WithError(err).
			WithField("vault", config.Vault).
			Errorf("failed to fetch vault balance")
		return err
	}

	vaultTokenABalance, err := strconv.ParseUint(balance.Value.Amount, 10, 64)
	if err != nil {
		logrus.
			WithError(err).
			WithField("vault", config.Vault).
			Errorf("failed to parse vault balance")
		return err
	}
	// Check if vault is empty
	if vaultTokenABalance == 0 || vaultTokenABalance < vaultData.DripAmount {
		logrus.
			WithField("tokenABalance", balance.Value.Amount).
			WithField("dripAmount", vaultData.DripAmount).
			WithField("vault", config.Vault).
			Errorf("exiting, token balance is too low")
		return nil
	}
	logrus.
		WithField("balance", balance.Value.Amount).
		WithField("vault", config.Vault).
		WithField("tokenAMint", config.TokenAMint).
		WithField("vaultTokenAAcount", config.VaultTokenAAccount).
		Info("fetched vault token a balance")

	var instructions []solana.Instruction
	lastVaultPeriod := int64(vaultData.LastDcaPeriod)
	vaultPeriodI, instruction, err := dca.fetchVaultPeriod(ctx, config, vaultPubKey, lastVaultPeriod)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get vaultPeriodI %d PDA", lastVaultPeriod)
		return err
	}
	if instruction != nil {
		instructions = append(instructions, instruction)
	}
	logrus.WithField("publicKey", vaultPeriodI.String()).Infof("fetched vaultPeriod %d PDA", lastVaultPeriod)

	currentVaultPeriod := lastVaultPeriod + 1
	vaultPeriodJ, instruction, err := dca.fetchVaultPeriod(ctx, config, vaultPubKey, currentVaultPeriod)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get vaultPeriodJ %d PDA", currentVaultPeriod)
		return err
	}
	if instruction != nil {
		instructions = append(instructions, instruction)
	}
	logrus.WithField("publicKey", vaultPeriodJ.String()).Infof("fetched vaultPeriod %d PDA", currentVaultPeriod)

	botTokenAAccount, instruction, err := dca.fetchBotTokenAAccount(ctx, config.TokenAMint)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get botTokenAAccount")
		return err
	}
	if instruction != nil {
		instructions = append(instructions, instruction)
	}
	logrus.WithField("publicKey", botTokenAAccount.String()).Infof("fetched botTokenAAccount")

	logrus.WithFields(logrus.Fields{
		"vault":            config.Vault,
		"tokenAMint":       config.TokenAMint,
		"tokenBMint":       config.TokenBMint,
		"i":                vaultData.LastDcaPeriod,
		"j":                vaultData.LastDcaPeriod + 1,
		"vaultPeriodI":     vaultPeriodI.String(),
		"vaultPeriodJ":     vaultPeriodJ.String(),
		"botTokenAAccount": botTokenAAccount.String(),
	}).Info("running dca")

	instruction, err = dca.walletProvider.TriggerDCA(ctx, config, vaultPeriodI, vaultPeriodJ, botTokenAAccount)
	if err != nil {
		logrus.
			WithError(err).
			WithField("dcaProgram", drip.ProgramID.String()).
			Errorf("failed to create TriggerDCA instruction")
		return err
	}
	instructions = append(instructions, instruction)
	if err := dca.walletProvider.Send(ctx, instructions...); err != nil {
		logrus.
			WithField("vault", config.Vault).
			WithField("numInstructions", len(instructions)).
			WithError(err).
			Errorf("failed to trigger dca")
		return err
	}
	logrus.
		WithFields(logrus.Fields{"vault": config.Vault}).
		Info("triggered DCA")
	return nil
}

func (dca *DCACronService) fetchVaultPeriod(
	ctx context.Context, config configs.TriggerDCAConfig, vaultPubKey solana.PublicKey, vaultPeriodID int64,
) (solana.PublicKey, solana.Instruction, error) {
	vaultPeriod, _, err := solana.FindProgramAddress([][]byte{
		[]byte("vault_period"),
		vaultPubKey[:],
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
		instruction, err = dca.walletProvider.InitVaultPeriod(ctx, config, vaultPeriod, vaultPeriodID)
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

func (dca *DCACronService) fetchBotTokenAAccount(
	ctx context.Context, tokenAMint string,
) (solana.PublicKey, solana.Instruction, error) {
	botTokenAAccount, _, err := solana.FindAssociatedTokenAddress(
		dca.walletProvider.Wallet.PublicKey(),
		solana.MustPublicKeyFromBase58(tokenAMint),
	)
	if err != nil {
		logrus.
			WithError(err).
			WithField("dcaProgram", drip.ProgramID.String()).
			WithField("mint", tokenAMint).
			Errorf("failed to get botTokenAAccount")
		return solana.PublicKey{}, nil, err
	}
	var instruction solana.Instruction

	if resp, err := dca.solClient.GetTokenAccountBalance(ctx, botTokenAAccount, "confirmed"); err != nil {
		instruction, err = dca.walletProvider.CreateTokenAccount(ctx, tokenAMint)
		if err != nil {
			logrus.
				WithError(err).
				WithField("dcaProgram", drip.ProgramID.String()).
				WithField("mint", tokenAMint).
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

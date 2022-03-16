package dca

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Dcaf-Protocol/keeper-bot/configs"
	dcaVault "github.com/Dcaf-Protocol/keeper-bot/generated/dca_vault"
	"github.com/Dcaf-Protocol/keeper-bot/pkg/wallet"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type DCACronService struct {
	DCACrons       []DCACron
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
	solClient *rpc.Client,
	walletProvider *wallet.WalletProvider,
) (*DCACronService, error) {

	dcaCronService := DCACronService{walletProvider: walletProvider, solClient: solClient}
	var dcaCrons []DCACron

	for i := range config.TriggerDCAConfigs {
		config := config.TriggerDCAConfigs[i]
		cron, err := dcaCronService.createCron(config)
		if err != nil {
			logrus.WithError(err).WithField("vault", config.Vault).Error("failed to create dca cron job")
			continue
		}
		dcaCron := DCACron{
			Config: config,
			Cron:   cron,
		}
		dcaCrons = append(dcaCrons, dcaCron)
	}
	dcaCronService.DCACrons = dcaCrons

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			for i := range dcaCronService.DCACrons {
				dcaCron := dcaCronService.DCACrons[i]
				dcaCron.Cron.Start()
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			for i := range dcaCronService.DCACrons {
				dcaCron := dcaCronService.DCACrons[i]
				if err := dcaCronService.stopCron(ctx, dcaCron.Cron); err != nil {
					logrus.WithError(err).WithField("vault", dcaCron.Config.Vault).Error("failed to stop dca cron job")
				}
			}
			return nil
		},
	})
	return nil, nil
}

// TODO(Mocha): We can cache the vault proto configs
func (dca *DCACronService) createCron(config configs.TriggerDCAConfig) (*cron.Cron, error) {
	var vaultProtoConfigData dcaVault.VaultProtoConfig
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

	cron := cron.New()
	runWithConfig := func() {
		dca.runWithRetry(config, 0, 5, 1)
	}
	if _, err := cron.AddFunc(fmt.Sprintf("@every %ds", vaultProtoConfigData.Granularity), runWithConfig); err != nil {
		return nil, err
	}
	return cron, nil
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
	if err := dca.run(config); err != nil {
		if try >= maxTry {
			logrus.WithField("tries", try).Info("failed to DCA with retry")
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

	var vaultData dcaVault.Vault
	vaultPubKey, err := solana.PublicKeyFromBase58(config.Vault)
	if err != nil {
		logrus.WithError(err).Errorf("failed to create vault pubkey from base58 string")
		return err
	}

	// Use GetAccountInfoWithOpts so we can pass in a commitment level
	resp, err := dca.solClient.GetAccountInfoWithOpts(ctx, vaultPubKey, &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: "confirmed",
		DataSlice:  nil,
	})
	if err != nil {
		logrus.WithError(err).Errorf("failed to get vault account info")
		return err
	}
	if err := bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&vaultData); err != nil {
		logrus.WithError(err).Errorf("failed to decode vault account data")
	}
	logrus.WithFields(
		logrus.Fields{
			"vaultData": fmt.Sprintf("%+v", vaultData),
		}).Infof("fetched vault")

	lastVaultPeriod := int64(vaultData.LastDcaPeriod)
	vaultPeriodI, _, err := solana.FindProgramAddress([][]byte{
		[]byte("vault_period"),
		vaultPubKey[:],
		[]byte(strconv.FormatInt(lastVaultPeriod, 10)),
	}, dcaVault.ProgramID)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get vaultPeriodI %d PDA", lastVaultPeriod)
		return err
	}
	logrus.WithField("publicKey", vaultPeriodI.String()).Infof("fetched vaultPeriod %d PDA", lastVaultPeriod)

	currentVaultPeriod := lastVaultPeriod + 1
	vaultPeriodJ, _, err := solana.FindProgramAddress([][]byte{
		[]byte("vault_period"),
		vaultPubKey[:],
		[]byte(strconv.FormatInt(currentVaultPeriod, 10)),
	}, dcaVault.ProgramID)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get vaultPeriodJ %d PDA", currentVaultPeriod)
		return err
	}
	logrus.WithField("publicKey", vaultPeriodJ.String()).Infof("fetched vaultPeriod %d PDA", currentVaultPeriod)

	logrus.WithFields(logrus.Fields{
		"vault":        config.Vault,
		"tokenAMint":   config.TokenAMint,
		"tokenBMint":   config.TokenBMint,
		"i":            vaultData.LastDcaPeriod,
		"j":            vaultData.LastDcaPeriod + 1,
		"vaultPeriodI": vaultPeriodI,
		"vaultPeriodJ": vaultPeriodJ,
	}).Info("running dca")

	if err := dca.walletProvider.TriggerDCA(ctx, config, vaultPeriodI, vaultPeriodJ); err != nil {
		logrus.
			WithFields(logrus.Fields{"vault": config.Vault}).
			WithError(err).
			Errorf("failed to trigger DCA")
		return err
	}
	logrus.
		WithFields(logrus.Fields{"vault": config.Vault}).
		Info("triggered DCA")
	return nil
}

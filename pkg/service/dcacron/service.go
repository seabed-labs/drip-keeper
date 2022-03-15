package dca

import (
	"context"
	"fmt"
	"time"

	"github.com/Dcaf-Protocol/keeper-bot/configs"
	dcaVault "github.com/Dcaf-Protocol/keeper-bot/pkg/generated/dca_vault"
	"github.com/Dcaf-Protocol/keeper-bot/pkg/wallet"
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

func (dca *DCACronService) createCron(config configs.TriggerDCAConfig) (*cron.Cron, error) {
	var vaultProtoConfigData dcaVault.VaultProtoConfig
	vaultProtoConfigPubKey, err := solana.PublicKeyFromBase58(config.VaultProtoConfig)
	if err != nil {
		return nil, err
	}

	if err := dca.solClient.GetAccountDataInto(context.TODO(), vaultProtoConfigPubKey, &vaultProtoConfigData); err != nil {
		return nil, err
	}

	cron := cron.New()
	runWithConfig := func() {
		dca.run(config)
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

func (dca *DCACronService) run(config configs.TriggerDCAConfig) {
	logrus.WithFields(logrus.Fields{
		"vault":      config.Vault,
		"tokenAMint": config.TokenAMint,
		"tokenBMint": config.TokenBMint,
	}).Info("running dca")
	// ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	// defer cancel()
	// vault := "TODO: VAULT PUBKEY"
	// if err := dca.walletProvider.TriggerDCA(ctx, vault); err != nil {
	// 	logrus.
	// 		WithFields(logrus.Fields{"vault": vault}).
	// 		WithError(err).
	// 		Errorf("failed to trigger DCA")
	// 	return
	// }
	// logrus.
	// 	WithFields(logrus.Fields{"vault": vault}).
	// 	Info("triggered DCA")
}

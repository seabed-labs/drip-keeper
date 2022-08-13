package vaultprovider

import (
	"context"
	"fmt"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/eventbus"
	"github.com/asaskevich/EventBus"
	"github.com/dcaf-labs/drip-client/drip-go"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type vaultProviderImpl struct {
	cron       *cron.Cron
	eventBus   EventBus.Bus
	dripClient *drip.APIClient
}

type VaultProvider interface {
	GetVaultChannel()
}

const (
	discoveryPeriod = 60
)

func NewVaultProvider(
	lc fx.Lifecycle,
	eventBus EventBus.Bus,
	config *configs.Config,
) (*VaultProvider, error) {
	cfg := drip.NewConfiguration()
	cfg.Host = config.DiscoveryURL
	cfg.UserAgent = "drip-keeper"
	// Debug is super noisy
	// cfg.Debug = config.Environment != configs.MainnetEnv
	cfg.Scheme = "https"
	vaultProviderImpl := vaultProviderImpl{
		cron:       cron.New(),
		eventBus:   eventBus,
		dripClient: drip.NewAPIClient(cfg),
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			for i := range config.TriggerDCAConfigs {
				dcaConfig := config.TriggerDCAConfigs[i]
				logrus.WithField("vault", dcaConfig.Vault).Info("publishing vault config")
				vaultProviderImpl.eventBus.Publish(string(eventbus.VaultConfigTopic), dcaConfig)
			}
			if config.ShouldDiscoverNewConfigs {
				if _, err := vaultProviderImpl.cron.AddFunc(fmt.Sprintf("@every %ds", discoveryPeriod), vaultProviderImpl.discoverConfigs); err != nil {
					return err
				}
				vaultProviderImpl.cron.Start()
				vaultProviderImpl.discoverConfigs()
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// TODO(Mocha): wait for context or return err if it takes too long
			if config.ShouldDiscoverNewConfigs {
				vaultProviderImpl.cron.Stop()
			}
			return nil
		},
	})

	return nil, nil
}

func (vaultProviderImpl vaultProviderImpl) discoverConfigs() {
	logrus.Info("searching for new configs")
	dripSPLTokenSwapConfigs, _, err := vaultProviderImpl.dripClient.DefaultApi.SpltokenswapconfigsGet(context.Background()).Execute()
	if err != nil {
		logrus.
			WithError(err).
			WithField("host", vaultProviderImpl.dripClient.GetConfig().Host).
			Error("failed to get spl token swaps configs from backend")
		return
	}
	dripOrcaWhirlpoolConfigs, _, err := vaultProviderImpl.dripClient.DefaultApi.OrcawhirlpoolconfigsGet(context.Background()).Execute()
	if err != nil {
		logrus.
			WithError(err).
			WithField("host", vaultProviderImpl.dripClient.GetConfig().Host).
			Error("failed to get orca whirlpool configs from backend")
		return
	}
	vaultSet := make(map[string]bool)
	splTokenSwapConfigsByVault := make(map[string]drip.SplTokenSwapConfig)
	for i := range dripSPLTokenSwapConfigs {
		splTokenSwapConfig := dripSPLTokenSwapConfigs[i]
		splTokenSwapConfigsByVault[splTokenSwapConfig.Vault] = splTokenSwapConfig
		vaultSet[splTokenSwapConfig.Vault] = true
	}
	orcaWhirlpoolConfigsByVault := make(map[string]drip.OrcaWhirlpoolConfig)
	for i := range dripOrcaWhirlpoolConfigs {
		orcaWhirlpoolConfig := dripOrcaWhirlpoolConfigs[i]
		orcaWhirlpoolConfigsByVault[orcaWhirlpoolConfig.Vault] = orcaWhirlpoolConfig
		vaultSet[orcaWhirlpoolConfig.Vault] = true
	}
	for vault := range vaultSet {
		dripSPLTokenSwapConfig, validTokenSwapConfig := splTokenSwapConfigsByVault[vault]
		dripOrcaWhirlpoolConfig, validOrcaWhirlpoolConfig := orcaWhirlpoolConfigsByVault[vault]
		dripConfig := configs.DripConfig{}
		if validTokenSwapConfig {
			dripConfig.Vault = dripSPLTokenSwapConfig.Vault
			dripConfig.VaultProtoConfig = dripSPLTokenSwapConfig.VaultProtoConfig
			dripConfig.VaultTokenAAccount = dripSPLTokenSwapConfig.VaultTokenAAccount
			dripConfig.VaultTokenBAccount = dripSPLTokenSwapConfig.VaultTokenBAccount
			dripConfig.SPLTokenSwapConfig = configs.SPLTokenSwapConfig{
				TokenAMint:        dripSPLTokenSwapConfig.TokenAMint,
				TokenBMint:        dripSPLTokenSwapConfig.TokenBMint,
				SwapTokenAAccount: dripSPLTokenSwapConfig.SwapTokenAAccount,
				SwapTokenBAccount: dripSPLTokenSwapConfig.SwapTokenBAccount,
				SwapTokenMint:     dripSPLTokenSwapConfig.SwapTokenMint,
				SwapFeeAccount:    dripSPLTokenSwapConfig.SwapFeeAccount,
				SwapAuthority:     dripSPLTokenSwapConfig.SwapAuthority,
				Swap:              dripSPLTokenSwapConfig.Swap,
			}
		}
		if validOrcaWhirlpoolConfig {
			dripConfig.Vault = dripOrcaWhirlpoolConfig.Vault
			dripConfig.VaultProtoConfig = dripOrcaWhirlpoolConfig.VaultProtoConfig
			dripConfig.VaultTokenAAccount = dripOrcaWhirlpoolConfig.VaultTokenAAccount
			dripConfig.VaultTokenBAccount = dripOrcaWhirlpoolConfig.VaultTokenBAccount
			dripConfig.OrcaWhirlpoolConfig = configs.OrcaWhirlpoolConfig{
				SwapTokenAAccount: dripOrcaWhirlpoolConfig.TokenVaultA,
				SwapTokenBAccount: dripOrcaWhirlpoolConfig.TokenVaultB,
				Oracle:            dripOrcaWhirlpoolConfig.Oracle,
				Whirlpool:         dripOrcaWhirlpoolConfig.Whirlpool,
			}
		}

		vaultProviderImpl.eventBus.Publish(string(eventbus.VaultConfigTopic), dripConfig)
	}
}

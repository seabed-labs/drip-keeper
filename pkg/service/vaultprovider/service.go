package vaultprovider

import (
	"context"
	"fmt"
	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/client/drip"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/eventbus"
	"github.com/asaskevich/EventBus"
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
	discoveryPeriod = 10
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
	res, _, err := vaultProviderImpl.dripClient.DefaultApi.SwapConfigsGet(context.Background()).Execute()
	if err != nil {
		logrus.
			WithError(err).
			WithField("host", vaultProviderImpl.dripClient.GetConfig().Host).
			Error("failed to get token swaps from backend")
		return
	}
	for i := range res {
		backendConfig := res[i]
		vaultProviderImpl.eventBus.Publish(string(eventbus.VaultConfigTopic), configs.TriggerDCAConfig{
			Vault:              backendConfig.Vault,
			VaultProtoConfig:   backendConfig.VaultProtoConfig,
			VaultTokenAAccount: backendConfig.VaultTokenAAccount,
			VaultTokenBAccount: backendConfig.VaultTokenBAccount,
			TokenAMint:         backendConfig.TokenAMint,
			TokenBMint:         backendConfig.TokenBMint,
			SwapTokenMint:      backendConfig.SwapTokenMint,
			SwapTokenAAccount:  backendConfig.SwapTokenAAccount,
			SwapTokenBAccount:  backendConfig.SwapTokenBAccount,
			SwapFeeAccount:     backendConfig.SwapFeeAccount,
			SwapAuthority:      backendConfig.SwapAuthority,
			Swap:               backendConfig.Swap,
		})
	}
}

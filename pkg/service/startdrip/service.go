package startdrip

import (
	"context"
	"fmt"
	"time"

	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/alert"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/keeper"
	"github.com/dcaf-labs/drip-client/drip-go"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type VaultProvider interface {
	getAllDripConfigs() ([]configs.DripConfig, error)
	dripAllVaults()
}

type vaultProviderImpl struct {
	dripClient   *drip.APIClient
	keeper       *keeper.KeeperService
	alertService alert.Service
}

const (
	discoveryPeriod = 300
)

func StartDrip(
	lc fx.Lifecycle,
	dripBackendClient *drip.APIClient,
	keeper *keeper.KeeperService,
	alertService alert.Service,
) {
	vaultProviderImpl := vaultProviderImpl{
		dripClient:   dripBackendClient,
		keeper:       keeper,
		alertService: alertService,
	}
	dripCron := cron.New()
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if _, err := dripCron.AddFunc(fmt.Sprintf("@every %ds", discoveryPeriod), vaultProviderImpl.dripAllVaults); err != nil {
				return err
			}
			dripCron.Start()
			vaultProviderImpl.dripAllVaults()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// TODO(Mocha): wait for context or return err if it takes too long
			dripCron.Stop()
			return nil
		},
	})
	return
}

func (impl vaultProviderImpl) dripAllVaults() {
	log := logrus.WithField("method", "dripAllVaults")
	log.Info("searching for new configs...")
	dripConfigs, err := impl.getAllDripConfigs()
	if err != nil {
		log.WithError(err).Errorf("failed to get configs")
	}
	log.WithField("len(dripConfigs)", len(dripConfigs)).Info("fetched drip configs")
	for _, dripConfig := range dripConfigs {
		log = logrus.WithField("vault", dripConfig.Vault)
		log.Info("starting drip...")

		startTime := time.Now()
		err := impl.keeper.Run(dripConfig)
		totalTime := time.Now().Unix() - startTime.Unix()
		log = log.WithField("totalTimeInSeconds", totalTime)

		if err != nil && err.Error() != keeper.ErrDripAmount0 && err.Error() != keeper.ErrDripAlreadyTriggered {
			log.WithError(err).Errorf("failed to drip")
			if err := impl.alertService.SendError(context.Background(), err); err != nil {
				log.WithError(err).Errorf("failed to alert error")
			}
		} else {
			log.Info("finished drip")
			msg := fmt.Sprintf("dripped vault %s", dripConfig.Vault)
			if err := impl.alertService.SendInfo(context.Background(), msg); err != nil {
				log.WithError(err).Errorf("failed to alert info")
			}
		}
	}
}

func (impl vaultProviderImpl) getAllDripConfigs() ([]configs.DripConfig, error) {
	dripSPLTokenSwapConfigs, _, err := impl.dripClient.DefaultApi.V1DripSpltokenswapconfigsGet(context.Background()).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get spl token swap configs from backend, err: %s", err.Error())
	}
	dripOrcaWhirlpoolConfigs, _, err := impl.dripClient.DefaultApi.V1DripOrcawhirlpoolconfigsGet(context.Background()).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get orca whirlpool configs from backend, err: %s", err.Error())
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
	dripConfigs := []configs.DripConfig{}
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
		dripConfigs = append(dripConfigs, dripConfig)
	}
	return dripConfigs, nil
}

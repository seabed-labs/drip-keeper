package main

import (
	"context"

	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/clients/dripclient"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/alert"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/clients/orcawhirlpool"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/clients/solana"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/eventbus"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/heartbeat"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/keeper"
	dca "github.com/Dcaf-Protocol/drip-keeper/pkg/service/schedule"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/vaultprovider"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	fxApp := fx.New(getDependencies()...)
	if err := fxApp.Start(context.Background()); err != nil {
		logrus.WithError(err).Fatalf("failed to start keeper bot")
	}
	logrus.Info("starting keeper bot")
	sig := <-fxApp.Done()
	logrus.WithFields(logrus.Fields{"signal": sig}).
		Infof("received exit signal, stoping keeper bot")
}

func getDependencies() []fx.Option {
	return []fx.Option{
		fx.Provide(
			configs.New,
			solana.New,
			orcawhirlpool.NewOrcaWhirlpoolClient,
			dripclient.NewDripBackendClient,
			eventbus.NewEventBus,
			alert.NewService,
			keeper.NewKeeperService,
		),
		fx.Invoke(
			// NewSchedulerService should be invoked first
			dca.NewSchedulerService,
			vaultprovider.NewVaultProvider,
			heartbeat.NewHeartbeatWorker,
		),
		fx.NopLogger,
	}
}

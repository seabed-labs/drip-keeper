package main

import (
	"context"

	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/keeper"

	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/alert"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/eventbus"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/heartbeat"
	dca "github.com/Dcaf-Protocol/drip-keeper/pkg/service/schedule"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/vaultprovider"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/solanaclient"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	fxApp := fx.New(
		fx.Provide(
			configs.New,
			solanaclient.New,
			eventbus.NewEventBus,
			alert.NewService,
			keeper.NewKeeperService,
		),
		fx.Invoke(
			// NewDCACron should be invoked first
			dca.NewDCACron,
			vaultprovider.NewVaultProvider,
			heartbeat.NewHeartbeatWorker,
		),
		fx.NopLogger,
	)
	if err := fxApp.Start(context.Background()); err != nil {
		logrus.WithError(err).Fatalf("failed to start keeper bot")
	}
	logrus.Info("starting keeper bot")
	sig := <-fxApp.Done()
	logrus.WithFields(logrus.Fields{"signal": sig}).
		Infof("received exit signal, stoping keeper bot")
}

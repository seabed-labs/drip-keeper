package main

import (
	"context"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/client"
	dca "github.com/Dcaf-Protocol/drip-keeper/pkg/service/dcacron"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/wallet"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	fxApp := fx.New(
		fx.Provide(
			configs.New,
			client.NewSolanaClient,
			wallet.New,
		),
		fx.Invoke(
			dca.NewDCACron,
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

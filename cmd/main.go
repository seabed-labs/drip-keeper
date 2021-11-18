package main

import (
	"context"

	"github.com/Dcaf-Protocol/keeper-bot/configs"
	"github.com/Dcaf-Protocol/keeper-bot/pkg/service/dca"
	"github.com/Dcaf-Protocol/keeper-bot/pkg/wallet"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func main() {
	logrus.Info("starting keeper bot")
	fxApp := fx.New(
		fx.Provide(configs.GetSecrets),
		fx.Provide(wallet.NewWallet),
		fx.Invoke(dca.NewDCA),
		fx.NopLogger,
	)
	if err := fxApp.Start(context.Background()); err != nil {
		logrus.WithError(err).Fatalf("failed to start keeper bot")
	}
	sig := <-fxApp.Done()
	logrus.WithFields(logrus.Fields{"signal": sig}).
		Infof("recieved exit signal, stoping keeper bot")
}

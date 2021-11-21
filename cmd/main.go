package main

import (
	"context"

	"github.com/Dcaf-Protocol/keeper-bot/configs"
	"github.com/Dcaf-Protocol/keeper-bot/pkg/client"
	"github.com/Dcaf-Protocol/keeper-bot/pkg/service/dca"
	"github.com/Dcaf-Protocol/keeper-bot/pkg/wallet"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func main() {
	fxApp := fx.New(
		fx.Provide(
			configs.GetSecrets,
			client.NewSolanaClient,
			wallet.NewWallet,
		),
		fx.Invoke(
			configs.InitLogrus,
			dca.NewDCA,
		),
		fx.NopLogger,
	)
	if err := fxApp.Start(context.Background()); err != nil {
		logrus.WithError(err).Fatalf("failed to start keeper bot")
	}
	logrus.Info("starting keeper bot")
	sig := <-fxApp.Done()
	logrus.WithFields(logrus.Fields{"signal": sig}).
		Infof("recieved exit signal, stoping keeper bot")
}

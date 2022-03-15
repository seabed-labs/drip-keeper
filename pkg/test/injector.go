package test

import (
	"context"

	"github.com/Dcaf-Protocol/keeper-bot/configs"
	"github.com/Dcaf-Protocol/keeper-bot/pkg/client"
	"github.com/Dcaf-Protocol/keeper-bot/pkg/wallet"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func InjectDependencies(
	testCase interface{},
) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	opts := []fx.Option{
		fx.Provide(
			configs.New,
			client.NewSolanaClient,
			wallet.New,
		),
		fx.Invoke(
			testCase,
		),
		fx.NopLogger,
	}
	app := fx.New(opts...)
	defer func() {
		if err := app.Stop(context.Background()); err != nil {
			logrus.WithError(err).Errorf("failed to stop test app")
		}
	}()
	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}
}

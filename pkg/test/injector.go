package test

import (
	"context"

	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/clients/solana"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
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
			solana.New,
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

package test

import (
	"context"

	"github.com/Dcaf-Protocol/keeper-bot/configs"
	"github.com/Dcaf-Protocol/keeper-bot/pkg/client"
	"github.com/Dcaf-Protocol/keeper-bot/pkg/wallet"
	"go.uber.org/fx"
)

func InjectDependencies(
	testCase interface{},
) {
	opts := []fx.Option{
		fx.Provide(
			configs.GetSecrets,
			client.NewSolanaClient,
			wallet.NewWallet,
		),
		fx.Invoke(
			configs.InitLogrus,
			testCase,
		),
		fx.NopLogger,
	}
	app := fx.New(opts...)
	defer app.Stop(context.Background())
	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}
}

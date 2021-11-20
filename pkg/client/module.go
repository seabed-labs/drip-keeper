package client

import "go.uber.org/fx"

func GetClientProviders() fx.Option {
	return fx.Options(
		fx.Provide(NewSolanaClient),
	)
}

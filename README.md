# drip-keeper

[![Build and Test](https://github.com/dcaf-labs/drip-keeper/actions/workflows/build-and-test.yaml/badge.svg?branch=main)](https://github.com/dcaf-labs/drip-keeper/actions/workflows/build-and-test.yaml)

## Getting Started

Setup the `.env` file:
```go
package configs

type Config struct {
	Network       Network `yaml:"network" env:"NETWORK" env-default:"DEVNET"`
	DripProgramID string  `yaml:"dripProgramID" env:"DRIP_PROGRAM_ID"  env-default:"dripTrkvSyQKvkyWg7oi4jmeEGMA5scSYowHArJ9Vwk"`
	Wallet        string  `yaml:"keeperBotWallet"      env:"KEEPER_BOT_WALLET"`
	FeeWallet     string  `yaml:"feeWallet"      env:"KEEPER_BOT_FEE_WALLET" env-default:"H7pb5fjhDygia45TeyhyE1JDmQXSuC5FrdXeSjA9kQ9T"`
	SolanaRPCURL  string  `yaml:"solanaRpcUrl" env:"SOLANARPCURL" env-default:"https://wiser-icy-bush.solana-devnet.discover.quiknode.pro/7288cc56d980336f6fc0508eb1aa73e44fd2efcd/"`
	DiscoveryURL  string  `yaml:"discoveryURL" env:"DISCOVERY_URL" env-default:"devnet.api.drip.dcaf.so"`
	HeartbeatURL  string  `yaml:"HeartbeatURL" env:"HEARTBEAT_URL"`
	// Discord Compatible webhook URL
	DiscordWebhookID          string `yaml:"DiscordWebhookID" env:"DISCORD_WEBHOOK_ID"`
	DiscordWebhookAccessToken string `yaml:"DiscordWebhookAccessToken" env:"DISCORD_ACCESS_TOKEN"`
	SlackWebhookURL           string `yaml:"slackWebhookURL" env:"SLACK_WEBHOOK_URL"`
}
```

The above config values are expected to be set via env vars for a production environment.

Run the Bot: `go run main.go`

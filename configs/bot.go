package configs

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type BotConfig struct {
	Environment Environment
	Account     string
}

const KEEPER_BOT_ACCOUNT = "KEEPER_BOT_ACCOUNT"

func GetBotConfig() (*BotConfig, error) {
	LoadEnv()
	accountString := os.Getenv(KEEPER_BOT_ACCOUNT)
	environment := Environment(os.Getenv(ENV))
	/* TODO: Load DCA Info
	- addresses needed for trigger DCA
	- granularity
	*/
	return &BotConfig{
		Account:     accountString,
		Environment: environment,
	}, nil
}

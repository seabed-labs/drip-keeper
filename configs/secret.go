package configs

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Environment string

type Secrets struct {
	Environment
	Account string
}

func GetSecrets() (*Secrets, error) {
	LoadEnv()
	accountString := os.Getenv("KEEPER_BOT_ACCOUNT")
	environment := Environment(os.Getenv("KEEPER_BOT_ENV"))
	return &Secrets{
		Account:     accountString,
		Environment: environment,
	}, nil
}

func (s *Secrets) IsLocal() bool {
	return s.Environment == LocalEnv || s.Environment == NilEnv
}

func (s *Secrets) IsTest() bool {
	return s.Environment == TestEnv
}

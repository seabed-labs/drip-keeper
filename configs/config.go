package configs

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Environment       Environment        `yaml:"environment" env:"ENV" env-default:"5432"`
	Wallet            string             `yaml:"wallet"      env:"KEEPER_BOT_WALLET"`
	TriggerDCAConfigs []TriggerDCAConfig `yaml:"triggerDCA"`
}

type TriggerDCAConfig struct {
	Vault              string `yaml:"vault"`
	VaultProtoConfig   string `yaml:"vaultProtoConfig"`
	VaultTokenAAccount string `yaml:"vaultTokenAAccount"`
	VaultTokenBAccount string `yaml:"vaultTokenBAccount"`
	TokenAMint         string `yaml:"tokenAMint"`
	TokenBMint         string `yaml:"tokenBMint"`
	SwapTokenMint      string `yaml:"swapTokenMint"`
	SwapTokenAAccount  string `yaml:"swapTokenAAccount"`
	SwapTokenBAccount  string `yaml:"swapTokenBAccount"`
	SwapFeeAccount     string `yaml:"swapFeeAccount"`
	SwapAuthority      string `yaml:"swapAuthority"`
	Swap               string `yaml:"swap"`
}

type Environment string

const (
	NilEnv   = Environment("")
	LocalEnv = Environment("LOCAL")
	DevEnv   = Environment("DEV")
	ProdEnv  = Environment("PROD")
)

const KEEPER_BOT_ACCOUNT = "KEEPER_BOT_ACCOUNT"
const ENV = "ENV"

func New() (*Config, error) {
	LoadEnv()

	environment := Environment(os.Getenv(ENV))
	configFileName := "./configs/local.yaml"
	if IsProd(environment) {
		configFileName = "./configs/prod.yaml"
	} else if IsDev(environment) {
		configFileName = "./configs/dev.yaml"
	}
	configFileName = fmt.Sprintf("%s/%s", GetProjectRoot(), configFileName)

	logrus.WithField("configFileName", configFileName).Infof("loading config file")
	configFile, err := os.Open(configFileName)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var config Config
	if err := cleanenv.ReadConfig(configFileName, &config); err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"TriggerDCAConfigs": fmt.Sprintf("%+v", config.TriggerDCAConfigs),
	}).Info("loaded trigger dca configs")

	return &config, nil
}

func IsLocal(env Environment) bool {
	return env == LocalEnv || env == NilEnv
}

func IsDev(env Environment) bool {
	return env == DevEnv
}

func IsProd(env Environment) bool {
	return env == ProdEnv
}

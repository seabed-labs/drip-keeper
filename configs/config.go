package configs

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Environment       Environment        `yaml:"environment" env:"ENV"`
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
	NilEnv      = Environment("")
	LocalnetEnv = Environment("LOCALNET")
	DevnetEnv   = Environment("DEVNET")
	MainnetEnv  = Environment("MAINNET")
)

const KEEPER_BOT_WALLET = "KEEPER_BOT_WALLET"
const ENV = "ENV"

func New() (*Config, error) {
	LoadEnv()

	environment := Environment(os.Getenv(ENV))
	if environment == NilEnv {
		environment = LocalnetEnv
	}
	configFileName := fmt.Sprintf("./configs/%s.yaml", environment)
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
	return env == LocalnetEnv || env == NilEnv
}

func IsDev(env Environment) bool {
	return env == DevnetEnv
}

func IsProd(env Environment) bool {
	return env == MainnetEnv
}

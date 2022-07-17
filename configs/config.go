package configs

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Environment              Environment        `yaml:"environment" env:"ENV" env-default:"DEVNET"`
	Wallet                   string             `yaml:"wallet"      env:"KEEPER_BOT_WALLET"`
	ShouldDiscoverNewConfigs bool               `yaml:"shouldDiscoverNewConfigs"      env-default:"true"`
	DiscoveryURL             string             `yaml:"discoveryURL" env:"DISCOVERY_URL" env-default:"devnet.api.drip.dcaf.so"`
	TriggerDCAConfigs        []TriggerDCAConfig `yaml:"triggerDCA"`
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
const PROJECT_ROOT_OVERRIDE = "PROJECT_ROOT_OVERRIDE"

func New() (*Config, error) {
	LoadEnv()

	environment := Environment(os.Getenv(ENV))
	if environment == NilEnv {
		environment = LocalnetEnv
	}
	// EXAMPLE: Load from config
	// configFileName := "config.yaml"
	// logrus.WithField("configFileName", configFileName).Infof("loading config file")
	// configFile, err := os.Open(configFileName)
	// if err != nil {
	// 	return nil, err
	// }
	// defer configFile.Close()

	var config Config
	if err := cleanenv.ReadEnv(&config); err != nil {
		return nil, err
	}
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

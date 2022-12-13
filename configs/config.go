package configs

import (
	"github.com/dcaf-labs/solana-go-clients/pkg/drip"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Environment              Environment  `yaml:"environment" env:"ENV" env-default:"STAGING"`
	Network                  Network      `yaml:"network" env:"NETWORK" env-default:"DEVNET"`
	DripProgramID            string       `yaml:"dripProgramID" env:"DRIP_PROGRAM_ID"  env-default:"dripTrkvSyQKvkyWg7oi4jmeEGMA5scSYowHArJ9Vwk"`
	Wallet                   string       `yaml:"keeperBotWallet"      env:"KEEPER_BOT_WALLET"`
	FeeWallet                string       `yaml:"feeWallet"      env:"KEEPER_BOT_FEE_WALLET" env-default:"H7pb5fjhDygia45TeyhyE1JDmQXSuC5FrdXeSjA9kQ9T"`
	ShouldDiscoverNewConfigs bool         `yaml:"shouldDiscoverNewConfigs"      env-default:"true"`
	TriggerDCAConfigs        []DripConfig `yaml:"triggerDCA"`
	DiscoveryURL             string       `yaml:"discoveryURL" env:"DISCOVERY_URL" env-default:"devnet.api.drip.dcaf.so"`
	HeartbeatURL             string       `yaml:"HeartbeatURL" env:"HEARTBEAT_URL"`
	// Discord Compatible webhook URL
	DiscordWebhookID          string `yaml:"DiscordWebhookID" env:"DISCORD_WEBHOOK_ID"`
	DiscordWebhookAccessToken string `yaml:"DiscordWebhookAccessToken" env:"DISCORD_ACCESS_TOKEN"`
}

type SPLTokenSwapConfig struct {
	TokenAMint        string `yaml:"tokenAMint"`
	TokenBMint        string `yaml:"tokenBMint"`
	SwapTokenAAccount string `yaml:"swapTokenAAccount"`
	SwapTokenBAccount string `yaml:"swapTokenBAccount"`
	SwapTokenMint     string `yaml:"swapTokenMint"`
	SwapFeeAccount    string `yaml:"swapFeeAccount"`
	SwapAuthority     string `yaml:"swapAuthority"`
	Swap              string `yaml:"swap"`
}

type OrcaWhirlpoolConfig struct {
	SwapTokenAAccount string `yaml:"swapTokenAAccount"`
	SwapTokenBAccount string `yaml:"swapTokenBAccount"`
	Oracle            string `yaml:"oracle"`
	Whirlpool         string `yaml:"whirlpool"`
}

type DripConfig struct {
	Vault               string              `yaml:"vault"`
	VaultProtoConfig    string              `yaml:"vaultProtoConfig"`
	VaultTokenAAccount  string              `yaml:"vaultTokenAAccount"`
	VaultTokenBAccount  string              `yaml:"vaultTokenBAccount"`
	OracleConfig        *string             `yaml:"oracleConfig"`
	SPLTokenSwapConfig  SPLTokenSwapConfig  `yaml:"SPLTokenSwapConfig"`
	OrcaWhirlpoolConfig OrcaWhirlpoolConfig `yaml:"orcaWhirlpoolConfig"`
}
type Network string

const (
	NilNetwork     = Network("")
	LocalNetwork   = Network("LOCALNET")
	DevnetNetwork  = Network("DEVNET")
	MainnetNetwork = Network("MAINNET")
)

type Environment string

const (
	NilEnv     = Environment("")
	LocalEnv   = Environment("LocalEnv")
	StagingEnv = Environment("STAGING")
	ProdEnv    = Environment("PROD")
)

const KEEPER_BOT_WALLET = "KEEPER_BOT_WALLET"
const ENV = "ENV"
const NETWORK = "NETWORK"
const PROJECT_ROOT_OVERRIDE = "PROJECT_ROOT_OVERRIDE"

func New() (*Config, error) {
	LoadEnv()

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
	drip.ProgramID = ag_solanago.MustPublicKeyFromBase58(config.DripProgramID)
	return &config, nil
}

func IsLocalnet(network Network) bool {
	return network == LocalNetwork || network == NilNetwork
}

func IsDevnet(network Network) bool {
	return network == DevnetNetwork
}

func IsMainnet(network Network) bool {
	return network == MainnetNetwork
}

func IsStagingEnv(env Environment) bool {
	return env == StagingEnv
}
func IsProductionEnv(env Environment) bool {
	return env == ProdEnv
}
func IsLocalEnv(env Environment) bool {
	return env == NilEnv || env == LocalEnv
}

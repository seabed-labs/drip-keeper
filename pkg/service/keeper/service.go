package keeper

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/solanaclient"
	"github.com/dcaf-labs/solana-go-clients/pkg/drip"
	"github.com/gagliardetto/solana-go"
	"github.com/sirupsen/logrus"
)

const ErrDripAmount0 = "drip amount is 0"
const ErrDripAlreadyTriggered = "drip already triggered"

type KeeperService struct {
	solanaClient *solanaclient.SolanaClient
	env          configs.Environment
}

func NewKeeperService(
	config *configs.Config,
	solanaClient *solanaclient.SolanaClient,
) *KeeperService {
	return &KeeperService{
		solanaClient: solanaClient,
		env:          config.Environment,
	}
}

func (dca *KeeperService) Run(dripConfig configs.DripConfig) error {
	logrus.WithField("vault", dripConfig.Vault).Info("preparing trigger dca")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	vaultData, err := dca.validateVault(ctx, dripConfig)
	if err != nil {
		return err
	}
	vaultAddress := solana.MustPublicKeyFromBase58(dripConfig.Vault)
	var instructions []solana.Instruction
	lastVaultPeriod := int64(vaultData.LastDripPeriod)
	vaultPeriodI, instruction, err := dca.solanaClient.GetMaybeUninitializedVaultPeriod(
		ctx, vaultAddress, vaultData.ProtoConfig, vaultData.TokenAMint, vaultData.TokenBMint, lastVaultPeriod)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get vaultPeriodI %d PDA", lastVaultPeriod)
		return err
	}
	logrus.WithField("publicKey", vaultPeriodI.String()).Infof("fetched vaultPeriod %d PDA", lastVaultPeriod)
	if instruction != nil {
		instructions = append(instructions, instruction)
	}

	currentVaultPeriod := lastVaultPeriod + 1
	vaultPeriodJ, instruction, err := dca.solanaClient.GetMaybeUninitializedVaultPeriod(
		ctx, vaultAddress, vaultData.ProtoConfig, vaultData.TokenAMint, vaultData.TokenBMint, currentVaultPeriod)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get vaultPeriodJ %d PDA", currentVaultPeriod)
		return err
	}
	logrus.WithField("publicKey", vaultPeriodJ.String()).Infof("fetched vaultPeriod %d PDA", currentVaultPeriod)
	if instruction != nil {
		instructions = append(instructions, instruction)
	}

	botTokenAFeeAccount, instruction, err := dca.solanaClient.GetMaybeUninitializedTokenAccount(ctx, dca.solanaClient.GetFeeWallet(), vaultData.TokenAMint)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get botTokenAFeeAccount")
		return err
	}
	logrus.WithField("publicKey", botTokenAFeeAccount.String()).Infof("fetched botTokenAFeeAccount")
	if instruction != nil {
		instructions = append(instructions, instruction)
	}

	switch {
	case dripConfig.OrcaWhirlpoolConfig.Whirlpool != "":
		newInstructions, err := dca.dripOrcaWhirlpool(ctx, dripConfig, vaultData, vaultPeriodI, vaultPeriodJ, botTokenAFeeAccount)
		if err != nil {
			return err
		}
		instructions = append(instructions, newInstructions...)
	case dripConfig.SPLTokenSwapConfig.Swap != "":
		newInstructions, err := dca.dripSplTokenSwap(ctx, dripConfig, vaultData, vaultPeriodI, vaultPeriodJ, botTokenAFeeAccount)
		if err != nil {
			return err
		}
		instructions = append(instructions, newInstructions...)
	default:
		logrus.WithField("vault", dripConfig.Vault).Infof("missing drip config")
	}
	if err := dca.solanaClient.Send(ctx, instructions...); err != nil {
		logrus.
			WithField("vault", dripConfig.Vault).
			WithField("numInstructions", len(instructions)).
			WithError(err).
			Errorf("failed to trigger dca")
		return err
	}
	logrus.
		WithFields(logrus.Fields{"vault": dripConfig.Vault}).
		Info("processed drip")
	return nil
}

func (dca *KeeperService) validateVault(ctx context.Context, dripConfig configs.DripConfig) (drip.Vault, error) {
	vaultAddress, err := solana.PublicKeyFromBase58(dripConfig.Vault)
	if err != nil {
		return drip.Vault{}, err
	}
	vaultData, err := dca.solanaClient.GetVault(ctx, vaultAddress)
	if err != nil {
		return drip.Vault{}, err
	}
	if vaultData.DripActivationTimestamp > time.Now().Unix() {
		return drip.Vault{}, errors.New(ErrDripAlreadyTriggered)
	}
	// Check if Vault can Drip
	if vaultData.DripAmount == 0 {
		//logrus.
		//	WithField("dripAmount", vaultData.DripAmount).
		//	WithField("vault", dripConfig.Vault).
		//	Info("drip amount is 0")
		return drip.Vault{}, errors.New(ErrDripAmount0)
	}
	vaultTokenAAccountAddress, err := solana.PublicKeyFromBase58(dripConfig.VaultTokenAAccount)
	if err != nil {
		return drip.Vault{}, err
	}
	tokenAccountData, err := dca.solanaClient.GetTokenAccount(ctx, vaultTokenAAccountAddress)
	if err != nil {
		logrus.
			WithError(err).
			WithField("vault", dripConfig.Vault).
			Errorf("failed to fetch vault balance")
		return drip.Vault{}, err
	}

	if tokenAccountData.Amount == 0 || tokenAccountData.Amount < vaultData.DripAmount {
		logrus.
			WithField("tokenABalance", tokenAccountData.Amount).
			WithField("dripAmount", vaultData.DripAmount).
			WithField("vault", dripConfig.Vault).
			Errorf("exiting, token balance is too low")
		return drip.Vault{}, fmt.Errorf("token balance is too low")
	}
	return vaultData, nil
}

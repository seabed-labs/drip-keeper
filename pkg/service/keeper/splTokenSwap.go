package keeper

import (
	"context"
	"fmt"

	"github.com/dcaf-labs/solana-go-clients/pkg/drip"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/gagliardetto/solana-go"
	"github.com/sirupsen/logrus"
)

func (dca *KeeperService) dripSplTokenSwap(
	ctx context.Context,
	dripConfig configs.DripConfig,
	vaultData drip.Vault,
	vaultPeriodI solana.PublicKey,
	vaultPeriodJ solana.PublicKey,
	botTokenAFeeAccount solana.PublicKey,
) ([]solana.Instruction, error) {
	var instructions []solana.Instruction
	swapTokenAAccount, swapTokenBAccount, err := dca.fetchSplTokenSwapTokenAccounts(ctx, dripConfig)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get swap token accounts")
		return []solana.Instruction{}, err
	}
	dripConfig.SPLTokenSwapConfig.SwapTokenAAccount = swapTokenAAccount
	dripConfig.SPLTokenSwapConfig.SwapTokenBAccount = swapTokenBAccount

	logrus.WithFields(logrus.Fields{
		"vault":               dripConfig.Vault,
		"tokenAMint":          vaultData.TokenAMint.String(),
		"tokenBMint":          vaultData.TokenBMint.String(),
		"swapTokenAAcount":    dripConfig.SPLTokenSwapConfig.SwapTokenAAccount,
		"swapTokenBAccount":   dripConfig.SPLTokenSwapConfig.SwapTokenBAccount,
		"i":                   vaultData.LastDripPeriod,
		"j":                   vaultData.LastDripPeriod + 1,
		"vaultPeriodI":        vaultPeriodI.String(),
		"vaultPeriodJ":        vaultPeriodJ.String(),
		"botTokenAFeeAccount": botTokenAFeeAccount.String(),
	}).Info("running drip")

	instruction, err := dca.solanaClient.DripSPLTokenSwap(ctx, dripConfig, vaultPeriodI, vaultPeriodJ, botTokenAFeeAccount)
	if err != nil {
		logrus.
			WithError(err).
			WithField("dcaProgram", drip.ProgramID.String()).
			Errorf("failed to create DripSPLTokenSwap instruction")
		return []solana.Instruction{}, err
	}
	instructions = append(instructions, instruction)

	return instructions, nil
}

func (dca *KeeperService) fetchSplTokenSwapTokenAccounts(
	ctx context.Context,
	dripConfig configs.DripConfig,
) (string, string, error) {
	swapTokenAAccountAddress, err := solana.PublicKeyFromBase58(dripConfig.SPLTokenSwapConfig.SwapTokenAAccount)
	if err != nil {
		return "", "", err
	}
	swapTokenAAccount, err := dca.solanaClient.GetTokenAccount(ctx, swapTokenAAccountAddress)
	if err != nil {
		return "", "", err
	}

	swapTokenBAccountAddress, err := solana.PublicKeyFromBase58(dripConfig.SPLTokenSwapConfig.SwapTokenBAccount)
	if err != nil {
		return "", "", err
	}
	swapTokenBAccount, err := dca.solanaClient.GetTokenAccount(ctx, swapTokenBAccountAddress)
	if err != nil {
		return "", "", err
	}

	if swapTokenAAccount.Mint.String() == dripConfig.SPLTokenSwapConfig.TokenAMint && swapTokenBAccount.Mint.String() == dripConfig.SPLTokenSwapConfig.TokenBMint {
		// Normal A -> b
		return dripConfig.SPLTokenSwapConfig.SwapTokenAAccount, dripConfig.SPLTokenSwapConfig.SwapTokenBAccount, nil
	} else if swapTokenAAccount.Mint.String() == dripConfig.SPLTokenSwapConfig.TokenBMint && swapTokenBAccount.Mint.String() == dripConfig.SPLTokenSwapConfig.TokenAMint {
		// Need to swap token accounts for inverse
		return dripConfig.SPLTokenSwapConfig.SwapTokenBAccount, dripConfig.SPLTokenSwapConfig.SwapTokenAAccount, nil
	}
	err = fmt.Errorf("token swap token accounts do not match config mints, or the inverse of the config mints")
	logrus.
		WithField("swapTokenAAccount", dripConfig.SPLTokenSwapConfig.SwapTokenAAccount).
		WithField("swapTokenBAccount", dripConfig.SPLTokenSwapConfig.SwapTokenBAccount).
		WithField("swapTokenAMint", swapTokenAAccount.Mint.String()).
		WithField("swapTokenBMint", swapTokenBAccount.Mint.String()).
		WithField("configTokenAMint", dripConfig.SPLTokenSwapConfig.TokenAMint).
		WithField("configTokenBMint", dripConfig.SPLTokenSwapConfig.TokenBMint).
		WithField("vault", dripConfig.Vault).
		WithError(err).
		Error("failed to get swap token accounts")
	return "", "", err
}

package dca

import (
	"context"
	"fmt"

	"github.com/dcaf-labs/solana-go-clients/pkg/drip"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

func (dca *DCACronService) dripSplTokenSwap(
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

	instruction, err := dca.walletProvider.DripSPLTokenSwap(ctx, dripConfig, vaultPeriodI, vaultPeriodJ, botTokenAFeeAccount)
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

func (dca *DCACronService) fetchSplTokenSwapTokenAccounts(
	ctx context.Context,
	config configs.DripConfig,
) (string, string, error) {
	// Fetch Token A
	resp, err := dca.solClient.GetAccountInfoWithOpts(ctx, solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapTokenAAccount), &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: "confirmed",
		DataSlice:  nil,
	})
	if err != nil {
		return "", "", err
	}
	var swapTokenAAccount token.Account
	if err := bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&swapTokenAAccount); err != nil {
		return "", "", err
	}

	// Fetch token B
	resp, err = dca.solClient.GetAccountInfoWithOpts(ctx, solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapTokenBAccount), &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: "confirmed",
		DataSlice:  nil,
	})
	if err != nil {
		return "", "", err
	}
	var swapTokenBAccount token.Account
	if err := bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&swapTokenBAccount); err != nil {
		return "", "", err
	}

	if swapTokenAAccount.Mint.String() == config.SPLTokenSwapConfig.TokenAMint && swapTokenBAccount.Mint.String() == config.SPLTokenSwapConfig.TokenBMint {
		// Normal A -> b
		return config.SPLTokenSwapConfig.SwapTokenAAccount, config.SPLTokenSwapConfig.SwapTokenBAccount, nil
	} else if swapTokenAAccount.Mint.String() == config.SPLTokenSwapConfig.TokenBMint && swapTokenBAccount.Mint.String() == config.SPLTokenSwapConfig.TokenAMint {
		// Need to swap token accounts for inverse
		return config.SPLTokenSwapConfig.SwapTokenBAccount, config.SPLTokenSwapConfig.SwapTokenAAccount, nil
	}
	err = fmt.Errorf("token swap token accounts do not match config mints, or the inverse of the config mints")
	logrus.
		WithField("swapTokenAAccount", config.SPLTokenSwapConfig.SwapTokenAAccount).
		WithField("swapTokenBAccount", config.SPLTokenSwapConfig.SwapTokenBAccount).
		WithField("swapTokenAMint", swapTokenAAccount.Mint.String()).
		WithField("swapTokenBMint", swapTokenBAccount.Mint.String()).
		WithField("configTokenAMint", config.SPLTokenSwapConfig.TokenAMint).
		WithField("configTokenBMint", config.SPLTokenSwapConfig.TokenBMint).
		WithField("vault", config.Vault).
		WithError(err).
		Error("failed to get swap token accounts")
	return "", "", err
}

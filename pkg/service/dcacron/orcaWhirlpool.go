package dca

import (
	"context"
	"math"
	"strconv"

	bin "github.com/gagliardetto/binary"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/Dcaf-Protocol/drip-keeper/pkg/wallet"
	"github.com/dcaf-labs/solana-go-clients/pkg/drip"
	"github.com/dcaf-labs/solana-go-clients/pkg/whirlpool"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

func (dca *DCACronService) dripOrcaWhirlpool(
	ctx context.Context,
	dripConfig configs.DripConfig,
	vaultData drip.Vault,
	vaultPeriodI solana.PublicKey,
	vaultPeriodJ solana.PublicKey,
	botTokenAFeeAccount solana.PublicKey,
) ([]solana.Instruction, error) {
	var instructions []solana.Instruction
	// Get WhirlpoolsConfig
	whirlpoolPubkey := solana.MustPublicKeyFromBase58(dripConfig.OrcaWhirlpoolConfig.Whirlpool)
	resp, err := dca.solClient.GetAccountInfoWithOpts(ctx, whirlpoolPubkey, &rpc.GetAccountInfoOpts{
		Encoding:   solana.EncodingBase64,
		Commitment: "confirmed",
		DataSlice:  nil,
	})
	if err != nil {
		return []solana.Instruction{}, err
	}
	var whirlpoolData whirlpool.Whirlpool
	if err := bin.NewBinDecoder(resp.Value.Data.GetBinary()).Decode(&whirlpoolData); err != nil {
		return []solana.Instruction{}, err
	}

	if err := dca.ensureTickArrays(ctx, dripConfig, vaultData, whirlpoolData); err != nil {
		return []solana.Instruction{}, err
	}
	quoteEstimate, err := wallet.GetOrcaWhirlpoolQuoteEstimate(
		whirlpoolData.WhirlpoolsConfig.String(),
		whirlpoolData.TokenMintA.String(),
		whirlpoolData.TokenMintB.String(),
		vaultData.TokenAMint.String(),
		whirlpoolData.TickSpacing,
		dca.env,
	)
	if err != nil {
		return []solana.Instruction{}, err
	}
	logrus.WithFields(logrus.Fields{
		"vault":               dripConfig.Vault,
		"tokenAMint":          vaultData.TokenAMint.String(),
		"tokenBMint":          vaultData.TokenBMint.String(),
		"swapTokenAAcount":    dripConfig.OrcaWhirlpoolConfig.SwapTokenAAccount,
		"swapTokenBAccount":   dripConfig.OrcaWhirlpoolConfig.SwapTokenBAccount,
		"i":                   vaultData.LastDripPeriod,
		"j":                   vaultData.LastDripPeriod + 1,
		"vaultPeriodI":        vaultPeriodI.String(),
		"vaultPeriodJ":        vaultPeriodJ.String(),
		"botTokenAFeeAccount": botTokenAFeeAccount.String(),
	}).Info("running drip")

	instruction, err := dca.walletProvider.DripOrcaWhirlpool(ctx,
		wallet.DripOrcaWhirlpoolParams{
			VaultData:           vaultData,
			Vault:               solana.MustPublicKeyFromBase58(dripConfig.Vault),
			VaultPeriodI:        vaultPeriodI,
			VaultPeriodJ:        vaultPeriodJ,
			BotTokenAFeeAccount: botTokenAFeeAccount,
			WhirlpoolData:       whirlpoolData,
			Whirlpool:           whirlpoolPubkey,
			TickArray0:          solana.MustPublicKeyFromBase58(quoteEstimate.TickArray0),
			TickArray1:          solana.MustPublicKeyFromBase58(quoteEstimate.TickArray1),
			TickArray2:          solana.MustPublicKeyFromBase58(quoteEstimate.TickArray2),
			Oracle:              solana.MustPublicKeyFromBase58(dripConfig.OrcaWhirlpoolConfig.Oracle),
		},
	)
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

func (dca *DCACronService) ensureTickArrays(
	ctx context.Context,
	dripConfig configs.DripConfig,
	vault drip.Vault,
	whirlpoolData whirlpool.Whirlpool,
) error {
	whirlpoolPubkey := solana.MustPublicKeyFromBase58(dripConfig.OrcaWhirlpoolConfig.Whirlpool)
	var instructions []solana.Instruction
	realIndex := math.Floor(float64(whirlpoolData.TickCurrentIndex) / float64(whirlpoolData.TickSpacing) / 88.0)
	startTickIndex := int32(realIndex) * int32(whirlpoolData.TickSpacing) * 88

	aToB := vault.TokenAMint.String() == whirlpoolData.TokenMintA.String()
	var tickArrayIndexs []int32
	if aToB {
		tickArrayIndexs = []int32{
			startTickIndex,
			startTickIndex - int32(whirlpoolData.TickSpacing*88)*1,
			startTickIndex - int32(whirlpoolData.TickSpacing*88)*2,
		}
	} else {
		tickArrayIndexs = []int32{
			startTickIndex,
			startTickIndex + int32(whirlpoolData.TickSpacing*88)*1,
			startTickIndex + int32(whirlpoolData.TickSpacing*88)*2,
		}
	}
	for _, tickArrayIndex := range tickArrayIndexs {
		tickArrayPubkey, _, _ := solana.FindProgramAddress([][]byte{
			[]byte("tick_array"),
			whirlpoolPubkey[:],
			[]byte(strconv.FormatInt(int64(tickArrayIndex), 10)),
		}, whirlpool.ProgramID)
		// Use GetAccountInfoWithOpts so we can pass in a commitment level
		if _, err := dca.solClient.GetAccountInfoWithOpts(ctx, tickArrayPubkey, &rpc.GetAccountInfoOpts{
			Encoding:   solana.EncodingBase64,
			Commitment: "confirmed",
			DataSlice:  nil,
		}); err != nil && err.Error() == "not found" {
			initTickArrayInstruction, err := dca.walletProvider.InitializeTickArray(ctx,
				wallet.InitializeTickArrayParams{
					Whirlpool:  whirlpoolPubkey,
					StartIndex: tickArrayIndex,
					TickArray:  tickArrayPubkey,
				})
			if err != nil {
				logrus.
					WithError(err).
					Errorf("failed to create InitializeTickArrayParams instruction")
				return err
			}
			instructions = append(instructions, initTickArrayInstruction)
		}
	}
	if err := dca.walletProvider.Send(ctx, instructions...); err != nil {
		logrus.
			WithError(err).
			WithField("whirlpool", dripConfig.OrcaWhirlpoolConfig.Whirlpool).
			WithField("numInstructions", len(instructions)).
			Errorf("failed to initialize tick arrays")
		return err
	}
	logrus.
		WithField("whirlpool", dripConfig.OrcaWhirlpoolConfig.Whirlpool).
		WithField("numInstructions", len(instructions)).
		Info("initialized tick arrays")
	return nil
}

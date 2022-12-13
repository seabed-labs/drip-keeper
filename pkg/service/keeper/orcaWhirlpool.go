package keeper

import (
	"context"
	"math"
	"strconv"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	solclient "github.com/Dcaf-Protocol/drip-keeper/pkg/service/clients/solana"
	dripextension "github.com/dcaf-labs/drip-client/drip-extension-go"
	"github.com/dcaf-labs/solana-go-clients/pkg/drip"
	"github.com/dcaf-labs/solana-go-clients/pkg/whirlpool"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

func (dca *KeeperService) dripOrcaWhirlpool(
	ctx context.Context,
	dripConfig configs.DripConfig,
	vaultAccount drip.Vault,
	vaultPeriodI solana.PublicKey,
	vaultPeriodJ solana.PublicKey,
	botTokenAFeeAccount solana.PublicKey,
) ([]solana.Instruction, error) {
	whirlpoolPub, err := solana.PublicKeyFromBase58(dripConfig.OrcaWhirlpoolConfig.Whirlpool)
	if err != nil {
		return nil, err
	}
	whirlpoolAccount, err := dca.solanaClient.GetOrcaWhirlpool(ctx, whirlpoolPub)
	if err != nil {
		return nil, err
	}
	if err := dca.ensureTickArrays(ctx, dripConfig, vaultAccount, whirlpoolAccount); err != nil {
		return []solana.Instruction{}, err
	}
	quoteEstimate, err := dca.orcaWhirlpoolClient.GetOrcaWhirlpoolQuoteEstimate(
		ctx,
		dripConfig.OrcaWhirlpoolConfig.Whirlpool,
		vaultAccount.TokenAMint.String(),
		strconv.FormatUint(vaultAccount.DripAmount, 10),
	)
	if err != nil {
		return []solana.Instruction{}, err
	}
	logrus.WithFields(logrus.Fields{
		"vault":               dripConfig.Vault,
		"tokenAMint":          vaultAccount.TokenAMint.String(),
		"tokenBMint":          vaultAccount.TokenBMint.String(),
		"swapTokenAAcount":    dripConfig.OrcaWhirlpoolConfig.SwapTokenAAccount,
		"swapTokenBAccount":   dripConfig.OrcaWhirlpoolConfig.SwapTokenBAccount,
		"i":                   vaultAccount.LastDripPeriod,
		"j":                   vaultAccount.LastDripPeriod + 1,
		"vaultPeriodI":        vaultPeriodI.String(),
		"vaultPeriodJ":        vaultPeriodJ.String(),
		"botTokenAFeeAccount": botTokenAFeeAccount.String(),
	}).Info("running drip")

	if dripConfig.OracleConfig == nil {
		return dca.dripV1OrcaWhirlpool(ctx,
			solana.MustPublicKeyFromBase58(dripConfig.Vault),
			vaultAccount,
			vaultPeriodI,
			vaultPeriodJ,
			botTokenAFeeAccount,
			whirlpoolPub,
			solana.MustPublicKeyFromBase58(dripConfig.OrcaWhirlpoolConfig.Oracle),
			whirlpoolAccount,
			*quoteEstimate,
		)
	} else {
		return dca.dripV2OrcaWhirlpool(ctx,
			solana.MustPublicKeyFromBase58(dripConfig.Vault),
			vaultAccount,
			vaultPeriodI,
			vaultPeriodJ,
			botTokenAFeeAccount,
			whirlpoolPub,
			solana.MustPublicKeyFromBase58(dripConfig.OrcaWhirlpoolConfig.Oracle),
			whirlpoolAccount,
			*quoteEstimate,
		)
	}
}

func (dca *KeeperService) dripV1OrcaWhirlpool(
	ctx context.Context,
	vaultPub solana.PublicKey,
	vaultAccount drip.Vault,
	vaultPeriodI solana.PublicKey,
	vaultPeriodJ solana.PublicKey,
	botTokenAFeeAccountPub solana.PublicKey,

	whirlpoolPub solana.PublicKey,
	whirlpoolOraclePub solana.PublicKey,
	whirlpoolAccount whirlpool.Whirlpool,
	quoteEstimate dripextension.V1OrcawhirlpoolQuote200Response,
) ([]solana.Instruction, error) {
	var instructions []solana.Instruction

	logrus.WithFields(logrus.Fields{
		"vault":               vaultPub.String(),
		"tokenAMint":          vaultAccount.TokenAMint.String(),
		"tokenBMint":          vaultAccount.TokenBMint.String(),
		"i":                   vaultAccount.LastDripPeriod,
		"j":                   vaultAccount.LastDripPeriod + 1,
		"vaultPeriodI":        vaultPeriodI.String(),
		"vaultPeriodJ":        vaultPeriodJ.String(),
		"botTokenAFeeAccount": botTokenAFeeAccountPub.String(),
	}).Info("running drip")

	instruction, err := dca.solanaClient.DripOrcaWhirlpool(ctx,
		solclient.DripOrcaWhirlpoolParams{
			VaultAccount:           vaultAccount,
			VaultPub:               vaultPub,
			VaultPeriodIPub:        vaultPeriodI,
			VaultPeriodJPub:        vaultPeriodJ,
			BotTokenAFeeAccountPub: botTokenAFeeAccountPub,
			WhirlpoolAccount:       whirlpoolAccount,
			WhirlpoolPub:           whirlpoolPub,
			TickArray0Pub:          solana.MustPublicKeyFromBase58(quoteEstimate.TickArray0),
			TickArray1Pub:          solana.MustPublicKeyFromBase58(quoteEstimate.TickArray1),
			TickArray2Pub:          solana.MustPublicKeyFromBase58(quoteEstimate.TickArray2),
			WhirlpoolOraclePub:     whirlpoolOraclePub,
		},
	)
	if err != nil {
		logrus.
			WithError(err).
			WithField("dripProgram", drip.ProgramID.String()).
			Errorf("failed to create dripV1OrcaWhirlpool instruction")
		return []solana.Instruction{}, err
	}
	instructions = append(instructions, instruction)
	return instructions, nil
}

func (dca *KeeperService) dripV2OrcaWhirlpool(
	ctx context.Context,
	vaultPub solana.PublicKey,
	vaultAccount drip.Vault,
	vaultPeriodI solana.PublicKey,
	vaultPeriodJ solana.PublicKey,
	botTokenAFeeAccountPub solana.PublicKey,

	whirlpoolPub solana.PublicKey,
	whirlpoolOraclePub solana.PublicKey,
	whirlpoolAccount whirlpool.Whirlpool,
	quoteEstimate dripextension.V1OrcawhirlpoolQuote200Response,
) ([]solana.Instruction, error) {
	var instructions []solana.Instruction
	// Get OracleConfig
	dripOracleConfigAccount, err := dca.solanaClient.GetDripOracleConfig(ctx, vaultAccount.OracleConfig)
	if err != nil {
		return nil, err
	}
	logrus.WithFields(logrus.Fields{
		"vault":               vaultPub,
		"tokenAMint":          vaultAccount.TokenAMint.String(),
		"tokenBMint":          vaultAccount.TokenBMint.String(),
		"i":                   vaultAccount.LastDripPeriod,
		"j":                   vaultAccount.LastDripPeriod + 1,
		"vaultPeriodI":        vaultPeriodI.String(),
		"vaultPeriodJ":        vaultPeriodJ.String(),
		"botTokenAFeeAccount": botTokenAFeeAccountPub.String(),
		"dripOracleConfig":    vaultAccount.OracleConfig.String(),
	}).Info("running drip")

	instruction, err := dca.solanaClient.DripV2OrcaWhirlpool(ctx,
		solclient.DripV2OrcaWhirlpoolParams{
			VaultAccount:           vaultAccount,
			VaultPub:               vaultPub,
			VaultPeriodIPub:        vaultPeriodI,
			VaultPeriodJPub:        vaultPeriodJ,
			BotTokenAFeeAccountPub: botTokenAFeeAccountPub,

			DripOraclePub:                   vaultAccount.OracleConfig,
			DripOracleTokenAMintPub:         dripOracleConfigAccount.TokenAMint,
			DripOracleTokenAPriceAccountPub: dripOracleConfigAccount.TokenAPrice,
			DripOracleTokenBMintPub:         dripOracleConfigAccount.TokenBMint,
			DripOracleTokenBPriceAccountPub: dripOracleConfigAccount.TokenBPrice,

			WhirlpoolAccount:   whirlpoolAccount,
			WhirlpoolPub:       whirlpoolPub,
			TickArray0Pub:      solana.MustPublicKeyFromBase58(quoteEstimate.TickArray0),
			TickArray1Pub:      solana.MustPublicKeyFromBase58(quoteEstimate.TickArray1),
			TickArray2Pub:      solana.MustPublicKeyFromBase58(quoteEstimate.TickArray2),
			WhirlpoolOraclePub: whirlpoolOraclePub,
		},
	)
	if err != nil {
		logrus.
			WithError(err).
			WithField("dripProgram", drip.ProgramID.String()).
			Errorf("failed to create DripV2OrcaWhirlpool instruction")
		return []solana.Instruction{}, err
	}
	instructions = append(instructions, instruction)
	return instructions, nil
}

func (dca *KeeperService) ensureTickArrays(
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
		if _, err := dca.solanaClient.GetOrcaWhirlpoolTickArray(ctx, tickArrayPubkey); err != nil && err == rpc.ErrNotFound {
			initTickArrayInstruction, err := dca.solanaClient.InitializeTickArray(ctx,
				solclient.InitializeTickArrayParams{
					WhirlpoolPub: whirlpoolPubkey,
					StartIndex:   tickArrayIndex,
					TickArray:    tickArrayPubkey,
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
	if len(instructions) > 0 {
		if err := dca.solanaClient.Send(ctx, instructions...); err != nil {
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
	}
	return nil
}

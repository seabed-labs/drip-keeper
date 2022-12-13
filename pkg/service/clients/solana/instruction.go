package solana

import (
	"context"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/dcaf-labs/solana-go-clients/pkg/drip"
	"github.com/dcaf-labs/solana-go-clients/pkg/whirlpool"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
)

type DripOrcaWhirlpoolParams struct {
	VaultAccount           drip.Vault
	VaultPub               solana.PublicKey
	VaultPeriodIPub        solana.PublicKey
	VaultPeriodJPub        solana.PublicKey
	BotTokenAFeeAccountPub solana.PublicKey

	WhirlpoolAccount   whirlpool.Whirlpool
	WhirlpoolPub       solana.PublicKey
	TickArray0Pub      solana.PublicKey
	TickArray1Pub      solana.PublicKey
	TickArray2Pub      solana.PublicKey
	WhirlpoolOraclePub solana.PublicKey
}

type DripV2OrcaWhirlpoolParams struct {
	VaultAccount           drip.Vault
	VaultPub               solana.PublicKey
	VaultPeriodIPub        solana.PublicKey
	VaultPeriodJPub        solana.PublicKey
	BotTokenAFeeAccountPub solana.PublicKey

	DripOraclePub                   solana.PublicKey
	DripOracleTokenAMintPub         solana.PublicKey
	DripOracleTokenAPriceAccountPub solana.PublicKey
	DripOracleTokenBMintPub         solana.PublicKey
	DripOracleTokenBPriceAccountPub solana.PublicKey

	WhirlpoolAccount   whirlpool.Whirlpool
	WhirlpoolPub       solana.PublicKey
	TickArray0Pub      solana.PublicKey
	TickArray1Pub      solana.PublicKey
	TickArray2Pub      solana.PublicKey
	WhirlpoolOraclePub solana.PublicKey
}

func (w *SolanaClient) DripOrcaWhirlpool(
	ctx context.Context,
	params DripOrcaWhirlpoolParams,
) (solana.Instruction, error) {
	commonBuilder := drip.NewDripOrcaWhirlpoolCommonAccountsBuilder()
	commonBuilder.SetDripTriggerSourceAccount(w.wallet.PublicKey())
	commonBuilder.SetVaultAccount(params.VaultPub)
	commonBuilder.SetVaultProtoConfigAccount(params.VaultAccount.ProtoConfig)
	commonBuilder.SetLastVaultPeriodAccount(params.VaultPeriodIPub)
	commonBuilder.SetCurrentVaultPeriodAccount(params.VaultPeriodJPub)
	commonBuilder.SetVaultTokenAAccountAccount(params.VaultAccount.TokenAAccount)
	commonBuilder.SetVaultTokenBAccountAccount(params.VaultAccount.TokenBAccount)
	commonBuilder.SetSwapTokenAAccountAccount(params.WhirlpoolAccount.TokenVaultA)
	commonBuilder.SetSwapTokenBAccountAccount(params.WhirlpoolAccount.TokenVaultB)
	commonBuilder.SetDripFeeTokenAAccountAccount(params.BotTokenAFeeAccountPub)
	commonBuilder.SetTokenProgramAccount(solana.TokenProgramID)

	txBuilder := drip.NewDripOrcaWhirlpoolInstructionBuilder()
	txBuilder.SetCommonAccountsFromBuilder(commonBuilder)
	txBuilder.SetWhirlpoolAccount(params.WhirlpoolPub)
	txBuilder.SetTickArray0Account(params.TickArray0Pub)
	txBuilder.SetTickArray1Account(params.TickArray1Pub)
	txBuilder.SetTickArray2Account(params.TickArray2Pub)
	txBuilder.SetOracleAccount(params.WhirlpoolOraclePub)

	txBuilder.SetWhirlpoolProgramAccount(whirlpool.ProgramID)

	return txBuilder.ValidateAndBuild()
}

func (w *SolanaClient) DripV2OrcaWhirlpool(
	ctx context.Context,
	params DripV2OrcaWhirlpoolParams,
) (solana.Instruction, error) {
	commonBuilder := drip.NewDripV2OrcaWhirlpoolCommonAccountsBuilder()
	commonBuilder.SetDripTriggerSourceAccount(w.wallet.PublicKey())
	commonBuilder.SetVaultAccount(params.VaultPub)
	commonBuilder.SetVaultProtoConfigAccount(params.VaultAccount.ProtoConfig)
	commonBuilder.SetLastVaultPeriodAccount(params.VaultPeriodIPub)
	commonBuilder.SetCurrentVaultPeriodAccount(params.VaultPeriodJPub)
	commonBuilder.SetVaultTokenAAccountAccount(params.VaultAccount.TokenAAccount)
	commonBuilder.SetVaultTokenBAccountAccount(params.VaultAccount.TokenBAccount)
	commonBuilder.SetSwapTokenAAccountAccount(params.WhirlpoolAccount.TokenVaultA)
	commonBuilder.SetSwapTokenBAccountAccount(params.WhirlpoolAccount.TokenVaultB)
	commonBuilder.SetDripFeeTokenAAccountAccount(params.BotTokenAFeeAccountPub)
	commonBuilder.SetTokenProgramAccount(solana.TokenProgramID)

	oracleBuilder := drip.NewDripV2OrcaWhirlpoolOracleCommonAccountsBuilder()
	oracleBuilder.SetOracleConfigAccount(params.DripOraclePub)
	oracleBuilder.SetTokenAMintAccount(params.DripOracleTokenAMintPub)
	oracleBuilder.SetTokenAPriceAccount(params.DripOracleTokenAPriceAccountPub)
	oracleBuilder.SetTokenBMintAccount(params.DripOracleTokenBMintPub)
	oracleBuilder.SetTokenBPriceAccount(params.DripOracleTokenBPriceAccountPub)

	txBuilder := drip.NewDripV2OrcaWhirlpoolInstructionBuilder()
	txBuilder.SetCommonAccountsFromBuilder(commonBuilder)
	txBuilder.SetOracleCommonAccountsFromBuilder(oracleBuilder)
	txBuilder.SetWhirlpoolAccount(params.WhirlpoolPub)
	txBuilder.SetTickArray0Account(params.TickArray0Pub)
	txBuilder.SetTickArray1Account(params.TickArray1Pub)
	txBuilder.SetTickArray2Account(params.TickArray2Pub)
	txBuilder.SetOracleAccount(params.WhirlpoolOraclePub)

	txBuilder.SetWhirlpoolProgramAccount(whirlpool.ProgramID)

	return txBuilder.ValidateAndBuild()
}

type InitializeTickArrayParams struct {
	WhirlpoolPub solana.PublicKey
	StartIndex   int32
	TickArray    solana.PublicKey
}

func (w *SolanaClient) InitializeTickArray(
	ctx context.Context,
	params InitializeTickArrayParams,
) (solana.Instruction, error) {
	txBuilder := whirlpool.NewInitializeTickArrayInstructionBuilder()
	txBuilder.SetWhirlpoolAccount(params.WhirlpoolPub)
	txBuilder.SetFunderAccount(w.wallet.PublicKey())
	txBuilder.SetTickArrayAccount(params.TickArray)
	txBuilder.SetSystemProgramAccount(solana.SystemProgramID)
	txBuilder.SetStartTickIndex(params.StartIndex)
	return txBuilder.ValidateAndBuild()
}

func (w *SolanaClient) DripSPLTokenSwap(
	ctx context.Context, config configs.DripConfig, vaultPeriodI, vaultPeriodJ, botTokenAAccount solana.PublicKey,
) (solana.Instruction, error) {
	commonBuilder := drip.NewDripSplTokenSwapCommonAccountsBuilder()
	commonBuilder.SetDripFeeTokenAAccountAccount(botTokenAAccount)
	commonBuilder.SetVaultAccount(solana.MustPublicKeyFromBase58(config.Vault))
	commonBuilder.SetVaultProtoConfigAccount(solana.MustPublicKeyFromBase58(config.VaultProtoConfig))
	commonBuilder.SetLastVaultPeriodAccount(solana.MustPublicKeyFromBase58(vaultPeriodI.String()))
	commonBuilder.SetCurrentVaultPeriodAccount(solana.MustPublicKeyFromBase58(vaultPeriodJ.String()))
	commonBuilder.SetVaultTokenAAccountAccount(solana.MustPublicKeyFromBase58(config.VaultTokenAAccount))
	commonBuilder.SetVaultTokenBAccountAccount(solana.MustPublicKeyFromBase58(config.VaultTokenBAccount))
	commonBuilder.SetSwapTokenAAccountAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapTokenAAccount))
	commonBuilder.SetSwapTokenBAccountAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapTokenBAccount))
	commonBuilder.SetDripTriggerSourceAccount(w.wallet.PublicKey())

	txBuilder := drip.NewDripSplTokenSwapInstructionBuilder()
	txBuilder.SetCommonAccountsFromBuilder(commonBuilder)
	txBuilder.SetSwapTokenMintAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapTokenMint))
	txBuilder.SetSwapFeeAccountAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapFeeAccount))
	txBuilder.SetSwapAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.Swap))
	txBuilder.SetSwapAuthorityAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapAuthority))
	txBuilder.SetTokenSwapProgramAccount(solana.TokenSwapProgramID)
	commonBuilder.SetTokenProgramAccount(solana.TokenProgramID)
	return txBuilder.ValidateAndBuild()
}

func (w *SolanaClient) InitVaultPeriod(
	ctx context.Context, vault, vaultPeriod string, vaultPeriodID int64,
) (solana.Instruction, error) {
	txBuilder := drip.NewInitVaultPeriodInstructionBuilder()
	txBuilder.SetVaultPeriodAccount(solana.MustPublicKeyFromBase58(vaultPeriod))
	txBuilder.SetVaultAccount(solana.MustPublicKeyFromBase58(vault))
	txBuilder.SetCreatorAccount(w.wallet.PublicKey())
	txBuilder.SetSystemProgramAccount(solana.SystemProgramID)
	txBuilder.SetParams(drip.InitializeVaultPeriodParams{
		PeriodId: uint64(vaultPeriodID),
	})
	return txBuilder.ValidateAndBuild()
}

func (w *SolanaClient) CreateTokenAccount(
	ctx context.Context, owner solana.PublicKey, tokenMint solana.PublicKey,
) (solana.Instruction, error) {
	txBuilder := associatedtokenaccount.NewCreateInstructionBuilder()
	txBuilder.SetMint(tokenMint)
	txBuilder.SetPayer(w.wallet.PublicKey())
	txBuilder.SetWallet(owner)
	return txBuilder.ValidateAndBuild()
}

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
	VaultData           drip.Vault
	Vault               solana.PublicKey
	VaultPeriodI        solana.PublicKey
	VaultPeriodJ        solana.PublicKey
	BotTokenAFeeAccount solana.PublicKey
	WhirlpoolData       whirlpool.Whirlpool
	Whirlpool           solana.PublicKey
	TickArray0          solana.PublicKey
	TickArray1          solana.PublicKey
	TickArray2          solana.PublicKey
	Oracle              solana.PublicKey
}

func (w *SolanaClient) DripOrcaWhirlpool(
	ctx context.Context,
	params DripOrcaWhirlpoolParams,
) (solana.Instruction, error) {
	commonBuilder := drip.NewDripOrcaWhirlpoolCommonAccountsBuilder()
	commonBuilder.SetDripTriggerSourceAccount(w.wallet.PublicKey())
	commonBuilder.SetVaultAccount(params.Vault)
	commonBuilder.SetVaultProtoConfigAccount(params.VaultData.ProtoConfig)
	commonBuilder.SetLastVaultPeriodAccount(params.VaultPeriodI)
	commonBuilder.SetCurrentVaultPeriodAccount(params.VaultPeriodJ)
	commonBuilder.SetVaultTokenAAccountAccount(params.VaultData.TokenAAccount)
	commonBuilder.SetVaultTokenBAccountAccount(params.VaultData.TokenBAccount)
	commonBuilder.SetSwapTokenAAccountAccount(params.WhirlpoolData.TokenVaultA)
	commonBuilder.SetSwapTokenBAccountAccount(params.WhirlpoolData.TokenVaultB)
	commonBuilder.SetDripFeeTokenAAccountAccount(params.BotTokenAFeeAccount)
	commonBuilder.SetTokenProgramAccount(solana.TokenProgramID)

	txBuilder := drip.NewDripOrcaWhirlpoolInstructionBuilder()
	txBuilder.SetCommonAccountsFromBuilder(commonBuilder)
	txBuilder.SetWhirlpoolAccount(params.Whirlpool)
	txBuilder.SetTickArray0Account(params.TickArray0)
	txBuilder.SetTickArray1Account(params.TickArray1)
	txBuilder.SetTickArray2Account(params.TickArray2)
	txBuilder.SetOracleAccount(params.Oracle)

	txBuilder.SetWhirlpoolProgramAccount(whirlpool.ProgramID)

	return txBuilder.ValidateAndBuild()
}

type InitializeTickArrayParams struct {
	Whirlpool  solana.PublicKey
	StartIndex int32
	TickArray  solana.PublicKey
}

func (w *SolanaClient) InitializeTickArray(
	ctx context.Context,
	params InitializeTickArrayParams,
) (solana.Instruction, error) {
	txBuilder := whirlpool.NewInitializeTickArrayInstructionBuilder()
	txBuilder.SetWhirlpoolAccount(params.Whirlpool)
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
	ctx context.Context, vault, vaultProtoConfig, vaultPeriod, tokenAMint, tokenBMint string, vaultPeriodID int64,
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

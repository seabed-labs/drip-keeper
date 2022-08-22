package solanaclient

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
	txBuilder := drip.NewDripOrcaWhirlpoolInstructionBuilder()
	txBuilder.SetDripTriggerSourceAccount(w.wallet.PublicKey())

	txBuilder.SetVaultAccount(params.Vault)
	txBuilder.SetVaultProtoConfigAccount(params.VaultData.ProtoConfig)
	txBuilder.SetVaultTokenAAccountAccount(params.VaultData.TokenAAccount)
	txBuilder.SetVaultTokenBAccountAccount(params.VaultData.TokenBAccount)
	txBuilder.SetDripFeeTokenAAccountAccount(params.BotTokenAFeeAccount)
	txBuilder.SetLastVaultPeriodAccount(params.VaultPeriodI)
	txBuilder.SetCurrentVaultPeriodAccount(params.VaultPeriodJ)

	txBuilder.SetWhirlpoolAccount(params.Whirlpool)
	txBuilder.SetSwapTokenAAccountAccount(params.WhirlpoolData.TokenVaultA)
	txBuilder.SetSwapTokenBAccountAccount(params.WhirlpoolData.TokenVaultB)
	txBuilder.SetTickArray0Account(params.TickArray0)
	txBuilder.SetTickArray1Account(params.TickArray1)
	txBuilder.SetTickArray2Account(params.TickArray2)
	txBuilder.SetOracleAccount(params.Oracle)

	txBuilder.SetTokenProgramAccount(solana.TokenProgramID)
	txBuilder.SetAssociatedTokenProgramAccount(solana.SPLAssociatedTokenAccountProgramID)
	txBuilder.SetWhirlpoolProgramAccount(whirlpool.ProgramID)
	txBuilder.SetSystemProgramAccount(solana.SystemProgramID)
	txBuilder.SetRentAccount(solana.SysVarRentPubkey)

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
	txBuilder := drip.NewDripSplTokenSwapInstructionBuilder()
	txBuilder.SetDripTriggerSourceAccount(w.wallet.PublicKey())

	txBuilder.SetDripFeeTokenAAccountAccount(botTokenAAccount)
	txBuilder.SetVaultAccount(solana.MustPublicKeyFromBase58(config.Vault))
	txBuilder.SetVaultProtoConfigAccount(solana.MustPublicKeyFromBase58(config.VaultProtoConfig))
	txBuilder.SetLastVaultPeriodAccount(solana.MustPublicKeyFromBase58(vaultPeriodI.String()))
	txBuilder.SetCurrentVaultPeriodAccount(solana.MustPublicKeyFromBase58(vaultPeriodJ.String()))
	txBuilder.SetSwapTokenMintAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapTokenMint))
	txBuilder.SetVaultTokenAAccountAccount(solana.MustPublicKeyFromBase58(config.VaultTokenAAccount))
	txBuilder.SetVaultTokenBAccountAccount(solana.MustPublicKeyFromBase58(config.VaultTokenBAccount))
	txBuilder.SetSwapTokenAAccountAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapTokenAAccount))
	txBuilder.SetSwapTokenBAccountAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapTokenBAccount))
	txBuilder.SetSwapFeeAccountAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapFeeAccount))
	txBuilder.SetSwapAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.Swap))
	txBuilder.SetSwapAuthorityAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapAuthority))
	txBuilder.SetTokenSwapProgramAccount(solana.TokenSwapProgramID)
	txBuilder.SetTokenProgramAccount(solana.TokenProgramID)
	txBuilder.SetAssociatedTokenProgramAccount(solana.SPLAssociatedTokenAccountProgramID)
	txBuilder.SetSystemProgramAccount(solana.SystemProgramID)
	txBuilder.SetRentAccount(solana.SysVarRentPubkey)
	return txBuilder.ValidateAndBuild()
}

func (w *SolanaClient) InitVaultPeriod(
	ctx context.Context, vault, vaultProtoConfig, vaultPeriod, tokenAMint, tokenBMint string, vaultPeriodID int64,
) (solana.Instruction, error) {
	txBuilder := drip.NewInitVaultPeriodInstructionBuilder()
	txBuilder.SetVaultAccount(solana.MustPublicKeyFromBase58(vault))
	txBuilder.SetTokenAMintAccount(solana.MustPublicKeyFromBase58(tokenAMint))
	txBuilder.SetTokenBMintAccount(solana.MustPublicKeyFromBase58(tokenBMint))
	txBuilder.SetVaultProtoConfigAccount(solana.MustPublicKeyFromBase58(vaultProtoConfig))
	txBuilder.SetCreatorAccount(w.wallet.PublicKey())
	txBuilder.SetSystemProgramAccount(solana.SystemProgramID)
	txBuilder.SetVaultPeriodAccount(solana.MustPublicKeyFromBase58(vaultPeriod))
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

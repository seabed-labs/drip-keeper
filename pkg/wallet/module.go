package wallet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/dcaf-labs/solana-go-clients/pkg/drip"
	"github.com/dcaf-labs/solana-go-clients/pkg/whirlpool"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"github.com/sirupsen/logrus"
)

type WalletProvider struct {
	Client          *rpc.Client
	Wallet          *solana.Wallet
	FeeWalletPubkey solana.PublicKey
}

func New(
	config *configs.Config,
	solClient *rpc.Client,
) (*WalletProvider, error) {
	WalletProvider := WalletProvider{Client: solClient}
	if configs.IsLocal(config.Environment) {
		logrus.Infof("creating and funding test wallet")
		WalletProvider.Wallet = solana.NewWallet()
		if _, err := InitTestWallet(solClient, &WalletProvider); err != nil {
			return nil, err
		}
	} else {
		var accountBytes []byte
		if err := json.Unmarshal([]byte(config.Wallet), &accountBytes); err != nil {
			return nil, err
		}
		priv := base58.Encode(accountBytes)
		solWallet, err := solana.WalletFromPrivateKeyBase58(priv)
		if err != nil {
			return nil, err
		}
		WalletProvider.Wallet = solWallet
	}
	if config.FeeWallet != "" {
		WalletProvider.FeeWalletPubkey = solana.MustPublicKeyFromBase58(config.FeeWallet)
	} else {
		WalletProvider.FeeWalletPubkey = WalletProvider.Wallet.PublicKey()
	}
	logrus.
		WithField("Wallet", WalletProvider.Wallet.PublicKey()).
		WithField("FeeWalletPubkey", WalletProvider.FeeWalletPubkey.String()).
		Infof("loaded wallets")
	return &WalletProvider, nil
}

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

func (w *WalletProvider) DripOrcaWhirlpool(
	ctx context.Context,
	params DripOrcaWhirlpoolParams,
) (solana.Instruction, error) {
	txBuilder := drip.NewDripOrcaWhirlpoolInstructionBuilder()
	txBuilder.SetDripTriggerSourceAccount(w.Wallet.PublicKey())

	txBuilder.SetVaultAccount(params.Vault)
	txBuilder.SetVaultProtoConfigAccount(params.VaultData.ProtoConfig)
	txBuilder.SetTokenAMintAccount(params.VaultData.TokenAMint)
	txBuilder.SetTokenBMintAccount(params.VaultData.TokenBMint)
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

func (w *WalletProvider) InitializeTickArray(
	ctx context.Context,
	params InitializeTickArrayParams,
) (solana.Instruction, error) {
	txBuilder := whirlpool.NewInitializeTickArrayInstructionBuilder()
	txBuilder.SetWhirlpoolAccount(params.Whirlpool)
	txBuilder.SetFunderAccount(w.Wallet.PublicKey())
	txBuilder.SetTickArrayAccount(params.TickArray)
	txBuilder.SetSystemProgramAccount(solana.SystemProgramID)
	txBuilder.SetStartTickIndex(params.StartIndex)
	return txBuilder.ValidateAndBuild()
}

func (w *WalletProvider) DripSPLTokenSwap(
	ctx context.Context, config configs.DripConfig, vaultPeriodI, vaultPeriodJ, botTokenAAccount solana.PublicKey,
) (solana.Instruction, error) {
	txBuilder := drip.NewDripSplTokenSwapInstructionBuilder()
	txBuilder.SetDripTriggerSourceAccount(w.Wallet.PublicKey())

	txBuilder.SetDripFeeTokenAAccountAccount(botTokenAAccount)
	txBuilder.SetVaultAccount(solana.MustPublicKeyFromBase58(config.Vault))
	txBuilder.SetVaultProtoConfigAccount(solana.MustPublicKeyFromBase58(config.VaultProtoConfig))
	txBuilder.SetLastVaultPeriodAccount(solana.MustPublicKeyFromBase58(vaultPeriodI.String()))
	txBuilder.SetCurrentVaultPeriodAccount(solana.MustPublicKeyFromBase58(vaultPeriodJ.String()))
	txBuilder.SetSwapTokenMintAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.SwapTokenMint))
	txBuilder.SetTokenAMintAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.TokenAMint))
	txBuilder.SetTokenBMintAccount(solana.MustPublicKeyFromBase58(config.SPLTokenSwapConfig.TokenBMint))
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

// TODO(Mocha): These don't need to rely on the wallet
func (w *WalletProvider) InitVaultPeriod(
	ctx context.Context, vault, vaultProtoConfig, vaultPeriod, tokenAMint, tokenBMint string, vaultPeriodID int64,
) (solana.Instruction, error) {
	txBuilder := drip.NewInitVaultPeriodInstructionBuilder()
	txBuilder.SetVaultAccount(solana.MustPublicKeyFromBase58(vault))
	txBuilder.SetTokenAMintAccount(solana.MustPublicKeyFromBase58(tokenAMint))
	txBuilder.SetTokenBMintAccount(solana.MustPublicKeyFromBase58(tokenBMint))
	txBuilder.SetVaultProtoConfigAccount(solana.MustPublicKeyFromBase58(vaultProtoConfig))
	txBuilder.SetCreatorAccount(w.Wallet.PublicKey())
	txBuilder.SetSystemProgramAccount(solana.SystemProgramID)
	txBuilder.SetVaultPeriodAccount(solana.MustPublicKeyFromBase58(vaultPeriod))
	txBuilder.SetParams(drip.InitializeVaultPeriodParams{
		PeriodId: uint64(vaultPeriodID),
	})
	return txBuilder.ValidateAndBuild()
}

func (w *WalletProvider) CreateTokenAccount(
	ctx context.Context, owner solana.PublicKey, tokenMint solana.PublicKey,
) (solana.Instruction, error) {
	txBuilder := associatedtokenaccount.NewCreateInstructionBuilder()
	txBuilder.SetMint(tokenMint)
	txBuilder.SetPayer(w.Wallet.PublicKey())
	txBuilder.SetWallet(owner)
	return txBuilder.ValidateAndBuild()
}

func (w *WalletProvider) Send(
	ctx context.Context, instructions ...solana.Instruction,
) error {
	if len(instructions) == 0 {
		return nil
	}
	recent, err := w.Client.GetRecentBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		return err
	}
	logFields := logrus.Fields{"numInstructions": len(instructions), "block": recent.Value.Blockhash}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(w.Wallet.PublicKey()),
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction, err %s", err)
	}
	logrus.WithFields(logFields).Infof("built transaction")

	if _, err := tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if w.Wallet.PublicKey().Equals(key) {
				return &w.Wallet.PrivateKey
			}
			return nil
		},
	); err != nil {
		return fmt.Errorf("failed to sign transaction, err %s", err)
	}
	logrus.WithFields(logFields).Info("signed transaction")

	txHash, err := w.Client.SendTransactionWithOpts(
		ctx, tx, false, rpc.CommitmentConfirmed,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction, err %s", err)
	}
	logFields["txHash"] = txHash

	logrus.WithFields(logFields).Info("waiting for transaction to confirm")
	errC := make(chan error)
	go checkTxHash(w.Client, txHash, errC)
	if err := <-errC; err != nil {
		return err
	}
	return nil
}

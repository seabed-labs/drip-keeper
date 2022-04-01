package wallet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Dcaf-Protocol/keeper-bot/configs"
	dcaVault "github.com/Dcaf-Protocol/keeper-bot/generated/dca_vault"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"github.com/sirupsen/logrus"
)

type WalletProvider struct {
	Client *rpc.Client
	Wallet *solana.Wallet
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
	logrus.
		WithFields(logrus.Fields{"publicKey": WalletProvider.Wallet.PublicKey()}).
		Infof("loaded wallet")
	return &WalletProvider, nil
}

// TODO(Mocha): These don't need to rely on the wallet
func (w *WalletProvider) TriggerDCA(
	ctx context.Context, config configs.TriggerDCAConfig, vaultPeriodI, vaultPeriodJ, botTokenAAccount solana.PublicKey,
) (solana.Instruction, error) {
	txBuilder := dcaVault.NewTriggerDcaInstructionBuilder()
	txBuilder.SetDcaTriggerSourceAccount(w.Wallet.PublicKey())
	txBuilder.SetDcaTriggerFeeTokenAAccountAccount(botTokenAAccount)
	txBuilder.SetVaultAccount(solana.MustPublicKeyFromBase58(config.Vault))
	txBuilder.SetVaultProtoConfigAccount(solana.MustPublicKeyFromBase58(config.VaultProtoConfig))
	txBuilder.SetLastVaultPeriodAccount(solana.MustPublicKeyFromBase58(vaultPeriodI.String()))
	txBuilder.SetCurrentVaultPeriodAccount(solana.MustPublicKeyFromBase58(vaultPeriodJ.String()))
	txBuilder.SetSwapTokenMintAccount(solana.MustPublicKeyFromBase58(config.SwapTokenMint))
	txBuilder.SetTokenAMintAccount(solana.MustPublicKeyFromBase58(config.TokenAMint))
	txBuilder.SetTokenBMintAccount(solana.MustPublicKeyFromBase58(config.TokenBMint))
	txBuilder.SetVaultTokenAAccountAccount(solana.MustPublicKeyFromBase58(config.VaultTokenAAccount))
	txBuilder.SetVaultTokenBAccountAccount(solana.MustPublicKeyFromBase58(config.VaultTokenBAccount))
	txBuilder.SetSwapTokenAAccountAccount(solana.MustPublicKeyFromBase58(config.SwapTokenAAccount))
	txBuilder.SetSwapTokenBAccountAccount(solana.MustPublicKeyFromBase58(config.SwapTokenBAccount))
	txBuilder.SetSwapFeeAccountAccount(solana.MustPublicKeyFromBase58(config.SwapFeeAccount))
	txBuilder.SetSwapAccount(solana.MustPublicKeyFromBase58(config.Swap))
	txBuilder.SetSwapAuthorityAccount(solana.MustPublicKeyFromBase58(config.SwapAuthority))
	txBuilder.SetTokenSwapProgramAccount(solana.TokenSwapProgramID)
	txBuilder.SetTokenProgramAccount(solana.TokenProgramID)
	txBuilder.SetAssociatedTokenProgramAccount(solana.SPLAssociatedTokenAccountProgramID)
	txBuilder.SetSystemProgramAccount(solana.SystemProgramID)
	txBuilder.SetRentAccount(solana.SysVarRentPubkey)
	return txBuilder.ValidateAndBuild()
}

// TODO(Mocha): These don't need to rely on the wallet
func (w *WalletProvider) InitVaultPeriod(
	ctx context.Context, config configs.TriggerDCAConfig, vaultPeriod solana.PublicKey, vaultPeriodID int64,
) (solana.Instruction, error) {
	txBuilder := dcaVault.NewInitVaultPeriodInstructionBuilder()
	txBuilder.SetVaultAccount(solana.MustPublicKeyFromBase58(config.Vault))
	txBuilder.SetTokenAMintAccount(solana.MustPublicKeyFromBase58(config.TokenAMint))
	txBuilder.SetTokenBMintAccount(solana.MustPublicKeyFromBase58(config.TokenBMint))
	txBuilder.SetVaultProtoConfigAccount(solana.MustPublicKeyFromBase58(config.VaultProtoConfig))
	txBuilder.SetCreatorAccount(w.Wallet.PublicKey())
	txBuilder.SetSystemProgramAccount(solana.SystemProgramID)
	txBuilder.SetVaultPeriodAccount(vaultPeriod)
	txBuilder.SetParams(dcaVault.InitializeVaultPeriodParams{
		PeriodId: uint64(vaultPeriodID),
	})
	return txBuilder.ValidateAndBuild()
}

func (w *WalletProvider) CreateTokenAccount(
	ctx context.Context, tokenMint string,
) (solana.Instruction, error) {
	txBuilder := associatedtokenaccount.NewCreateInstructionBuilder()
	txBuilder.SetMint(solana.MustPublicKeyFromBase58(tokenMint))
	txBuilder.SetPayer(w.Wallet.PublicKey())
	txBuilder.SetWallet(w.Wallet.PublicKey())
	return txBuilder.ValidateAndBuild()
}

func (w *WalletProvider) Send(
	ctx context.Context, instructions ...solana.Instruction,
) error {
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

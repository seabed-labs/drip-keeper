package wallet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Dcaf-Protocol/keeper-bot/configs"
	dcaVault "github.com/Dcaf-Protocol/keeper-bot/pkg/generated/dca_vault"
	"github.com/Dcaf-Protocol/keeper-bot/pkg/util"
	"github.com/gagliardetto/solana-go"
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
	if !configs.IsProd(config.Environment) {
		logrus.Infof("creating & funding test wallet")
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

// // TODO(Mocha): Missing Tests
// TODO(Mocha): Decouple transaction building and transaction signing
func (w *WalletProvider) TriggerDCA(
	ctx context.Context, vaultPubkey string,
) error {
	goID := util.GoRoutineID()
	logFields := logrus.Fields{"vault": vaultPubkey, "id": goID}
	logrus.WithFields(logFields).Infof("starting DCA")

	txBuilder := dcaVault.NewTriggerDcaInstructionBuilder()
	txBuilder.SetClientAccount(w.Wallet.PublicKey())
	txBuilder.SetVaultAccount(solana.MustPublicKeyFromBase58("TODO"))
	recent, err := w.Client.GetRecentBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		return err
	}
	logFields["block"] = recent.Value.Blockhash
	logrus.WithFields(logFields).Infof("DCA recrny block")

	tx, err := solana.NewTransaction(
		[]solana.Instruction{txBuilder.Build()},
		recent.Value.Blockhash,
		solana.TransactionPayer(w.Wallet.PublicKey()),
	)
	if err != nil {
		return fmt.Errorf("failed to create Trigger DCA transaction, err %s", err)
	}
	logrus.WithFields(logFields).Infof("built Trigger DCA transaction")

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if w.Wallet.PublicKey().Equals(key) {
				return &w.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to sign transaction, err %s", err)
	}
	logrus.WithFields(logFields).Info("signed Trigger DCA transaction")

	sig, err := w.Client.SendTransactionWithOpts(
		ctx, tx, false, rpc.CommitmentConfirmed,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction, err %s", err)
	}
	logFields["sig"] = sig
	logrus.WithFields(logFields).Info("broadcast Trigger DCA transaction")

	return nil
}

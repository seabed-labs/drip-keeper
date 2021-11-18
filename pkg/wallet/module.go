package wallet

import (
	"context"
	"encoding/json"

	"github.com/Dcaf-Protocol/keeper-bot/configs"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/types"
	"github.com/sirupsen/logrus"
)

type Wallet struct {
	client  *client.Client
	account types.Account
}

func NewWallet(
	secrets *configs.Secrets,
) (*Wallet, error) {
	wallet := Wallet{}
	if secrets.IsLocal() {
		wallet.client = client.NewClient(rpc.LocalnetRPCEndpoint)
	} else {
		wallet.client = client.NewClient(rpc.MainnetRPCEndpoint)
	}
	if resp, err := wallet.client.GetVersion(context.Background()); err != nil {
		logrus.WithError(err).Fatalf("failed to get client version info")
		return nil, err
	} else {
		logrus.
			WithFields(logrus.Fields{"version": resp.SolanaCore}).
			Info("created solClient")
	}

	var accountBytes []byte
	if err := json.Unmarshal([]byte(secrets.Account), &accountBytes); err != nil {
		return nil, err
	}
	account, err := types.AccountFromBytes(accountBytes)
	if err != nil {
		return nil, err
	}
	wallet.account = account
	logrus.
		WithFields(logrus.Fields{"publicKey": wallet.account.PublicKey.ToBase58()}).
		Infof("loaded wallet")
	return &wallet, nil
}

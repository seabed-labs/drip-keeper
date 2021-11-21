package wallet

import (
	"encoding/json"

	"github.com/Dcaf-Protocol/keeper-bot/configs"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"github.com/sirupsen/logrus"
)

type Wallet struct {
	Client  *rpc.Client
	Account *solana.Wallet
}

func NewWallet(
	secrets *configs.Secrets,
	solClient *rpc.Client,
) (*Wallet, error) {
	wallet := Wallet{Client: solClient}
	if secrets.Environment != configs.ProdEnv && secrets.Account == "" {
		logrus.Infof("creating & funding test wallet")
		wallet.Account = solana.NewWallet()
		if _, err := InitTestWallet(solClient, &wallet); err != nil {
			return nil, err
		}
	} else {
		var accountBytes []byte
		if err := json.Unmarshal([]byte(secrets.Account), &accountBytes); err != nil {
			return nil, err
		}
		priv := base58.Encode(accountBytes)
		solWallet, err := solana.WalletFromPrivateKeyBase58(priv)
		if err != nil {
			return nil, err
		}
		wallet.Account = solWallet
	}
	logrus.
		WithFields(logrus.Fields{"publicKey": wallet.Account.PublicKey()}).
		Infof("loaded wallet")
	return &wallet, nil
}

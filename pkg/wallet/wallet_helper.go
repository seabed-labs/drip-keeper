package wallet

import (
	"context"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

func InitTestWallet(
	solClient *rpc.Client, walletProvider *WalletProvider,
) (uint64, error) {
	txHash, err := solClient.RequestAirdrop(
		context.Background(), walletProvider.Wallet.PublicKey(),
		solana.LAMPORTS_PER_SOL*50, rpc.CommitmentConfirmed)
	if err != nil {
		return 0, err
	}
	errC := make(chan error)
	go checkTxHash(solClient, txHash, errC)
	if err := <-errC; err != nil {
		return 0, err
	}
	balance, err := solClient.GetBalance(
		context.Background(), walletProvider.Wallet.PublicKey(),
		rpc.CommitmentConfirmed)
	if err != nil {
		return 0, err
	}
	logrus.WithField("balance", balance.Value).Infof("intiazlied wallet")
	return balance.Value, nil
}

// TODO: This can be a generic function to check for a transaction till its confirmed
func checkTxHash(
	client *rpc.Client, txHash solana.Signature, done chan error,
) {
	// Should take max 13s
	timeout := time.Second * 15
	ticker := time.NewTicker(timeout)
	for {
		select {
		case <-ticker.C:
			done <- fmt.Errorf("timeout waiting for tx to confirm")
			return
		default:
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		res, err := client.GetSignatureStatuses(
			ctx, false, txHash)
		cancel()
		if err == nil && res != nil && len(res.Value) == 1 && res.Value[0] != nil && res.Value[0].ConfirmationStatus == rpc.ConfirmationStatusConfirmed {
			done <- nil
			return
		}
		if err != nil {
			logrus.WithError(err).Warnf("error getting signature status, retrying")
		}
	}
}

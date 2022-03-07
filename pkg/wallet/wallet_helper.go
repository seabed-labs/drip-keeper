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
	solClient *rpc.Client, wallet *Wallet,
) (uint64, error) {
	if _, err := solClient.RequestAirdrop(
		context.Background(), wallet.Account.PublicKey(),
		solana.LAMPORTS_PER_SOL*5, rpc.CommitmentConfirmed); err != nil {
		return 0, err
	}
	errC := make(chan error)
	go checkAirDrop(solClient, wallet.Account.PublicKey(), errC)
	if err := <-errC; err != nil {
		return 0, err
	}
	balance, err := solClient.GetBalance(
		context.Background(), wallet.Account.PublicKey(),
		rpc.CommitmentConfirmed)
	if err != nil {
		return 0, err
	}
	logrus.WithField("balance", balance.Value).Infof("intiazlied wallet")
	return balance.Value, nil
}

// TODO: This can be a generic function to check for a transaction till its confirmed
func checkAirDrop(
	client *rpc.Client, pubkey solana.PublicKey, done chan error,
) {
	// Should take max 13s
	timeout := time.Second * 15
	ticker := time.NewTicker(timeout)
	for {
		select {
		case <-ticker.C:
			done <- fmt.Errorf("timeout waiting for airdrop")
			return
		default:
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		balance, err := client.GetBalance(
			ctx, pubkey, rpc.CommitmentFinalized)
		cancel()
		if err == nil && balance.Value != 0 {
			done <- nil
			return
		}
		if err != nil {
			logrus.WithError(err).Warnf("error getting balance, retrying")
		}
	}
}

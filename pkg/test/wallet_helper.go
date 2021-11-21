package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Dcaf-Protocol/keeper-bot/pkg/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func InitTestWallet(
	t *testing.T, solClient *rpc.Client, wallet *wallet.Wallet,
) {
	_, err := solClient.GetHealth(context.Background())
	assert.NoError(t, err)
	_, err = solClient.RequestAirdrop(
		context.Background(), wallet.Account.PublicKey(),
		solana.LAMPORTS_PER_SOL*5, rpc.CommitmentFinalized)
	if err != nil {
		logrus.WithError(err).Errorf("%v", err)
	}
	assert.NoError(t, err)
	errC := make(chan error)
	go checkAirDrop(solClient, wallet.Account.PublicKey(), errC)
	err = <-errC
	assert.NoError(t, err)
	balance, err := solClient.GetBalance(
		context.Background(), wallet.Account.PublicKey(),
		rpc.CommitmentFinalized)
	assert.NoError(t, err)
	assert.NoError(t, err)
	logrus.WithField("balance", balance.Value).Infof("intiazlied wallet")
	assert.NotZero(t, balance.Value)
}

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
		ctx, _ := context.WithTimeout(context.Background(), timeout)
		balance, err := client.GetBalance(
			ctx, pubkey, rpc.CommitmentFinalized)
		if err == nil && balance.Value != 0 {
			done <- nil
			return
		}
		if err != nil {
			logrus.WithError(err).Warnf("error getting balance, retrying")
		}
	}
}

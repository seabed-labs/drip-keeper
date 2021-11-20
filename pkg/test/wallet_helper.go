package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Dcaf-Protocol/keeper-bot/pkg/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/assert"
)

func InitTestWallet(t *testing.T, solClient *rpc.Client, wallet *wallet.Wallet) {
	_, err := solClient.RequestAirdrop(
		context.Background(), wallet.Account.PublicKey(),
		solana.LAMPORTS_PER_SOL*5, rpc.CommitmentFinalized)
	assert.NoError(t, err)
	errC := make(chan error)
	go checkAirDrop(solClient, wallet.Account.PublicKey(), errC)
	err = <-errC
	assert.NoError(t, err)
	balance, err := solClient.GetBalance(context.Background(), wallet.Account.PublicKey(), "")
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.NotZero(t, balance.Value)
}

func checkAirDrop(
	client *rpc.Client, pubkey solana.PublicKey, done chan error,
) {
	timeout := time.Second * 10
	ticker := time.NewTicker(timeout)
	for {
		select {
		case <-ticker.C:
			done <- fmt.Errorf("timeout waiting for airdrop")
			return
		default:
		}
		ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(timeout))
		balance, err := client.GetBalance(ctx, pubkey, "")
		if err == nil && balance.Value != 0 {
			done <- nil
			return
		}
		if err != nil {
			fmt.Println(err)
		}
	}
}

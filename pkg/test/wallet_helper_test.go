package test

import (
	"testing"

	"github.com/Dcaf-Protocol/keeper-bot/pkg/wallet"
	"github.com/gagliardetto/solana-go/rpc"
)

func TestInitWallet(t *testing.T) {
	InjectDependencies(func(
		solClient *rpc.Client,
		wallet *wallet.Wallet,
	) {
		InitTestWallet(t, solClient, wallet)
	})
}

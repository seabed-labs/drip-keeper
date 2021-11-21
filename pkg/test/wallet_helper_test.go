package test

import (
	"testing"

	"github.com/Dcaf-Protocol/keeper-bot/pkg/wallet"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/assert"
)

func TestInitWallet(t *testing.T) {
	InjectDependencies(func(
		solClient *rpc.Client,
		wallet *wallet.Wallet,
	) {
		balance, err := InitTestWallet(solClient, wallet)
		assert.NoError(t, err)
		assert.NotZero(t, balance)
	})
}

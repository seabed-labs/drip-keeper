package wallet_test

import (
	"context"
	"testing"

	"github.com/Dcaf-Protocol/keeper-bot/pkg/test"
	walletPkg "github.com/Dcaf-Protocol/keeper-bot/pkg/wallet"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/assert"
)

func TestInitWallet(t *testing.T) {
	test.InjectDependencies(func(
		solClient *rpc.Client,
		wallet *walletPkg.Wallet,
	) {
		originalBalance, err := solClient.GetBalance(context.Background(), wallet.Account.PublicKey(), rpc.CommitmentConfirmed)
		assert.NoError(t, err)
		balance, err := walletPkg.InitTestWallet(solClient, wallet)
		assert.NoError(t, err)
		assert.NotZero(t, balance)
		assert.NotEqual(t, originalBalance, balance)
	})
}

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
		walletProvider *walletPkg.WalletProvider,
	) {
		originalBalance, err := solClient.GetBalance(context.Background(), walletProvider.Wallet.PublicKey(), rpc.CommitmentConfirmed)
		assert.NoError(t, err)
		balance, err := walletPkg.InitTestWallet(solClient, walletProvider)
		assert.NoError(t, err)
		assert.NotZero(t, balance)
		assert.NotEqual(t, originalBalance, balance)
	})
}

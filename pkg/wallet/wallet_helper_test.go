package wallet_test

import (
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
		balance, err := walletPkg.InitTestWallet(solClient, wallet)
		assert.NoError(t, err)
		assert.NotZero(t, balance)
	})
}

package client

import (
	"context"

	"github.com/Dcaf-Protocol/keeper-bot/configs"

	// "github.com/portto/solana-go-sdk/client"
	// "github.com/portto/solana-go-sdk/rpc"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

func NewSolanaClient(
	secrets *configs.Secrets,
) (*rpc.Client, error) {
	var solClient rpc.Client
	if secrets.IsLocal() {
		solClient = *rpc.New(rpc.LocalNet_RPC)
	} else {
		solClient = *rpc.New(rpc.MainNetBetaSerum_RPC)
	}
	if resp, err := solClient.GetVersion(context.Background()); err != nil {
		logrus.WithError(err).Fatalf("failed to get client version info")
		return nil, err
	} else {
		logrus.
			WithFields(logrus.Fields{"version": resp.SolanaCore}).
			Info("created solClient")
	}
	return &solClient, nil
}

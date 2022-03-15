package client

import (
	"context"

	"github.com/Dcaf-Protocol/keeper-bot/configs"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

func NewSolanaClient(
	config *configs.Config,
) (*rpc.Client, error) {
	url := getURL(config.Environment)
	solClient := *rpc.New(url)
	resp, err := solClient.GetVersion(context.Background())
	if err != nil {
		logrus.WithError(err).Fatalf("failed to get client version info")
		return nil, err
	}
	logrus.
		WithFields(logrus.Fields{
			"version": resp.SolanaCore,
			"url":     url}).
		Info("created solClient")
	return &solClient, nil
}

func getURL(env configs.Environment) string {
	switch env {
	case configs.DevEnv:
		return rpc.DevNet_RPC
	case configs.ProdEnv:
		return rpc.MainNetBeta_RPC
	case configs.NilEnv:
		fallthrough
	case configs.LocalEnv:
		fallthrough
	default:
		return rpc.LocalNet_RPC
	}
}

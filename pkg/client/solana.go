package client

import (
	"context"

	"github.com/Dcaf-Protocol/keeper-bot/configs"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

func NewSolanaClient(
	secrets *configs.Secrets,
) (*rpc.Client, error) {
	url := getURL(secrets.Environment)
	solClient := *rpc.New(url)
	if resp, err := solClient.GetVersion(context.Background()); err != nil {
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
	case configs.LocalEnv:
	case configs.TestEnv:
	default:
		return rpc.LocalNet_RPC
	}
	return rpc.LocalNet_RPC
}

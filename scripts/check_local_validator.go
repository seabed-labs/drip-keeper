package main

import (
	"context"
	"os"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// TODO(Mocha): Pass in custom amount
	maxTries := 10
	try := 0

	timeout := time.Second * 2
	url := rpc.LocalNet_RPC

	for try < maxTries {
		logrus.
			WithFields(logrus.Fields{
				"try": try,
				"url": url}).
			Info("getting client version ....")
		solClient := rpc.New(rpc.LocalNet_RPC)
		resp, err := solClient.GetVersion(context.Background())
		if err != nil {
			logrus.WithError(err).Errorf("failed to get client version, retrying...")
			time.Sleep(timeout)
			try += 1
		} else {
			logrus.
				WithFields(logrus.Fields{
					"version": resp.SolanaCore,
					"url":     url}).
				Info("")
			return
		}
	}
	logrus.
		WithFields(logrus.Fields{
			"try": try}).
		Errorf("failed to get client version info")
	os.Exit(1)
}

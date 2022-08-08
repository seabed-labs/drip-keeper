package solana

import (
	"context"
	"net/http"
	"time"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

func NewSolanaClient(
	config *configs.Config,
) (*rpc.Client, error) {
	url := getURL(config.Environment)
	// Maximum number of requests per 10 seconds per IP for a single RPC: 40
	rateLimiter := rate.NewLimiter(rate.Every(time.Second*10/40), 1)
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = nil
	httpClient.RetryWaitMin = time.Second * 5
	httpClient.RetryMax = 3
	httpClient.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retry int) {
		if err := rateLimiter.Wait(context.Background()); err != nil {
			logrus.WithField("err", err.Error()).Warnf("waiting for rate limit")
			return
		}
	}
	solClient := rpc.NewWithCustomRPCClient(jsonrpc.NewClientWithOpts(url, &jsonrpc.RPCClientOpts{
		HTTPClient:    httpClient.StandardClient(),
		CustomHeaders: nil,
	}))
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
	return solClient, nil
}

func getURL(env configs.Environment) string {
	switch env {
	case configs.DevnetEnv:
		return rpc.DevNet_RPC
	case configs.MainnetEnv:
		return rpc.MainNetBeta_RPC
	case configs.NilEnv:
		fallthrough
	case configs.LocalnetEnv:
		fallthrough
	default:
		return rpc.LocalNet_RPC
	}
}

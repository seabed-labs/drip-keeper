package orcawhirlpool

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Dcaf-Protocol/drip-keeper/pkg/service/clients"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
	dripextension "github.com/dcaf-labs/drip-client/drip-extension-go"
)

type OrcaWhirlpoolClient interface {
	GetOrcaWhirlpoolQuoteEstimate(ctx context.Context, whirlpool, inputTokenMint, inputAmount string) (*dripextension.V1OrcawhirlpoolQuote200Response, error)
}

func NewOrcaWhirlpoolClient(
	config *configs.Config,
) OrcaWhirlpoolClient {
	return newClient(config.SolanaRPCURL)
}

type client struct {
	*dripextension.APIClient
	connectionUrl string
}

func newClient(rpcUrl string) *client {
	httpClient := clients.GetRateLimitedHTTPClient(callsPerSecond)

	config := dripextension.NewConfiguration()
	config.HTTPClient = httpClient.StandardClient()
	config.Host = host
	config.UserAgent = "drip-keeper"
	config.Scheme = "https"
	return &client{
		APIClient:     dripextension.NewAPIClient(config),
		connectionUrl: rpcUrl,
	}
}

func (c *client) GetOrcaWhirlpoolQuoteEstimate(
	ctx context.Context,
	whirlpool, inputTokenMint, inputAmount string,
) (res *dripextension.V1OrcawhirlpoolQuote200Response, err error) {
	res, httpRes, err := c.DefaultApi.
		V1OrcawhirlpoolQuote(ctx).
		V1OrcawhirlpoolQuoteRequest(dripextension.V1OrcawhirlpoolQuoteRequest{
			ConnectionUrl:  c.connectionUrl,
			Whirlpool:      whirlpool,
			InputTokenMint: inputTokenMint,
			InputAmount:    inputAmount,
		}).
		Execute()
	if err != nil {
		return nil, err
	} else if httpRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("V1OrcawhirlpoolQuoteExecute returned non-200, statusCode: %d", httpRes.StatusCode)
	} else if res == nil {
		return nil, fmt.Errorf("nil V1OrcawhirlpoolQuote200Response with 200 status")
	}
	return res, nil
}

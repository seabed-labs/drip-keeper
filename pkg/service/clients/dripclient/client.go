package dripclient

import (
	"github.com/Dcaf-Protocol/drip-keeper/configs"
	"github.com/dcaf-labs/drip-client/drip-go"
)

func NewDripBackendClient(
	appConfig *configs.Config,
) *drip.APIClient {
	config := drip.NewConfiguration()
	config.Host = appConfig.DiscoveryURL
	config.UserAgent = "drip-keeper"
	config.Scheme = "https"
	return drip.NewAPIClient(config)
}

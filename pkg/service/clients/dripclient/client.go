package dripclient

import (
	"strings"

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
	if strings.Contains(appConfig.DiscoveryURL, "localhost") {
		config.Scheme = "http"
	}
	return drip.NewAPIClient(config)
}

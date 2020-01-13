package portalclient

import "github.com/kubitre/diplom/configuration"

type PortalClient struct {
	Config *configuration.ConfigurationExecutor
}

func NewPortalClient(config *configuration.ConfigurationExecutor) *PortalClient {
	return &PortalClient{
		Config: config,
	}
}

package configapi

import (
	"context"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/configs"
)

type ConfigAPIPlugin struct {
	Name string `toml:"name"`

	api    *api
	cancel context.CancelFunc
}

func (a *ConfigAPIPlugin) GetName() string {
	if a.Name != "" {
		return "api." + a.Name
	}
	return "api"
}

func (a *ConfigAPIPlugin) Init(ctx context.Context, cfg *config.Config, agent config.AgentController) error {
	a.api, a.cancel = newAPI(ctx, cfg, agent)

	// start listening for requests

	return nil
}

func (a *ConfigAPIPlugin) Close() error {
	// stop accepting new requests
	// wait until all requests finish
	// shut down server
	// trigger all plugins to stop and wait for them to exit
	return nil
}

func init() {
	configs.Add("api", func() config.ConfigPlugin {
		return &ConfigAPIPlugin{}
	})
}

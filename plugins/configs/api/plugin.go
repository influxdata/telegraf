package api

import (
	"context"
	"fmt"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/configs"
)

type ConfigAPIPlugin struct {
	// Name string `toml:"name"`
	Protocol []string `toml:"protocol"` // protocol = ["http", "grpc", "websocket"]
	// Storage string `toml:"storage"` // storage = "config_state"
	Storage config.StoragePlugin `toml:"storage"`
	// [config.api.storage.internal]
	//   file = "config_state.db"

	api    *api
	cancel context.CancelFunc

	plugins []PluginConfig
}

// type RunningPlugins struct {
// 	Plugins []Plugin `json:"plugins"`
// }

// func (a *ConfigAPIPlugin) GetName() string {
// 	if a.Name != "" {
// 		return "api." + a.Name
// 	}
// 	return "api"
// }

func (a *ConfigAPIPlugin) GetName() string {
	return "api"
}

func (a *ConfigAPIPlugin) Init(ctx context.Context, cfg *config.Config, agent config.AgentController) error {
	a.api, a.cancel = newAPI(ctx, cfg, agent)

	// TODO: is this needed?
	if err := a.Storage.Init(); err != nil {
		return nil
	}

	if err := a.Storage.Load("config-api", "plugins", &a.plugins); err != nil {
		return fmt.Errorf("loading plugin state: %w", err)
	}

	// start listening for HTTP requests

	return nil
}

func (a *ConfigAPIPlugin) Close() error {
	fmt.Println("api closing")
	// stop accepting new requests
	// wait until all requests finish
	// store state
	m := map[string]interface{}{}
	for _, plugin := range a.plugins {
		m[plugin.ID] = map[string]interface{}{
			"name":   plugin.Name,
			"config": plugin.Config,
		}
	}

	if err := a.Storage.Save("config-api", "plugins", &a.plugins); err != nil {
		return fmt.Errorf("saving plugin state: %w", err)
	}
	// shut down server
	// trigger all plugins to stop and wait for them to exit
	return nil
}

func init() {
	configs.Add("api", func() config.ConfigPlugin {
		return &ConfigAPIPlugin{}
	})
}

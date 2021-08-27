package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/configs"
)

type ConfigAPIPlugin struct {
	ServiceAddress string               `toml:"service_address"`
	Storage        config.StoragePlugin `toml:"storage"`
	tls.ServerConfig

	api    *api
	server *ConfigAPIService

	Log          telegraf.Logger `toml:"-"`
	plugins      []PluginConfig
	pluginsMutex sync.Mutex
}

func (a *ConfigAPIPlugin) GetName() string {
	return "api"
}

// Init initializes the config api plugin.
// nolint:revive
func (a *ConfigAPIPlugin) Init(ctx context.Context, outputCtx context.Context, cfg *config.Config, agent config.AgentController) error {
	a.api = newAPI(ctx, outputCtx, cfg, agent)
	if a.Storage == nil {
		a.Log.Warn("initializing config-api without storage, changes via the api will not be persisted.")
	} else {
		if err := a.Storage.Init(); err != nil {
			return fmt.Errorf("initializing storage: %w", err)
		}

		if err := a.Storage.Load("config-api", "plugins", &a.plugins); err != nil {
			return fmt.Errorf("loading plugin state: %w", err)
		}
	}

	a.Log.Info(fmt.Sprintf("Loading %d stored plugins", len(a.plugins)))
	for _, plug := range a.plugins {
		id, err := a.api.CreatePlugin(plug.PluginConfigCreate, models.PluginID(plug.ID))
		if err != nil {
			a.Log.Errorf("Couldn't recreate plugin %q: %s", id, err)
		}
	}

	a.api.OnPluginAdded(func(p PluginConfig) {
		a.pluginsMutex.Lock()
		defer a.pluginsMutex.Unlock()
		a.plugins = append(a.plugins, p)
		if err := a.Storage.Save("config-api", "plugins", &a.plugins); err != nil {
			a.Log.Error("saving plugin state: " + err.Error())
		}
	})
	a.api.OnPluginRemoved(func(p PluginConfig) {
		a.pluginsMutex.Lock()
		defer a.pluginsMutex.Unlock()
		changed := false
		for i, plug := range a.plugins {
			if plug.ID == p.ID {
				a.plugins = append(a.plugins[:i], a.plugins[i+1:]...)
				changed = true
			}
		}
		if changed {
			if err := a.Storage.Save("config-api", "plugins", &a.plugins); err != nil {
				a.Log.Error("saving plugin state: " + err.Error())
			}
		}
	})

	// start listening for HTTP requests
	tlsConfig, err := a.TLSConfig()
	if err != nil {
		return err
	}
	if a.ServiceAddress == "" {
		a.ServiceAddress = ":7551"
	}
	a.server = newConfigAPIService(&http.Server{
		Addr:      a.ServiceAddress,
		TLSConfig: tlsConfig,
	}, a.api, a.Log)

	a.server.Start()
	return nil
}

func (a *ConfigAPIPlugin) Close() error {
	// shut down server
	// stop accepting new requests
	// wait until all requests finish
	a.server.Stop()

	a.pluginsMutex.Lock()
	defer a.pluginsMutex.Unlock()

	// store state
	if a.Storage != nil {
		if err := a.Storage.Save("config-api", "plugins", &a.plugins); err != nil {
			return fmt.Errorf("saving plugin state: %w", err)
		}
	}
	return nil
}

func init() {
	configs.Add("api", func() config.ConfigPlugin {
		return &ConfigAPIPlugin{}
	})
}

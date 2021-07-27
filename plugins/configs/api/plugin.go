package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/configs"
)

type ConfigAPIPlugin struct {
	ServiceAddress string               `toml:"service_address"`
	Storage        config.StoragePlugin `toml:"storage"`
	tls.ServerConfig

	api    *api
	cancel context.CancelFunc
	server *ConfigAPIService

	Log     telegraf.Logger `toml:"-"`
	plugins []PluginConfig
}

func (a *ConfigAPIPlugin) GetName() string {
	return "api"
}

func (a *ConfigAPIPlugin) Init(ctx context.Context, cfg *config.Config, agent config.AgentController) error {
	a.api, a.cancel = newAPI(ctx, cfg, agent)

	if err := a.Storage.Init(); err != nil {
		return fmt.Errorf("initializing storage: %w", err)
	}

	if err := a.Storage.Load("config-api", "plugins", &a.plugins); err != nil {
		return fmt.Errorf("loading plugin state: %w", err)
	}

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

	// store state
	if err := a.Storage.Save("config-api", "plugins", &a.plugins); err != nil {
		return fmt.Errorf("saving plugin state: %w", err)
	}
	return nil
}

func init() {
	configs.Add("api", func() config.ConfigPlugin {
		return &ConfigAPIPlugin{}
	})
}

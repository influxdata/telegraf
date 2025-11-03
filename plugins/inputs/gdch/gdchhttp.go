package gdch

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	http_plugin "github.com/influxdata/telegraf/plugins/inputs/http"
	"github.com/influxdata/telegraf/plugins/secretstores/gdch"
)

//go:embed sample.conf
var sampleConfig string

// GdchHttp is the main plugin struct
type GdchHttp struct {
	Http *http_plugin.HTTP `toml:"http"` // Embedded http plugin
	Auth *gdch.GdchAuth    `toml:"auth"` // GDCH authenticator

	Log telegraf.Logger `toml:"-"`
}

// --- Telegraf Plugin Interface Methods ---

// Description returns a one-sentence description of the plugin
func (g *GdchHttp) Description() string {
	return "Wraps the http input plugin to add GDCH service account auth"
}

func (g *GdchHttp) SampleConfig() string {
	return sampleConfig
}

// Init is called once when the plugin starts.
// This is where we load the key file and initialize the embedded http plugin.
func (g *GdchHttp) Init() error {
	if g.Http == nil {
		return errors.New("http plugin configuration is missing")
	}
	if g.Auth == nil {
		return errors.New("auth configuration is missing")
	}

	g.Auth.SetLogger(g.Log)
	if err := g.Auth.Init(); err != nil {
		return fmt.Errorf("failed to initialize auth module: %w", err)
	}

	g.Log.Info("GDCH HTTP plugin initialized. Calling Init() on embedded http plugin.")
	return g.Http.Init()
}

// Gather is the main method called by Telegraf at each interval
func (g *GdchHttp) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	token, err := g.Auth.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}
	g.Http.Token = config.NewSecret([]byte(token))

	return g.Http.Gather(acc)
}

// SetParserFunc passes the parser function to the embedded http plugin.
func (g *GdchHttp) SetParserFunc(fn telegraf.ParserFunc) {
	g.Http.SetParserFunc(fn)
}

// --- Telegraf Plugin Registration ---

// init registers the plugin with Telegraf
func init() {
	inputs.Add("gdch_http",
		func() telegraf.Input {
			return &GdchHttp{ //nolint:staticcheck // Setting HTTP is required for the plugin to function.
				Http: &http_plugin.HTTP{},
				Auth: &gdch.GdchAuth{},
			}
		})
}

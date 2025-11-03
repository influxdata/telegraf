package gdchhttp

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	http_plugin "github.com/influxdata/telegraf/plugins/inputs/http"
	"github.com/influxdata/telegraf/plugins/secretstores/gdchauth"
)

//go:embed sample.conf
var sampleConfig string

// GdchHTTP is the main plugin struct
type GdchHTTP struct {
	HTTP *http_plugin.HTTP  `toml:"http"` // Embedded http plugin
	Auth *gdchauth.GdchAuth `toml:"auth"` // GDCH authenticator

	Log telegraf.Logger `toml:"-"`
}

// --- Telegraf Plugin Interface Methods ---

// Description returns a one-sentence description of the plugin
func (*GdchHTTP) Description() string {
	return "Wraps the http input plugin to add GDCH service account auth"
}

func (*GdchHTTP) SampleConfig() string {
	return sampleConfig
}

// Init is called once when the plugin starts.
// This is where we load the key file and initialize the embedded http plugin.
func (g *GdchHTTP) Init() error {
	if g.HTTP == nil {
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
	return g.HTTP.Init()
}

// Gather is the main method called by Telegraf at each interval
func (g *GdchHTTP) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	token, err := g.Auth.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}
	g.HTTP.Token = config.NewSecret([]byte(token))

	return g.HTTP.Gather(acc)
}

// SetParserFunc passes the parser function to the embedded http plugin.
func (g *GdchHTTP) SetParserFunc(fn telegraf.ParserFunc) {
	g.HTTP.SetParserFunc(fn)
}

// --- Telegraf Plugin Registration ---

// init registers the plugin with Telegraf
func init() {
	inputs.Add("gdch_http",
		func() telegraf.Input {
			return &GdchHTTP{
				HTTP: &http_plugin.HTTP{},
				Auth: &gdchauth.GdchAuth{},
			}
		})
}

package gdch

import (
	"context"
	"errors"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	http_plugin "github.com/influxdata/telegraf/plugins/inputs/http" // Alias the http plugin
)

// GdchHttp is the main plugin struct
type GdchHttp struct {
	Http *http_plugin.HTTP `toml:"http"` // Embedded http plugin
	Auth *GdchAuth         `toml:"auth"` // GDCH authenticator

	Log telegraf.Logger `toml:"-"`
}

// --- Telegraf Plugin Interface Methods ---

// Description returns a one-sentence description of the plugin
func (g *GdchHttp) Description() string {
	return "Wraps the http input plugin to add GDCH service account auth"
}

func (g *GdchHttp) SampleConfig() string {
	return `
  [[inputs.gdch_http]]
  data_format = "json_v2"

  [inputs.gdch_http.auth]
    ## Path to the GDCH service account JSON key file
    service_account_file = "/etc/telegraf/gdch-key.json"
    audience = "https://{GDCH_URL}"
	## Time before token expiry to fetch a new one.
	# token_expiry_buffer = "5m"

    [inputs.gdch_http.auth.tls]
      insecure_skip_verify = true
	  ## Optional TLS configuration for the token endpoint.
  	  # tls_ca = "/etc/telegraf/ca.pem"
  
  ## Embedded HTTP Input Plugin Configuration.
  [inputs.gdch_http.http] 
    ## A list of URLs to pull data from.
    urls = [
      "https://{GDCH_URL}/{PROJECT}/metrics"
    ]
    ## ... other http plugin options ...
`
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

	g.Auth.Log = g.Log
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
			return &GdchHttp{
				Http: &http_plugin.HTTP{},
				Auth: &GdchAuth{},
			}
		})
}

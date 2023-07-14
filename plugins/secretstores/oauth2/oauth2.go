//go:generate ../../../tools/readme_config_includer/generator
package oauth2

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/endpoints"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/secretstores"
)

//go:embed sample.conf
var sampleConfig string

type TokenConfig struct {
	Key          string            `toml:"key"`
	ClientID     config.Secret     `toml:"client_id"`
	ClientSecret config.Secret     `toml:"client_secret"`
	Scopes       []string          `toml:"scopes"`
	Params       map[string]string `toml:"parameters"`
}

type OAuth2 struct {
	Service      string          `toml:"service"`
	Endpoint     string          `toml:"token_endpoint"`
	Tenant       string          `toml:"tenant_id"`
	ExpiryMargin config.Duration `toml:"token_expiry_margin"`
	TokenConfigs []TokenConfig   `toml:"token"`
	Log          telegraf.Logger `toml:"-"`

	sources map[string]oauth2.TokenSource
	cancel  context.CancelFunc
}

func (*OAuth2) SampleConfig() string {
	return sampleConfig
}

// Init initializes all internals of the secret-store
func (o *OAuth2) Init() error {
	ctx, cancel := context.WithCancel(context.Background())
	o.cancel = cancel

	// Check the service setting and determine the endpoint
	var endpoint oauth2.Endpoint
	var requireTenant, acceptCustomEndpoint bool
	switch strings.ToLower(o.Service) {
	case "", "custom":
		if o.Endpoint == "" {
			return errors.New("'token_endpoint' required for custom service")
		}
		endpoint.TokenURL = o.Endpoint
		endpoint.AuthStyle = oauth2.AuthStyleAutoDetect
		acceptCustomEndpoint = true
	case "auth0":
		if o.Endpoint == "" {
			return errors.New("'token_endpoint' required for Auth0")
		}
		endpoint = oauth2.Endpoint{
			TokenURL:  o.Endpoint,
			AuthStyle: oauth2.AuthStyleInParams,
		}
		acceptCustomEndpoint = true
	case "azuread":
		if o.Tenant == "" {
			return errors.New("'tenant_id' required for AzureAD")
		}
		requireTenant = true
		endpoint = endpoints.AzureAD(o.Tenant)
	default:
		return fmt.Errorf("service %q not supported", o.Service)
	}

	if !requireTenant && o.Tenant != "" {
		o.Log.Warnf("'tenant_id' set but ignored by service %q", o.Service)
	}

	if !acceptCustomEndpoint && o.Endpoint != "" {
		return fmt.Errorf("'token_endpoint' cannot be set for service %q", o.Service)
	}

	// Setup the token sources
	o.sources = make(map[string]oauth2.TokenSource, len(o.TokenConfigs))
	for _, c := range o.TokenConfigs {
		if c.Key == "" {
			return errors.New("'key' not specified")
		}
		if c.ClientID.Empty() {
			return fmt.Errorf("'client_id' not specified for key %q", c.Key)
		}
		if c.ClientSecret.Empty() {
			return fmt.Errorf("'client_secret' not specified for key %q", c.Key)
		}

		// Check service specific parameters
		if strings.ToLower(o.Service) == "auth0" {
			if audience := c.Params["audience"]; audience == "" {
				return fmt.Errorf("'audience' parameter in key %q missing for service Auth0", c.Key)
			}
		}

		if _, found := o.sources[c.Key]; found {
			return fmt.Errorf("token with key %q already defined", c.Key)
		}

		// Get the secrets
		cid, err := c.ClientID.Get()
		if err != nil {
			return fmt.Errorf("getting client ID for %q failed: %w", c.Key, err)
		}

		csecret, err := c.ClientSecret.Get()
		if err != nil {
			return fmt.Errorf("getting client secret for %q failed: %w", c.Key, err)
		}

		// Setup the configuration
		cfg := &clientcredentials.Config{
			ClientID:     string(cid),
			ClientSecret: string(csecret),
			TokenURL:     endpoint.TokenURL,
			Scopes:       c.Scopes,
			AuthStyle:    endpoint.AuthStyle,
		}
		config.ReleaseSecret(cid)
		config.ReleaseSecret(csecret)

		// Add the parameters if any
		for k, v := range c.Params {
			cfg.EndpointParams.Add(k, v)
		}
		src := cfg.TokenSource(ctx)
		o.sources[c.Key] = oauth2.ReuseTokenSourceWithExpiry(nil, src, time.Duration(o.ExpiryMargin))
	}

	return nil
}

// Get searches for the given key and return the secret
func (o *OAuth2) Get(key string) ([]byte, error) {
	src, found := o.sources[key]
	if !found {
		return nil, fmt.Errorf("token %q not found", key)
	}

	// Return the token from the token-source. The token will be automatically
	// renewed if the token expires.
	token, err := src.Token()
	if err != nil {
		return nil, err
	}

	if !token.Valid() {
		return nil, errors.New("token invalid")
	}

	return []byte(token.AccessToken), nil
}

// Set sets the given secret for the given key
func (o *OAuth2) Set(_, _ string) error {
	return errors.New("not supported")
}

// List lists all known secret keys
func (o *OAuth2) List() ([]string, error) {
	keys := make([]string, 0, len(o.sources))
	for k := range o.sources {
		keys = append(keys, k)
	}
	return keys, nil
}

// GetResolver returns a function to resolve the given key.
func (o *OAuth2) GetResolver(key string) (telegraf.ResolveFunc, error) {
	resolver := func() ([]byte, bool, error) {
		s, err := o.Get(key)
		return s, true, err
	}
	return resolver, nil
}

// Register the secret-store on load.
func init() {
	secretstores.Add("oauth2", func(_ string) telegraf.SecretStore {
		return &OAuth2{ExpiryMargin: config.Duration(time.Second)}
	})
}

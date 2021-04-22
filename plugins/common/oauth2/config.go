package oauth2

import (
	"context"
	"net/http"

	"github.com/influxdata/telegraf"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type OAuth2Config struct {
	// OAuth2 Credentials
	ClientID     string   `toml:"client_id"`
	ClientSecret string   `toml:"client_secret"`
	TokenURL     string   `toml:"token_url"`
	Scopes       []string `toml:"scopes"`

	Log telegraf.Logger
}

func (o *OAuth2Config) CreateOauth2Client(client *http.Client, ctx context.Context) *http.Client {
	if o.ClientID != "" && o.ClientSecret != "" && o.TokenURL != "" {
		oauthConfig := clientcredentials.Config{
			ClientID:     o.ClientID,
			ClientSecret: o.ClientSecret,
			TokenURL:     o.TokenURL,
			Scopes:       o.Scopes,
		}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
		client = oauthConfig.Client(ctx)
	} else if o.ClientID != "" || o.ClientSecret != "" || o.TokenURL != "" {
		o.Log.Warnf("One of the following fields is empty: Client ID, Client Secret or Token URL. Skipping OAuth.")
	}

	return client
}

package oauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/influxdata/telegraf/config"
)

type OAuth2Config struct {
	// OAuth2 Credentials
	ClientID     config.Secret `toml:"client_id"`
	ClientSecret config.Secret `toml:"client_secret"`
	TokenURL     string        `toml:"token_url"`
	Audience     string        `toml:"audience"`
	Scopes       []string      `toml:"scopes"`
}

func (o *OAuth2Config) CreateOauth2Client(ctx context.Context, client *http.Client) (*http.Client, error) {
	if o.ClientID.Empty() || o.ClientSecret.Empty() || o.TokenURL == "" {
		return client, nil
	}

	clientID, err := o.ClientID.Get()
	if err != nil {
		return nil, fmt.Errorf("getting client ID failed: %w", err)
	}
	defer clientID.Destroy()

	clientSecret, err := o.ClientSecret.Get()
	if err != nil {
		return nil, fmt.Errorf("getting client secret failed: %w", err)
	}
	defer clientSecret.Destroy()

	oauthConfig := clientcredentials.Config{
		ClientID:       clientID.String(),
		ClientSecret:   clientSecret.String(),
		TokenURL:       o.TokenURL,
		Scopes:         o.Scopes,
		EndpointParams: make(url.Values),
	}

	if o.Audience != "" {
		oauthConfig.EndpointParams.Add("audience", o.Audience)
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
	client = oauthConfig.Client(ctx)

	return client, nil
}

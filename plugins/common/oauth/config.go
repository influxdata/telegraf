package oauth

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type OAuth2Config struct {
	// OAuth2 Credentials
	ClientID     string   `toml:"client_id"`
	ClientSecret string   `toml:"client_secret"`
	TokenURL     string   `toml:"token_url"`
	Audience     string   `toml:"audience"`
	Scopes       []string `toml:"scopes"`
}

func (o *OAuth2Config) CreateOauth2Client(ctx context.Context, client *http.Client) *http.Client {
	if o.ClientID == "" || o.ClientSecret == "" || o.TokenURL == "" {
		return client
	}

	oauthConfig := clientcredentials.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		TokenURL:     o.TokenURL,
		Scopes:       o.Scopes,
	}

	if o.Audience != "" {
		oauthConfig.EndpointParams.Add("audience", o.Audience)
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
	client = oauthConfig.Client(ctx)

	return client
}

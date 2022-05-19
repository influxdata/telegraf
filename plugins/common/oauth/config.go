package oauth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/api/idtoken"
)

type OAuth2Config struct {
	// OAuth2 Credentials
	ClientID     string   `toml:"client_id"`
	ClientSecret string   `toml:"client_secret"`
	TokenURL     string   `toml:"token_url"`
	Scopes       []string `toml:"scopes"`
	// Google API Auth
	CredentialsFile string `toml:"google_application_credentials"`
	oauth2Token     *oauth2.Token
}

func (o *OAuth2Config) CreateOauth2Client(ctx context.Context, client *http.Client) (*http.Client, error) {
	if o.ClientID != "" && o.ClientSecret != "" && o.TokenURL != "" {
		oauthConfig := clientcredentials.Config{
			ClientID:     o.ClientID,
			ClientSecret: o.ClientSecret,
			TokenURL:     o.TokenURL,
			Scopes:       o.Scopes,
		}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
		client = oauthConfig.Client(ctx)
	}

	return client, nil
}

func (o *OAuth2Config) GetAccessToken(ctx context.Context, audience string) (*oauth2.Token, error) {
	if o.oauth2Token.Valid() {
		return o.oauth2Token, nil
	}

	ts, err := idtoken.NewTokenSource(ctx, audience, idtoken.WithCredentialsFile(o.CredentialsFile))
	if err != nil {
		return nil, fmt.Errorf("error creating oauth2 token source: %s", err)
	}

	token, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("error fetching oauth2 token: %s", err)
	}

	o.oauth2Token = token

	return token, nil
}

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
	CredentialsFile string `toml:"credentials_file"`
	AccessToken     *oauth2.Token
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

func (o *OAuth2Config) GetAccessToken(ctx context.Context, audience string) error {
	ts, err := idtoken.NewTokenSource(ctx, audience, idtoken.WithCredentialsFile(o.CredentialsFile))
	if err != nil {
		return fmt.Errorf("error creating oauth2 token source: %s", err)
	}

	token, err := ts.Token()
	if err != nil {
		return fmt.Errorf("error fetching oauth2 token: %s", err)
	}

	o.AccessToken = token

	return nil
}

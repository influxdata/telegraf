package oauth

import (
	"context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/api/idtoken"
	"net/http"
)

type OAuth2Config struct {
	// OAuth2 Credentials
	ClientID     string   `toml:"client_id"`
	ClientSecret string   `toml:"client_secret"`
	TokenURL     string   `toml:"token_url"`
	Scopes       []string `toml:"scopes"`

	CredentialsFile string `toml:"credentials_file"`
	//AccessToken     oauth2.Token
	AccessToken string
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

	// google api auth
	if o.CredentialsFile != "" {
		err := o.GetAccessToken(ctx, o.TokenURL)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (o *OAuth2Config) GetAccessToken(ctx context.Context, audience string) error {
	ts, err := idtoken.NewTokenSource(ctx, audience)
	if err != nil {
		return err
	}

	token, err := ts.Token()
	if err != nil {
		return err
	}
	o.AccessToken = token.AccessToken

	return nil
}

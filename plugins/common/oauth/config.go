package oauth

import (
	"context"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

type OAuth2Config struct {
	// OAuth2 Credentials
	ClientID     string   `toml:"client_id"`
	ClientSecret string   `toml:"client_secret"`
	TokenURL     string   `toml:"token_url"`
	Scopes       []string `toml:"scopes"`

	CredentialsFile string `toml:"credentials_file"`
	AccessToken     string
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

	if o.CredentialsFile != "" {
		err := o.GetAccessToken(ctx, o.TokenURL)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (o *OAuth2Config) GetAccessToken(ctx context.Context, audience string) error {
	data, err := ioutil.ReadFile(o.CredentialsFile)
	if err != nil {
		return err
	}

	conf, err := google.JWTConfigFromJSON(data, audience)
	if err != nil {
		return err
	}

	jwtConfig := &jwt.Config{
		Email:         conf.Email,
		TokenURL:      conf.TokenURL,
		PrivateKey:    conf.PrivateKey,
		PrivateClaims: map[string]interface{}{"target_audience": audience},
	}

	token, err := jwtConfig.TokenSource(ctx).Token()
	if err != nil {
		return err
	}

	o.AccessToken = token.Extra("id_token").(string)
	return nil
}

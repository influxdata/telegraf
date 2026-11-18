package oauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
)

func TestCreateOauth2ClientIncompleteConfig(t *testing.T) {
	tests := []struct {
		name        string
		oauthConfig OAuth2Config
	}{
		{
			name: "no credentials",
		},
		{
			name: "no token URL",
			oauthConfig: OAuth2Config{
				ClientID:     config.NewSecret([]byte("howdy")),
				ClientSecret: config.NewSecret([]byte("secret")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{}
			got, err := tt.oauthConfig.CreateOauth2Client(context.Background(), client)
			require.NoError(t, err)
			require.Same(t, client, got)
		})
	}
}

func TestCreateOauth2ClientWithSecrets(t *testing.T) {
	const token = "test-token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			username, password, ok := r.BasicAuth()
			if !ok || username != "howdy" || password != "secret" {
				t.Error("unexpected token request credentials")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			values := url.Values{
				"access_token": {token},
				"token_type":   {"bearer"},
				"expires_in":   {"3600"},
			}
			w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
			if _, err := w.Write([]byte(values.Encode())); err != nil {
				t.Error(err)
			}
		case "/write":
			if auth := r.Header.Get("Authorization"); auth != "Bearer "+token {
				t.Errorf("unexpected Authorization header: %q", auth)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	o := OAuth2Config{
		ClientID:     config.NewSecret([]byte("howdy")),
		ClientSecret: config.NewSecret([]byte("secret")),
		TokenURL:     server.URL + "/token",
	}
	client, err := o.CreateOauth2Client(context.Background(), &http.Client{})
	require.NoError(t, err)

	resp, err := client.Get(server.URL + "/write")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

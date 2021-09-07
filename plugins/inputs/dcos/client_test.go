package dcos

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var privateKey = testutil.NewPKI("../../../testutil/pki").ReadServerKey()

func TestLogin(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	var tests = []struct {
		name          string
		responseCode  int
		responseBody  string
		expectedError error
		expectedToken string
	}{
		{
			name:          "Login successful",
			responseCode:  http.StatusOK,
			responseBody:  `{"token": "XXX.YYY.ZZZ"}`,
			expectedError: nil,
			expectedToken: "XXX.YYY.ZZZ",
		},
		{
			name:         "Unauthorized Error",
			responseCode: http.StatusUnauthorized,
			responseBody: `{"title": "x", "description": "y"}`,
			expectedError: &APIError{
				URL:         ts.URL + "/acs/api/v1/auth/login",
				StatusCode:  http.StatusUnauthorized,
				Title:       "x",
				Description: "y",
			},
			expectedToken: "",
		},
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				fmt.Fprintln(w, tt.responseBody)
			})

			u, err := url.Parse(ts.URL)
			require.NoError(t, err)

			ctx := context.Background()
			sa := &ServiceAccount{
				AccountID:  "telegraf",
				PrivateKey: key,
			}
			client := NewClusterClient(u, defaultResponseTimeout, 1, nil)
			auth, err := client.Login(ctx, sa)

			require.Equal(t, tt.expectedError, err)

			if tt.expectedToken != "" {
				require.Equal(t, tt.expectedToken, auth.Text)
			} else {
				require.Nil(t, auth)
			}
		})
	}
}

func TestGetSummary(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	var tests = []struct {
		name          string
		responseCode  int
		responseBody  string
		expectedValue *Summary
		expectedError error
	}{
		{
			name:          "No nodes",
			responseCode:  http.StatusOK,
			responseBody:  `{"cluster": "a", "slaves": []}`,
			expectedValue: &Summary{Cluster: "a", Slaves: []Slave{}},
			expectedError: nil,
		},
		{
			name:          "Unauthorized Error",
			responseCode:  http.StatusUnauthorized,
			responseBody:  `<html></html>`,
			expectedValue: nil,
			expectedError: &APIError{
				URL:        ts.URL + "/mesos/master/state-summary",
				StatusCode: http.StatusUnauthorized,
				Title:      "401 Unauthorized",
			},
		},
		{
			name:         "Has nodes",
			responseCode: http.StatusOK,
			responseBody: `{"cluster": "a", "slaves": [{"id": "a"}, {"id": "b"}]}`,
			expectedValue: &Summary{
				Cluster: "a",
				Slaves: []Slave{
					{ID: "a"},
					{ID: "b"},
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// check the path
				w.WriteHeader(tt.responseCode)
				fmt.Fprintln(w, tt.responseBody)
			})

			u, err := url.Parse(ts.URL)
			require.NoError(t, err)

			ctx := context.Background()
			client := NewClusterClient(u, defaultResponseTimeout, 1, nil)
			summary, err := client.GetSummary(ctx)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedValue, summary)
		})
	}
}

func TestGetNodeMetrics(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	var tests = []struct {
		name          string
		responseCode  int
		responseBody  string
		expectedValue *Metrics
		expectedError error
	}{
		{
			name:          "Empty Body",
			responseCode:  http.StatusOK,
			responseBody:  `{}`,
			expectedValue: &Metrics{},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// check the path
				w.WriteHeader(tt.responseCode)
				fmt.Fprintln(w, tt.responseBody)
			})

			u, err := url.Parse(ts.URL)
			require.NoError(t, err)

			ctx := context.Background()
			client := NewClusterClient(u, defaultResponseTimeout, 1, nil)
			m, err := client.GetNodeMetrics(ctx, "foo")

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedValue, m)
		})
	}
}

func TestGetContainerMetrics(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	var tests = []struct {
		name          string
		responseCode  int
		responseBody  string
		expectedValue *Metrics
		expectedError error
	}{
		{
			name:          "204 No Content",
			responseCode:  http.StatusNoContent,
			responseBody:  ``,
			expectedValue: &Metrics{},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// check the path
				w.WriteHeader(tt.responseCode)
				fmt.Fprintln(w, tt.responseBody)
			})

			u, err := url.Parse(ts.URL)
			require.NoError(t, err)

			ctx := context.Background()
			client := NewClusterClient(u, defaultResponseTimeout, 1, nil)
			m, err := client.GetContainerMetrics(ctx, "foo", "bar")

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedValue, m)
		})
	}
}

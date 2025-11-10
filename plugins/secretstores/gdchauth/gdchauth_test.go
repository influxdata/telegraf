package gdchauth

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

func TestSampleConfig(t *testing.T) {
	require.NoError(t, config.NewConfig().LoadConfigData(testutil.DefaultSampleConfig((&GdchAuth{}).SampleConfig()), config.EmptySourcePath))
}

func TestInit(t *testing.T) {
	tests := []struct {
		name          string
		plugin        *GdchAuth
		expectedError string
	}{
		{
			name: "missing service account file should fail",
			plugin: &GdchAuth{
				Audience: "https://localhost",
				Log:      testutil.Logger{},
			},
			expectedError: "service_account_file is required",
		},
		{
			name: "non-existent service account file should fail",
			plugin: &GdchAuth{
				Audience:           "https://localhost",
				Log:                testutil.Logger{},
				ServiceAccountFile: "non-existent-file.json",
			},
			expectedError: "failed to read service account file",
		},
		{
			name: "invalid service account file json should fail",
			plugin: &GdchAuth{
				Audience:           "https://localhost",
				Log:                testutil.Logger{},
				ServiceAccountFile: "./testdata/invalid-json-sa-key.json",
			},
			expectedError: "failed to parse service account JSON",
		},
		{
			name: "invalid private key pem should fail",
			plugin: &GdchAuth{
				Audience:           "https://localhost",
				Log:                testutil.Logger{},
				ServiceAccountFile: "./testdata/invalid-pem-sa-key.json",
			},
			expectedError: "failed to decode PEM block from private key",
		},
		{
			name: "missing audience should fail",
			plugin: &GdchAuth{
				Log:                testutil.Logger{},
				ServiceAccountFile: "./testdata/valid-sa-key.json",
			},
			expectedError: "audience is required",
		},
		{
			name: "successful init",
			plugin: &GdchAuth{
				Audience:           "https://localhost",
				Log:                testutil.Logger{},
				ServiceAccountFile: "./testdata/valid-sa-key.json",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.plugin.Init()
			if tc.expectedError != "" {
				require.ErrorContains(t, err, tc.expectedError, "error mismatch")
			} else {
				require.NoError(t, err)
				require.NotNil(t, tc.plugin.account)
				require.NotNil(t, tc.plugin.Audience)
				require.NotNil(t, tc.plugin.client)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name               string
		audience           string
		serviceAccountFile string
		tokenExpiryBuffer  config.Duration
		token              string
		expiry             time.Time
		httpClient         *http.Client
		expectedToken      string
		expectedError      string
	}{
		{
			name:               "successful token retrieval",
			audience:           "https://localhost",
			serviceAccountFile: "testdata/valid-sa-key.json",
			httpClient: &http.Client{
				Transport: &mockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"access_token": "test_token", "expires_in": 3600}`)),
					},
				},
			},
			expectedToken: "test_token",
		},
		{
			name:               "cached token retrieval",
			audience:           "https://localhost",
			serviceAccountFile: "testdata/valid-sa-key.json",
			tokenExpiryBuffer:  config.Duration(5 * time.Minute),
			token:              "cached_token",
			expiry:             time.Now().Add(1 * time.Hour),
			httpClient:         &http.Client{}, // No HTTP call expected
			expectedToken:      "cached_token",
		},
		{
			name:               "token refresh due to expiry",
			audience:           "https://localhost",
			serviceAccountFile: "testdata/valid-sa-key.json",
			tokenExpiryBuffer:  config.Duration(5 * time.Minute),
			expiry:             time.Now().Add(1 * time.Minute),
			httpClient: &http.Client{
				Transport: &mockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"access_token": "refreshed_token", "expires_in": 3600}`)),
					},
				},
			},
			expectedToken: "refreshed_token",
		},
		{
			name:               "http request fails",
			audience:           "https://localhost",
			serviceAccountFile: "testdata/valid-sa-key.json",
			httpClient: &http.Client{
				Transport: &mockRoundTripper{
					Err: errors.New("http request failed"),
				},
			},
			expectedError: "http request failed",
		},
		{
			name:               "invalid token response",
			audience:           "https://localhost",
			serviceAccountFile: "testdata/valid-sa-key.json",
			httpClient: &http.Client{
				Transport: &mockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"invalid_response": "oops"}`)),
					},
				},
			},
			expectedError: "token response did not contain 'access_token'",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plugin := &GdchAuth{
				Audience:           tc.audience,
				ServiceAccountFile: tc.serviceAccountFile,
				TokenExpiryBuffer:  tc.tokenExpiryBuffer,
				Log:                testutil.Logger{},
				token:              tc.token,
				expiry:             tc.expiry,
			}

			require.NoError(t, plugin.Init())
			plugin.client = tc.httpClient // set mock after Init()

			token, err := plugin.Get("token")

			if tc.expectedError != "" {
				require.ErrorContains(t, err, tc.expectedError, "error mismatch")
				require.Nil(t, token)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedToken, string(token))
			}
		})
	}
}

type mockRoundTripper struct {
	Response *http.Response
	Err      error
}

// RoundTrip provides the mock response or error.
func (m *mockRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return m.Response, m.Err
}

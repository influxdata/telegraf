package googlecloud

import (
	"context"
	"errors"
	"testing"
	"time"

	"cloud.google.com/go/auth"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

func TestSampleConfig(t *testing.T) {
	require.NoError(t, config.NewConfig().LoadConfigData(testutil.DefaultSampleConfig((&GoogleCloud{}).SampleConfig()), config.EmptySourcePath))
}

func TestInitSuccess(t *testing.T) {
	plugin := &GoogleCloud{
		STSAudience:     "https://localhost",
		Log:             testutil.Logger{},
		CredentialsFile: "./testdata/gdch.json",
	}
	require.NoError(t, plugin.Init())
	require.NotNil(t, plugin.credentials)
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *GoogleCloud
		expected string
	}{
		{
			name: "non-existent service account file should fail",
			plugin: &GoogleCloud{
				STSAudience:     "https://localhost",
				Log:             testutil.Logger{},
				CredentialsFile: "non-existent-file.json",
			},
			expected: "cannot load the credential file:",
		},
		{
			name: "invalid service account file json should fail",
			plugin: &GoogleCloud{
				STSAudience:     "https://localhost",
				Log:             testutil.Logger{},
				CredentialsFile: "./testdata/invalid-json-sa-key.json",
			},
			expected: "cannot parse the credential file: invalid character",
		},
		{
			name: "missing service account type should fail",
			plugin: &GoogleCloud{
				STSAudience:     "https://localhost",
				Log:             testutil.Logger{},
				CredentialsFile: "./testdata/missing-type-sa-key.json",
			},
			expected: "credentials search failed: credentials: unsupported unidentified file type",
		},
		{
			name: "missing audience should fail",
			plugin: &GoogleCloud{
				Log:             testutil.Logger{},
				CredentialsFile: "./testdata/gdch.json",
			},
			expected: "credentials search failed: credentials: STSAudience must be set for the GDCH auth flows",
		},
		{
			name: "missing ca cert path should fail",
			plugin: &GoogleCloud{
				Log:             testutil.Logger{},
				CredentialsFile: "./testdata/gdch-missing-ca-cert-path.json",
				STSAudience:     "https://localhost",
			},
			expected: "credentials search failed: credentials: failed to read certificate:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorContains(t, tc.plugin.Init(), tc.expected, "error mismatch")
			require.Nil(t, tc.plugin.credentials)
		})
	}
}

func TestGetSuccess(t *testing.T) {
	plugin := &GoogleCloud{
		credentials: auth.NewCredentials(&auth.CredentialsOptions{
			TokenProvider: mockTokenProvider{
				token: &auth.Token{Value: "token", Expiry: time.Now().Add(time.Hour)},
			},
		}),
	}
	token, err := plugin.Get("token")
	require.NoError(t, err)
	require.Equal(t, []byte("token"), token)
}

func TestGetFail(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		provider auth.TokenProvider
		expected string
	}{
		{
			name: "token provider error",
			key:  "token",
			provider: mockTokenProvider{
				err: errors.New("token provider error"),
			},
			expected: "token provider error",
		},
		{
			name: "unsupported key",
			key:  "invalid_key",
			provider: mockTokenProvider{
				token: &auth.Token{Value: "token", Expiry: time.Now().Add(time.Hour)},
			},
			expected: "invalid key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &GoogleCloud{
				credentials: auth.NewCredentials(&auth.CredentialsOptions{
					TokenProvider: tt.provider,
				}),
			}
			token, err := plugin.Get(tt.key)
			require.ErrorContains(t, err, tt.expected)
			require.Nil(t, token)
		})
	}
}

type mockTokenProvider struct {
	token *auth.Token
	err   error
}

func (tp mockTokenProvider) Token(context.Context) (*auth.Token, error) {
	if tp.err != nil {
		return nil, tp.err
	}
	return tp.token, nil
}

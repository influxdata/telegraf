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
		STSAudience:        "https://localhost",
		Log:                testutil.Logger{},
		ServiceAccountFile: "./testdata/gdch.json",
	}
	err := plugin.Init()
	require.NoError(t, err)
	require.NotNil(t, plugin.credentials)
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name        string
		plugin      *GoogleCloud
		wantErr     bool
		errContains string
	}{
		{
			name: "non-existent service account file should fail",
			plugin: &GoogleCloud{
				STSAudience:        "https://localhost",
				Log:                testutil.Logger{},
				ServiceAccountFile: "non-existent-file.json",
			},
			errContains: "credentials search failed:",
		},
		{
			name: "invalid service account file json should fail",
			plugin: &GoogleCloud{
				STSAudience:        "https://localhost",
				Log:                testutil.Logger{},
				ServiceAccountFile: "./testdata/invalid-json-sa-key.json",
			},
			wantErr:     true,
			errContains: "credentials search failed: invalid character",
		},
		{
			name: "missing service account type should fail",
			plugin: &GoogleCloud{
				STSAudience:        "https://localhost",
				Log:                testutil.Logger{},
				ServiceAccountFile: "./testdata/missing-type-sa-key.json",
			},
			wantErr:     true,
			errContains: "credentials search failed: credentials: unsupported unidentified file type",
		},
		{
			name: "missing audience should fail",
			plugin: &GoogleCloud{
				Log:                testutil.Logger{},
				ServiceAccountFile: "./testdata/gdch.json",
			},
			errContains: "credentials search failed: credentials: STSAudience must be set for the GDCH auth flows",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.plugin.Init()
			require.Error(t, err)
			require.ErrorContains(t, err, tc.errContains, "error mismatch")
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
		name        string
		key         string
		provider    auth.TokenProvider
		wantToken   []byte
		errContains string
	}{
		{
			name: "token provider error",
			key:  "token",
			provider: mockTokenProvider{
				err: errors.New("token provider error"),
			},
			errContains: "token provider error",
		},
		{
			name: "unsupported key",
			key:  "invalid_key",
			provider: mockTokenProvider{
				token: &auth.Token{Value: "token", Expiry: time.Now().Add(time.Hour)},
			},
			errContains: "invalid key",
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
			require.Error(t, err)
			require.ErrorContains(t, err, tt.errContains)
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

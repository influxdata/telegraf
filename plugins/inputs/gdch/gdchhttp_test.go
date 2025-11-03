package gdch

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	http_plugin "github.com/influxdata/telegraf/plugins/inputs/http"
	"github.com/influxdata/telegraf/plugins/secretstores/gdch"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const (
	testProject           = "test-project"
	testServiceIdentity   = "test-identity"
	testPrivateKeyID      = "test-key-id"
	testAudience          = "test-audience"
	testAccessToken       = "this-is-a-fake-access-token"
	testAccessTokenExpiry = 3600 // seconds
)

// generateTestKeyFile creates a temporary service account JSON file for testing.
func generateTestKeyFile(t *testing.T, tokenURI string) string {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	x509Encoded, err := x509.MarshalECPrivateKey(privateKey)
	require.NoError(t, err)

	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})

	saKey := gdch.ServiceAccountKey{
		PrivateKeyID:        testPrivateKeyID,
		PrivateKey:          string(pemEncoded),
		ServiceIdentityName: testServiceIdentity,
		TokenURI:            tokenURI,
		Project:             testProject,
	}

	keyData, err := json.Marshal(saKey)
	require.NoError(t, err)

	tmpfile, err := os.CreateTemp("", "test-sa-key-*.json")
	require.NoError(t, err)
	defer tmpfile.Close()

	_, err = tmpfile.Write(keyData)
	require.NoError(t, err)

	return tmpfile.Name()
}

// --- Test Cases ---

func TestInit(t *testing.T) {
	t.Run("missing service account file should fail", func(t *testing.T) {
		plugin := &GdchHttp{
			Http: &http_plugin.HTTP{Log: &testutil.Logger{}},
		}
		require.ErrorContains(t, plugin.Init(), "auth configuration is missing")
	})

	t.Run("missing http config should fail", func(t *testing.T) {
		plugin := &GdchHttp{ //nolint:staticcheck // We are testing the error case where Auth is nil
			Auth: &gdch.GdchAuth{},
		}
		err := plugin.Init()
		require.Error(t, err)
		require.Contains(t, err.Error(), "http plugin configuration is missing")
	})

	t.Run("auth init fails", func(t *testing.T) {
		plugin := &GdchHttp{ //nolint:staticcheck // We are testing the error case where Auth is nil
			Auth: &gdch.GdchAuth{},
			Http: &http_plugin.HTTP{Log: &testutil.Logger{}},
			Log:  testutil.Logger{},
		}
		err := plugin.Init()
		require.Error(t, err)
		require.Contains(t, err.Error(), "service_account_file is required")
	})

	t.Run("successful init", func(t *testing.T) {
		plugin := &GdchHttp{
			Auth: &gdch.GdchAuth{
				ServiceAccountFile: generateTestKeyFile(t, "http://localhost/token"),
			},
			Http: &http_plugin.HTTP{Log: &testutil.Logger{}},
			Log:  &testutil.Logger{},
		}
		err := plugin.Init()
		require.NoError(t, err)

		require.Equal(t, config.Duration(5*time.Minute), plugin.Auth.TokenExpiryBuffer)
	})
}

func TestGather(t *testing.T) {
	// --- Setup Mock Server ---
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": testAccessToken,
			"expire_time":  "2025-01-01T00:00:00Z",
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	// --- Setup Plugin for Test ---
	// Use the real http plugin, but we won't actually call its Gather method.
	// We just need to check that the token is set on it.
	httpPlugin := &http_plugin.HTTP{Log: &testutil.Logger{}}
	keyFile := generateTestKeyFile(t, server.URL)
	defer os.Remove(keyFile)

	plugin := &GdchHttp{
		Auth: &gdch.GdchAuth{
			ServiceAccountFile: keyFile,
			Audience:           testAudience,
			Log:                &testutil.Logger{},
		},
		Http: httpPlugin,
		Log:  &testutil.Logger{},
	}
	err := plugin.Init()
	require.NoError(t, err)
	plugin.Auth.SetLogger(testutil.Logger{})

	// --- Run Test ---
	var acc testutil.Accumulator
	// We do not care about the return value
	plugin.Gather(&acc)

	// Verify that the token was set on the embedded http plugin.
	// The token is a secret, so we need to get it to check its value.
	require.NotNil(t, httpPlugin.Token)
	tokenBytes, err := httpPlugin.Token.Get()
	require.NoError(t, err)
	require.Equal(t, testAccessToken, tokenBytes.String())
}

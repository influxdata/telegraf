package gdch

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	http_plugin "github.com/influxdata/telegraf/plugins/inputs/http"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

// --- Test Cases ---

func TestInit(t *testing.T) {
	t.Run("missing service account file should fail", func(t *testing.T) {
		plugin := &GdchHttp{
			Http: &http_plugin.HTTP{},
		}
		err := plugin.Init()
		require.Error(t, err)
		require.Contains(t, err.Error(), "auth configuration is missing")
	})

	t.Run("missing http config should fail", func(t *testing.T) {
		plugin := &GdchHttp{
			Auth: &GdchAuth{
				ServiceAccountFile: "dummy.json",
			},
		}
		err := plugin.Init()
		require.Error(t, err)
		require.Contains(t, err.Error(), "http plugin configuration is missing")
	})

	t.Run("invalid service account file should fail", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "invalid-sa-key-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())
		_, err = tmpfile.WriteString("this is not json")
		require.NoError(t, err)
		require.NoError(t, tmpfile.Close())

		plugin := &GdchHttp{
			Auth: &GdchAuth{
				ServiceAccountFile: tmpfile.Name(),
			},
			Http: &http_plugin.HTTP{},
			Log:  testutil.Logger{},
		}
		err = plugin.Init()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse service account JSON")
	})

	t.Run("successful init", func(t *testing.T) {
		keyFile := generateTestKeyFile(t, "http://localhost/token")
		defer os.Remove(keyFile)

		plugin := &GdchHttp{
			Auth: &GdchAuth{
				ServiceAccountFile: keyFile,
			},
			Http: &http_plugin.HTTP{},
			Log:  testutil.Logger{},
		}
		err := plugin.Init()
		require.NoError(t, err)
		require.NotNil(t, plugin.Auth.saKey)
		require.NotNil(t, plugin.Auth.httpClient)
		require.Equal(t, config.Duration(5*time.Minute), plugin.Auth.TokenExpiryBuffer)
	})
}

func TestGather(t *testing.T) {
	// --- Setup Mock Token Server ---
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintf(w, `{"access_token": "%s", "expires_in": %d}`, testAccessToken, testAccessTokenExpiry)
		require.NoError(t, err)
	}))
	defer server.Close()

	// --- Setup Plugin for Test ---
	keyFile := generateTestKeyFile(t, server.URL)
	defer os.Remove(keyFile)

	// Use the real http plugin, but we won't actually call its Gather method.
	// We just need to check that the token is set on it.
	httpPlugin := &http_plugin.HTTP{}

	plugin := &GdchHttp{
		Auth: &GdchAuth{
			ServiceAccountFile: keyFile,
			Audience:           testAudience,
		},
		Http: httpPlugin,
		Log:  testutil.Logger{},
	}
	err := plugin.Init()
	require.NoError(t, err)

	// --- Run Test ---
	var acc testutil.Accumulator
	// We do not care about the return value
	plugin.Gather(&acc)

	// Verify that the token was set on the embedded http plugin
	token, err := httpPlugin.Token.Get()
	require.NoError(t, err)
	require.Equal(t, testAccessToken, string(token.Bytes()))
}

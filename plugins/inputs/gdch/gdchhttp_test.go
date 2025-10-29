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

// --- Test Helper Functions ---

// // generateTestKeyFile creates a temporary service account JSON file for testing.
// func generateTestKeyFile(t *testing.T, tokenURI string) (string, *ecdsa.PrivateKey) {
// 	t.Helper()

// 	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	require.NoError(t, err)

// 	x509Encoded, err := x509.MarshalECPrivateKey(privateKey)
// 	require.NoError(t, err)

// 	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})

// 	saKey := serviceAccountKey{
// 		PrivateKeyID:        testPrivateKeyID,
// 		PrivateKey:          string(pemEncoded),
// 		ServiceIdentityName: testServiceIdentity,
// 		TokenURI:            tokenURI,
// 		Project:             testProject,
// 	}

// 	keyData, err := json.Marshal(saKey)
// 	require.NoError(t, err)

// 	tmpfile, err := os.CreateTemp("", "test-sa-key-*.json")
// 	require.NoError(t, err)

// 	_, err = tmpfile.Write(keyData)
// 	require.NoError(t, err)
// 	require.NoError(t, tmpfile.Close())

// 	return tmpfile.Name(), privateKey
// }

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

// func TestGetToken(t *testing.T) {
// 	// --- Setup Mock Token Server ---
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Check request method and content type
// 		require.Equal(t, "POST", r.Method)
// 		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

// 		// Decode request body to verify claims
// 		var reqBody map[string]string
// 		err := json.NewDecoder(r.Body).Decode(&reqBody)
// 		require.NoError(t, err)
// 		require.Equal(t, testAudience, reqBody["audience"])

// 		// Send back a successful token response
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		_, err = fmt.Fprintf(w, `{"access_token": "%s", "expires_in": %d}`, testAccessToken, testAccessTokenExpiry)
// 		require.NoError(t, err)
// 	}))
// 	defer server.Close()

// 	// --- Setup Plugin for Tests ---
// 	keyFile, _ := generateTestKeyFile(t, server.URL)
// 	defer os.Remove(keyFile)

// 	plugin := &GdchHttp{
// 		Auth: &GdchAuth{
// 			ServiceAccountFile: keyFile,
// 			Audience:           testAudience,
// 		},
// 		Http: &http_plugin.HTTP{},
// 		Log:  testutil.Logger{},
// 	}
// 	err := plugin.Init()
// 	require.NoError(t, err)

// 	// --- Run Tests ---

// 	t.Run("fetches new token successfully", func(t *testing.T) {
// 		token, err := plugin.getToken(t.Context())
// 		require.NoError(t, err)
// 		require.Equal(t, testAccessToken, token)
// 		require.Equal(t, testAccessToken, plugin.token) // Check internal state
// 		require.WithinDuration(t, time.Now().Add(testAccessTokenExpiry*time.Second), plugin.tokenExpiry, 5*time.Second)
// 	})

// 	t.Run("uses cached token", func(t *testing.T) {
// 		// Ensure token is already cached from previous test
// 		require.Equal(t, testAccessToken, plugin.token)

// 		// This call should not hit the server, it should return the cached token
// 		token, err := plugin.getToken(t.Context())
// 		require.NoError(t, err)
// 		require.Equal(t, testAccessToken, token)
// 	})

// 	t.Run("refreshes expired token", func(t *testing.T) {
// 		// Manually expire the token by setting its expiry time to the past
// 		plugin.tokenExpiry = time.Now().Add(-1 * time.Hour)

// 		// This call should detect the expiry and fetch a new token
// 		token, err := plugin.getToken(t.Context())
// 		require.NoError(t, err)
// 		require.Equal(t, testAccessToken, token)

// 		// Verify the expiry has been updated to a future time
// 		require.True(t, plugin.tokenExpiry.After(time.Now()))
// 	})

// 	t.Run("refreshes token within expiry buffer", func(t *testing.T) {
// 		// Set a long buffer
// 		plugin.TokenExpiryBuffer = config.Duration(2 * time.Hour)
// 		// Set the token to expire in less time than the buffer (e.g., 1 hour)
// 		plugin.tokenExpiry = time.Now().Add(1 * time.Hour)

// 		// This call should detect the token is inside the buffer window and fetch a new one
// 		token, err := plugin.getToken(t.Context())
// 		require.NoError(t, err)
// 		require.Equal(t, testAccessToken, token)

// 		// Verify the expiry has been updated to a future time
// 		require.True(t, plugin.tokenExpiry.After(time.Now()))

// 		// Reset buffer for other tests
// 		plugin.TokenExpiryBuffer = config.Duration(5 * time.Minute)
// 	})
// }

// func TestGetToken_ServerError(t *testing.T) {
// 	// --- Setup Mock Server that always fails ---
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		_, err := w.Write([]byte("internal server error"))
// 		require.NoError(t, err)
// 	}))
// 	defer server.Close()

// 	// --- Setup Plugin ---
// 	keyFile, _ := generateTestKeyFile(t, server.URL)
// 	defer os.Remove(keyFile)

// 	plugin := &GdchHttp{
// 		ServiceAccountFile: keyFile,
// 		Audience:           testAudience,
// 		Http:               &http_plugin.HTTP{},
// 		Log:                testutil.Logger{},
// 	}
// 	err := plugin.Init()
// 	require.NoError(t, err)

// 	// --- Run Test ---
// 	t.Run("handles token fetch error", func(t *testing.T) {
// 		token, err := plugin.getToken(t.Context())
// 		require.Error(t, err)
// 		require.Empty(t, token)
// 		require.Contains(t, err.Error(), "token request returned non-200 status 500")

// 		// Ensure internal token state was not updated on error
// 		require.Empty(t, plugin.token)
// 	})
// }adh

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

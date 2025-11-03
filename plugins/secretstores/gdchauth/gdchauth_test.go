package gdchauth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
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

// --- Test Helper Functions ---

// generateTestKeyFile creates a temporary service account JSON file for testing.
func generateTestKeyFile(t *testing.T, tokenURI string) string {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	x509Encoded, err := x509.MarshalECPrivateKey(privateKey)
	require.NoError(t, err)

	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: x509Encoded})

	saKey := ServiceAccountKey{
		PrivateKeyID:        testPrivateKeyID,
		PrivateKey:          string(pemEncoded),
		ServiceIdentityName: testServiceIdentity,
		TokenURI:            tokenURI,
		Project:             testProject,
	}

	keyData, err := json.Marshal(saKey)
	require.NoError(t, err)

	tmpfile, err := os.CreateTemp(t.TempDir(), "test-sa-key-*.json")
	require.NoError(t, err)
	defer tmpfile.Close()

	_, err = tmpfile.Write(keyData)
	require.NoError(t, err)

	return tmpfile.Name()
}

// --- Test Cases ---

func TestGdchAuth_Init(t *testing.T) {
	t.Run("missing service account file should fail", func(t *testing.T) {
		g := &GdchAuth{Log: testutil.Logger{}}
		err := g.Init()
		require.Error(t, err)
		require.EqualError(t, err, "service_account_file is required")
	})

	t.Run("non-existent service account file should fail", func(t *testing.T) {
		g := &GdchAuth{
			ServiceAccountFile: "non-existent-file.json",
			Log:                testutil.Logger{},
		}
		err := g.Init()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to read service account file")
	})

	t.Run("invalid service account file json should fail", func(t *testing.T) {
		tmpfile, err := os.CreateTemp(t.TempDir(), "invalid-sa-key-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())
		_, err = tmpfile.WriteString("this is not json")
		require.NoError(t, err)
		require.NoError(t, tmpfile.Close())

		g := &GdchAuth{
			ServiceAccountFile: tmpfile.Name(),
			Log:                testutil.Logger{},
		}
		err = g.Init()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse service account JSON")
	})

	t.Run("invalid private key pem should fail", func(t *testing.T) {
		saKey := ServiceAccountKey{
			PrivateKey: "this is not a pem",
		}
		keyData, err := json.Marshal(saKey)
		require.NoError(t, err)

		tmpfile, err := os.CreateTemp(t.TempDir(), "test-sa-key-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())
		_, err = tmpfile.Write(keyData)
		require.NoError(t, err)
		require.NoError(t, tmpfile.Close())

		g := &GdchAuth{
			ServiceAccountFile: tmpfile.Name(),
			Log:                testutil.Logger{},
		}
		err = g.Init()
		require.Error(t, err)
		require.EqualError(t, err, "failed to decode PEM block from private key")
	})

	t.Run("successful init should set defaults", func(t *testing.T) {
		keyFile := generateTestKeyFile(t, "http://localhost/token")
		defer os.Remove(keyFile)

		g := &GdchAuth{
			ServiceAccountFile: keyFile,
			Log:                testutil.Logger{},
		}
		err := g.Init()
		require.NoError(t, err)
		require.NotNil(t, g.saKey)
		require.NotNil(t, g.httpClient)
		require.Equal(t, config.Duration(5*time.Minute), g.TokenExpiryBuffer)
	})
}

func TestGdchAuth_GetToken(t *testing.T) {
	// --- Setup Mock Token Server ---
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and content type
		if !assert.Equal(t, "POST", r.Method) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !assert.Equal(t, "application/json", r.Header.Get("Content-Type")) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Decode request body to verify claims
		var reqBody map[string]string
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if !assert.NoError(t, err) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !assert.Equal(t, testAudience, reqBody["audience"]) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Send back a successful token response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(w, `{"access_token": "%s", "expires_in": %d}`, testAccessToken, testAccessTokenExpiry)
		assert.NoError(t, err)
	}))
	defer server.Close()

	// --- Setup Plugin for Tests ---
	keyFile := generateTestKeyFile(t, server.URL)
	defer os.Remove(keyFile)

	g := &GdchAuth{
		ServiceAccountFile: keyFile,
		Audience:           testAudience,
		Log:                testutil.Logger{},
	}
	err := g.Init()
	require.NoError(t, err)

	// --- Run Tests ---

	t.Run("fetches new token successfully", func(t *testing.T) {
		token, err := g.GetToken(context.Background())
		require.NoError(t, err)
		require.Equal(t, testAccessToken, token)
		require.Equal(t, testAccessToken, g.token) // Check internal state
		require.WithinDuration(t, time.Now().Add(testAccessTokenExpiry*time.Second), g.tokenExpiry, 5*time.Second)
	})

	t.Run("uses cached token", func(t *testing.T) {
		// Ensure token is already cached from previous test
		require.Equal(t, testAccessToken, g.token)

		// This call should not hit the server, it should return the cached token
		token, err := g.GetToken(context.Background())
		require.NoError(t, err)
		require.Equal(t, testAccessToken, token)
	})

	t.Run("refreshes expired token", func(t *testing.T) {
		// Manually expire the token by setting its expiry time to the past
		g.tokenExpiry = time.Now().Add(-1 * time.Hour)

		// This call should detect the expiry and fetch a new token
		token, err := g.GetToken(context.Background())
		require.NoError(t, err)
		require.Equal(t, testAccessToken, token)

		// Verify the expiry has been updated to a future time
		require.True(t, g.tokenExpiry.After(time.Now()))
	})

	t.Run("refreshes token within expiry buffer", func(t *testing.T) {
		// Set a long buffer
		g.TokenExpiryBuffer = config.Duration(2 * time.Hour)
		// Set the token to expire in less time than the buffer (e.g., 1 hour)
		g.tokenExpiry = time.Now().Add(1 * time.Hour)

		// This call should detect the token is inside the buffer window and fetch a new one
		token, err := g.GetToken(context.Background())
		require.NoError(t, err)
		require.Equal(t, testAccessToken, token)

		// Verify the expiry has been updated to a future time
		require.True(t, g.tokenExpiry.After(time.Now()))

		// Reset buffer for other tests
		g.TokenExpiryBuffer = config.Duration(5 * time.Minute)
	})
}

func TestGdchAuth_GetToken_ServerError(t *testing.T) {
	// --- Setup Mock Server that always fails ---
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("internal server error")); err != nil {
			t.Logf("error writing response: %v", err)
		}
	}))
	defer server.Close()

	// --- Setup Plugin ---
	keyFile := generateTestKeyFile(t, server.URL)
	defer os.Remove(keyFile)

	g := &GdchAuth{
		ServiceAccountFile: keyFile,
		Audience:           testAudience,
		Log:                testutil.Logger{},
	}
	err := g.Init()
	require.NoError(t, err)

	// --- Run Test ---
	t.Run("handles token fetch error", func(t *testing.T) {
		token, err := g.GetToken(context.Background())
		require.Error(t, err)
		require.Empty(t, token)
		require.Contains(t, err.Error(), "token request returned non-200 status 500")

		// Ensure internal token state was not updated on error
		require.Empty(t, g.token)
	})
}

func TestGdchAuth_GetToken_Concurrent(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		var err error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(w, `{"access_token": "%s", "expires_in": %d}`, testAccessToken, testAccessTokenExpiry)
		assert.NoError(t, err, "error writing response in mock server")
	}))
	defer server.Close()

	keyFile := generateTestKeyFile(t, server.URL)
	defer os.Remove(keyFile)

	g := &GdchAuth{
		ServiceAccountFile: keyFile,
		Audience:           testAudience,
		Log:                testutil.Logger{},
	}
	err := g.Init()
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			token, err := g.GetToken(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, testAccessToken, token)
		}()
	}
	wg.Wait()

	// The mock server should only be called once due to the lock.
	require.Equal(t, 1, callCount)
}

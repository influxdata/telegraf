package gdchhttp

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	http_plugin "github.com/influxdata/telegraf/plugins/inputs/http" // Alias the http plugin
)

const (
	TOKEN_EXCHANGE_TYPE        = "urn:ietf:params:oauth:token-type:token-exchange"
	ACCESS_TOKEN_TOKEN_TYPE    = "urn:ietf:params:oauth:token-type:access_token"
	SERVICE_ACCOUNT_TOKEN_TYPE = "urn:k8s:params:oauth:token-type:serviceaccount"
)

// GdchHttp is the main plugin struct
type GdchHttp struct {
	// Http holds the configuration for the embedded http plugin
	// The `toml:"http"` tag tells Telegraf to load the [inputs.gdch_http.http]
	// config table into this struct.
	Http *http_plugin.HTTP `toml:"http"`

	// --- Custom GDCH Auth Configuration ---
	ServiceAccountFile string          `toml:"service_account_file"`
	Audience           string          `toml:"audience"`
	TokenExpiryBuffer  config.Duration `toml:"token_expiry_buffer"`
	tls.ClientConfig   `toml:"tls"`

	// --- Internal State ---
	Log telegraf.Logger

	token       string
	tokenExpiry time.Time // The time the token actually expires.
	tokenMutex  sync.RWMutex
	saKey       *serviceAccountKey // Holds parsed SA key data
	httpClient  *http.Client       // Client for *fetching the token*
}

// serviceAccountKey struct to parse the Google SA JSON
type serviceAccountKey struct {
	PrivateKeyID        string `json:"private_key_id"`
	PrivateKey          string `json:"private_key"`
	ServiceIdentityName string `json:"name"`
	TokenURI            string `json:"token_uri"`
	Project             string `json:"project"`

	// --- Internal fields ---
	// Use the generic crypto.Signer interface to hold either
	// *rsa.PrivateKey or *ecdsa.PrivateKey
	parsedKey     crypto.Signer
	signingMethod jwt.SigningMethod
}

// --- Telegraf Plugin Interface Methods ---

// Description returns a one-sentence description of the plugin
func (g *GdchHttp) Description() string {
	return "Wraps the http input plugin to add GDCH service account auth"
}

func (g *GdchHttp) SampleConfig() string {
	return `
  ## Path to the GDCH service account JSON key file
  service_account_file = "/etc/telegraf/gdch-key.json"
  ## Audience for the token request
  audience = "https://monitoring.gdc.goog/api/v1/metrics"

  ## Time before token expiry to fetch a new one.
  # token_expiry_buffer = "5m"

  ## Optional TLS configuration for the token endpoint.
  # tls_ca = "/etc/telegraf/ca.pem"

  ## Embedded HTTP Input Plugin Configuration
  [inputs.gdch_http.http]
    ## A list of URLs to pull data from.
    urls = [
      "https://{GDCH_URL}/{PROJECT}/metrics"
    ]
    ## ... other http plugin options ...
`
}

// Init is called once when the plugin starts.
// This is where we load the key file and initialize the embedded http plugin.
func (g *GdchHttp) Init() error {
	// Validate our custom config
	if g.ServiceAccountFile == "" {
		return errors.New("service_account_file is required for gdch_http plugin")
	}
	if g.Http == nil {
		return errors.New("http plugin configuration is missing")
	}
	if g.TokenExpiryBuffer == 0 {
		g.TokenExpiryBuffer = config.Duration(5 * time.Minute)
	}

	// Load and parse the service account key
	keyData, err := os.ReadFile(g.ServiceAccountFile)
	if err != nil {
		return fmt.Errorf("failed to read service account file: %w", err)
	}

	g.saKey = &serviceAccountKey{}
	if err := json.Unmarshal(keyData, g.saKey); err != nil {
		return fmt.Errorf("failed to parse service account JSON: %w", err)
	}

	// Parse the private key from PEM format
	if err := g.parsePrivateKey(); err != nil {
		return err
	}

	if err := g.buildHttpClient(); err != nil {
		return err
	}

	// Initialize the embedded http plugin
	g.Log.Info("GDCH HTTP plugin initialized. Calling Init() on embedded http plugin.")
	return g.Http.Init()
}

func (g *GdchHttp) buildHttpClient() error {
	tlsConfig, err := g.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	g.httpClient = &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}
	return nil
}

func (g *GdchHttp) parsePrivateKey() error {
	block, _ := pem.Decode([]byte(g.saKey.PrivateKey))
	if block == nil {
		return errors.New("failed to decode PEM block from private key")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if key == nil {
		return errors.New("private key could not be parsed")
	}
	if err == nil {
		g.saKey.parsedKey = key
		g.saKey.signingMethod = jwt.SigningMethodES256
		g.Log.Debug("successfully parsed EC private key")
	}
	return nil
}

// Gather is the main method called by Telegraf at each interval
func (g *GdchHttp) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	// 1. Get a valid auth token (from cache or by fetching a new one)
	token, err := g.getToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}
	// 2. Inject the token into the embedded http plugin's input
	g.Http.Token = config.NewSecret([]byte(token))

	// 3. Call the embedded http plugin's Gather method
	return g.Http.Gather(acc)
}

// SetParserFunc passes the parser function to the embedded http plugin.
func (g *GdchHttp) SetParserFunc(fn telegraf.ParserFunc) {
	g.Http.SetParserFunc(fn)
}

// --- Token Generation & Caching Logic ---

// getToken handles caching, concurrent requests, and fetching
func (g *GdchHttp) getToken(ctx context.Context) (string, error) {
	g.tokenMutex.RLock()

	if g.token != "" && time.Now().Before(g.tokenExpiry.Add(-time.Duration(g.TokenExpiryBuffer))) {
		g.tokenMutex.RUnlock()
		g.Log.Debug("Using cached auth token")
		return g.token, nil
	}
	g.tokenMutex.RUnlock()

	// Token is invalid or expired, need to fetch a new one.
	// Acquire a full write-lock to prevent multiple concurrent fetches.
	g.tokenMutex.Lock()
	defer g.tokenMutex.Unlock()

	// Double-check: another goroutine might have refreshed the token
	// while we were waiting for the lock.
	if g.token != "" && time.Now().Before(g.tokenExpiry.Add(-time.Duration(g.TokenExpiryBuffer))) {
		return g.token, nil
	}

	// Fetch a new token
	g.Log.Debug("Auth token expired or missing, fetching new one...")
	newToken, expiry, err := g.fetchNewToken(ctx)
	if err != nil {
		return "", err // Don't cache on error
	}

	g.token = newToken
	g.tokenExpiry = expiry
	g.Log.Info("Successfully fetched new auth token")
	return g.token, nil
}

// fetchNewToken performs the actual JWT creation and HTTP call
func (g *GdchHttp) fetchNewToken(ctx context.Context) (string, time.Time, error) {
	// 1. Create JWT (as per the Python reference)
	now := time.Now()
	// The JWT itself is valid for 1 hour
	jwtExpiry := now.Add(1 * time.Hour)

	iss_sub_value := fmt.Sprintf("system:serviceaccount:%s:%s",
		g.saKey.Project,
		g.saKey.ServiceIdentityName)

	claims := jwt.MapClaims{
		"iss": iss_sub_value,
		"sub": iss_sub_value,
		"aud": g.saKey.TokenURI,
		"iat": now.Unix(),
		"exp": jwtExpiry.Unix(),
	}

	token := jwt.NewWithClaims(g.saKey.signingMethod, claims)
	token.Header["kid"] = g.saKey.PrivateKeyID // Add the key ID to the header

	signedJWT, err := token.SignedString(g.saKey.parsedKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign JWT: %w", err)
	}

	// 2. POST to token URI
	// Create a JSON body instead of a form
	tokenRequestBody := map[string]string{
		"grant_type":           TOKEN_EXCHANGE_TYPE,
		"audience":             g.Audience,
		"requested_token_type": ACCESS_TOKEN_TOKEN_TYPE,
		"subject_token":        signedJWT,
		"subject_token_type":   SERVICE_ACCOUNT_TOKEN_TYPE,
	}

	jsonBody, err := json.Marshal(tokenRequestBody)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to marshal token request JSON: %w", err)
	}

	// Create the request with the JSON body
	req, err := http.NewRequestWithContext(ctx, "POST", g.saKey.TokenURI, bytes.NewReader(jsonBody))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create token request: %w", err)
	}

	// Set the Content-Type to application/json
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to read token response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("token request returned non-200 status %d: %s", resp.StatusCode, string(body))
	}

	// 3. Parse response to get the access token
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"` // Token expiry in seconds
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse token response JSON: %w", err)
	}

	if tokenResponse.AccessToken == "" {
		return "", time.Time{}, errors.New("token response did not contain 'access_token'")
	}

	// Calculate the *access token's* actual expiry time
	var finalExpiry time.Time
	if tokenResponse.ExpiresIn > 0 {
		finalExpiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	} else {
		finalExpiry = jwtExpiry // Fallback to the JWT expiry
	}

	return tokenResponse.AccessToken, finalExpiry, nil
}

// --- Telegraf Plugin Registration ---

// init registers the plugin with Telegraf
func init() {
	inputs.Add("gdch_http",
		func() telegraf.Input {
			// We must initialize the embedded plugin struct
			return &GdchHttp{
				Http: &http_plugin.HTTP{},
			}
		})
}

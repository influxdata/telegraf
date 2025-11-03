package gdchauth

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	_ "embed"
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
	"github.com/influxdata/telegraf/plugins/secretstores"
)

//go:embed sample.conf
var sampleConfig string

/* #nosec G101 */
const (
	tokenExchangeType       = "urn:ietf:params:oauth:token-type:token-exchange"
	accessTokenTokenType    = "urn:ietf:params:oauth:token-type:access_token"
	serviceAccountTokenType = "urn:k8s:params:oauth:token-type:serviceaccount"
)

func (*GdchAuth) SampleConfig() string {
	return sampleConfig
}

// GdchAuth is the main authenticator struct
type GdchAuth struct {
	ServiceAccountFile string          `toml:"service_account_file"`
	Audience           string          `toml:"audience"`
	TokenExpiryBuffer  config.Duration `toml:"token_expiry_buffer"`
	tls.ClientConfig   `toml:"tls"`

	Log telegraf.Logger `toml:"-"`

	token       string
	tokenExpiry time.Time
	tokenMutex  sync.RWMutex
	saKey       *ServiceAccountKey
	httpClient  *http.Client
}

type ServiceAccountKey struct {
	PrivateKeyID        string `json:"private_key_id"`
	PrivateKey          string `json:"private_key"`
	ServiceIdentityName string `json:"name"`
	TokenURI            string `json:"token_uri"`
	Project             string `json:"project"`

	parsedKey     crypto.Signer
	signingMethod jwt.SigningMethod
}

func (g *GdchAuth) Init() error {
	if g.ServiceAccountFile == "" {
		return errors.New("service_account_file is required")
	}
	if g.TokenExpiryBuffer == 0 {
		g.TokenExpiryBuffer = config.Duration(5 * time.Minute)
	}

	keyData, err := os.ReadFile(g.ServiceAccountFile)
	if err != nil {
		return fmt.Errorf("failed to read service account file: %w", err)
	}

	g.saKey = &ServiceAccountKey{}
	if err := json.Unmarshal(keyData, g.saKey); err != nil {
		return fmt.Errorf("failed to parse service account JSON: %w", err)
	}

	if err := g.parsePrivateKey(); err != nil {
		return err
	}

	return g.buildHTTPClient()
}

// Get retrieves the token. The key is ignored as this secret store only provides one secret.
func (g *GdchAuth) Get(_ string) ([]byte, error) {
	token, err := g.GetToken(context.Background())
	if err != nil {
		return nil, err
	}
	return []byte(token), nil
}

// List returns the list of secrets provided by this store.
func (*GdchAuth) List() ([]string, error) {
	return []string{"token"}, nil
}

// Set is not supported for the gdchauth secret store.
func (*GdchAuth) Set(_, _ string) error {
	return errors.New("setting secrets is not supported")
}

// GetResolver returns a resolver function for the secret.
func (g *GdchAuth) GetResolver(key string) (telegraf.ResolveFunc, error) {
	return func() ([]byte, bool, error) {
		s, err := g.Get(key)
		return s, true, err
	}, nil
}

// SetLogger sets the logger for the authenticator.
func (g *GdchAuth) SetLogger(log telegraf.Logger) {
	g.Log = log
}

// GetToken retrieves a GDCH auth token. It caches the token and reuses it
// until it is within the 'token_expiry_buffer' of its expiry time.
func (g *GdchAuth) GetToken(ctx context.Context) (string, error) {
	g.tokenMutex.RLock()

	if g.token != "" && time.Now().Before(g.tokenExpiry.Add(-time.Duration(g.TokenExpiryBuffer))) {
		g.tokenMutex.RUnlock()
		g.Log.Debug("Using cached auth token")
		return g.token, nil
	}
	g.tokenMutex.RUnlock()

	g.tokenMutex.Lock()
	defer g.tokenMutex.Unlock()

	if g.token != "" && time.Now().Before(g.tokenExpiry.Add(-time.Duration(g.TokenExpiryBuffer))) {
		return g.token, nil
	}

	g.Log.Debug("Auth token expired or missing, fetching new one...")
	newToken, expiry, err := g.fetchNewToken(ctx)
	if err != nil {
		return "", err
	}

	g.token = newToken
	g.tokenExpiry = expiry
	g.Log.Info("Successfully fetched new auth token")
	return g.token, nil
}

func (g *GdchAuth) buildHTTPClient() error {
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

func (g *GdchAuth) parsePrivateKey() error {
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

func (g *GdchAuth) fetchNewToken(ctx context.Context) (string, time.Time, error) {
	now := time.Now()
	jwtExpiry := now.Add(1 * time.Hour)

	issSubValue := fmt.Sprintf("system:serviceaccount:%s:%s",
		g.saKey.Project,
		g.saKey.ServiceIdentityName)

	claims := jwt.MapClaims{
		"iss": issSubValue,
		"sub": issSubValue,
		"aud": g.saKey.TokenURI,
		"iat": now.Unix(),
		"exp": jwtExpiry.Unix(),
	}

	token := jwt.NewWithClaims(g.saKey.signingMethod, claims)
	token.Header["kid"] = g.saKey.PrivateKeyID

	signedJWT, err := token.SignedString(g.saKey.parsedKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign JWT: %w", err)
	}

	tokenRequestBody := map[string]string{
		"grant_type":           tokenExchangeType,
		"audience":             g.Audience,
		"requested_token_type": accessTokenTokenType,
		"subject_token":        signedJWT,
		"subject_token_type":   serviceAccountTokenType,
	}

	jsonBody, err := json.Marshal(tokenRequestBody)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to marshal token request JSON: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.saKey.TokenURI, bytes.NewReader(jsonBody))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create token request: %w", err)
	}

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

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse token response JSON: %w", err)
	}

	if tokenResponse.AccessToken == "" {
		return "", time.Time{}, errors.New("token response did not contain 'access_token'")
	}

	var finalExpiry time.Time
	if tokenResponse.ExpiresIn > 0 {
		finalExpiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	} else {
		finalExpiry = jwtExpiry
	}

	return tokenResponse.AccessToken, finalExpiry, nil
}

func init() {
	secretstores.Add("gdchauth", func(_ string) telegraf.SecretStore {
		return &GdchAuth{}
	})
}

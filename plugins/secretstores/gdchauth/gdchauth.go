package gdchauth

import (
	"bytes"
	"context"

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

	"github.com/golang-jwt/jwt/v5"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/secretstores"
)

//go:embed sample.conf
var sampleConfig string

func (*GdchAuth) SampleConfig() string {
	return sampleConfig
}

// GdchAuth is the main authenticator struct
type GdchAuth struct {
	Audience           string          `toml:"audience"`
	Log                telegraf.Logger `toml:"-"`
	ServiceAccountFile string          `toml:"service_account_file"`
	TokenExpiryBuffer  config.Duration `toml:"token_expiry_buffer"`
	HTTPClientConfig   httpconfig.HTTPClientConfig

	account *serviceAccountKey
	client  *http.Client
	expiry  time.Time
	token   string
	sync.Mutex
}

func (g *GdchAuth) Init() error {
	if g.ServiceAccountFile == "" {
		return errors.New("service_account_file is required")
	}

	if g.Audience == "" {
		return errors.New("audience is required")
	}

	keyData, err := os.ReadFile(g.ServiceAccountFile)
	if err != nil {
		return fmt.Errorf("failed to read service account file: %w", err)
	}

	g.account = &serviceAccountKey{}
	if err := json.Unmarshal(keyData, g.account); err != nil {
		return fmt.Errorf("failed to parse service account JSON: %w", err)
	}

	if err := g.parsePrivateKey(); err != nil {
		return err
	}

	return g.buildHTTPClient()
}

// Get retrieves the token. The key is ignored as this secret store only provides one secret.
func (g *GdchAuth) Get(_ string) ([]byte, error) {
	if err := g.getToken(context.Background()); err != nil {
		return nil, err
	}
	return []byte(g.token), nil
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

// getToken retrieves a GDCH auth token. It caches the token and reuses it
// until it is within the 'token_expiry_buffer' of its expiry time.
func (g *GdchAuth) getToken(ctx context.Context) error {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	if g.token != "" && time.Now().Before(g.expiry.Add(-time.Duration(g.TokenExpiryBuffer))) {
		return nil
	}

	if err := g.fetchNewToken(ctx); err != nil {
		return err
	}
	return nil
}

func (g *GdchAuth) buildHTTPClient() error {
	client, err := g.HTTPClientConfig.CreateClient(context.Background(), g.Log)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}
	g.client = client
	return nil
}

func (g *GdchAuth) parsePrivateKey() error {
	block, _ := pem.Decode([]byte(g.account.PrivateKey))
	if block == nil {
		return errors.New("failed to decode PEM block from private key")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil || key == nil {
		return fmt.Errorf("private key could not be parsed %w", err)
	}

	g.account.parsedKey = key
	g.account.signingMethod = jwt.SigningMethodES256

	return nil
}

// genertes a jwt token and requests access token from provided service-account tokenURI
func (g *GdchAuth) fetchNewToken(ctx context.Context) error {
	now := time.Now()

	claims := jwt.MapClaims{
		"iss": "system:serviceaccount:" + g.account.Project + ":" + g.account.ServiceIdentityName,
		"sub": "system:serviceaccount:" + g.account.Project + ":" + g.account.ServiceIdentityName,
		"aud": g.account.TokenURI,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(g.account.signingMethod, claims)
	token.Header["kid"] = g.account.PrivateKeyID

	signedJWT, err := token.SignedString(g.account.parsedKey)
	if err != nil {
		return fmt.Errorf("failed to sign JWT: %w", err)
	}

	tokenRequestBody := map[string]string{
		"grant_type":           "urn:ietf:params:oauth:token-type:token-exchange",
		"audience":             g.Audience,
		"requested_token_type": "urn:ietf:params:oauth:token-type:access_token",
		"subject_token":        signedJWT,
		"subject_token_type":   "urn:k8s:params:oauth:token-type:serviceaccount",
	}

	jsonBody, err := json.Marshal(tokenRequestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal token request JSON: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.account.TokenURI, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request returned non-200 status %d: %s", resp.StatusCode, string(body))
	}

	var response tokenResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse token response JSON: %w", err)
	}

	if response.AccessToken == "" {
		return errors.New("token response did not contain 'access_token'")
	}
	g.token = response.AccessToken

	if response.ExpiresIn > 0 {
		g.expiry = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)
	} else {
		g.expiry = now.Add(time.Hour)
	}

	return nil
}

func init() {
	secretstores.Add("gdchauth", func(_ string) telegraf.SecretStore {
		return &GdchAuth{
			TokenExpiryBuffer: config.Duration(5 * time.Minute),
		}
	})
}

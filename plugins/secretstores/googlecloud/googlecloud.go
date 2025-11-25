//go:generate ../../../tools/readme_config_includer/generator
package googlecloud

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"cloud.google.com/go/auth"
	"cloud.google.com/go/auth/credentials"

	"github.com/influxdata/telegraf"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/common/slog"
	"github.com/influxdata/telegraf/plugins/secretstores"
)

//go:embed sample.conf
var sampleConfig string

func (*GoogleCloud) SampleConfig() string {
	return sampleConfig
}

type GoogleCloud struct {
	STSAudience     string          `toml:"sts_audience"`
	CredentialsFile string          `toml:"credentials_file"`
	Log             telegraf.Logger `toml:"-"`
	common_http.HTTPClientConfig

	credentials *auth.Credentials
}

func (g *GoogleCloud) Init() error {
	client, err := g.HTTPClientConfig.CreateClient(context.Background(), g.Log)
	if err != nil {
		return fmt.Errorf("creating HTTP client failed: %w", err)
	}
	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		STSAudience:     g.STSAudience,
		CredentialsFile: g.CredentialsFile,
		Client:          client,
		Logger:          slog.NewLogger(g.Log),
	})
	if err != nil {
		return fmt.Errorf("credentials search failed: %w", err)
	}
	g.credentials = creds
	return nil
}

// Get retrieves the token. The key is ignored as this secret store only provides one secret.
func (g *GoogleCloud) Get(key string) ([]byte, error) {
	if key != "token" {
		return nil, fmt.Errorf("invalid key %q, only 'token' is supported", key)
	}
	token, err := g.credentials.Token(context.Background())
	if err != nil {
		return nil, fmt.Errorf("token retrieval failed: %w", err)
	}
	return []byte(token.Value), nil
}

// List returns the list of secrets provided by this store.
func (*GoogleCloud) List() ([]string, error) {
	return []string{"token"}, nil
}

// Set is not supported for the gcloud secret store.
func (*GoogleCloud) Set(_, _ string) error {
	return errors.New("setting secrets is not supported")
}

// GetResolver returns a resolver function for the secret.
func (g *GoogleCloud) GetResolver(key string) (telegraf.ResolveFunc, error) {
	return func() ([]byte, bool, error) {
		s, err := g.Get(key)
		return s, true, err
	}, nil
}

func init() {
	secretstores.Add("googlecloud", func(string) telegraf.SecretStore {
		return &GoogleCloud{}
	})
}

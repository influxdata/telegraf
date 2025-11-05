package googlecloud

import (
	"context"
	"fmt"

	_ "embed"
	"errors"

	"github.com/influxdata/telegraf"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/common/slog"
	"github.com/influxdata/telegraf/plugins/secretstores"

	"cloud.google.com/go/auth"
	creds "cloud.google.com/go/auth/credentials"
)

//go:embed sample.conf
var sampleConfig string

func (*GoogleCloudOptions) SampleConfig() string {
	return sampleConfig
}

type GoogleCloudOptions struct {
	STSAudience        string `toml:"sts_audience"`
	ServiceAccountFile string `toml:"service_account_file"`
	common_http.HTTPClientConfig

	creds *auth.Credentials
	Log   telegraf.Logger `toml:"-"`
}

func (g *GoogleCloudOptions) Init() error {
	httpClient, err := g.HTTPClientConfig.CreateClient(context.Background(), g.Log)
	if err != nil {
		return err
	}
	creds, err := creds.DetectDefault(&creds.DetectOptions{
		STSAudience:     g.STSAudience,
		CredentialsFile: g.ServiceAccountFile,
		Client:          httpClient,
		Logger:          slog.NewLogger(g.Log),
	})
	if err != nil {
		return err
	}
	g.creds = creds
	return nil
}

// Get retrieves the token. The key is ignored as this secret store only provides one secret.
func (g *GoogleCloudOptions) Get(key string) ([]byte, error) {
	if key != "token" {
		return nil, fmt.Errorf("invalid key %q, only 'token' is supported", key)
	}
	token, err := g.creds.Token(context.Background())
	if err != nil {
		return nil, err
	}
	return []byte(token.Value), nil
}

// List returns the list of secrets provided by this store.
func (*GoogleCloudOptions) List() ([]string, error) {
	return []string{"token"}, nil
}

// Set is not supported for the gcloud secret store.
func (*GoogleCloudOptions) Set(_, _ string) error {
	return errors.New("setting secrets is not supported")
}

// GetResolver returns a resolver function for the secret.
func (g *GoogleCloudOptions) GetResolver(key string) (telegraf.ResolveFunc, error) {
	return func() ([]byte, bool, error) {
		s, err := g.Get(key)
		return s, true, err
	}, nil
}

func init() {
	secretstores.Add("googlecloud", func(_ string) telegraf.SecretStore {
		return &GoogleCloudOptions{}
	})
}

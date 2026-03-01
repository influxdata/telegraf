//go:generate ../../../tools/readme_config_includer/generator
package vault

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"maps"
	"slices"

	vault "github.com/hashicorp/vault/api"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/secretstores"
	"github.com/influxdata/telegraf/plugins/secretstores/vault/auth"
)

//go:embed sample.conf
var sampleConfig string

type Vault struct {
	ID         string `toml:"id"`
	Address    string `toml:"address"`
	MountPath  string `toml:"mount_path"`
	SecretPath string `toml:"secret_path"`
	Engine     string `toml:"engine"`

	AppRole    *auth.AppRole    `toml:"approle"`
	AwsEC2     *auth.AwsEC2     `toml:"aws_ec2"`
	AwsIAM     *auth.AwsIAM     `toml:"aws_iam"`
	Azure      *auth.Azure      `toml:"azure"`
	Kubernetes *auth.Kubernetes `toml:"kubernetes"`
	UserPass   *auth.UserPass   `toml:"userpass"`

	auth   auth.VaultAuth
	client *vault.Client
}

func (*Vault) SampleConfig() string {
	return sampleConfig
}

func (v *Vault) Init() error {
	switch v.Engine {
	case "kv-v1", "kv-v2":
	case "":
		v.Engine = "kv-v2"
	default:
		return fmt.Errorf("unsupported engine: %s", v.Engine)
	}

	if err := v.validateAuth(); err != nil {
		return err
	}

	if v.ID == "" {
		return errors.New("id missing")
	}
	if v.Address == "" {
		return errors.New("address missing")
	}
	if v.MountPath == "" {
		return errors.New("mount_path missing")
	}
	if v.SecretPath == "" {
		return errors.New("secret_path missing")
	}

	cfg := vault.DefaultConfig()
	cfg.Address = v.Address
	client, err := vault.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("error creating Vault client: %w", err)
	}

	v.client = client

	authInfo, err := v.auth.Authenticate(v.client)
	if err != nil {
		return err
	}

	if renewable, err := authInfo.TokenIsRenewable(); renewable && err == nil {
		watcher, err := v.client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{Secret: authInfo})
		if err != nil {
			return fmt.Errorf("unable to initialize Vault lifetime watcher: %w", err)
		}
		go watcher.Start()
	}

	return nil
}

func (v *Vault) validateAuth() error {
	if v.AppRole != nil {
		if v.auth != nil {
			return errors.New("must only specify one authentication method")
		}
		v.auth = v.AppRole
	}
	if v.AwsEC2 != nil {
		if v.auth != nil {
			return errors.New("must only specify one authentication method")
		}
		v.auth = v.AwsEC2
	}
	if v.AwsIAM != nil {
		if v.auth != nil {
			return errors.New("must only specify one authentication method")
		}
		v.auth = v.AwsIAM
	}
	if v.Azure != nil {
		if v.auth != nil {
			return errors.New("must only specify one authentication method")
		}
		v.auth = v.Azure
	}
	if v.Kubernetes != nil {
		if v.auth != nil {
			return errors.New("must only specify one authentication method")
		}
		v.auth = v.Kubernetes
	}
	if v.UserPass != nil {
		if v.auth != nil {
			return errors.New("must only specify one authentication method")
		}
		v.auth = v.UserPass
	}

	if v.auth == nil {
		return errors.New("no auth method set")
	}
	return v.auth.Validate()
}

func (v *Vault) Get(key string) ([]byte, error) {
	secret, err := v.getSecret()
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %w", err)
	}

	// Secret can exist but have no data if all secrets at the path were
	// deleted. Return an empty array if this is the case, or if the requested
	// key does not exist at the specified path.
	if secret.Data == nil || secret.Data[key] == nil {
		return make([]byte, 0), nil
	}

	value, ok := secret.Data[key].(string)
	if !ok {
		return nil, fmt.Errorf("value type assertion failed: %T", secret.Data[key])
	}
	return []byte(value), nil
}

func (v *Vault) List() ([]string, error) {
	secret, err := v.getSecret()
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %w", err)
	}

	// Secret can exist but have no data if all secrets at the path were deleted
	if secret.Data == nil {
		return nil, errors.New("no secret data found")
	}

	return slices.Collect(maps.Keys(secret.Data)), nil
}

func (v *Vault) Set(key, value string) error {
	secretsData := map[string]interface{}{key: value}

	if v.Engine == "kv-v1" {
		return v.client.KVv1(v.MountPath).Put(context.Background(), v.SecretPath, secretsData)
	}

	_, err := v.client.KVv2(v.MountPath).Put(context.Background(), v.SecretPath, secretsData)
	return err
}

func (v *Vault) GetResolver(key string) (telegraf.ResolveFunc, error) {
	resolver := func() ([]byte, bool, error) {
		s, err := v.Get(key)
		return s, true, err
	}
	return resolver, nil
}

func (v *Vault) getSecret() (*vault.KVSecret, error) {
	if v.Engine == "kv-v1" {
		return v.client.KVv1(v.MountPath).Get(context.Background(), v.SecretPath)
	}
	return v.client.KVv2(v.MountPath).Get(context.Background(), v.SecretPath)
}

func init() {
	secretstores.Add("vault", func(id string) telegraf.SecretStore {
		return &Vault{ID: id}
	})
}

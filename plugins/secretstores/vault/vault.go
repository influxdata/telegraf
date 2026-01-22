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
	"github.com/hashicorp/vault/api/auth/approle"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/secretstores"
)

//go:embed sample.conf
var sampleConfig string

type Vault struct {
	ID         string   `toml:"id"`
	Address    string   `toml:"address"`
	MountPath  string   `toml:"mount_path"`
	SecretPath string   `toml:"secret_path"`
	Engine     string   `toml:"engine"`
	AppRole    *appRole `toml:"approle"`

	client *vault.Client
}

type appRole struct {
	RoleID          string        `toml:"role_id"`
	ResponseWrapped bool          `toml:"response_wrapped"`
	Secret          config.Secret `toml:"secret"`
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

	if v.AppRole == nil {
		return errors.New("approle configuration missing")
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

	return v.authenticate()
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

func (v *Vault) authenticate() error {
	secret, err := v.AppRole.Secret.Get()
	if err != nil {
		return fmt.Errorf("getting secret failed: %w", err)
	}
	secretID := &approle.SecretID{FromString: secret.String()}
	defer secret.Destroy()

	opts := make([]approle.LoginOption, 0)
	if v.AppRole.ResponseWrapped {
		opts = append(opts, approle.WithWrappingToken())
	}

	appRoleAuth, err := approle.NewAppRoleAuth(v.AppRole.RoleID, secretID, opts...)
	if err != nil {
		return fmt.Errorf("unable to initialize AppRole auth method: %w", err)
	}

	authInfo, err := v.client.Auth().Login(context.Background(), appRoleAuth)
	if err != nil {
		return fmt.Errorf("unable to login to AppRole auth method: %w", err)
	}
	if authInfo == nil {
		return errors.New("no auth info was returned after login")
	}

	watcher, err := v.client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{Secret: authInfo})
	if err != nil {
		return fmt.Errorf("unable to initialize Vault lifetime watcher: %w", err)
	}
	go watcher.Start()

	return nil
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

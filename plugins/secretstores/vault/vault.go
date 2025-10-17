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
	"github.com/influxdata/telegraf/plugins/secretstores"
)

//go:embed sample.conf
var sampleConfig string

type Vault struct {
	ID         string   `toml:"id"`
	Address    string   `toml:"address"`
	MountPath  string   `toml:"mount_path"`
	SecretPath string   `toml:"secret_path"`
	UseKVv1    bool     `toml:"use_kv_v1"`
	AppRole    *appRole `toml:"approle"`

	client *vault.Client
}

type appRole struct {
	RoleID          string `toml:"role_id"`
	ResponseWrapped bool   `toml:"response_wrapped"`
	SecretFile      string `toml:"secret_file"`
	SecretEnv       string `toml:"secret_env"`
	SecretID        string `toml:"secret_id"`
}

func (*Vault) SampleConfig() string {
	return sampleConfig
}

func (v *Vault) Init() error {
	if v.ID == "" {
		return errors.New("id missing")
	}
	if v.Address == "" {
		return errors.New("address missing")
	}

	config := vault.DefaultConfig()
	config.Address = v.Address
	client, err := vault.NewClient(config)
	if err != nil {
		return fmt.Errorf("error creating Vault client: %w", err)
	}

	v.client = client

	return v.authenticate()
}

func (v *Vault) authenticate() error {
	secretID := &approle.SecretID{}
	if v.AppRole.SecretFile != "" {
		secretID.FromFile = v.AppRole.SecretFile
	} else if v.AppRole.SecretEnv != "" {
		secretID.FromEnv = v.AppRole.SecretEnv
	} else if v.AppRole.SecretID != "" {
		secretID.FromString = v.AppRole.SecretID
	} else {
		return errors.New("no AppRole credentials specified")
	}

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

	return nil
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

	// Secret can exist but have no data if all secrets at the path were
	// deleted. Return an empty array if this is the case
	if secret.Data == nil {
		return make([]string, 0), nil
	}

	return slices.Collect(maps.Keys(secret.Data)), nil
}

func (v *Vault) getSecret() (*vault.KVSecret, error) {
	if v.UseKVv1 {
		return v.client.KVv1(v.MountPath).Get(context.Background(), v.SecretPath)
	}
	return v.client.KVv2(v.MountPath).Get(context.Background(), v.SecretPath)
}

func (v *Vault) Set(key, value string) error {
	secretsData := map[string]interface{}{key: value}

	var err error
	if v.UseKVv1 {
		err = v.client.KVv1(v.MountPath).Put(context.Background(), v.SecretPath, secretsData)
	} else {
		_, err = v.client.KVv2(v.MountPath).Put(context.Background(), v.SecretPath, secretsData)
	}
	return err
}

func (v *Vault) GetResolver(key string) (telegraf.ResolveFunc, error) {
	resolver := func() ([]byte, bool, error) {
		s, err := v.Get(key)
		return s, true, err
	}
	return resolver, nil
}

func init() {
	secretstores.Add("vault", func(id string) telegraf.SecretStore {
		return &Vault{ID: id}
	})
}

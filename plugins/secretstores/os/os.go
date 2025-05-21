//go:build darwin || linux || windows

//go:generate ../../../tools/readme_config_includer/generator
package os

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/99designs/keyring"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/secretstores"
)

//go:embed sample.conf
var sampleConfig string

type OS struct {
	ID         string        `toml:"id"`
	Keyring    string        `toml:"keyring"`
	Collection string        `toml:"collection"`
	Dynamic    bool          `toml:"dynamic"`
	Password   config.Secret `toml:"password"`

	ring keyring.Keyring
}

func (*OS) SampleConfig() string {
	return sampleConfig
}

func (o *OS) Init() error {
	defer o.Password.Destroy()

	if o.ID == "" {
		return errors.New("id missing")
	}

	// Set defaults
	if o.Keyring == "" {
		o.Keyring = "telegraf"
	}

	// Setup the actual keyring
	cfg, err := o.createKeyringConfig()
	if err != nil {
		return fmt.Errorf("getting keyring config failed: %w", err)
	}
	kr, err := keyring.Open(cfg)
	if err != nil {
		return fmt.Errorf("opening keyring failed: %w", err)
	}
	o.ring = kr

	return nil
}

func (o *OS) Get(key string) ([]byte, error) {
	item, err := o.ring.Get(key)
	if err != nil {
		return nil, err
	}

	return item.Data, nil
}

func (o *OS) Set(key, value string) error {
	item := keyring.Item{
		Key:  key,
		Data: []byte(value),
	}

	return o.ring.Set(item)
}

func (o *OS) List() ([]string, error) {
	return o.ring.Keys()
}

func (o *OS) GetResolver(key string) (telegraf.ResolveFunc, error) {
	resolver := func() ([]byte, bool, error) {
		s, err := o.Get(key)
		return s, o.Dynamic, err
	}
	return resolver, nil
}

func init() {
	secretstores.Add("os", func(id string) telegraf.SecretStore {
		return &OS{ID: id}
	})
}

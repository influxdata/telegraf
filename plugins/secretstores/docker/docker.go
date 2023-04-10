//go:generate ../../../tools/readme_config_includer/generator
package docker

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/secretstores"
)

//go:embed sample.conf
var sampleConfig string

type Docker struct {
	ID   string `toml:"id"`
	Path string `toml:"path"`
}

func (*Docker) SampleConfig() string {
	return sampleConfig
}

// Init initializes all internals of the secret-store
func (d *Docker) Init() error {
	if d.ID == "" {
		return errors.New("id missing")
	}
	if d.Path == "" {
		return errors.New("path missing")
	}
	if _, err := os.Stat(d.Path); os.IsNotExist(err) {
		return errors.New("directory does not exist")
	}
	return nil
}

func (d *Docker) Get(key string) ([]byte, error) {
	secretFile := filepath.Join(d.Path, key)
	value, err := os.ReadFile(secretFile)
	if err != nil {
		return nil, errors.New("cannot find the secrets file under the directory mentioned in path parameter")
	}
	return value, nil
}

func (d *Docker) List() ([]string, error) {
	secretFiles, err := os.ReadDir(d.Path)
	if err != nil {
		return nil, errors.New("cannot read files under the directory mentioned in path")
	}
	secrets := make([]string, 0, len(secretFiles))
	for _, entry := range secretFiles {
		secrets = append(secrets, entry.Name())
	}
	return secrets, nil
}

func (d *Docker) Set(_, _ string) error {
	return errors.New("secret-store does not support creating secrets")
}

// GetResolver returns a function to resolve the given key.
func (d *Docker) GetResolver(key string) (telegraf.ResolveFunc, error) {
	resolver := func() ([]byte, bool, error) {
		s, err := d.Get(key)
		return s, false, err
	}
	return resolver, nil
}

// Register the secret-store on load.
func init() {
	secretstores.Add("docker", func(id string) telegraf.SecretStore {
		return &Docker{ID: id}
	})
}

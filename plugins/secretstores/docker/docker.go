//go:generate ../../../tools/readme_config_includer/generator
package docker

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/secretstores"
)

//go:embed sample.conf
var sampleConfig string

type Docker struct {
	ID      string `toml:"id"`
	Path    string `toml:"path"`
	Dynamic bool   `toml:"dynamic"`
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
		// setting the default directory for Docker Secrets
		// if no explicit path mentioned in configuration
		d.Path = "/run/secrets"
	}
	if _, err := os.Stat(d.Path); err != nil {
		// if there is no /run/secrets directory for default Path value
		// this implies that there are no secrets.
		// Or for any explicit path definitions for that matter.
		return fmt.Errorf("accessing directory %q failed: %w", d.Path, err)
	}
	return nil
}

func (d *Docker) Get(key string) ([]byte, error) {
	secretFile, err := filepath.Abs(filepath.Join(d.Path, key))
	if err != nil {
		return nil, err
	}
	if filepath.Dir(secretFile) != d.Path {
		return nil, fmt.Errorf("directory traversal detected for key %q", key)
	}
	value, err := os.ReadFile(secretFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read the secret's value under the directory: %w", err)
	}
	return value, nil
}

func (d *Docker) List() ([]string, error) {
	secretFiles, err := os.ReadDir(d.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot read files under the directory: %w", err)
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
		return s, d.Dynamic, err
	}
	return resolver, nil
}

// Register the secret-store on load.
func init() {
	secretstores.Add("docker", func(id string) telegraf.SecretStore {
		return &Docker{ID: id}
	})
}

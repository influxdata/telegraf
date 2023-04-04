//go:generate ../../../tools/readme_config_includer/generator
package docker

import (
	_ "embed"
	"errors"
	"os"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/secretstores"
)

//go:embed sample.conf
var sampleConfig string

// mount directory created by Docker when using Docker Secrets
const dockerSecretsDir string = "/run/secrets"

type Docker struct {
	ID string `toml:"id"`
}

func (*Docker) SampleConfig() string {
	return sampleConfig
}

// Init initializes all internals of the secret-store
func (d *Docker) Init() error {
	if d.ID == "" {
		return errors.New("id missing")
	}
	if _, err := os.Stat(dockerSecretsDir); os.IsNotExist(err) {
		return errors.New("/run/secrets directory does not exist")
	}
	return nil
}

func (d *Docker) Get(key string) ([]byte, error) {
	secretFile := dockerSecretsDir + "/" + key
	value, err := os.ReadFile(secretFile)
	if err != nil {
		return nil, errors.New("cannot find the secret file under /run/secrets")
	}
	if string(value) == "" {
		return nil, errors.New("the value of the secrets file is empty")
	}
	return value, nil
}

func (d *Docker) List() ([]string, error) {
	var secrets []string
	secretFiles, err := os.ReadDir(dockerSecretsDir)
	if err != nil {
		return nil, errors.New("cannot read files under /run/secrets directory")
	}
	if len(secretFiles) == 0 {
		return nil, errors.New("cannot find any secrets under /run/secrets")
	}
	for _, entry := range secretFiles {
		secrets = append(secrets, entry.Name())
	}
	return secrets, nil
}

func (d *Docker) Set(key, value string) error {
	return nil
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

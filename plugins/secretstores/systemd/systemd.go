//go:build linux

//go:generate ../../../tools/readme_config_includer/generator
package systemd

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/secretstores"
)

const systemdMinimumVersion = 250

// Required to be a variable to mock the version in tests
var getSystemdVersion = getSystemdMajorVersion

//go:embed sample.conf
var sampleConfig string

type Systemd struct {
	Path   string          `toml:"path"`
	Prefix string          `toml:"prefix"`
	Log    telegraf.Logger `toml:"-"`
}

func (*Systemd) SampleConfig() string {
	return sampleConfig
}

// Init initializes all internals of the secret-store
func (s *Systemd) Init() error {
	version, err := getSystemdVersion()
	if err != nil {
		return fmt.Errorf("unable to detect systemd version: %w", err)
	}
	s.Log.Debugf("Found systemd version %d...", version)
	if version < systemdMinimumVersion {
		return fmt.Errorf("systemd version %d below minimum version %d", version, systemdMinimumVersion)
	}

	// By default the credentials directory is passed in by systemd
	// via the "CREDENTIALS_DIRECTORY" environment variable.
	defaultPath := os.Getenv("CREDENTIALS_DIRECTORY")
	if defaultPath == "" {
		s.Log.Warn("CREDENTIALS_DIRECTORY environment variable undefined. Make sure credentials are setup correctly!")
		if s.Path == "" {
			return errors.New("'path' required without CREDENTIALS_DIRECTORY")
		}
	}

	// Use default path if no explicit was specified. This should be the common case.
	if s.Path == "" {
		s.Path = defaultPath
	}
	s.Path, err = filepath.Abs(s.Path)
	if err != nil {
		return fmt.Errorf("cannot determine absolute path of %q: %w", s.Path, err)
	}

	// Check if we can access the target directory
	if _, err := os.Stat(s.Path); err != nil {
		return fmt.Errorf("accessing credentials directory %q failed: %w", s.Path, err)
	}
	return nil
}

func (s *Systemd) Get(key string) ([]byte, error) {
	secretFile, err := filepath.Abs(filepath.Join(s.Path, s.Prefix+key))
	if err != nil {
		return nil, err
	}
	if filepath.Dir(secretFile) != s.Path {
		return nil, fmt.Errorf("invalid directory detected for key %q", key)
	}
	value, err := os.ReadFile(secretFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read the secret's value: %w", err)
	}
	return value, nil
}

func (s *Systemd) List() ([]string, error) {
	secretFiles, err := os.ReadDir(s.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot read files: %w", err)
	}
	secrets := make([]string, 0, len(secretFiles))
	for _, entry := range secretFiles {
		key := strings.TrimPrefix(entry.Name(), s.Prefix)
		secrets = append(secrets, key)
	}
	return secrets, nil
}

func (*Systemd) Set(_, _ string) error {
	return errors.New("secret-store does not support creating secrets")
}

// GetResolver returns a function to resolve the given key.
func (s *Systemd) GetResolver(key string) (telegraf.ResolveFunc, error) {
	resolver := func() ([]byte, bool, error) {
		s, err := s.Get(key)
		return s, false, err
	}
	return resolver, nil
}

func getSystemdMajorVersion() (int, error) {
	ctx := context.Background()
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	fullVersion, err := conn.GetManagerProperty("Version")
	if err != nil {
		return 0, err
	}
	fullVersion = strings.Trim(fullVersion, "\"")
	return strconv.Atoi(strings.SplitN(fullVersion, ".", 2)[0])
}

// Register the secret-store on load.
func init() {
	secretstores.Add("systemd", func(_ string) telegraf.SecretStore {
		return &Systemd{Prefix: "telegraf."}
	})
}

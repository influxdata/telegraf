package secretstore

import (
	"fmt"
	"net/url"
	"strings"
)

type storeImpl interface {
	Get(key string) (string, error)
	Set(key, value string) error
	List() ([]string, error)
}

type SecretStore struct {
	Name     string `toml:"name"`
	Service  string `toml:"service"`
	Password string `toml:"password"`

	store storeImpl
}

// Init initializes all internals of the secret-store
func (s *SecretStore) Init() error {
	// Remove the password from memory when leaving
	defer func() {
		s.Password = strings.Repeat("*", len(s.Password))
		s.Password = ""
	}()

	if s.Name == "" {
		return fmt.Errorf("name missing")
	}

	// Default
	if s.Service == "" {
		s.Service = "os://telegraf"
	}

	// Determine the service type and map it to the implementation configuration
	u, err := url.Parse(s.Service)
	if err != nil {
		return fmt.Errorf("parsing service failed: %v", err)
	}

	// Determine the additional arguments
	path := strings.TrimPrefix(strings.TrimPrefix(s.Service, u.Scheme+":"), "//")
	if path == "" {
		path = "telegraf"
	}

	switch u.Scheme {
	case "file", "kwallet", "os", "secret-service":
		s.store, err = NewKeyringStore(s.Name, u.Scheme, path, s.Password)
		if err != nil {
			return fmt.Errorf("creating keyring store for service %q failed: %v", u.Scheme, err)
		}
	default:
		return fmt.Errorf("unknown service %q", u.Scheme)
	}

	return nil
}

// Get searches for the given key and return the secret
func (s *SecretStore) Get(key string) (string, error) {
	return s.store.Get(key)
}

// Set sets the given secret for the given key
func (s *SecretStore) Set(key, value string) error {
	return s.store.Set(key, value)
}

// List lists all known secret keys
func (s *SecretStore) List() ([]string, error) {
	return s.store.List()
}

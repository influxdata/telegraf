package secretstore

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/99designs/keyring"
)

type KeyringStore struct {
	keyring.Keyring
}

// Init initializes all internals of the secret-store
func NewKeyringStore(name, scheme, path, passwd string) (*KeyringStore, error) {
	config := keyring.Config{
		ServiceName: "telegraf",
	}

	// Create the prompt-function in case we need it
	promptFunc := keyring.TerminalPrompt
	if passwd != "" {
		promptFunc = keyring.FixedStringPrompt(passwd)
	}

	switch scheme {
	case "file":
		config.AllowedBackends = []keyring.BackendType{keyring.FileBackend}
		config.FileDir = path
		config.FilePasswordFunc = promptFunc
	case "kwallet":
		params := strings.SplitN(path, "/", 2)
		folder := ""
		if len(params) > 1 {
			folder = params[1]
		}
		config.AllowedBackends = []keyring.BackendType{keyring.KWalletBackend}
		config.KWalletAppID = params[0]
		config.KWalletFolder = folder
	case "os":
		switch runtime.GOOS {
		case "darwin":
			config.AllowedBackends = []keyring.BackendType{keyring.KeychainBackend}
			config.KeychainName = path
			config.KeychainPasswordFunc = promptFunc
		case "linux":
			config.AllowedBackends = []keyring.BackendType{keyring.KeyCtlBackend}
			config.KeyCtlScope = "user"
			config.KeyCtlPerm = 0x3f3f0000 // "alswrvalswrv------------"
			config.ServiceName = path
		case "windows":
			config.AllowedBackends = []keyring.BackendType{keyring.WinCredBackend}
			config.ServiceName = path
		default:
			return nil, fmt.Errorf("'os' service not supported for OS %q", runtime.GOOS)
		}
	case "secret-service":
		config.AllowedBackends = []keyring.BackendType{keyring.SecretServiceBackend}
		config.LibSecretCollectionName = path
	default:
		return nil, fmt.Errorf("service not supported")
	}

	kr, err := keyring.Open(config)
	if err != nil {
		return nil, err
	}
	return &KeyringStore{Keyring: kr}, nil
}

func (s *KeyringStore) Get(key string) (string, error) {
	item, err := s.Keyring.Get(key)
	if err != nil {
		return "", err
	}

	return string(item.Data), nil
}

func (s *KeyringStore) Set(key, value string) error {
	item := keyring.Item{
		Key:  key,
		Data: []byte(value),
	}

	return s.Keyring.Set(item)
}

func (s *KeyringStore) List() ([]string, error) {
	return s.Keyring.Keys()
}

//go:build windows

package os

import (
	"github.com/99designs/keyring"
)

func (o *OS) createKeyringConfig() (keyring.Config, error) {
	return keyring.Config{
		ServiceName:     o.Keyring,
		AllowedBackends: []keyring.BackendType{keyring.WinCredBackend},
		WinCredPrefix:   o.Collection,
	}, nil
}

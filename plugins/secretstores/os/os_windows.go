//go:build windows

package os

import (
	_ "embed"

	"github.com/99designs/keyring"
)

//go:embed sample_windows.conf
var sampleConfig string

func (o *OS) createKeyringConfig() (keyring.Config, error) {
	return keyring.Config{
		ServiceName:     o.Keyring,
		AllowedBackends: []keyring.BackendType{keyring.WinCredBackend},
		WinCredPrefix:   o.Collection,
	}, nil
}

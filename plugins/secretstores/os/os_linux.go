//go:build linux

package os

import (
	"github.com/99designs/keyring"
)

func (o *OS) createKeyringConfig() (keyring.Config, error) {
	if o.Keyring == "" {
		o.Keyring = "telegraf"
	}
	return keyring.Config{
		ServiceName:     o.Keyring,
		AllowedBackends: []keyring.BackendType{keyring.KeyCtlBackend},
		KeyCtlScope:     "user",
		KeyCtlPerm:      0x3f3f0000, // "alswrvalswrv------------"
	}, nil
}

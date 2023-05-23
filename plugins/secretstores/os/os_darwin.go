//go:build darwin

package os

import (
	"fmt"

	"github.com/99designs/keyring"

	"github.com/influxdata/telegraf/config"
)

func (o *OS) createKeyringConfig() (keyring.Config, error) {
	passwd, err := o.Password.Get()
	if err != nil {
		return keyring.Config{}, fmt.Errorf("getting password failed: %w", err)
	}
	defer config.ReleaseSecret(passwd)

	// Create the prompt-function in case we need it
	promptFunc := keyring.TerminalPrompt
	if len(passwd) != 0 {
		promptFunc = keyring.FixedStringPrompt(string(passwd))
	}

	return keyring.Config{
		ServiceName:          o.Collection,
		AllowedBackends:      []keyring.BackendType{keyring.KeychainBackend},
		KeychainName:         o.Keyring,
		KeychainPasswordFunc: promptFunc,
	}, nil
}

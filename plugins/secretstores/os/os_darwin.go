//go:build darwin

package os

import (
	"fmt"

	"github.com/99designs/keyring"
)

func (o *OS) createKeyringConfig() (keyring.Config, error) {
	// Create the prompt-function in case we need it
	promptFunc := keyring.TerminalPrompt
	if !o.Password.Empty() {
		passwd, err := o.Password.Get()
		if err != nil {
			return keyring.Config{}, fmt.Errorf("getting password failed: %w", err)
		}
		promptFunc = keyring.FixedStringPrompt(passwd.String())
		passwd.Destroy()
	}

	return keyring.Config{
		ServiceName:          o.Collection,
		AllowedBackends:      []keyring.BackendType{keyring.KeychainBackend},
		KeychainName:         o.Keyring,
		KeychainPasswordFunc: promptFunc,
	}, nil
}

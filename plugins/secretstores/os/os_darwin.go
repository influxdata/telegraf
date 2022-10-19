//go:build darwin
// +build darwin

package os

import (
	_ "embed"
	"fmt"

	"github.com/99designs/keyring"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//
//go:embed sample_darwin.conf
var sampleConfig string

func (o *OS) createKeyringConfig() (keyring.Config, error) {
	passwd, err := o.Password.Get()
	if err != nil {
		return keyring.Config{}, fmt.Errorf("getting password failed: %v", err)
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

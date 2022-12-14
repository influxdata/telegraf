//go:build linux
// +build linux

package os

import (
	_ "embed"

	"github.com/99designs/keyring"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//
//go:embed sample_linux.conf
var sampleConfig string

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

//go:build !linux

package config

import (
	"github.com/awnumar/memguard"
)

func protect(secret []byte) error {
	return nil
}

func ReleaseSecret(secret []byte) {
	memguard.WipeBytes(secret)
}

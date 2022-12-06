//go:build linux

package config

import (
	"syscall"

	"github.com/awnumar/memguard"
)

func protect(secret []byte) error {
	return syscall.Mlock(secret)
}

func ReleaseSecret(secret []byte) {
	memguard.WipeBytes(secret)
	if err := syscall.Munlock(secret); err != nil {
		panic(err)
	}
}

//go:build linux

package config

import (
	"syscall"

	"github.com/awnumar/memguard"
)

func protect(secret []byte) error {
	return syscall.Mlock(secret)
}

func ReleaseSecret(secret []byte) error {
	memguard.WipeBytes(secret)
	return syscall.Munlock(secret)
}

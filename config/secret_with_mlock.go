//go:build linux

package config

import (
	"github.com/awnumar/memguard"
)

func protect(_ []byte) error {
	//return syscall.Mlock(secret)
	return nil
}

func ReleaseSecret(secret []byte) {
	memguard.WipeBytes(secret)
	// if err := syscall.Munlock(secret); err != nil {
	// 	panic(err)
	// }
}

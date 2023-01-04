package encryption

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"strings"

	"golang.org/x/crypto/pbkdf2"

	"github.com/influxdata/telegraf/config"
)

type KDFConfig struct {
	Algorithm  string        `toml:"kdf_algorithm"`
	Passwd     config.Secret `toml:"password"`
	Salt       config.Secret `toml:"salt"`
	Iterations int           `toml:"iterations"`
}

type hashFunc func() hash.Hash

func (k *KDFConfig) NewKey(keylen int) (key, iv []byte, err error) {
	switch strings.ToUpper(k.Algorithm) {
	case "", "PBKDF2-HMAC-SHA256":
		return k.generatePBKDF2HMAC(sha256.New, keylen)
	}
	return nil, nil, fmt.Errorf("unknown key-derivation function %q", k.Algorithm)
}

func (k *KDFConfig) generatePBKDF2HMAC(hf hashFunc, keylen int) ([]byte, []byte, error) {
	passwd, err := k.Passwd.Get()
	if err != nil {
		return nil, nil, fmt.Errorf("getting password failed: %w", err)
	}
	defer config.ReleaseSecret(passwd)

	salt, err := k.Salt.Get()
	if err != nil {
		return nil, nil, fmt.Errorf("getting salt failed: %w", err)
	}
	defer config.ReleaseSecret(salt)

	key := pbkdf2.Key(passwd, salt, k.Iterations, keylen, hf)
	return key, nil, nil
}

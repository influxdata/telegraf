package http

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"strings"

	"golang.org/x/crypto/pbkdf2"

	"github.com/influxdata/telegraf/config"
)

type kdfConfig struct {
	Algorithm  string        `toml:"kdf_algorithm"`
	Passwd     config.Secret `toml:"password"`
	Salt       config.Secret `toml:"salt"`
	Iterations int           `toml:"iterations"`
}

type hashFunc func() hash.Hash

func (k *kdfConfig) newKey(keyLen int) (key, iv config.Secret, err error) {
	switch strings.ToUpper(k.Algorithm) {
	case "", "PBKDF2-HMAC-SHA256":
		return k.generatePBKDF2HMAC(sha256.New, keyLen)
	}
	return config.Secret{}, config.Secret{}, fmt.Errorf("unknown key-derivation function %q", k.Algorithm)
}

func (k *kdfConfig) generatePBKDF2HMAC(hf hashFunc, keyLen int) (key, iv config.Secret, err error) {
	if k.Iterations == 0 {
		return config.Secret{}, config.Secret{}, errors.New("'iteration value not set")
	}

	passwd, err := k.Passwd.Get()
	if err != nil {
		return config.Secret{}, config.Secret{}, fmt.Errorf("getting password failed: %w", err)
	}
	defer passwd.Destroy()

	salt, err := k.Salt.Get()
	if err != nil {
		return config.Secret{}, config.Secret{}, fmt.Errorf("getting salt failed: %w", err)
	}
	defer salt.Destroy()

	rawkey := pbkdf2.Key(passwd.Bytes(), salt.Bytes(), k.Iterations, keyLen, hf)
	key = config.NewSecret([]byte(hex.EncodeToString(rawkey)))
	return key, config.Secret{}, nil
}

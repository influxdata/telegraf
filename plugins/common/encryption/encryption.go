package encryption

import (
	"fmt"
	"strings"
)

type Decrypter interface {
	Init() error
	Decrypt(data []byte) ([]byte, error)
}

type EncryptionConfig struct {
	Cipher string       `toml:"cipher"`
	Aes    AesEncryptor `toml:"aes"`
}

func (c *EncryptionConfig) CreateDecrypter() (Decrypter, error) {
	// For ciphers that allowing variants (e.g. AES256/CBC/PKCS#5Padding)
	// can specify the variant using <algorithm>[/param 1>[/<param 2>]...]
	// where all parameters will be passed on to the decryptor.
	parts := strings.Split(c.Cipher, "/")
	switch strings.ToLower(parts[0]) {
	case "aes", "aes128", "aes192", "aes256":
		c.Aes.Variant = parts
		if err := c.Aes.Init(); err != nil {
			return nil, fmt.Errorf("init of AES decrypter failed: %w", err)
		}
		return &c.Aes, nil
	}
	return nil, fmt.Errorf("unknown cipher %q", c.Cipher)
}

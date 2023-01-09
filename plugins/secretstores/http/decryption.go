package http

import (
	_ "embed"
	"fmt"
	"strings"
)

type Decrypter interface {
	Decrypt(data []byte) ([]byte, error)
}

type DecryptionConfig struct {
	Cipher string       `toml:"cipher"`
	Aes    AesEncryptor `toml:"aes"`
}

func (c *DecryptionConfig) CreateDecrypter() (Decrypter, error) {
	// For ciphers that allowing variants (e.g. AES256/CBC/PKCS#5Padding)
	// can specify the variant using <algorithm>[/param 1>[/<param 2>]...]
	// where all parameters will be passed on to the decrypter.
	parts := strings.Split(c.Cipher, "/")
	switch strings.ToLower(parts[0]) {
	case "", "none":
		return nil, nil
	case "aes", "aes128", "aes192", "aes256":
		c.Aes.Variant = parts
		if err := c.Aes.Init(); err != nil {
			return nil, fmt.Errorf("init of AES decrypter failed: %w", err)
		}
		return &c.Aes, nil
	}
	return nil, fmt.Errorf("unknown cipher %q", c.Cipher)
}

func PKCS7Trimming(in []byte) []byte {
	// 'count' number of bytes where padded to the end of the clear-text
	// each containing the value of 'count'
	count := int(in[len(in)-1])
	return in[:len(in)-count]
}

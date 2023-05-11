package http

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/influxdata/telegraf/config"
)

type AesEncryptor struct {
	Variant []string      `toml:"-"`
	Key     config.Secret `toml:"key"`
	Vec     config.Secret `toml:"init_vector"`
	KDFConfig

	mode string
	trim func([]byte) ([]byte, error)
}

func (a *AesEncryptor) Init() error {
	var cipherName, mode, padding string

	switch len(a.Variant) {
	case 3:
		padding = strings.ToLower(a.Variant[2])
		fallthrough
	case 2:
		mode = strings.ToLower(a.Variant[1])
		fallthrough
	case 1:
		cipherName = strings.ToLower(a.Variant[0])
		if !strings.HasPrefix(cipherName, "aes") {
			return fmt.Errorf("requested AES but specified %q", cipherName)
		}
	case 0:
		return errors.New("please specify cipher")
	default:
		return errors.New("too many variant elements")
	}

	var keylen int
	switch cipherName {
	case "aes128":
		keylen = 16
	case "aes192":
		keylen = 24
	case "aes256":
		keylen = 32
	default:
		return fmt.Errorf("unsupported AES cipher %q", cipherName)
	}

	switch mode {
	case "", "none": // pure AES
		mode = "none"
	case "cbc": // AES block mode
	default:
		return fmt.Errorf("unsupported cipher mode %q", a.Variant[1])
	}
	a.mode = mode

	// Setup the trimming function to revert padding
	switch padding {
	case "", "none":
		// identity, no padding
		a.trim = func(in []byte) ([]byte, error) { return in, nil }
	case "pkcs#5", "pkcs#7":
		a.trim = PKCS5or7Trimming
	default:
		return fmt.Errorf("unsupported padding %q", padding)
	}

	// Generate the key using password-based-keys
	if a.Key.Empty() {
		if a.Passwd.Empty() {
			return errors.New("either key or password has to be specified")
		}
		if a.Salt.Empty() || a.Iterations == 0 {
			return errors.New("salt and iterations required for password-based-keys")
		}

		key, iv, err := a.KDFConfig.NewKey(keylen)
		if err != nil {
			return fmt.Errorf("generating key failed: %w", err)
		}
		a.Key.Destroy()
		a.Key = key

		if a.Vec.Empty() && iv.Empty() {
			a.Vec.Destroy()
			a.Vec = iv
		}
	} else {
		encodedKey, err := a.Key.Get()
		if err != nil {
			return fmt.Errorf("getting key failed: %w", err)
		}
		key := make([]byte, hex.DecodedLen(len(encodedKey)))
		if _, err := hex.Decode(key, encodedKey); err != nil {
			config.ReleaseSecret(encodedKey)
			return fmt.Errorf("decoding key failed: %w", err)
		}
		config.ReleaseSecret(encodedKey)

		if len(key) != keylen {
			return fmt.Errorf("key length (%d bit) does not match cipher (%d bit)", len(key)*8, keylen*8)
		}
	}

	if a.Vec.Empty() {
		return errors.New("'init_vector' has to be specified or derived from password")
	}

	return nil
}

func (a *AesEncryptor) Decrypt(data []byte) ([]byte, error) {
	if len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("invalid data size %d", len(data))
	}

	// Setup the cipher and return the decoded data
	encodedKey, err := a.Key.Get()
	if err != nil {
		return nil, fmt.Errorf("getting key failed: %w", err)
	}
	key := make([]byte, hex.DecodedLen(len(encodedKey)))
	if _, err := hex.Decode(key, encodedKey); err != nil {
		config.ReleaseSecret(encodedKey)
		return nil, fmt.Errorf("decoding key failed: %w", err)
	}
	config.ReleaseSecret(encodedKey)

	// Setup AES
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating AES cipher failed: %w", err)
	}

	// Setup the block/stream cipher and decode the data
	iv, err := a.Vec.Get()
	if err != nil {
		return nil, fmt.Errorf("getting initialization-vector failed: %w", err)
	}
	defer config.ReleaseSecret(iv)

	switch a.mode {
	case "none":
		block.Decrypt(data, data)
	case "cbc":
		cipher.NewCBCDecrypter(block, iv).CryptBlocks(data, data)
	default:
		return nil, fmt.Errorf("unsupported block mode %q", a.mode)
	}

	return a.trim(data)
}

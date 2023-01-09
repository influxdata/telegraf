package http

import (
	"crypto/aes"
	"crypto/cipher"
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
	trim func([]byte) []byte
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
	case "aes", "aes128":
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
	case "cfb", "ctr", "ofb": // AES stream mode
	default:
		return fmt.Errorf("unsupported block mode %q", a.Variant[1])
	}
	a.mode = mode

	// Setup the trimming function to revert padding
	switch padding {
	case "", "none":
		// identity, no padding
		a.trim = func(in []byte) []byte { return in }
	case "pkcs5", "pkcs#5", "pkcs5padding", "pkcs#5padding":
		// The implementation can handle both variants, so fallthrough to
		// the PKCS#7 case
		fallthrough
	case "pkcs7", "pkcs#7", "pkcs7padding", "pkcs#7padding":
		a.trim = PKCS7Trimming
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
		a.Key = config.NewSecret(key)
		if a.Vec.Empty() && len(iv) > 0 {
			a.Vec.Destroy()
			a.Vec = config.NewSecret(iv)
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
	key, err := a.Key.Get()
	if err != nil {
		return nil, fmt.Errorf("getting key failed: %w", err)
	}
	defer config.ReleaseSecret(key)

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
	case "cfb":
		cipher.NewCFBDecrypter(block, iv).XORKeyStream(data, data)
	case "ctr":
		cipher.NewCTR(block, iv).XORKeyStream(data, data)
	case "ofb":
		cipher.NewOFB(block, iv).XORKeyStream(data, data)
	default:
		return nil, fmt.Errorf("unsupported block mode %q", a.mode)
	}

	return a.trim(data), nil
}

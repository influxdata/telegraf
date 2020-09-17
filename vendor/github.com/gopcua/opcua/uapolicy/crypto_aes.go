package uapolicy

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/gopcua/opcua/errors"
)

const (
	AESBlockSize  = aes.BlockSize
	AESMinPadding = 0
)

type AES struct {
	KeyLength int
	IV        []byte
	Secret    []byte
}

func (a *AES) Decrypt(src []byte) ([]byte, error) {
	paddedKey := make([]byte, a.KeyLength/8)
	copy(paddedKey, a.Secret)

	block, err := aes.NewCipher(a.Secret)
	if err != nil {
		return nil, err
	}

	if len(src) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	// CBC mode always works in whole blocks.
	if len(src)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	dst := make([]byte, len(src))
	mode := cipher.NewCBCDecrypter(block, a.IV)
	mode.CryptBlocks(dst, src)
	return dst, nil
}

func (a *AES) Encrypt(src []byte) ([]byte, error) {
	paddedKey := make([]byte, a.KeyLength/8)
	copy(paddedKey, a.Secret)

	// CBC mode always works in whole blocks.
	if len(src)%aes.BlockSize != 0 {
		return nil, errors.New("plaintext is not a multiple of the block size")
	}

	block, err := aes.NewCipher(paddedKey)
	if err != nil {
		return nil, err
	}

	dst := make([]byte, len(src))
	mode := cipher.NewCBCEncrypter(block, a.IV)
	mode.CryptBlocks(dst, src)
	return dst, nil
}

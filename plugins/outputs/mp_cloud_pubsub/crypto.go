package mp_cloud_pubsub

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// errMalformedCipherText is returned by decrypt when ciphertext is malformed.
var errMalformedCipherText = errors.New("malformed ciphertext")

// decrypt decrypts data using 256-bit AES-
//
// This both hides the content of the data and provides a check that it hasn't
// been altered. Expects input form of base64 urlencoded nonce|ciphertext|tag
// where '|' indicates concatenation.
func decrypt(cryptoText string, key [32]byte) ([]byte, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(cryptoText)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errMalformedCipherText
	}

	pt, err := gcm.Open(nil, ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():], nil)
	if err != nil {
		return nil, err
	}

	return pt, nil
}

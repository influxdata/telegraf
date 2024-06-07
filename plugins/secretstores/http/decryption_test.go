package http

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateAESFail(t *testing.T) {
	cfg := DecryptionConfig{Cipher: "aes128/CBC/PKCS#5/garbage"}
	decrypt, err := cfg.CreateDecrypter()
	require.ErrorContains(t, err, "init of AES decrypter failed")
	require.Nil(t, decrypt)
}

func TestTrimPKCSFail(t *testing.T) {
	_, err := PKCS5or7Trimming([]byte{})
	require.ErrorContains(t, err, "empty value to trim")

	_, err = PKCS5or7Trimming([]byte{0x00, 0x05})
	require.ErrorContains(t, err, "length 2 shorter than trim value 5")
}

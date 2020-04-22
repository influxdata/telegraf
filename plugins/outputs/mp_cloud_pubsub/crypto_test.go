package mp_cloud_pubsub

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key, err := newEncryptionKey()
	if err != nil {
		t.Fatal(err)
	}

	input := []byte("Hello, world!")

	ciphertext, err := encrypt(string(input), key)
	if err != nil {
		t.Fatal(err)
	}

	output, err := decrypt(ciphertext, key)
	if err != nil {
		t.Fatal(err)
	}

	if string(input) != string(output) {
		t.Fatalf("expected %q got %q", input, output)
	}
}

func TestEncryptDecrypt_badCipher(t *testing.T) {
	key, err := newEncryptionKey()
	if err != nil {
		t.Fatal(err)
	}

	input := []byte("Hello, world!")

	ciphertext, err := encrypt(string(input), key)
	if err != nil {
		t.Fatal(err)
	}

	bt := []byte(ciphertext)
	bt[0] ^= 0xff

	if _, err = decrypt(string(bt), key); err == nil {
		t.Fatalf("expected illegal base 64 error but got none")
	}
}

func newEncryptionKey() ([32]byte, error) {
	key := [32]byte{}
	_, err := io.ReadFull(rand.Reader, key[:])
	return key, err
}

func encrypt(plaintext string, key [32]byte) (string, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ct := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.URLEncoding.EncodeToString(ct), nil
}

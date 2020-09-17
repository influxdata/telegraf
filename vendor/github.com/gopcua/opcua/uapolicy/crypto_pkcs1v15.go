package uapolicy

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"

	// Force compilation of required hashing algorithms, although we don't directly use the packages
	_ "crypto/sha1"
	_ "crypto/sha256"

	"github.com/gopcua/opcua/ua"
)

const PKCS1v15MinPadding = 11

type PKCS1v15 struct {
	Hash       crypto.Hash
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

func (c *PKCS1v15) Decrypt(src []byte) ([]byte, error) {
	if c.PrivateKey == nil {
		return nil, ua.StatusBadSecurityChecksFailed
	}

	rng := rand.Reader

	var plaintext []byte

	blockSize := c.PrivateKey.PublicKey.Size()
	srcRemaining := len(src)
	start := 0

	for srcRemaining > 0 {
		end := start + blockSize
		if end > len(src) {
			end = len(src)
		}

		p, err := rsa.DecryptPKCS1v15(rng, c.PrivateKey, src[start:end])
		if err != nil {
			return nil, err
		}

		plaintext = append(plaintext, p...)
		start = end
		srcRemaining = len(src) - start
	}

	return plaintext, nil
}

func (c *PKCS1v15) Encrypt(src []byte) ([]byte, error) {
	if c.PublicKey == nil {
		return nil, ua.StatusBadSecurityChecksFailed
	}

	rng := rand.Reader

	var ciphertext []byte

	maxBlock := c.PublicKey.Size() - PKCS1v15MinPadding
	srcRemaining := len(src)
	start := 0
	for srcRemaining > 0 {
		end := start + maxBlock
		if end > len(src) {
			end = len(src)
		}

		c, err := rsa.EncryptPKCS1v15(rng, c.PublicKey, src[start:end])
		if err != nil {
			return nil, err
		}

		ciphertext = append(ciphertext, c...)
		start = end
		srcRemaining = len(src) - start
	}

	return ciphertext, nil
}

func (s *PKCS1v15) Signature(msg []byte) ([]byte, error) {
	if s.PrivateKey == nil {
		return nil, ua.StatusBadSecurityChecksFailed
	}

	rng := rand.Reader

	h := s.Hash.New()
	if _, err := h.Write(msg); err != nil {
		return nil, err
	}
	hashed := h.Sum(nil)

	return rsa.SignPKCS1v15(rng, s.PrivateKey, s.Hash, hashed[:])
}

func (s *PKCS1v15) Verify(msg, signature []byte) error {
	if s.PublicKey == nil {
		return ua.StatusBadSecurityChecksFailed
	}

	h := s.Hash.New()
	if _, err := h.Write(msg); err != nil {
		return err
	}
	hashed := h.Sum(nil)
	return rsa.VerifyPKCS1v15(s.PublicKey, s.Hash, hashed[:], signature)
}

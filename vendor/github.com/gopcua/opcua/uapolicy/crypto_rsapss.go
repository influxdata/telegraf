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

type RSAPSS struct {
	Hash       crypto.Hash
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

func (s *RSAPSS) Signature(msg []byte) ([]byte, error) {
	if s.PrivateKey == nil {
		return nil, ua.StatusBadSecurityChecksFailed
	}

	rng := rand.Reader

	h := s.Hash.New()
	if _, err := h.Write(msg); err != nil {
		return nil, err
	}
	hashed := h.Sum(nil)

	return rsa.SignPSS(rng, s.PrivateKey, s.Hash, hashed[:], &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash})
}

func (s *RSAPSS) Verify(msg, signature []byte) error {
	if s.PublicKey == nil {
		return ua.StatusBadSecurityChecksFailed
	}

	h := s.Hash.New()
	if _, err := h.Write(msg); err != nil {
		return err
	}
	hashed := h.Sum(nil)
	return rsa.VerifyPSS(s.PublicKey, s.Hash, hashed[:], signature, nil)
}

package uapolicy

import (
	"crypto"
	"crypto/hmac"

	"github.com/gopcua/opcua/errors"
)

type HMAC struct {
	Hash   crypto.Hash
	Secret []byte
}

func (s *HMAC) Signature(msg []byte) ([]byte, error) {
	h := hmac.New(s.Hash.New, s.Secret)
	if _, err := h.Write(msg); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func (s *HMAC) Verify(msg, signature []byte) error {
	sig, err := s.Signature(msg)
	if err != nil {
		return err
	}
	if !hmac.Equal(sig, signature) {
		return errors.New("signature validation failed")
	}
	return nil
}

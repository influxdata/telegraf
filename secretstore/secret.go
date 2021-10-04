package secretstore

import (
	"bytes"

	"github.com/awnumar/memguard"
)

type Secret struct {
	*memguard.Enclave
}

// UnmarshalTOML creates a secret from a toml value
func (s *Secret) UnmarshalTOML(b []byte) error {
	s.Enclave = memguard.NewEnclave(unquote(b))

	return nil
}

func unquote(b []byte) []byte {
	if bytes.HasPrefix(b, []byte("'''")) && bytes.HasSuffix(b, []byte("'''")) {
		return b[3 : len(b)-3]
	}
	if bytes.HasPrefix(b, []byte("'")) && bytes.HasSuffix(b, []byte("'")) {
		return b[1 : len(b)-1]
	}
	if bytes.HasPrefix(b, []byte("\"\"\"")) && bytes.HasSuffix(b, []byte("\"\"\"")) {
		return b[3 : len(b)-3]
	}
	if bytes.HasPrefix(b, []byte("\"")) && bytes.HasSuffix(b, []byte("\"")) {
		return b[1 : len(b)-1]
	}
	return b
}

package secretstore

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/awnumar/memguard"
)

// secretPattern is a regex to extract references to secrets stored in a secret-store.
var secretPattern = regexp.MustCompile(`@\{(\w+:\w+)\}`)

// NewSecret creates a new secret from the given bytes
func NewSecret(b []byte) (Secret, error) {
	s := Secret{}
	err := s.initialize(b)
	return s, err
}

// Secret safely stores sensitive data such as a password or token
type Secret struct {
	enclave    *memguard.Enclave
	initialzed bool
	resolver   func() (string, error)
}

// staticResolver returns static secrets that do not change over time
func (s *Secret) staticResolver() (string, error) {
	lockbuf, err := s.enclave.Open()
	if err != nil {
		return "", fmt.Errorf("opening enclave failed: %v", err)
	}

	return lockbuf.String(), nil
}

// dynamicResolver returns dynamic secrets that change over time e.g. TOTP
func (s *Secret) dynamicResolver() (string, error) {
	return s.replace(true, false)
}

// UnmarshalTOML creates a secret from a toml value
func (s *Secret) UnmarshalTOML(b []byte) error {
	return s.initialize(b)
}

// Get return the string representation of the secret
func (s *Secret) Get() (string, error) {
	if s.initialzed {
		return s.resolver()
	}
	return "", nil
}

// Destroy the secret content
func (s *Secret) Destroy() {
	if s.enclave == nil {
		return
	}

	// Wipe the secret from memory
	lockbuf, err := s.enclave.Open()
	if err == nil {
		lockbuf.Destroy()
	}

	s.initialzed = false
}

func (s *Secret) initialize(b []byte) error {
	s.enclave = memguard.NewEnclave(unquote(b))
	s.resolver = s.staticResolver
	s.initialzed = true

	// We don't need to know the secret
	_, err := s.replace(false, true)
	return err
}

func (s *Secret) replace(replaceDynamic, save bool) (string, error) {
	lockbuf, err := s.enclave.Open()
	if err != nil {
		return "", fmt.Errorf("opening enclave failed: %v", err)
	}

	replaceErrs := make([]string, 0)
	newsecret := secretPattern.ReplaceAllStringFunc(lockbuf.String(), func(match string) string {
		// There should _ALWAYS_ be two parts due to the regular expression match
		parts := strings.SplitN(match[2:len(match)-1], ":", 2)
		storename := parts[0]
		keyname := parts[1]

		store, found := stores[storename]
		if !found {
			replaceErrs = append(replaceErrs, fmt.Sprintf("unknown store %q for %q", storename, match))
			return match
		}

		// Do not replace secrets from a dynamic store and remember their stores
		if replaceDynamic && store.IsDynamic() {
			s.resolver = s.dynamicResolver
			return match
		}

		// Replace all secrets from static stores
		replacement, err := store.Get(keyname)
		if err != nil {
			replaceErrs = append(replaceErrs, fmt.Sprintf("getting secret %q for %q: %v", keyname, match, err))
			return match
		}
		return replacement
	})
	if len(replaceErrs) > 0 {
		return "", fmt.Errorf("replacing secrets failed: %s", strings.Join(replaceErrs, ";"))
	}

	if save && lockbuf.String() != newsecret {
		s.enclave = memguard.NewEnclave([]byte(newsecret))
		lockbuf.Destroy()
	}

	return newsecret, nil
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

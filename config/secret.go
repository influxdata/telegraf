package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/awnumar/memguard"

	"github.com/influxdata/telegraf"
)

// unlinkedSecrets contains the list of secrets that contain
// references not yet linked to their corresponding secret-store.
// Those secrets must later (after reading the config) be linked
// by the config to their respective secret-stores.
// Secrets containing constant strings will not be found in this
// list.
var unlinkedSecrets = make([]*Secret, 0)

// secretPattern is a regex to extract references to secrets stored
// in a secret-store.
var secretPattern = regexp.MustCompile(`@\{(\w+:\w+)\}`)

// Secret safely stores sensitive data such as a password or token
type Secret struct {
	enclave   *memguard.Enclave
	resolvers map[string]telegraf.ResolveFunc
	// unlinked contains all references in the secret that are not yet
	// linked to the corresponding secret store.
	unlinked []string
}

// NewSecret creates a new secret from the given bytes
func NewSecret(b []byte) Secret {
	s := Secret{}
	s.init(b)
	return s
}

// UnmarshalTOML creates a secret from a toml value.
func (s *Secret) UnmarshalTOML(b []byte) error {
	// Unmarshal raw secret from TOML and put it into protected memory
	s.init(b)

	// Keep track of secrets that contain references to secret-stores
	// for later resolving by the config.
	if len(s.unlinked) > 0 {
		unlinkedSecrets = append(unlinkedSecrets, s)
	}

	return nil
}

// Initialize the secret content
func (s *Secret) init(b []byte) {
	secret := unquoteTomlString(b)

	// Find all parts that need to be resolved and return them
	s.unlinked = secretPattern.FindAllString(string(secret), -1)

	// Setup the enclave
	s.enclave = memguard.NewEnclave(secret)
	s.resolvers = nil
}

// Destroy the secret content
func (s *Secret) Destroy() {
	s.resolvers = nil
	s.unlinked = nil

	if s.enclave == nil {
		return
	}

	// Wipe the secret from memory
	lockbuf, err := s.enclave.Open()
	if err == nil {
		lockbuf.Destroy()
	}
	s.enclave = nil
}

// Get return the string representation of the secret
func (s *Secret) Get() ([]byte, error) {
	if s.enclave == nil {
		return nil, nil
	}

	if len(s.unlinked) > 0 {
		return nil, fmt.Errorf("unlinked parts in secret: %v", strings.Join(s.unlinked, ";"))
	}

	// Decrypt the secret so we can return it
	lockbuf, err := s.enclave.Open()
	if err != nil {
		return nil, fmt.Errorf("opening enclave failed: %v", err)
	}
	defer lockbuf.Destroy()
	secret := lockbuf.Bytes()

	if len(s.resolvers) == 0 {
		// Make a copy as we cannot access lockbuf after Destroy, i.e.
		// after this function finishes.
		newsecret := append([]byte{}, secret...)
		return newsecret, protect(newsecret)
	}

	replaceErrs := make([]string, 0)
	newsecret := secretPattern.ReplaceAllFunc(secret, func(match []byte) []byte {
		resolver, found := s.resolvers[string(match)]
		if !found {
			replaceErrs = append(replaceErrs, fmt.Sprintf("no resolver for %q", match))
			return match
		}
		replacement, _, err := resolver()
		if err != nil {
			replaceErrs = append(replaceErrs, fmt.Sprintf("resolving %q failed: %v", match, err))
			return match
		}

		return replacement
	})
	if len(replaceErrs) > 0 {
		return nil, fmt.Errorf("replacing secrets failed: %s", strings.Join(replaceErrs, ";"))
	}

	return newsecret, protect(newsecret)
}

// GetUnlinked return the parts of the secret that is not yet linked to a resolver
func (s *Secret) GetUnlinked() []string {
	return s.unlinked
}

// Link used the given resolver map to link the secret parts to their
// secret-store resolvers.
func (s *Secret) Link(resolvers map[string]telegraf.ResolveFunc) error {
	// Setup the resolver map
	s.resolvers = make(map[string]telegraf.ResolveFunc)

	// Decrypt the secret so we can return it
	if s.enclave == nil {
		return nil
	}
	lockbuf, err := s.enclave.Open()
	if err != nil {
		return fmt.Errorf("opening enclave failed: %v", err)
	}
	defer lockbuf.Destroy()
	secret := lockbuf.Bytes()

	// Iterate through the parts and try to resolve them. For static parts
	// we directly replace them, while for dynamic ones we store the resolver.
	replaceErrs := make([]string, 0)
	newsecret := secretPattern.ReplaceAllFunc(secret, func(match []byte) []byte {
		resolver, found := resolvers[string(match)]
		if !found {
			replaceErrs = append(replaceErrs, fmt.Sprintf("unlinked part %q", match))
			return match
		}
		replacement, dynamic, err := resolver()
		if err != nil {
			replaceErrs = append(replaceErrs, fmt.Sprintf("resolving %q failed: %v", match, err))
			return match
		}

		// Replace static parts right away
		if !dynamic {
			return replacement
		}

		// Keep the resolver for dynamic secrets
		s.resolvers[string(match)] = resolver
		return match
	})
	if len(replaceErrs) > 0 {
		return fmt.Errorf("linking secrets failed: %s", strings.Join(replaceErrs, ";"))
	}

	// Store the secret if it has changed
	if string(secret) != string(newsecret) {
		s.enclave = memguard.NewEnclave(newsecret)
	}

	// All linked now
	s.unlinked = nil

	return nil
}

func splitLink(s string) (storeid string, key string) {
	// There should _ALWAYS_ be two parts due to the regular expression match
	parts := strings.SplitN(s[2:len(s)-1], ":", 2)
	return parts[0], parts[1]
}

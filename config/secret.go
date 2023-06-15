package config

import (
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"

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

// secretStorePattern is a regex to validate secret-store IDs
var secretStorePattern = regexp.MustCompile(`^\w+$`)

// secretPattern is a regex to extract references to secrets stored
// in a secret-store.
var secretPattern = regexp.MustCompile(`@\{(\w+:\w+)\}`)

var secretCount atomic.Int64

// Secret safely stores sensitive data such as a password or token
type Secret struct {
	enclave   *memguard.Enclave
	resolvers map[string]telegraf.ResolveFunc
	// unlinked contains all references in the secret that are not yet
	// linked to the corresponding secret store.
	unlinked []string

	// Denotes if the secret is completely empty
	notempty bool
}

// NewSecret creates a new secret from the given bytes
func NewSecret(b []byte) Secret {
	s := Secret{}
	s.init(b)
	return s
}

// UnmarshalText creates a secret from a toml value following the "string" rule.
func (s *Secret) UnmarshalText(b []byte) error {
	// Unmarshal secret from TOML and put it into protected memory
	s.init(b)

	// Keep track of secrets that contain references to secret-stores
	// for later resolving by the config.
	if len(s.unlinked) > 0 && s.notempty {
		unlinkedSecrets = append(unlinkedSecrets, s)
	}

	return nil
}

// Initialize the secret content
func (s *Secret) init(secret []byte) {
	// Keep track of the number of secrets...
	secretCount.Add(1)

	// Remember if the secret is completely empty
	s.notempty = len(secret) != 0

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
	s.notempty = false

	if s.enclave == nil {
		return
	}

	// Wipe the secret from memory
	lockbuf, err := s.enclave.Open()
	if err == nil {
		lockbuf.Destroy()
	}
	s.enclave = nil

	// Keep track of the number of secrets...
	secretCount.Add(-1)
}

// Empty return if the secret is completely empty
func (s *Secret) Empty() bool {
	return !s.notempty
}

// EqualTo performs a constant-time comparison of the secret to the given reference
func (s *Secret) EqualTo(ref []byte) (bool, error) {
	if s.enclave == nil {
		return false, nil
	}

	if len(s.unlinked) > 0 {
		return false, fmt.Errorf("unlinked parts in secret: %v", strings.Join(s.unlinked, ";"))
	}

	// Get a locked-buffer of the secret to perform the comparison
	lockbuf, err := s.enclave.Open()
	if err != nil {
		return false, fmt.Errorf("opening enclave failed: %w", err)
	}
	defer lockbuf.Destroy()

	return lockbuf.EqualTo(ref), nil
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
		return nil, fmt.Errorf("opening enclave failed: %w", err)
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
		memguard.WipeBytes(newsecret)
		return nil, fmt.Errorf("replacing secrets failed: %s", strings.Join(replaceErrs, ";"))
	}

	return newsecret, protect(newsecret)
}

// Set overwrites the secret's value with a new one. Please note, the secret
// is not linked again, so only references to secret-stores can be used, e.g. by
// adding more clear-text or reordering secrets.
func (s *Secret) Set(value []byte) error {
	// Link the new value can be resolved
	secret, res, replaceErrs := resolve(value, s.resolvers)
	if len(replaceErrs) > 0 {
		return fmt.Errorf("linking new secrets failed: %s", strings.Join(replaceErrs, ";"))
	}

	// Set the new secret
	s.enclave = memguard.NewEnclave(secret)
	s.resolvers = res
	s.notempty = len(value) > 0

	return nil
}

// GetUnlinked return the parts of the secret that is not yet linked to a resolver
func (s *Secret) GetUnlinked() []string {
	return s.unlinked
}

// Link used the given resolver map to link the secret parts to their
// secret-store resolvers.
func (s *Secret) Link(resolvers map[string]telegraf.ResolveFunc) error {
	// Decrypt the secret so we can return it
	if s.enclave == nil {
		return nil
	}
	lockbuf, err := s.enclave.Open()
	if err != nil {
		return fmt.Errorf("opening enclave failed: %w", err)
	}
	defer lockbuf.Destroy()
	secret := lockbuf.Bytes()

	// Iterate through the parts and try to resolve them. For static parts
	// we directly replace them, while for dynamic ones we store the resolver.
	newsecret, res, replaceErrs := resolve(secret, resolvers)
	if len(replaceErrs) > 0 {
		return fmt.Errorf("linking secrets failed: %s", strings.Join(replaceErrs, ";"))
	}
	s.resolvers = res

	// Store the secret if it has changed
	if string(secret) != string(newsecret) {
		s.enclave = memguard.NewEnclave(newsecret)
	}

	// All linked now
	s.unlinked = nil

	return nil
}

func resolve(secret []byte, resolvers map[string]telegraf.ResolveFunc) ([]byte, map[string]telegraf.ResolveFunc, []string) {
	// Iterate through the parts and try to resolve them. For static parts
	// we directly replace them, while for dynamic ones we store the resolver.
	replaceErrs := make([]string, 0)
	remaining := make(map[string]telegraf.ResolveFunc)
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
		remaining[string(match)] = resolver
		return match
	})
	return newsecret, remaining, replaceErrs
}

func splitLink(s string) (storeid string, key string) {
	// There should _ALWAYS_ be two parts due to the regular expression match
	parts := strings.SplitN(s[2:len(s)-1], ":", 2)
	return parts[0], parts[1]
}

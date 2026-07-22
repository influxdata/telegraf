package telegraf

// SecretStore is an interface defining functions that a secret store plugin must satisfy
type SecretStore interface {
	Initializer
	PluginDescriber

	// Get searches for the given key and return the secret
	Get(key string) ([]byte, error)

	// List lists all known secret keys
	List() ([]string, error)

	// GetResolver returns a function to resolve the given key.
	GetResolver(key string) (ResolveFunc, error)
}

// SecretStoreEditor is an optional interface for secret stores that can create
// or modify secrets. Stores backed by a read-only source do not implement it.
type SecretStoreEditor interface {
	// Set creates or modifies the secret for the given key
	Set(key, value string) error
}

// ResolveFunc is a function to resolve the secret.
// The returned flag indicates if the resolver is static (false), i.e.
// the secret will not change over time, or dynamic (true) to handle
// secrets that change over time (e.g. TOTP).
type ResolveFunc func() ([]byte, bool, error)

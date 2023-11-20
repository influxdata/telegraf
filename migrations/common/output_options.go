package common

type OutputOptions struct {
	// General options
	Alias string `toml:"alias,omitempty"`

	// Filter options
	FilterOptions
}

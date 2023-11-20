package common

type InputOptions struct {
	// General options
	Interval         string            `toml:"interval,omitempty"`
	Precision        string            `toml:"precision,omitempty"`
	CollectionJitter string            `toml:"collection_jitter,omitempty"`
	CollectionOffset string            `toml:"collection_offset,omitempty"`
	NamePrefix       string            `toml:"name_prefix,omitempty"`
	NameSuffix       string            `toml:"name_suffix,omitempty"`
	NameOverride     string            `toml:"name_override,omitempty"`
	Alias            string            `toml:"alias,omitempty"`
	Tags             map[string]string `toml:"tags,omitempty"`

	// Filter options
	FilterOptions
}

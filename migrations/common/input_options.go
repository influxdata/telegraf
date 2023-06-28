package common

type InputOptions struct {
	Interval         string            `toml:"interval"`
	Precision        string            `toml:"precision"`
	CollectionJitter string            `toml:"collection_jitter"`
	CollectionOffset string            `toml:"collection_offset"`
	NamePrefix       string            `toml:"name_prefix"`
	NameSuffix       string            `toml:"name_suffix"`
	NameOverride     string            `toml:"name_override"`
	Alias            string            `toml:"alias"`
	Tags             map[string]string `toml:"tags"`
	FilterOptions
}

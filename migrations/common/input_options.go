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
	NamePass       []string            `toml:"namepass,omitempty"`
	NameDrop       []string            `toml:"namedrop,omitempty"`
	FieldPassOld   []string            `toml:"pass,omitempty"`
	FieldPass      []string            `toml:"fieldpass,omitempty"`
	FieldDropOld   []string            `toml:"drop,omitempty"`
	FieldDrop      []string            `toml:"fielddrop,omitempty"`
	TagPassFilters map[string][]string `toml:"tagpass,omitempty"`
	TagDropFilters map[string][]string `toml:"tagdrop,omitempty"`
	TagExclude     []string            `toml:"tagexclude,omitempty"`
	TagInclude     []string            `toml:"taginclude,omitempty"`
	MetricPass     string              `toml:"metricpass,omitempty"`
}

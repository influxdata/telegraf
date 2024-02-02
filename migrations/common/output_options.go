package common

type OutputOptions struct {
	// General options
	Alias string `toml:"alias,omitempty"`

	// Filter options
	NamePass       []string            `toml:"namepass,omitempty"`
	NameDrop       []string            `toml:"namedrop,omitempty"`
	FieldPassOld   []string            `toml:"pass,omitempty"`
	FieldPass      []string            `toml:"fieldpass,omitempty"`
	FieldInclude   []string            `toml:"fieldinclude,omitempty"`
	FieldDropOld   []string            `toml:"drop,omitempty"`
	FieldDrop      []string            `toml:"fielddrop,omitempty"`
	FieldExclude   []string            `toml:"fieldexclude,omitempty"`
	TagPassFilters map[string][]string `toml:"tagpass,omitempty"`
	TagDropFilters map[string][]string `toml:"tagdrop,omitempty"`
	TagExclude     []string            `toml:"tagexclude,omitempty"`
	TagInclude     []string            `toml:"taginclude,omitempty"`
	MetricPass     string              `toml:"metricpass,omitempty"`
}

func (oo *OutputOptions) Migrate() {
	oo.FieldInclude = append(oo.FieldInclude, oo.FieldPassOld...)
	oo.FieldInclude = append(oo.FieldInclude, oo.FieldPass...)

	oo.FieldPassOld = nil
	oo.FieldPass = nil

	oo.FieldExclude = append(oo.FieldExclude, oo.FieldDropOld...)
	oo.FieldExclude = append(oo.FieldExclude, oo.FieldDrop...)

	oo.FieldDropOld = nil
	oo.FieldDrop = nil
}

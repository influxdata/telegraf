package common

type FilterOptions struct {
	NamePass       []string            `toml:"namepass"`
	NameDrop       []string            `toml:"namedrop"`
	FieldPassOld   []string            `toml:"pass"`
	FieldPass      []string            `toml:"fieldpass"`
	FieldDropOld   []string            `toml:"drop"`
	FieldDrop      []string            `toml:"fielddrop"`
	TagPassFilters map[string][]string `toml:"tagpass"`
	TagDropFilters map[string][]string `toml:"tagdrop"`
	TagExclude     []string            `toml:"tagexclude"`
	TagInclude     []string            `toml:"taginclude"`
	MetricPass     string              `toml:"metricpass"`
}

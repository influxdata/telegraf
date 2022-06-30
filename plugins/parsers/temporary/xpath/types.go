package xpath

// Config definition for backward compatibility ONLY.
// We need this here to avoid cyclic dependencies. However, we need
// to move this to plugins/parsers/xpath once we deprecate parser
// construction via `NewParser()`.
type Config struct {
	MetricQuery  string            `toml:"metric_name"`
	Selection    string            `toml:"metric_selection"`
	Timestamp    string            `toml:"timestamp"`
	TimestampFmt string            `toml:"timestamp_format"`
	Tags         map[string]string `toml:"tags"`
	Fields       map[string]string `toml:"fields"`
	FieldsInt    map[string]string `toml:"fields_int"`

	FieldSelection  string `toml:"field_selection"`
	FieldNameQuery  string `toml:"field_name"`
	FieldValueQuery string `toml:"field_value"`
	FieldNameExpand bool   `toml:"field_name_expansion"`

	TagSelection  string `toml:"tag_selection"`
	TagNameQuery  string `toml:"tag_name"`
	TagValueQuery string `toml:"tag_value"`
	TagNameExpand bool   `toml:"tag_name_expansion"`
}

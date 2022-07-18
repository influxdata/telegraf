package json_v2

// Config definition for backward compatibility ONLY.
// We need this here to avoid cyclic dependencies. However, we need
// to move this to plugins/parsers/json_v2 once we deprecate parser
// construction via `NewParser()`.
type Config struct {
	MeasurementName     string `toml:"measurement_name"`      // OPTIONAL
	MeasurementNamePath string `toml:"measurement_name_path"` // OPTIONAL
	TimestampPath       string `toml:"timestamp_path"`        // OPTIONAL
	TimestampFormat     string `toml:"timestamp_format"`      // OPTIONAL, but REQUIRED when timestamp_path is defined
	TimestampTimezone   string `toml:"timestamp_timezone"`    // OPTIONAL, but REQUIRES timestamp_path

	Fields      []DataSet `toml:"field"`
	Tags        []DataSet `toml:"tag"`
	JSONObjects []Object  `toml:"object"`
}

type DataSet struct {
	Path     string `toml:"path"` // REQUIRED
	Type     string `toml:"type"` // OPTIONAL, can't be set for tags they will always be a string
	Rename   string `toml:"rename"`
	Optional bool   `toml:"optional"` // Will suppress errors if there isn't a match with Path
}

type Object struct {
	Path               string            `toml:"path"`     // REQUIRED
	Optional           bool              `toml:"optional"` // Will suppress errors if there isn't a match with Path
	TimestampKey       string            `toml:"timestamp_key"`
	TimestampFormat    string            `toml:"timestamp_format"`   // OPTIONAL, but REQUIRED when timestamp_path is defined
	TimestampTimezone  string            `toml:"timestamp_timezone"` // OPTIONAL, but REQUIRES timestamp_path
	Renames            map[string]string `toml:"renames"`
	Fields             map[string]string `toml:"fields"`
	Tags               []string          `toml:"tags"`
	IncludedKeys       []string          `toml:"included_keys"`
	ExcludedKeys       []string          `toml:"excluded_keys"`
	DisablePrependKeys bool              `toml:"disable_prepend_keys"`
	FieldPaths         []DataSet         `toml:"field"`
	TagPaths           []DataSet         `toml:"tag"`
}

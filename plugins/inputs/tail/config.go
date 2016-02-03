package tail

import "github.com/influxdata/telegraf/internal/encoding/graphite"

const (
	// DefaultSeparator is the default join character to use when joining multiple
	// measurment parts in a template.
	DefaultSeparator = "."
)

// Config represents the configuration for Graphite endpoints.
type Config struct {
	Files []string

	graphite.InnerConfig
}

// WithDefaults takes the given config and returns a new config with any required
// default values set.
func (c *Config) WithDefaults() *Config {
	d := *c
	if d.Separator == "" {
		d.Separator = DefaultSeparator
	}
	return &d
}

// New Config instance.
func NewConfig(files []string, separator string, tags []string, templates []string) *Config {
	c := &Config{}
	c.Files = files
	c.Separator = separator
	c.Tags = tags
	c.Templates = templates

	return c
}

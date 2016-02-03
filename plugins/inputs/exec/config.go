package exec

import (
	"github.com/influxdata/telegraf/internal/encoding/graphite"
)

const (
	// DefaultSeparator is the default join character to use when joining multiple
	// measurment parts in a template.
	DefaultSeparator = "."
)

// Config represents the configuration for Graphite endpoints.
type Config struct {
	Commands []string
	graphite.InnerConfig
}

// New Config instance.
func NewConfig(commands, tags, templates []string, separator string) *Config {
	c := &Config{}
	c.Commands = commands
	c.Tags = tags
	c.Templates = templates
	c.Separator = separator
	return c
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

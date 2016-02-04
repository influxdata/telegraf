package exec

import (
	"github.com/influxdata/telegraf/internal/encoding/graphite"
)

// Config represents the configuration for Graphite endpoints.
type Config struct {
	Commands []string
	graphite.Config
}

// New Config instance.
func NewConfig(commands, templates []string, separator string) *Config {
	c := &Config{}
	if separator == "" {
		separator = graphite.DefaultSeparator
	}

	c.Commands = commands
	c.Templates = templates
	c.Separator = separator

	return c
}

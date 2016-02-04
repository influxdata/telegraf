package tail

import "github.com/influxdata/telegraf/internal/encoding/graphite"

// Config represents the configuration for Graphite endpoints.
type Config struct {
	Files []string

	graphite.Config
}

// New Config instance.
func NewConfig(files []string, separator string, templates []string) *Config {
	c := &Config{}
	if separator == "" {
		separator = graphite.DefaultSeparator
	}

	c.Files = files
	c.Templates = templates
	c.Separator = separator

	return c
}

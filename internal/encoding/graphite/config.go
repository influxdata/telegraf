package graphite

import (
	"fmt"
	"strings"
)

const (
	// DefaultSeparator is the default join character to use when joining multiple
	// measurment parts in a template.
	DefaultSeparator = "."
)

// Config represents the configuration for Graphite endpoints.
type Config struct {
	Separator string
	Templates []string
}

// Validate validates the config's templates and tags.
func (c *Config) Validate() error {
	if err := c.validateTemplates(); err != nil {
		return err
	}

	return nil
}

func (c *Config) validateTemplates() error {
	// map to keep track of filters we see
	filters := map[string]struct{}{}

	for i, t := range c.Templates {
		parts := strings.Fields(t)
		// Ensure template string is non-empty
		if len(parts) == 0 {
			return fmt.Errorf("missing template at position: %d", i)
		}
		if len(parts) == 1 && parts[0] == "" {
			return fmt.Errorf("missing template at position: %d", i)
		}

		if len(parts) > 3 {
			return fmt.Errorf("invalid template format: '%s'", t)
		}

		template := t
		filter := ""
		tags := ""
		if len(parts) >= 2 {
			// We could have <filter> <template>  or <template> <tags>.  Equals is only allowed in
			// tags section.
			if strings.Contains(parts[1], "=") {
				template = parts[0]
				tags = parts[1]
			} else {
				filter = parts[0]
				template = parts[1]
			}
		}

		if len(parts) == 3 {
			tags = parts[2]
		}

		// Validate the template has one and only one measurement
		if err := c.validateTemplate(template); err != nil {
			return err
		}

		// Prevent duplicate filters in the config
		if _, ok := filters[filter]; ok {
			return fmt.Errorf("duplicate filter '%s' found at position: %d", filter, i)
		}
		filters[filter] = struct{}{}

		if filter != "" {
			// Validate filter expression is valid
			if err := c.validateFilter(filter); err != nil {
				return err
			}
		}

		if tags != "" {
			// Validate tags
			for _, tagStr := range strings.Split(tags, ",") {
				if err := c.validateTag(tagStr); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *Config) validateTemplate(template string) error {
	hasMeasurement := false
	for _, p := range strings.Split(template, ".") {
		if p == "measurement" || p == "measurement*" {
			hasMeasurement = true
		}
	}

	if !hasMeasurement {
		return fmt.Errorf("no measurement in template `%s`", template)
	}

	return nil
}

func (c *Config) validateFilter(filter string) error {
	for _, p := range strings.Split(filter, ".") {
		if p == "" {
			return fmt.Errorf("filter contains blank section: %s", filter)
		}

		if strings.Contains(p, "*") && p != "*" {
			return fmt.Errorf("invalid filter wildcard section: %s", filter)
		}
	}
	return nil
}

func (c *Config) validateTag(keyValue string) error {
	parts := strings.Split(keyValue, "=")
	if len(parts) != 2 {
		return fmt.Errorf("invalid template tags: '%s'", keyValue)
	}

	if parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid template tags: %s'", keyValue)
	}

	return nil
}

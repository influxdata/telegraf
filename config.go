package tivan

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"time"

	"github.com/influxdb/tivan/plugins"
	"github.com/naoina/toml"
	"github.com/naoina/toml/ast"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalTOML(b []byte) error {
	dur, err := time.ParseDuration(string(b[1 : len(b)-1]))
	if err != nil {
		return err
	}

	d.Duration = dur

	return nil
}

type Config struct {
	URL       string
	Username  string
	Password  string
	Database  string
	UserAgent string
	Tags      map[string]string

	plugins map[string]*ast.Table
}

func (c *Config) Plugins() map[string]*ast.Table {
	return c.plugins
}

func (c *Config) Apply(name string, v interface{}) error {
	if tbl, ok := c.plugins[name]; ok {
		return toml.UnmarshalTable(tbl, v)
	}

	return nil
}

func (c *Config) PluginsDeclared() []string {
	var plugins []string

	for name, _ := range c.plugins {
		plugins = append(plugins, name)
	}

	sort.Strings(plugins)

	return plugins
}

func DefaultConfig() *Config {
	return &Config{}
}

var ErrInvalidConfig = errors.New("invalid configuration")

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tbl, err := toml.Parse(data)
	if err != nil {
		return nil, err
	}

	c := &Config{
		plugins: make(map[string]*ast.Table),
	}

	for name, val := range tbl.Fields {
		subtbl, ok := val.(*ast.Table)
		if !ok {
			return nil, ErrInvalidConfig
		}

		if name == "influxdb" {
			err := toml.UnmarshalTable(subtbl, c)
			if err != nil {
				return nil, err
			}
		} else {
			c.plugins[name] = subtbl
		}
	}

	return c, nil
}

func (c *Config) ListTags() string {
	var tags []string

	for k, v := range c.Tags {
		tags = append(tags, fmt.Sprintf("%s=%s", k, v))
	}

	sort.Strings(tags)

	return strings.Join(tags, " ")
}

type hasConfig interface {
	BasicConfig() string
}

type hasDescr interface {
	Description() string
}

var header = `# Tivan configuration

# Tivan is entirely plugin driven. All metrics are gathered from the
# declared plugins.

# Even if a plugin has no configuration, it must be declared in here
# to be active. Declaring a plugin means just specifying the name
# as a section with no variables.

# Use 'tivan -config tivan.toml -test' to see what metrics a config
# file would generate.

# One rule that plugins conform is wherever a connection string
# can be passed, the values '' and 'localhost' are treated specially.
# They indicate to the plugin to use their own builtin configuration to
# connect to the local system.

# Configuration for influxdb server to send metrics to
# [influxdb]
# url = "http://10.20.2.4"
# username = "tivan"
# password = "metricsmetricsmetricsmetrics"
# database = "tivan"
# user_agent = "tivan"
# tags = { "dc": "us-east-1" }

# Tags can also be specified via a normal map, but only one form at a time:

# [influxdb.tags]
# dc = "us-east-1"

# PLUGINS

`

func PrintSampleConfig() {
	fmt.Printf(header)

	var names []string

	for name, _ := range plugins.Plugins {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		creator := plugins.Plugins[name]

		plugin := creator()

		fmt.Printf("# %s\n[%s]\n", plugin.Description(), name)

		var config string

		config = strings.TrimSpace(plugin.SampleConfig())

		if config == "" {
			fmt.Printf("  # no configuration\n\n")
		} else {
			fmt.Printf("\n")
			lines := strings.Split(config, "\n")
			for _, line := range lines {
				fmt.Printf("%s\n", line)
			}

			fmt.Printf("\n")
		}
	}
}

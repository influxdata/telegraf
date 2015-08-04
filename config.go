package telegraf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"time"

	"github.com/influxdb/telegraf/plugins"
	"github.com/naoina/toml"
	"github.com/naoina/toml/ast"
)

// Duration just wraps time.Duration
type Duration struct {
	time.Duration
}

// UnmarshalTOML parses the duration from the TOML config file
func (d *Duration) UnmarshalTOML(b []byte) error {
	dur, err := time.ParseDuration(string(b[1 : len(b)-1]))
	if err != nil {
		return err
	}

	d.Duration = dur

	return nil
}

// Config specifies the URL/user/password for the database that telegraf
// will be logging to, as well as all the plugins that the user has
// specified
type Config struct {
	URL       string
	Username  string
	Password  string
	Database  string
	UserAgent string
	Timeout   Duration
	Tags      map[string]string

	agent   *ast.Table
	plugins map[string]*ast.Table
}

// Plugins returns the configured plugins as a map of name -> plugin toml
func (c *Config) Plugins() map[string]*ast.Table {
	return c.plugins
}

// ConfiguredPlugin containing a name, interval, and drop/pass prefix lists
type ConfiguredPlugin struct {
	Name string

	Drop []string
	Pass []string

	Interval time.Duration
}

// ShouldPass returns true if the metric should pass, false if should drop
func (cp *ConfiguredPlugin) ShouldPass(measurement string) bool {
	if cp.Pass != nil {
		for _, pat := range cp.Pass {
			if strings.HasPrefix(measurement, pat) {
				return true
			}
		}

		return false
	}

	if cp.Drop != nil {
		for _, pat := range cp.Drop {
			if strings.HasPrefix(measurement, pat) {
				return false
			}
		}

		return true
	}

	return true
}

// ApplyAgent loads the toml config into the given interface
func (c *Config) ApplyAgent(v interface{}) error {
	if c.agent != nil {
		return toml.UnmarshalTable(c.agent, v)
	}

	return nil
}

// ApplyPlugin takes defined plugin names and applies them to the given
// interface, returning a ConfiguredPlugin object in the end that can
// be inserted into a runningPlugin by the agent.
func (c *Config) ApplyPlugin(name string, v interface{}) (*ConfiguredPlugin, error) {
	cp := &ConfiguredPlugin{Name: name}

	if tbl, ok := c.plugins[name]; ok {

		if node, ok := tbl.Fields["pass"]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if ary, ok := kv.Value.(*ast.Array); ok {
					for _, elem := range ary.Value {
						if str, ok := elem.(*ast.String); ok {
							cp.Pass = append(cp.Pass, str.Value)
						}
					}
				}
			}
		}

		if node, ok := tbl.Fields["drop"]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if ary, ok := kv.Value.(*ast.Array); ok {
					for _, elem := range ary.Value {
						if str, ok := elem.(*ast.String); ok {
							cp.Drop = append(cp.Drop, str.Value)
						}
					}
				}
			}
		}

		if node, ok := tbl.Fields["interval"]; ok {
			if kv, ok := node.(*ast.KeyValue); ok {
				if str, ok := kv.Value.(*ast.String); ok {
					dur, err := time.ParseDuration(str.Value)
					if err != nil {
						return nil, err
					}

					cp.Interval = dur
				}
			}
		}

		delete(tbl.Fields, "drop")
		delete(tbl.Fields, "pass")
		delete(tbl.Fields, "interval")
		return cp, toml.UnmarshalTable(tbl, v)
	}

	return cp, nil
}

// PluginsDeclared returns the name of all plugins declared in the config.
func (c *Config) PluginsDeclared() []string {
	var plugins []string

	for name := range c.plugins {
		plugins = append(plugins, name)
	}

	sort.Strings(plugins)

	return plugins
}

// DefaultConfig returns an empty default configuration
func DefaultConfig() *Config {
	return &Config{}
}

var errInvalidConfig = errors.New("invalid configuration")

// LoadConfig loads the given config file and returns a *Config pointer
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
			return nil, errInvalidConfig
		}

		switch name {
		case "influxdb":
			err := toml.UnmarshalTable(subtbl, c)
			if err != nil {
				return nil, err
			}
		case "agent":
			c.agent = subtbl
		default:
			c.plugins[name] = subtbl
		}
	}

	return c, nil
}

// ListTags returns a string of tags specified in the config,
// line-protocol style
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

var header = `# Telegraf configuration

# If this file is missing an [agent] section, you must first generate a
# valid config with 'telegraf -sample-config > telegraf.toml'

# Telegraf is entirely plugin driven. All metrics are gathered from the
# declared plugins.

# Even if a plugin has no configuration, it must be declared in here
# to be active. Declaring a plugin means just specifying the name
# as a section with no variables. To deactivate a plugin, comment
# out the name and any variables.

# Use 'telegraf -config telegraf.toml -test' to see what metrics a config
# file would generate.

# One rule that plugins conform to is wherever a connection string
# can be passed, the values '' and 'localhost' are treated specially.
# They indicate to the plugin to use their own builtin configuration to
# connect to the local system.

# NOTE: The configuration has a few required parameters. They are marked
# with 'required'. Be sure to edit those to make this configuration work.

# Configuration for influxdb server to send metrics to
[influxdb]
# The full HTTP endpoint URL for your InfluxDB instance
url = "http://localhost:8086" # required.

# The target database for metrics. This database must already exist
database = "telegraf" # required.

# Connection timeout (for the connection with InfluxDB), formatted as a string.
# Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
# If not provided, will default to 0 (no timeout)
# timeout = "5s"

# username = "telegraf"
# password = "metricsmetricsmetricsmetrics"

# Set the user agent for the POSTs (can be useful for log differentiation)
# user_agent = "telegraf"
# tags = { "dc": "us-east-1" }

# Tags can also be specified via a normal map, but only one form at a time:

# [influxdb.tags]
# dc = "us-east-1"

# Configuration for telegraf itself
# [agent]
# interval = "10s"
# debug = false
# hostname = "prod3241"

# PLUGINS

`

// PrintSampleConfig prints the sample config!
func PrintSampleConfig() {
	fmt.Printf(header)

	var names []string

	for name := range plugins.Plugins {
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

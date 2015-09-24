package telegraf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"time"

	"github.com/influxdb/telegraf/outputs"
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
	Tags map[string]string

	agent   *ast.Table
	plugins map[string]*ast.Table
	outputs map[string]*ast.Table
}

// Plugins returns the configured plugins as a map of name -> plugin toml
func (c *Config) Plugins() map[string]*ast.Table {
	return c.plugins
}

// Outputs returns the configured outputs as a map of name -> output toml
func (c *Config) Outputs() map[string]*ast.Table {
	return c.outputs
}

// TagFilter is the name of a tag, and the values on which to filter
type TagFilter struct {
	Name   string
	Filter []string
}

// ConfiguredPlugin containing a name, interval, and drop/pass prefix lists
// Also lists the tags to filter
type ConfiguredPlugin struct {
	Name string

	Drop []string
	Pass []string

	TagDrop []TagFilter
	TagPass []TagFilter

	Interval time.Duration
}

// ShouldPass returns true if the metric should pass, false if should drop
func (cp *ConfiguredPlugin) ShouldPass(measurement string, tags map[string]string) bool {
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

	if cp.TagPass != nil {
		for _, pat := range cp.TagPass {
			if tagval, ok := tags[pat.Name]; ok {
				for _, filter := range pat.Filter {
					if filter == tagval {
						return true
					}
				}
			}
		}
		return false
	}

	if cp.TagDrop != nil {
		for _, pat := range cp.TagDrop {
			if tagval, ok := tags[pat.Name]; ok {
				for _, filter := range pat.Filter {
					if filter == tagval {
						return false
					}
				}
			}
		}
		return true
	}

	return true
}

// ApplyOutput loads the toml config into the given interface
func (c *Config) ApplyOutput(name string, v interface{}) error {
	if c.outputs[name] != nil {
		return toml.UnmarshalTable(c.outputs[name], v)
	}
	return nil
}

// ApplyAgent loads the toml config into the given Agent object, overriding
// defaults (such as collection duration) with the values from the toml config.
func (c *Config) ApplyAgent(a *Agent) error {
	if c.agent != nil {
		return toml.UnmarshalTable(c.agent, a)
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

		if node, ok := tbl.Fields["tagpass"]; ok {
			if subtbl, ok := node.(*ast.Table); ok {
				for name, val := range subtbl.Fields {
					if kv, ok := val.(*ast.KeyValue); ok {
						tagfilter := &TagFilter{Name: name}
						if ary, ok := kv.Value.(*ast.Array); ok {
							for _, elem := range ary.Value {
								if str, ok := elem.(*ast.String); ok {
									tagfilter.Filter = append(tagfilter.Filter, str.Value)
								}
							}
						}
						cp.TagPass = append(cp.TagPass, *tagfilter)
					}
				}
			}
		}

		if node, ok := tbl.Fields["tagdrop"]; ok {
			if subtbl, ok := node.(*ast.Table); ok {
				for name, val := range subtbl.Fields {
					if kv, ok := val.(*ast.KeyValue); ok {
						tagfilter := &TagFilter{Name: name}
						if ary, ok := kv.Value.(*ast.Array); ok {
							for _, elem := range ary.Value {
								if str, ok := elem.(*ast.String); ok {
									tagfilter.Filter = append(tagfilter.Filter, str.Value)
								}
							}
						}
						cp.TagDrop = append(cp.TagDrop, *tagfilter)
					}
				}
			}
		}

		delete(tbl.Fields, "drop")
		delete(tbl.Fields, "pass")
		delete(tbl.Fields, "interval")
		delete(tbl.Fields, "tagdrop")
		delete(tbl.Fields, "tagpass")
		return cp, toml.UnmarshalTable(tbl, v)
	}

	return cp, nil
}

// PluginsDeclared returns the name of all plugins declared in the config.
func (c *Config) PluginsDeclared() []string {
	return declared(c.plugins)
}

// OutputsDeclared returns the name of all outputs declared in the config.
func (c *Config) OutputsDeclared() []string {
	return declared(c.outputs)
}

func declared(endpoints map[string]*ast.Table) []string {
	var names []string

	for name := range endpoints {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
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
		Tags:    make(map[string]string),
		plugins: make(map[string]*ast.Table),
		outputs: make(map[string]*ast.Table),
	}

	for name, val := range tbl.Fields {
		subtbl, ok := val.(*ast.Table)
		if !ok {
			return nil, errInvalidConfig
		}

		switch name {
		case "agent":
			c.agent = subtbl
		case "tags":
			if err := toml.UnmarshalTable(subtbl, c.Tags); err != nil {
				return nil, errInvalidConfig
			}
		case "outputs":
			for outputName, outputVal := range subtbl.Fields {
				outputSubtbl, ok := outputVal.(*ast.Table)
				if !ok {
					return nil, errInvalidConfig
				}
				c.outputs[outputName] = outputSubtbl
			}
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

# Tags can also be specified via a normal map, but only one form at a time:
[tags]
	# dc = "us-east-1"

# Configuration for telegraf agent
[agent]
	# Default data collection interval for all plugins
	interval = "10s"

	# If utc = false, uses local time (utc is highly recommended)
	utc = true

	# Precision of writes, valid values are n, u, ms, s, m, and h
	# note: using second precision greatly helps InfluxDB compression
	precision = "s"

	# run telegraf in debug mode
	debug = false

	# Override default hostname, if empty use os.Hostname()
	hostname = ""


###############################################################################
#                                  OUTPUTS                                    #
###############################################################################

[outputs]
`

var pluginHeader = `

###############################################################################
#                                  PLUGINS                                    #
###############################################################################
`

var servicePluginHeader = `

###############################################################################
#                              SERVICE PLUGINS                                #
###############################################################################
`

// PrintSampleConfig prints the sample config
func PrintSampleConfig(pluginFilters []string, outputFilters []string) {
	fmt.Printf(header)

	// Filter outputs
	var onames []string
	for oname := range outputs.Outputs {
		if len(outputFilters) == 0 || sliceContains(oname, outputFilters) {
			onames = append(onames, oname)
		}
	}
	sort.Strings(onames)

	// Print Outputs
	for _, oname := range onames {
		creator := outputs.Outputs[oname]
		output := creator()

		fmt.Printf("\n# %s\n[outputs.%s]", output.Description(), oname)

		config := output.SampleConfig()
		if config == "" {
			fmt.Printf("\n	# no configuration\n")
		} else {
			fmt.Printf(config)
		}
	}

	// Filter plugins
	var pnames []string
	for pname := range plugins.Plugins {
		if len(pluginFilters) == 0 || sliceContains(pname, pluginFilters) {
			pnames = append(pnames, pname)
		}
	}
	sort.Strings(pnames)

	// Print Plugins
	fmt.Printf(pluginHeader)
	servPlugins := make(map[string]plugins.ServicePlugin)
	for _, pname := range pnames {
		creator := plugins.Plugins[pname]
		plugin := creator()

		switch p := plugin.(type) {
		case plugins.ServicePlugin:
			servPlugins[pname] = p
			continue
		}

		printConfig(pname, plugin)
	}

	// Print Service Plugins
	fmt.Printf(servicePluginHeader)
	for name, plugin := range servPlugins {
		printConfig(name, plugin)
	}
}

func printConfig(name string, plugin plugins.Plugin) {
	fmt.Printf("\n# %s\n[%s]", plugin.Description(), name)
	config := plugin.SampleConfig()
	if config == "" {
		fmt.Printf("\n	# no configuration\n")
	} else {
		fmt.Printf(config)
	}
}

func sliceContains(name string, list []string) bool {
	for _, b := range list {
		if b == name {
			return true
		}
	}
	return false
}

// PrintPluginConfig prints the config usage of a single plugin.
func PrintPluginConfig(name string) error {
	if creator, ok := plugins.Plugins[name]; ok {
		printConfig(name, creator())
	} else {
		return errors.New(fmt.Sprintf("Plugin %s not found", name))
	}
	return nil
}

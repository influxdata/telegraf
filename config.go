package telegraf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/influxdb/telegraf/outputs"
	"github.com/influxdb/telegraf/plugins"
	"github.com/naoina/toml"
	"github.com/naoina/toml/ast"
)

// Config specifies the URL/user/password for the database that telegraf
// will be logging to, as well as all the plugins that the user has
// specified
type Config struct {
	Tags map[string]string

	agent   *ast.Table
	plugins map[string]*ast.Table
	outputs map[string]*ast.Table
}

func NewConfig() *Config {
	c := &Config{
		Tags:    make(map[string]string),
		plugins: make(map[string]*ast.Table),
		outputs: make(map[string]*ast.Table),
	}
	return c
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
// based on the drop/pass plugin parameters
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

// ShouldTagsPass returns true if the metric should pass, false if should drop
// based on the tagdrop/tagpass plugin parameters
func (cp *ConfiguredPlugin) ShouldTagsPass(tags map[string]string) bool {
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
  # Rounds collection interval to 'interval'
  # ie, if interval="10s" then always collect on :00, :10, :20, etc.
  round_interval = true

  # Default data flushing interval for all outputs. You should not set this below
  # interval. Maximum flush_interval will be flush_interval + flush_jitter
  flush_interval = "10s"
  # Jitter the flush interval by a random amount. This is primarily to avoid
  # large write spikes for users running a large number of telegraf instances.
  # ie, a jitter of 5s and interval 10s means flushes will happen every 10-15s
  flush_jitter = "0s"

  # Run telegraf in debug mode
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

[plugins]
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

		fmt.Printf("\n# %s\n[[outputs.%s]]", output.Description(), oname)

		config := output.SampleConfig()
		if config == "" {
			fmt.Printf("\n  # no configuration\n")
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

type printer interface {
	Description() string
	SampleConfig() string
}

func printConfig(name string, p printer) {
	fmt.Printf("\n# %s\n[[plugins.%s]]", p.Description(), name)
	config := p.SampleConfig()
	if config == "" {
		fmt.Printf("\n  # no configuration\n")
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

// PrintOutputConfig prints the config usage of a single output.
func PrintOutputConfig(name string) error {
	if creator, ok := outputs.Outputs[name]; ok {
		printConfig(name, creator())
	} else {
		return errors.New(fmt.Sprintf("Output %s not found", name))
	}
	return nil
}

func (c *Config) LoadDirectory(path string) error {
	directoryEntries, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range directoryEntries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name[len(name)-5:] != ".conf" {
			continue
		}
		err := c.LoadConfig(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadConfig loads the given config file and returns a *Config pointer
func (c *Config) LoadConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	tbl, err := toml.Parse(data)
	if err != nil {
		return err
	}

	for name, val := range tbl.Fields {
		subTable, ok := val.(*ast.Table)
		if !ok {
			return errors.New("invalid configuration")
		}

		switch name {
		case "agent":
			c.agent = subTable
		case "tags":
			if err = toml.UnmarshalTable(subTable, c.Tags); err != nil {
				log.Printf("Could not parse [tags] config\n")
				return err
			}
		case "outputs":
			for outputName, outputVal := range subTable.Fields {
				switch outputSubTable := outputVal.(type) {
				case *ast.Table:
					c.outputs[outputName] = outputSubTable
				case []*ast.Table:
					for id, t := range outputSubTable {
						nameID := fmt.Sprintf("%s-%d", outputName, id)
						c.outputs[nameID] = t
					}
				default:
					return fmt.Errorf("Unsupported config format: %s",
						outputName)
				}
			}
		case "plugins":
			for pluginName, pluginVal := range subTable.Fields {
				switch pluginSubTable := pluginVal.(type) {
				case *ast.Table:
					c.plugins[pluginName] = pluginSubTable
				case []*ast.Table:
					for id, t := range pluginSubTable {
						nameID := fmt.Sprintf("%s-%d", pluginName, id)
						c.plugins[nameID] = t
					}
				default:
					return fmt.Errorf("Unsupported config format: %s",
						pluginName)
				}
			}
		// Assume it's a plugin for legacy config file support if no other
		// identifiers are present
		default:
			c.plugins[name] = subTable
		}
	}
	return nil
}

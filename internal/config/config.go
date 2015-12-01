package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/influxdb/telegraf/internal"
	"github.com/influxdb/telegraf/outputs"
	"github.com/influxdb/telegraf/plugins"

	"github.com/naoina/toml"
	"github.com/naoina/toml/ast"

	"github.com/influxdb/influxdb/client/v2"
)

// Config specifies the URL/user/password for the database that telegraf
// will be logging to, as well as all the plugins that the user has
// specified
type Config struct {
	Tags          map[string]string
	PluginFilters []string
	OutputFilters []string

	Agent   *AgentConfig
	Plugins []*RunningPlugin
	Outputs []*RunningOutput
}

func NewConfig() *Config {
	c := &Config{
		// Agent defaults:
		Agent: &AgentConfig{
			Interval:      internal.Duration{10 * time.Second},
			RoundInterval: true,
			FlushInterval: internal.Duration{10 * time.Second},
			FlushRetries:  2,
			FlushJitter:   internal.Duration{5 * time.Second},
		},

		Tags:          make(map[string]string),
		Plugins:       make([]*RunningPlugin, 0),
		Outputs:       make([]*RunningOutput, 0),
		PluginFilters: make([]string, 0),
		OutputFilters: make([]string, 0),
	}
	return c
}

type AgentConfig struct {
	// Interval at which to gather information
	Interval internal.Duration

	// RoundInterval rounds collection interval to 'interval'.
	//     ie, if Interval=10s then always collect on :00, :10, :20, etc.
	RoundInterval bool

	// Interval at which to flush data
	FlushInterval internal.Duration

	// FlushRetries is the number of times to retry each data flush
	FlushRetries int

	// FlushJitter tells
	FlushJitter internal.Duration

	// TODO(cam): Remove UTC and Precision parameters, they are no longer
	// valid for the agent config. Leaving them here for now for backwards-
	// compatability
	UTC       bool `toml:"utc"`
	Precision string

	// Option for running in debug mode
	Debug    bool
	Hostname string
}

// TagFilter is the name of a tag, and the values on which to filter
type TagFilter struct {
	Name   string
	Filter []string
}

type RunningOutput struct {
	Name   string
	Output outputs.Output
	Config *OutputConfig
}

type RunningPlugin struct {
	Name   string
	Plugin plugins.Plugin
	Config *PluginConfig
}

// Filter containing drop/pass and tagdrop/tagpass rules
type Filter struct {
	Drop []string
	Pass []string

	TagDrop []TagFilter
	TagPass []TagFilter

	IsActive bool
}

// PluginConfig containing a name, interval, and filter
type PluginConfig struct {
	Name     string
	Filter   Filter
	Interval time.Duration
}

// OutputConfig containing name and filter
type OutputConfig struct {
	Name   string
	Filter Filter
}

// Filter returns filtered slice of client.Points based on whether filters
// are active for this RunningOutput.
func (ro *RunningOutput) FilterPoints(points []*client.Point) []*client.Point {
	if !ro.Config.Filter.IsActive {
		return points
	}

	var filteredPoints []*client.Point
	for i := range points {
		if !ro.Config.Filter.ShouldPass(points[i].Name()) || !ro.Config.Filter.ShouldTagsPass(points[i].Tags()) {
			continue
		}
		filteredPoints = append(filteredPoints, points[i])
	}
	return filteredPoints
}

// ShouldPass returns true if the metric should pass, false if should drop
// based on the drop/pass filter parameters
func (f Filter) ShouldPass(measurement string) bool {
	if f.Pass != nil {
		for _, pat := range f.Pass {
			if strings.HasPrefix(measurement, pat) {
				return true
			}
		}
		return false
	}

	if f.Drop != nil {
		for _, pat := range f.Drop {
			if strings.HasPrefix(measurement, pat) {
				return false
			}
		}

		return true
	}
	return true
}

// ShouldTagsPass returns true if the metric should pass, false if should drop
// based on the tagdrop/tagpass filter parameters
func (f Filter) ShouldTagsPass(tags map[string]string) bool {
	if f.TagPass != nil {
		for _, pat := range f.TagPass {
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

	if f.TagDrop != nil {
		for _, pat := range f.TagDrop {
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

// Plugins returns a list of strings of the configured plugins.
func (c *Config) PluginNames() []string {
	var name []string
	for _, plugin := range c.Plugins {
		name = append(name, plugin.Name)
	}
	return name
}

// Outputs returns a list of strings of the configured plugins.
func (c *Config) OutputNames() []string {
	var name []string
	for _, output := range c.Outputs {
		name = append(name, output.Name)
	}
	return name
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
		printConfig(oname, output, "outputs")
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

		printConfig(pname, plugin, "plugins")
	}

	// Print Service Plugins
	fmt.Printf(servicePluginHeader)
	for name, plugin := range servPlugins {
		printConfig(name, plugin, "plugins")
	}
}

type printer interface {
	Description() string
	SampleConfig() string
}

func printConfig(name string, p printer, op string) {
	fmt.Printf("\n# %s\n[[%s.%s]]", p.Description(), op, name)
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
		printConfig(name, creator(), "plugins")
	} else {
		return errors.New(fmt.Sprintf("Plugin %s not found", name))
	}
	return nil
}

// PrintOutputConfig prints the config usage of a single output.
func PrintOutputConfig(name string) error {
	if creator, ok := outputs.Outputs[name]; ok {
		printConfig(name, creator(), "outputs")
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
		if len(name) < 6 || name[len(name)-5:] != ".conf" {
			continue
		}
		err := c.LoadConfig(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadConfig loads the given config file and applies it to c
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
			if err = toml.UnmarshalTable(subTable, c.Agent); err != nil {
				log.Printf("Could not parse [agent] config\n")
				return err
			}
		case "tags":
			if err = toml.UnmarshalTable(subTable, c.Tags); err != nil {
				log.Printf("Could not parse [tags] config\n")
				return err
			}
		case "outputs":
			for outputName, outputVal := range subTable.Fields {
				switch outputSubTable := outputVal.(type) {
				case *ast.Table:
					if err = c.addOutput(outputName, outputSubTable); err != nil {
						return err
					}
				case []*ast.Table:
					for _, t := range outputSubTable {
						if err = c.addOutput(outputName, t); err != nil {
							return err
						}
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
					if err = c.addPlugin(pluginName, pluginSubTable); err != nil {
						return err
					}
				case []*ast.Table:
					for _, t := range pluginSubTable {
						if err = c.addPlugin(pluginName, t); err != nil {
							return err
						}
					}
				default:
					return fmt.Errorf("Unsupported config format: %s",
						pluginName)
				}
			}
		// Assume it's a plugin for legacy config file support if no other
		// identifiers are present
		default:
			if err = c.addPlugin(name, subTable); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) addOutput(name string, table *ast.Table) error {
	if len(c.OutputFilters) > 0 && !sliceContains(name, c.OutputFilters) {
		return nil
	}
	creator, ok := outputs.Outputs[name]
	if !ok {
		return fmt.Errorf("Undefined but requested output: %s", name)
	}
	output := creator()

	outputConfig, err := buildOutput(name, table)
	if err != nil {
		return err
	}

	if err := toml.UnmarshalTable(table, output); err != nil {
		return err
	}

	ro := &RunningOutput{
		Name:   name,
		Output: output,
		Config: outputConfig,
	}
	c.Outputs = append(c.Outputs, ro)
	return nil
}

func (c *Config) addPlugin(name string, table *ast.Table) error {
	if len(c.PluginFilters) > 0 && !sliceContains(name, c.PluginFilters) {
		return nil
	}
	creator, ok := plugins.Plugins[name]
	if !ok {
		return fmt.Errorf("Undefined but requested plugin: %s", name)
	}
	plugin := creator()

	pluginConfig, err := buildPlugin(name, table)
	if err != nil {
		return err
	}

	if err := toml.UnmarshalTable(table, plugin); err != nil {
		return err
	}

	rp := &RunningPlugin{
		Name:   name,
		Plugin: plugin,
		Config: pluginConfig,
	}
	c.Plugins = append(c.Plugins, rp)
	return nil
}

// buildFilter builds a Filter (tagpass/tagdrop/pass/drop) to
// be inserted into the OutputConfig/PluginConfig to be used for prefix
// filtering on tags and measurements
func buildFilter(tbl *ast.Table) Filter {
	f := Filter{}

	if node, ok := tbl.Fields["pass"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						f.Pass = append(f.Pass, str.Value)
						f.IsActive = true
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
						f.Drop = append(f.Drop, str.Value)
						f.IsActive = true
					}
				}
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
					f.TagPass = append(f.TagPass, *tagfilter)
					f.IsActive = true
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
					f.TagDrop = append(f.TagDrop, *tagfilter)
					f.IsActive = true
				}
			}
		}
	}

	delete(tbl.Fields, "drop")
	delete(tbl.Fields, "pass")
	delete(tbl.Fields, "tagdrop")
	delete(tbl.Fields, "tagpass")
	return f
}

// buildPlugin parses plugin specific items from the ast.Table, builds the filter and returns a
// PluginConfig to be inserted into RunningPlugin
func buildPlugin(name string, tbl *ast.Table) (*PluginConfig, error) {
	cp := &PluginConfig{Name: name}
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
	delete(tbl.Fields, "interval")
	cp.Filter = buildFilter(tbl)
	return cp, nil

}

// buildOutput parses output specific items from the ast.Table, builds the filter and returns an
// OutputConfig to be inserted into RunningPlugin
// Note: error exists in the return for future calls that might require error
func buildOutput(name string, tbl *ast.Table) (*OutputConfig, error) {
	oc := &OutputConfig{
		Name:   name,
		Filter: buildFilter(tbl),
	}
	return oc, nil

}

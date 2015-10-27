package telegraf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
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
	// This lives outside the agent because mergeStruct doesn't need to handle maps normally.
	// We just copy the elements manually in ApplyAgent.
	Tags map[string]string

	agent                *Agent
	plugins              map[string]plugins.Plugin
	pluginConfigurations map[string]*ConfiguredPlugin
	outputs              map[string]outputs.Output

	agentFieldsSet               []string
	pluginFieldsSet              map[string][]string
	pluginConfigurationFieldsSet map[string][]string
	outputFieldsSet              map[string][]string
}

// Plugins returns the configured plugins as a map of name -> plugins.Plugin
func (c *Config) Plugins() map[string]plugins.Plugin {
	return c.plugins
}

// Outputs returns the configured outputs as a map of name -> outputs.Output
func (c *Config) Outputs() map[string]outputs.Output {
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

// ApplyOutput loads the Output struct built from the config into the given Output struct.
// Overrides only values in the given struct that were set in the config.
func (c *Config) ApplyOutput(name string, v interface{}) error {
	if c.outputs[name] != nil {
		return mergeStruct(v, c.outputs[name], c.outputFieldsSet[name])
	}
	return nil
}

// ApplyAgent loads the Agent struct built from the config into the given Agent struct.
// Overrides only values in the given struct that were set in the config.
func (c *Config) ApplyAgent(a *Agent) error {
	if c.agent != nil {
		for key, value := range c.Tags {
			a.Tags[key] = value
		}
		return mergeStruct(a, c.agent, c.agentFieldsSet)
	}

	return nil
}

// ApplyPlugin loads the Plugin struct built from the config into the given Plugin struct.
// Overrides only values in the given struct that were set in the config.
// Additionally return a ConfiguredPlugin, which is always generated from the config.
func (c *Config) ApplyPlugin(name string, v interface{}) (*ConfiguredPlugin, error) {
	if c.plugins[name] != nil {
		err := mergeStruct(v, c.plugins[name], c.pluginFieldsSet[name])
		if err != nil {
			return nil, err
		}
		return c.pluginConfigurations[name], nil
	}

	return nil, nil
}

// Couldn't figure out how to get this to work with the declared function.

// PluginsDeclared returns the name of all plugins declared in the config.
func (c *Config) PluginsDeclared() []string {
	var names []string
	for name := range c.plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// OutputsDeclared returns the name of all outputs declared in the config.
func (c *Config) OutputsDeclared() []string {
	var names []string
	for name := range c.outputs {
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
	fmt.Printf("\n# %s\n[%s]", p.Description(), name)
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

// Used for fuzzy matching struct field names in FieldByNameFunc calls below
func fieldMatch(field string) func(string) bool {
	return func(name string) bool {
		r := strings.NewReplacer("_", "")
		return strings.ToLower(name) == strings.ToLower(r.Replace(field))
	}
}

// A very limited merge. Merges the fields named in the fields parameter, replacing most values, but appending to arrays.
func mergeStruct(base, overlay interface{}, fields []string) error {
	baseValue := reflect.ValueOf(base).Elem()
	overlayValue := reflect.ValueOf(overlay).Elem()
	if baseValue.Kind() != reflect.Struct {
		return fmt.Errorf("Tried to merge something that wasn't a struct: type %v was %v", baseValue.Type(), baseValue.Kind())
	}
	if baseValue.Type() != overlayValue.Type() {
		return fmt.Errorf("Tried to merge two different types: %v and %v", baseValue.Type(), overlayValue.Type())
	}
	for _, field := range fields {
		overlayFieldValue := overlayValue.FieldByNameFunc(fieldMatch(field))
		if !overlayFieldValue.IsValid() {
			return fmt.Errorf("could not find field in %v matching %v", overlayValue.Type(), field)
		}
		if overlayFieldValue.Kind() == reflect.Slice {
			baseFieldValue := baseValue.FieldByNameFunc(fieldMatch(field))
			baseFieldValue.Set(reflect.AppendSlice(baseFieldValue, overlayFieldValue))
		} else {
			baseValue.FieldByNameFunc(fieldMatch(field)).Set(overlayFieldValue)
		}
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
		subConfig, err := LoadConfig(filepath.Join(path, name))
		if err != nil {
			return err
		}
		if subConfig.agent != nil {
			err = mergeStruct(c.agent, subConfig.agent, subConfig.agentFieldsSet)
			if err != nil {
				return err
			}
			for _, field := range subConfig.agentFieldsSet {
				if !sliceContains(field, c.agentFieldsSet) {
					c.agentFieldsSet = append(c.agentFieldsSet, field)
				}
			}
		}
		for pluginName, plugin := range subConfig.plugins {
			if _, ok := c.plugins[pluginName]; !ok {
				c.plugins[pluginName] = plugin
				c.pluginFieldsSet[pluginName] = subConfig.pluginFieldsSet[pluginName]
				c.pluginConfigurations[pluginName] = subConfig.pluginConfigurations[pluginName]
				c.pluginConfigurationFieldsSet[pluginName] = subConfig.pluginConfigurationFieldsSet[pluginName]
				continue
			}
			err = mergeStruct(c.plugins[pluginName], plugin, subConfig.pluginFieldsSet[pluginName])
			if err != nil {
				return err
			}
			for _, field := range subConfig.pluginFieldsSet[pluginName] {
				if !sliceContains(field, c.pluginFieldsSet[pluginName]) {
					c.pluginFieldsSet[pluginName] = append(c.pluginFieldsSet[pluginName], field)
				}
			}
			err = mergeStruct(c.pluginConfigurations[pluginName], subConfig.pluginConfigurations[pluginName], subConfig.pluginConfigurationFieldsSet[pluginName])
			if err != nil {
				return err
			}
			for _, field := range subConfig.pluginConfigurationFieldsSet[pluginName] {
				if !sliceContains(field, c.pluginConfigurationFieldsSet[pluginName]) {
					c.pluginConfigurationFieldsSet[pluginName] = append(c.pluginConfigurationFieldsSet[pluginName], field)
				}
			}
		}
		for outputName, output := range subConfig.outputs {
			if _, ok := c.outputs[outputName]; !ok {
				c.outputs[outputName] = output
				c.outputFieldsSet[outputName] = subConfig.outputFieldsSet[outputName]
				continue
			}
			err = mergeStruct(c.outputs[outputName], output, subConfig.outputFieldsSet[outputName])
			if err != nil {
				return err
			}
			for _, field := range subConfig.outputFieldsSet[outputName] {
				if !sliceContains(field, c.outputFieldsSet[outputName]) {
					c.outputFieldsSet[outputName] = append(c.outputFieldsSet[outputName], field)
				}
			}
		}
	}
	return nil
}

// hazmat area. Keeping the ast parsing here.

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
		Tags:                         make(map[string]string),
		plugins:                      make(map[string]plugins.Plugin),
		pluginConfigurations:         make(map[string]*ConfiguredPlugin),
		outputs:                      make(map[string]outputs.Output),
		pluginFieldsSet:              make(map[string][]string),
		pluginConfigurationFieldsSet: make(map[string][]string),
		outputFieldsSet:              make(map[string][]string),
	}

	for name, val := range tbl.Fields {
		subtbl, ok := val.(*ast.Table)
		if !ok {
			return nil, errors.New("invalid configuration")
		}

		switch name {
		case "agent":
			err := c.parseAgent(subtbl)
			if err != nil {
				return nil, err
			}
		case "tags":
			if err = toml.UnmarshalTable(subtbl, c.Tags); err != nil {
				return nil, err
			}
		case "outputs":
			for outputName, outputVal := range subtbl.Fields {
				outputSubtbl, ok := outputVal.(*ast.Table)
				if !ok {
					return nil, err
				}
				err = c.parseOutput(outputName, outputSubtbl)
				if err != nil {
					return nil, err
				}
			}
		default:
			err = c.parsePlugin(name, subtbl)
			if err != nil {
				return nil, err
			}
		}
	}

	return c, nil
}

// Needs to have the field names, for merging later.
func extractFieldNames(ast *ast.Table) []string {
	// A reasonable capacity?
	var names []string
	for name := range ast.Fields {
		names = append(names, name)
	}
	return names
}

// Parse the agent config out of the given *ast.Table.
func (c *Config) parseAgent(agentAst *ast.Table) error {
	c.agentFieldsSet = extractFieldNames(agentAst)
	agent := &Agent{}
	err := toml.UnmarshalTable(agentAst, agent)
	if err != nil {
		return err
	}
	c.agent = agent
	return nil
}

// Parse an output config out of the given *ast.Table.
func (c *Config) parseOutput(name string, outputAst *ast.Table) error {
	c.outputFieldsSet[name] = extractFieldNames(outputAst)
	creator, ok := outputs.Outputs[name]
	if !ok {
		return fmt.Errorf("Undefined but requested output: %s", name)
	}
	output := creator()
	err := toml.UnmarshalTable(outputAst, output)
	if err != nil {
		return err
	}
	c.outputs[name] = output
	return nil
}

// Parse a plugin config, plus plugin meta-config, out of the given *ast.Table.
func (c *Config) parsePlugin(name string, pluginAst *ast.Table) error {
	creator, ok := plugins.Plugins[name]
	if !ok {
		return fmt.Errorf("Undefined but requested plugin: %s", name)
	}
	plugin := creator()
	cp := &ConfiguredPlugin{Name: name}
	cpFields := make([]string, 0, 5)

	if node, ok := pluginAst.Fields["pass"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						cp.Pass = append(cp.Pass, str.Value)
					}
				}
				cpFields = append(cpFields, "pass")
			}
		}
	}

	if node, ok := pluginAst.Fields["drop"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						cp.Drop = append(cp.Drop, str.Value)
					}
				}
				cpFields = append(cpFields, "drop")
			}
		}
	}

	if node, ok := pluginAst.Fields["interval"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				dur, err := time.ParseDuration(str.Value)
				if err != nil {
					return err
				}

				cp.Interval = dur
				cpFields = append(cpFields, "interval")
			}
		}
	}

	if node, ok := pluginAst.Fields["tagpass"]; ok {
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
					cpFields = append(cpFields, "tagpass")
				}
			}
		}
	}

	if node, ok := pluginAst.Fields["tagdrop"]; ok {
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
					cpFields = append(cpFields, "tagdrop")
				}
			}
		}
	}

	delete(pluginAst.Fields, "drop")
	delete(pluginAst.Fields, "pass")
	delete(pluginAst.Fields, "interval")
	delete(pluginAst.Fields, "tagdrop")
	delete(pluginAst.Fields, "tagpass")
	c.pluginFieldsSet[name] = extractFieldNames(pluginAst)
	c.pluginConfigurationFieldsSet[name] = cpFields
	err := toml.UnmarshalTable(pluginAst, plugin)
	if err != nil {
		return err
	}
	c.plugins[name] = plugin
	c.pluginConfigurations[name] = cp
	return nil
}

package telegraf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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
	// This lives outside the agent because mergeStruct doesn't need to handle
	// maps normally. We just copy the elements manually in ApplyAgent.
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

// Returns a new, empty config object.
func NewConfig() *Config {
	c := &Config{
		Tags:                         make(map[string]string),
		plugins:                      make(map[string]plugins.Plugin),
		pluginConfigurations:         make(map[string]*ConfiguredPlugin),
		outputs:                      make(map[string]outputs.Output),
		pluginFieldsSet:              make(map[string][]string),
		pluginConfigurationFieldsSet: make(map[string][]string),
		outputFieldsSet:              make(map[string][]string),
	}
	return c
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

func (c *Config) GetPluginConfig(name string) *ConfiguredPlugin {
	return c.pluginConfigurations[name]
}

// Couldn't figure out how to get this to work with the declared function.

// PluginsDeclared returns the name of all plugins declared in the config.
func (c *Config) PluginsDeclared() map[string]plugins.Plugin {
	return c.plugins
}

// OutputsDeclared returns the name of all outputs declared in the config.
func (c *Config) OutputsDeclared() map[string]outputs.Output {
	return c.outputs
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

// Find the field with a name matching fieldName, respecting the struct tag and ignoring case and underscores.
// If no field is found, return the zero reflect.Value, which should be checked for with .IsValid().
func findField(fieldName string, value reflect.Value) reflect.Value {
	r := strings.NewReplacer("_", "")
	vType := value.Type()
	for i := 0; i < vType.NumField(); i++ {
		fieldType := vType.Field(i)

		// if we have toml tag, use it
		if tag := fieldType.Tag.Get("toml"); tag != "" {
			if tag == "-" { // omit
				continue
			}
			if tag == fieldName {
				return value.Field(i)
			}
		} else {
			if strings.ToLower(fieldType.Name) == strings.ToLower(r.Replace(fieldName)) {
				return value.Field(i)
			}
		}
	}
	return reflect.Value{}
}

// A very limited merge. Merges the fields named in the fields parameter,
// replacing most values, but appending to arrays.
func mergeStruct(base, overlay interface{}, fields []string) error {
	baseValue := reflect.ValueOf(base).Elem()
	overlayValue := reflect.ValueOf(overlay).Elem()
	if baseValue.Kind() != reflect.Struct {
		return fmt.Errorf("Tried to merge something that wasn't a struct: type %v was %v",
			baseValue.Type(), baseValue.Kind())
	}
	if baseValue.Type() != overlayValue.Type() {
		return fmt.Errorf("Tried to merge two different types: %v and %v",
			baseValue.Type(), overlayValue.Type())
	}
	for _, field := range fields {
		overlayFieldValue := findField(field, overlayValue)
		if !overlayFieldValue.IsValid() {
			return fmt.Errorf("could not find field in %v matching %v",
				overlayValue.Type(), field)
		}
		baseFieldValue := findField(field, baseValue)
		baseFieldValue.Set(overlayFieldValue)
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

// hazmat area. Keeping the ast parsing here.

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
			err := c.parseAgent(subTable)
			if err != nil {
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
					err = c.parseOutput(outputName, outputSubTable, 0)
					if err != nil {
						log.Printf("Could not parse config for output: %s\n",
							outputName)
						return err
					}
				case []*ast.Table:
					for id, t := range outputSubTable {
						err = c.parseOutput(outputName, t, id)
						if err != nil {
							log.Printf("Could not parse config for output: %s\n",
								outputName)
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
					err = c.parsePlugin(pluginName, pluginSubTable, 0)
					if err != nil {
						log.Printf("Could not parse config for plugin: %s\n",
							pluginName)
						return err
					}
				case []*ast.Table:
					for id, t := range pluginSubTable {
						err = c.parsePlugin(pluginName, t, id)
						if err != nil {
							log.Printf("Could not parse config for plugin: %s\n",
								pluginName)
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
			err = c.parsePlugin(name, subTable, 0)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
func (c *Config) parseOutput(name string, outputAst *ast.Table, id int) error {
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
	c.outputs[fmt.Sprintf("%s-%d", name, id)] = output
	return nil
}

// Parse a plugin config, plus plugin meta-config, out of the given *ast.Table.
func (c *Config) parsePlugin(name string, pluginAst *ast.Table, id int) error {
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
	nameID := fmt.Sprintf("%s-%d", name, id)
	c.pluginFieldsSet[nameID] = extractFieldNames(pluginAst)
	c.pluginConfigurationFieldsSet[nameID] = cpFields
	err := toml.UnmarshalTable(pluginAst, plugin)
	if err != nil {
		return err
	}
	c.plugins[nameID] = plugin
	c.pluginConfigurations[nameID] = cp
	return nil
}

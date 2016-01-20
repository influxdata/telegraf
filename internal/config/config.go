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

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"

	"github.com/naoina/toml"
	"github.com/naoina/toml/ast"

	"github.com/influxdata/influxdb/client/v2"
)

// Config specifies the URL/user/password for the database that telegraf
// will be logging to, as well as all the plugins that the user has
// specified
type Config struct {
	Tags          map[string]string
	InputFilters  []string
	OutputFilters []string

	Agent   *AgentConfig
	Inputs  []*RunningInput
	Outputs []*RunningOutput
}

func NewConfig() *Config {
	c := &Config{
		// Agent defaults:
		Agent: &AgentConfig{
			Interval:      internal.Duration{Duration: 10 * time.Second},
			RoundInterval: true,
			FlushInterval: internal.Duration{Duration: 10 * time.Second},
			FlushRetries:  2,
			FlushJitter:   internal.Duration{Duration: 5 * time.Second},
		},

		Tags:          make(map[string]string),
		Inputs:        make([]*RunningInput, 0),
		Outputs:       make([]*RunningOutput, 0),
		InputFilters:  make([]string, 0),
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

	// CollectionJitter is used to jitter the collection by a random amount.
	// Each plugin will sleep for a random time within jitter before collecting.
	// This can be used to avoid many plugins querying things like sysfs at the
	// same time, which can have a measurable effect on the system.
	CollectionJitter internal.Duration

	// Interval at which to flush data
	FlushInterval internal.Duration

	// FlushRetries is the number of times to retry each data flush
	FlushRetries int

	// FlushJitter Jitters the flush interval by a random amount.
	// This is primarily to avoid large write spikes for users running a large
	// number of telegraf instances.
	// ie, a jitter of 5s and interval 10s means flushes will happen every 10-15s
	FlushJitter internal.Duration

	// TODO(cam): Remove UTC and Precision parameters, they are no longer
	// valid for the agent config. Leaving them here for now for backwards-
	// compatability
	UTC       bool `toml:"utc"`
	Precision string

	// Debug is the option for running in debug mode
	Debug bool

	// Quiet is the option for running in quiet mode
	Quiet    bool
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

type RunningInput struct {
	Name   string
	Input  inputs.Input
	Config *InputConfig
}

// Filter containing drop/pass and tagdrop/tagpass rules
type Filter struct {
	Drop []string
	Pass []string

	TagDrop []TagFilter
	TagPass []TagFilter

	IsActive bool
}

// InputConfig containing a name, interval, and filter
type InputConfig struct {
	Name              string
	NameOverride      string
	MeasurementPrefix string
	MeasurementSuffix string
	Tags              map[string]string
	Filter            Filter
	Interval          time.Duration
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
func (f Filter) ShouldPass(fieldkey string) bool {
	if f.Pass != nil {
		for _, pat := range f.Pass {
			// TODO remove HasPrefix check, leaving it for now for legacy support.
			// Cam, 2015-12-07
			if strings.HasPrefix(fieldkey, pat) || internal.Glob(pat, fieldkey) {
				return true
			}
		}
		return false
	}

	if f.Drop != nil {
		for _, pat := range f.Drop {
			// TODO remove HasPrefix check, leaving it for now for legacy support.
			// Cam, 2015-12-07
			if strings.HasPrefix(fieldkey, pat) || internal.Glob(pat, fieldkey) {
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
					if internal.Glob(filter, tagval) {
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
					if internal.Glob(filter, tagval) {
						return false
					}
				}
			}
		}
		return true
	}

	return true
}

// Inputs returns a list of strings of the configured inputs.
func (c *Config) InputNames() []string {
	var name []string
	for _, input := range c.Inputs {
		name = append(name, input.Name)
	}
	return name
}

// Outputs returns a list of strings of the configured inputs.
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
# declared inputs.

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
  # Default data collection interval for all inputs
  interval = "10s"
  # Rounds collection interval to 'interval'
  # ie, if interval="10s" then always collect on :00, :10, :20, etc.
  round_interval = true
  # Collection jitter is used to jitter the collection by a random amount.
  # Each plugin will sleep for a random time within jitter before collecting.
  # This can be used to avoid many plugins querying things like sysfs at the
  # same time, which can have a measurable effect on the system.
  collection_jitter = "0s"

  # Default data flushing interval for all outputs. You should not set this below
  # interval. Maximum flush_interval will be flush_interval + flush_jitter
  flush_interval = "10s"
  # Jitter the flush interval by a random amount. This is primarily to avoid
  # large write spikes for users running a large number of telegraf instances.
  # ie, a jitter of 5s and interval 10s means flushes will happen every 10-15s
  flush_jitter = "0s"

  # Run telegraf in debug mode
  debug = false
  # Run telegraf in quiet mode
  quiet = false
  # Override default hostname, if empty use os.Hostname()
  hostname = ""


###############################################################################
#                                  OUTPUTS                                    #
###############################################################################

`

var pluginHeader = `

###############################################################################
#                                  INPUTS                                     #
###############################################################################

`

var serviceInputHeader = `

###############################################################################
#                              SERVICE INPUTS                                 #
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

	// Filter inputs
	var pnames []string
	for pname := range inputs.Inputs {
		if len(pluginFilters) == 0 || sliceContains(pname, pluginFilters) {
			pnames = append(pnames, pname)
		}
	}
	sort.Strings(pnames)

	// Print Inputs
	fmt.Printf(pluginHeader)
	servInputs := make(map[string]inputs.ServiceInput)
	for _, pname := range pnames {
		creator := inputs.Inputs[pname]
		input := creator()

		switch p := input.(type) {
		case inputs.ServiceInput:
			servInputs[pname] = p
			continue
		}

		printConfig(pname, input, "inputs")
	}

	// Print Service Inputs
	fmt.Printf(serviceInputHeader)
	for name, input := range servInputs {
		printConfig(name, input, "inputs")
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

// PrintInputConfig prints the config usage of a single input.
func PrintInputConfig(name string) error {
	if creator, ok := inputs.Inputs[name]; ok {
		printConfig(name, creator(), "inputs")
	} else {
		return errors.New(fmt.Sprintf("Input %s not found", name))
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
			for pluginName, pluginVal := range subTable.Fields {
				switch pluginSubTable := pluginVal.(type) {
				case *ast.Table:
					if err = c.addOutput(pluginName, pluginSubTable); err != nil {
						return err
					}
				case []*ast.Table:
					for _, t := range pluginSubTable {
						if err = c.addOutput(pluginName, t); err != nil {
							return err
						}
					}
				default:
					return fmt.Errorf("Unsupported config format: %s",
						pluginName)
				}
			}
		case "inputs", "plugins":
			for pluginName, pluginVal := range subTable.Fields {
				switch pluginSubTable := pluginVal.(type) {
				case *ast.Table:
					if err = c.addInput(pluginName, pluginSubTable); err != nil {
						return err
					}
				case []*ast.Table:
					for _, t := range pluginSubTable {
						if err = c.addInput(pluginName, t); err != nil {
							return err
						}
					}
				default:
					return fmt.Errorf("Unsupported config format: %s",
						pluginName)
				}
			}
		// Assume it's an input input for legacy config file support if no other
		// identifiers are present
		default:
			if err = c.addInput(name, subTable); err != nil {
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

func (c *Config) addInput(name string, table *ast.Table) error {
	if len(c.InputFilters) > 0 && !sliceContains(name, c.InputFilters) {
		return nil
	}
	// Legacy support renaming io input to diskio
	if name == "io" {
		name = "diskio"
	}

	creator, ok := inputs.Inputs[name]
	if !ok {
		return fmt.Errorf("Undefined but requested input: %s", name)
	}
	input := creator()

	pluginConfig, err := buildInput(name, table)
	if err != nil {
		return err
	}

	if err := toml.UnmarshalTable(table, input); err != nil {
		return err
	}

	rp := &RunningInput{
		Name:   name,
		Input:  input,
		Config: pluginConfig,
	}
	c.Inputs = append(c.Inputs, rp)
	return nil
}

// buildFilter builds a Filter (tagpass/tagdrop/pass/drop) to
// be inserted into the OutputConfig/InputConfig to be used for prefix
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

// buildInput parses input specific items from the ast.Table,
// builds the filter and returns a
// InputConfig to be inserted into RunningInput
func buildInput(name string, tbl *ast.Table) (*InputConfig, error) {
	cp := &InputConfig{Name: name}
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

	if node, ok := tbl.Fields["name_prefix"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				cp.MeasurementPrefix = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["name_suffix"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				cp.MeasurementSuffix = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["name_override"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				cp.NameOverride = str.Value
			}
		}
	}

	cp.Tags = make(map[string]string)
	if node, ok := tbl.Fields["tags"]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			if err := toml.UnmarshalTable(subtbl, cp.Tags); err != nil {
				log.Printf("Could not parse tags for input %s\n", name)
			}
		}
	}

	delete(tbl.Fields, "name_prefix")
	delete(tbl.Fields, "name_suffix")
	delete(tbl.Fields, "name_override")
	delete(tbl.Fields, "interval")
	delete(tbl.Fields, "tags")
	cp.Filter = buildFilter(tbl)
	return cp, nil
}

// buildOutput parses output specific items from the ast.Table, builds the filter and returns an
// OutputConfig to be inserted into RunningInput
// Note: error exists in the return for future calls that might require error
func buildOutput(name string, tbl *ast.Table) (*OutputConfig, error) {
	oc := &OutputConfig{
		Name:   name,
		Filter: buildFilter(tbl),
	}
	return oc, nil
}

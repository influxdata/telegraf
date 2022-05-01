package config

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/json_v2"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

var (
	// Default sections
	sectionDefaults = []string{"global_tags", "agent", "outputs",
		"processors", "aggregators", "inputs"}

	// Default input plugins
	inputDefaults = []string{"cpu", "mem", "swap", "system", "kernel",
		"processes", "disk", "diskio"}

	// Default output plugins
	outputDefaults = []string{"influxdb"}

	// envVarRe is a regex to find environment variables in the config file
	envVarRe = regexp.MustCompile(`\$\{(\w+)\}|\$(\w+)`)

	envVarEscaper = strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	)
	httpLoadConfigRetryInterval = 10 * time.Second

	// fetchURLRe is a regex to determine whether the requested file should
	// be fetched from a remote or read from the filesystem.
	fetchURLRe = regexp.MustCompile(`^\w+://`)
)

// Config specifies the URL/user/password for the database that telegraf
// will be logging to, as well as all the plugins that the user has
// specified
type Config struct {
	toml         *toml.Config
	errs         []error // config load errors.
	UnusedFields map[string]bool

	Tags          map[string]string
	InputFilters  []string
	OutputFilters []string

	Agent       *AgentConfig
	Inputs      []*models.RunningInput
	Outputs     []*models.RunningOutput
	Aggregators []*models.RunningAggregator
	Parsers     []*models.RunningParser
	// Processors have a slice wrapper type because they need to be sorted
	Processors    models.RunningProcessors
	AggProcessors models.RunningProcessors

	Deprecations map[string][]int64
	version      *semver.Version
}

// NewConfig creates a new struct to hold the Telegraf config.
// For historical reasons, It holds the actual instances of the running plugins
// once the configuration is parsed.
func NewConfig() *Config {
	c := &Config{
		UnusedFields: map[string]bool{},

		// Agent defaults:
		Agent: &AgentConfig{
			Interval:                   Duration(10 * time.Second),
			RoundInterval:              true,
			FlushInterval:              Duration(10 * time.Second),
			LogTarget:                  "file",
			LogfileRotationMaxArchives: 5,
		},

		Tags:          make(map[string]string),
		Inputs:        make([]*models.RunningInput, 0),
		Outputs:       make([]*models.RunningOutput, 0),
		Parsers:       make([]*models.RunningParser, 0),
		Processors:    make([]*models.RunningProcessor, 0),
		AggProcessors: make([]*models.RunningProcessor, 0),
		InputFilters:  make([]string, 0),
		OutputFilters: make([]string, 0),
		Deprecations:  make(map[string][]int64),
	}

	// Handle unknown version
	version := internal.Version()
	if version == "" || version == "unknown" {
		version = "0.0.0-unknown"
	}
	c.version = semver.New(version)

	tomlCfg := &toml.Config{
		NormFieldName: toml.DefaultConfig.NormFieldName,
		FieldToKey:    toml.DefaultConfig.FieldToKey,
		MissingField:  c.missingTomlField,
	}
	c.toml = tomlCfg

	return c
}

// AgentConfig defines configuration that will be used by the Telegraf agent
type AgentConfig struct {
	// Interval at which to gather information
	Interval Duration

	// RoundInterval rounds collection interval to 'interval'.
	//     ie, if Interval=10s then always collect on :00, :10, :20, etc.
	RoundInterval bool

	// Collected metrics are rounded to the precision specified. Precision is
	// specified as an interval with an integer + unit (e.g. 0s, 10ms, 2us, 4s).
	// Valid time units are "ns", "us" (or "µs"), "ms", "s".
	//
	// By default or when set to "0s", precision will be set to the same
	// timestamp order as the collection interval, with the maximum being 1s:
	//   ie, when interval = "10s", precision will be "1s"
	//       when interval = "250ms", precision will be "1ms"
	//
	// Precision will NOT be used for service inputs. It is up to each individual
	// service input to set the timestamp at the appropriate precision.
	Precision Duration

	// CollectionJitter is used to jitter the collection by a random amount.
	// Each plugin will sleep for a random time within jitter before collecting.
	// This can be used to avoid many plugins querying things like sysfs at the
	// same time, which can have a measurable effect on the system.
	CollectionJitter Duration

	// CollectionOffset is used to shift the collection by the given amount.
	// This can be be used to avoid many plugins querying constraint devices
	// at the same time by manually scheduling them in time.
	CollectionOffset Duration

	// FlushInterval is the Interval at which to flush data
	FlushInterval Duration

	// FlushJitter Jitters the flush interval by a random amount.
	// This is primarily to avoid large write spikes for users running a large
	// number of telegraf instances.
	// ie, a jitter of 5s and interval 10s means flushes will happen every 10-15s
	FlushJitter Duration

	// MetricBatchSize is the maximum number of metrics that is wrote to an
	// output plugin in one call.
	MetricBatchSize int

	// MetricBufferLimit is the max number of metrics that each output plugin
	// will cache. The buffer is cleared when a successful write occurs. When
	// full, the oldest metrics will be overwritten. This number should be a
	// multiple of MetricBatchSize. Due to current implementation, this could
	// not be less than 2 times MetricBatchSize.
	MetricBufferLimit int

	// FlushBufferWhenFull tells Telegraf to flush the metric buffer whenever
	// it fills up, regardless of FlushInterval. Setting this option to true
	// does _not_ deactivate FlushInterval.
	FlushBufferWhenFull bool `toml:"flush_buffer_when_full" deprecated:"0.13.0;2.0.0;option is ignored"`

	// TODO(cam): Remove UTC and parameter, they are no longer
	// valid for the agent config. Leaving them here for now for backwards-
	// compatibility
	UTC bool `toml:"utc" deprecated:"1.0.0;option is ignored"`

	// Debug is the option for running in debug mode
	Debug bool `toml:"debug"`

	// Quiet is the option for running in quiet mode
	Quiet bool `toml:"quiet"`

	// Log target controls the destination for logs and can be one of "file",
	// "stderr" or, on Windows, "eventlog".  When set to "file", the output file
	// is determined by the "logfile" setting.
	LogTarget string `toml:"logtarget"`

	// Name of the file to be logged to when using the "file" logtarget.  If set to
	// the empty string then logs are written to stderr.
	Logfile string `toml:"logfile"`

	// The file will be rotated after the time interval specified.  When set
	// to 0 no time based rotation is performed.
	LogfileRotationInterval Duration `toml:"logfile_rotation_interval"`

	// The logfile will be rotated when it becomes larger than the specified
	// size.  When set to 0 no size based rotation is performed.
	LogfileRotationMaxSize Size `toml:"logfile_rotation_max_size"`

	// Maximum number of rotated archives to keep, any older logs are deleted.
	// If set to -1, no archives are removed.
	LogfileRotationMaxArchives int `toml:"logfile_rotation_max_archives"`

	// Pick a timezone to use when logging or type 'local' for local time.
	LogWithTimezone string `toml:"log_with_timezone"`

	Hostname     string
	OmitHostname bool

	// Method for translating SNMP objects. 'netsnmp' to call external programs,
	// 'gosmi' to use the built-in library.
	SnmpTranslator string `toml:"snmp_translator"`
}

// InputNames returns a list of strings of the configured inputs.
func (c *Config) InputNames() []string {
	var name []string
	for _, input := range c.Inputs {
		name = append(name, input.Config.Name)
	}
	return PluginNameCounts(name)
}

// AggregatorNames returns a list of strings of the configured aggregators.
func (c *Config) AggregatorNames() []string {
	var name []string
	for _, aggregator := range c.Aggregators {
		name = append(name, aggregator.Config.Name)
	}
	return PluginNameCounts(name)
}

// ParserNames returns a list of strings of the configured parsers.
func (c *Config) ParserNames() []string {
	var name []string
	for _, parser := range c.Parsers {
		name = append(name, parser.Config.DataFormat)
	}
	return PluginNameCounts(name)
}

// ProcessorNames returns a list of strings of the configured processors.
func (c *Config) ProcessorNames() []string {
	var name []string
	for _, processor := range c.Processors {
		name = append(name, processor.Config.Name)
	}
	return PluginNameCounts(name)
}

// OutputNames returns a list of strings of the configured outputs.
func (c *Config) OutputNames() []string {
	var name []string
	for _, output := range c.Outputs {
		name = append(name, output.Config.Name)
	}
	return PluginNameCounts(name)
}

// PluginNameCounts returns a list of sorted plugin names and their count
func PluginNameCounts(plugins []string) []string {
	names := make(map[string]int)
	for _, plugin := range plugins {
		names[plugin]++
	}

	var namecount []string
	for name, count := range names {
		if count == 1 {
			namecount = append(namecount, name)
		} else {
			namecount = append(namecount, fmt.Sprintf("%s (%dx)", name, count))
		}
	}

	sort.Strings(namecount)
	return namecount
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

var header = `# Telegraf Configuration
#
# Telegraf is entirely plugin driven. All metrics are gathered from the
# declared inputs, and sent to the declared outputs.
#
# Plugins must be declared in here to be active.
# To deactivate a plugin, comment out the name and any variables.
#
# Use 'telegraf -config telegraf.conf -test' to see what metrics a config
# file would generate.
#
# Environment variables can be used anywhere in this config file, simply surround
# them with ${}. For strings the variable must be within quotes (ie, "${STR_VAR}"),
# for numbers and booleans they should be plain (ie, ${INT_VAR}, ${BOOL_VAR})

`
var globalTagsConfig = `
# Global tags can be specified here in key="value" format.
[global_tags]
  # dc = "us-east-1" # will tag all metrics with dc=us-east-1
  # rack = "1a"
  ## Environment variables can be used as tags, and throughout the config file
  # user = "$USER"

`

var agentConfig = `
# Configuration for telegraf agent
[agent]
  ## Default data collection interval for all inputs
  interval = "10s"
  ## Rounds collection interval to 'interval'
  ## ie, if interval="10s" then always collect on :00, :10, :20, etc.
  round_interval = true

  ## Telegraf will send metrics to outputs in batches of at most
  ## metric_batch_size metrics.
  ## This controls the size of writes that Telegraf sends to output plugins.
  metric_batch_size = 1000

  ## Maximum number of unwritten metrics per output.  Increasing this value
  ## allows for longer periods of output downtime without dropping metrics at the
  ## cost of higher maximum memory usage.
  metric_buffer_limit = 10000

  ## Collection jitter is used to jitter the collection by a random amount.
  ## Each plugin will sleep for a random time within jitter before collecting.
  ## This can be used to avoid many plugins querying things like sysfs at the
  ## same time, which can have a measurable effect on the system.
  collection_jitter = "0s"

  ## Collection offset is used to shift the collection by the given amount.
  ## This can be be used to avoid many plugins querying constraint devices
  ## at the same time by manually scheduling them in time.
  # collection_offset = "0s"

  ## Default flushing interval for all outputs. Maximum flush_interval will be
  ## flush_interval + flush_jitter
  flush_interval = "10s"
  ## Jitter the flush interval by a random amount. This is primarily to avoid
  ## large write spikes for users running a large number of telegraf instances.
  ## ie, a jitter of 5s and interval 10s means flushes will happen every 10-15s
  flush_jitter = "0s"

  ## Collected metrics are rounded to the precision specified. Precision is
  ## specified as an interval with an integer + unit (e.g. 0s, 10ms, 2us, 4s).
  ## Valid time units are "ns", "us" (or "µs"), "ms", "s".
  ##
  ## By default or when set to "0s", precision will be set to the same
  ## timestamp order as the collection interval, with the maximum being 1s:
  ##   ie, when interval = "10s", precision will be "1s"
  ##       when interval = "250ms", precision will be "1ms"
  ##
  ## Precision will NOT be used for service inputs. It is up to each individual
  ## service input to set the timestamp at the appropriate precision.
  precision = "0s"

  ## Log at debug level.
  # debug = false
  ## Log only error level messages.
  # quiet = false

  ## Log target controls the destination for logs and can be one of "file",
  ## "stderr" or, on Windows, "eventlog".  When set to "file", the output file
  ## is determined by the "logfile" setting.
  # logtarget = "file"

  ## Name of the file to be logged to when using the "file" logtarget.  If set to
  ## the empty string then logs are written to stderr.
  # logfile = ""

  ## The logfile will be rotated after the time interval specified.  When set
  ## to 0 no time based rotation is performed.  Logs are rotated only when
  ## written to, if there is no log activity rotation may be delayed.
  # logfile_rotation_interval = "0h"

  ## The logfile will be rotated when it becomes larger than the specified
  ## size.  When set to 0 no size based rotation is performed.
  # logfile_rotation_max_size = "0MB"

  ## Maximum number of rotated archives to keep, any older logs are deleted.
  ## If set to -1, no archives are removed.
  # logfile_rotation_max_archives = 5

  ## Pick a timezone to use when logging or type 'local' for local time.
  ## Example: America/Chicago
  # log_with_timezone = ""

  ## Override default hostname, if empty use os.Hostname()
  hostname = ""
  ## If set to true, do no set the "host" tag in the telegraf agent.
  omit_hostname = false

  ## Method of translating SNMP objects. Can be "netsnmp" which
  ## translates by calling external programs snmptranslate and snmptable,
  ## or "gosmi" which translates using the built-in gosmi library.
  # snmp_translator = "netsnmp"
`

var outputHeader = `
###############################################################################
#                            OUTPUT PLUGINS                                   #
###############################################################################

`

var processorHeader = `
###############################################################################
#                            PROCESSOR PLUGINS                                #
###############################################################################

`

var aggregatorHeader = `
###############################################################################
#                            AGGREGATOR PLUGINS                               #
###############################################################################

`

var inputHeader = `
###############################################################################
#                            INPUT PLUGINS                                    #
###############################################################################

`

var serviceInputHeader = `
###############################################################################
#                            SERVICE INPUT PLUGINS                            #
###############################################################################

`

// PrintSampleConfig prints the sample config
func PrintSampleConfig(
	sectionFilters []string,
	inputFilters []string,
	outputFilters []string,
	aggregatorFilters []string,
	processorFilters []string,
) {
	// print headers
	fmt.Print(header)

	if len(sectionFilters) == 0 {
		sectionFilters = sectionDefaults
	}
	printFilteredGlobalSections(sectionFilters)

	// print output plugins
	if sliceContains("outputs", sectionFilters) {
		if len(outputFilters) != 0 {
			if len(outputFilters) >= 3 && outputFilters[1] != "none" {
				fmt.Print(outputHeader)
			}
			printFilteredOutputs(outputFilters, false)
		} else {
			fmt.Print(outputHeader)
			printFilteredOutputs(outputDefaults, false)
			// Print non-default outputs, commented
			var pnames []string
			for pname := range outputs.Outputs {
				if !sliceContains(pname, outputDefaults) {
					pnames = append(pnames, pname)
				}
			}
			sort.Strings(pnames)
			printFilteredOutputs(pnames, true)
		}
	}

	// print processor plugins
	if sliceContains("processors", sectionFilters) {
		if len(processorFilters) != 0 {
			if len(processorFilters) >= 3 && processorFilters[1] != "none" {
				fmt.Print(processorHeader)
			}
			printFilteredProcessors(processorFilters, false)
		} else {
			fmt.Print(processorHeader)
			pnames := []string{}
			for pname := range processors.Processors {
				pnames = append(pnames, pname)
			}
			sort.Strings(pnames)
			printFilteredProcessors(pnames, true)
		}
	}

	// print aggregator plugins
	if sliceContains("aggregators", sectionFilters) {
		if len(aggregatorFilters) != 0 {
			if len(aggregatorFilters) >= 3 && aggregatorFilters[1] != "none" {
				fmt.Print(aggregatorHeader)
			}
			printFilteredAggregators(aggregatorFilters, false)
		} else {
			fmt.Print(aggregatorHeader)
			pnames := []string{}
			for pname := range aggregators.Aggregators {
				pnames = append(pnames, pname)
			}
			sort.Strings(pnames)
			printFilteredAggregators(pnames, true)
		}
	}

	// print input plugins
	if sliceContains("inputs", sectionFilters) {
		if len(inputFilters) != 0 {
			if len(inputFilters) >= 3 && inputFilters[1] != "none" {
				fmt.Print(inputHeader)
			}
			printFilteredInputs(inputFilters, false)
		} else {
			fmt.Print(inputHeader)
			printFilteredInputs(inputDefaults, false)
			// Print non-default inputs, commented
			var pnames []string
			for pname := range inputs.Inputs {
				if !sliceContains(pname, inputDefaults) {
					pnames = append(pnames, pname)
				}
			}
			sort.Strings(pnames)
			printFilteredInputs(pnames, true)
		}
	}
}

func printFilteredProcessors(processorFilters []string, commented bool) {
	// Filter processors
	var pnames []string
	for pname := range processors.Processors {
		if sliceContains(pname, processorFilters) {
			pnames = append(pnames, pname)
		}
	}
	sort.Strings(pnames)

	// Print Outputs
	for _, pname := range pnames {
		creator := processors.Processors[pname]
		output := creator()
		printConfig(pname, output, "processors", commented, processors.Deprecations[pname])
	}
}

func printFilteredAggregators(aggregatorFilters []string, commented bool) {
	// Filter outputs
	var anames []string
	for aname := range aggregators.Aggregators {
		if sliceContains(aname, aggregatorFilters) {
			anames = append(anames, aname)
		}
	}
	sort.Strings(anames)

	// Print Outputs
	for _, aname := range anames {
		creator := aggregators.Aggregators[aname]
		output := creator()
		printConfig(aname, output, "aggregators", commented, aggregators.Deprecations[aname])
	}
}

func printFilteredInputs(inputFilters []string, commented bool) {
	// Filter inputs
	var pnames []string
	for pname := range inputs.Inputs {
		if sliceContains(pname, inputFilters) {
			pnames = append(pnames, pname)
		}
	}
	sort.Strings(pnames)

	// cache service inputs to print them at the end
	servInputs := make(map[string]telegraf.ServiceInput)
	// for alphabetical looping:
	servInputNames := []string{}

	// Print Inputs
	for _, pname := range pnames {
		// Skip inputs that are registered twice for backward compatibility
		switch pname {
		case "cisco_telemetry_gnmi", "io", "KNXListener":
			continue
		}
		creator := inputs.Inputs[pname]
		input := creator()

		if p, ok := input.(telegraf.ServiceInput); ok {
			servInputs[pname] = p
			servInputNames = append(servInputNames, pname)
			continue
		}

		printConfig(pname, input, "inputs", commented, inputs.Deprecations[pname])
	}

	// Print Service Inputs
	if len(servInputs) == 0 {
		return
	}
	sort.Strings(servInputNames)

	fmt.Print(serviceInputHeader)
	for _, name := range servInputNames {
		printConfig(name, servInputs[name], "inputs", commented, inputs.Deprecations[name])
	}
}

func printFilteredOutputs(outputFilters []string, commented bool) {
	// Filter outputs
	var onames []string
	for oname := range outputs.Outputs {
		if sliceContains(oname, outputFilters) {
			onames = append(onames, oname)
		}
	}
	sort.Strings(onames)

	// Print Outputs
	for _, oname := range onames {
		creator := outputs.Outputs[oname]
		output := creator()
		printConfig(oname, output, "outputs", commented, outputs.Deprecations[oname])
	}
}

func printFilteredGlobalSections(sectionFilters []string) {
	if sliceContains("global_tags", sectionFilters) {
		fmt.Print(globalTagsConfig)
	}

	if sliceContains("agent", sectionFilters) {
		fmt.Print(agentConfig)
	}
}

func printConfig(name string, p telegraf.PluginDescriber, op string, commented bool, di telegraf.DeprecationInfo) {
	comment := ""
	if commented {
		comment = "# "
	}

	if di.Since != "" {
		removalNote := ""
		if di.RemovalIn != "" {
			removalNote = " and will be removed in " + di.RemovalIn
		}
		fmt.Printf("\n%s ## DEPRECATED: The '%s' plugin is deprecated in version %s%s, %s.", comment, name, di.Since, removalNote, di.Notice)
	}

	config := p.SampleConfig()
	if config == "" {
		fmt.Printf("\n#[[%s.%s]]", op, name)
		fmt.Printf("\n%s  # no configuration\n\n", comment)
	} else {
		lines := strings.Split(config, "\n")
		fmt.Print("\n")
		for i, line := range lines {
			if i == len(lines)-1 {
				fmt.Print("\n")
				continue
			}
			fmt.Print(strings.TrimRight(comment+line, " ") + "\n")
		}
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
	creator, ok := inputs.Inputs[name]
	if !ok {
		return fmt.Errorf("input %s not found", name)
	}

	printConfig(name, creator(), "inputs", false, inputs.Deprecations[name])
	return nil
}

// PrintOutputConfig prints the config usage of a single output.
func PrintOutputConfig(name string) error {
	creator, ok := outputs.Outputs[name]
	if !ok {
		return fmt.Errorf("output %s not found", name)
	}

	printConfig(name, creator(), "outputs", false, outputs.Deprecations[name])
	return nil
}

// LoadDirectory loads all toml config files found in the specified path, recursively.
func (c *Config) LoadDirectory(path string) error {
	walkfn := func(thispath string, info os.FileInfo, _ error) error {
		if info == nil {
			log.Printf("W! Telegraf is not permitted to read %s", thispath)
			return nil
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), "..") {
				// skip Kubernetes mounts, prevening loading the same config twice
				return filepath.SkipDir
			}

			return nil
		}
		name := info.Name()
		if len(name) < 6 || name[len(name)-5:] != ".conf" {
			return nil
		}
		err := c.LoadConfig(thispath)
		if err != nil {
			return err
		}
		return nil
	}
	return filepath.Walk(path, walkfn)
}

// Try to find a default config file at these locations (in order):
//   1. $TELEGRAF_CONFIG_PATH
//   2. $HOME/.telegraf/telegraf.conf
//   3. /etc/telegraf/telegraf.conf
//
func getDefaultConfigPath() (string, error) {
	envfile := os.Getenv("TELEGRAF_CONFIG_PATH")
	homefile := os.ExpandEnv("${HOME}/.telegraf/telegraf.conf")
	etcfile := "/etc/telegraf/telegraf.conf"
	if runtime.GOOS == "windows" {
		programFiles := os.Getenv("ProgramFiles")
		if programFiles == "" { // Should never happen
			programFiles = `C:\Program Files`
		}
		etcfile = programFiles + `\Telegraf\telegraf.conf`
	}
	for _, path := range []string{envfile, homefile, etcfile} {
		if isURL(path) {
			log.Printf("I! Using config url: %s", path)
			return path, nil
		}
		if _, err := os.Stat(path); err == nil {
			log.Printf("I! Using config file: %s", path)
			return path, nil
		}
	}

	// if we got here, we didn't find a file in a default location
	return "", fmt.Errorf("No config file specified, and could not find one"+
		" in $TELEGRAF_CONFIG_PATH, %s, or %s", homefile, etcfile)
}

// isURL checks if string is valid url
func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// LoadConfig loads the given config file and applies it to c
func (c *Config) LoadConfig(path string) error {
	var err error
	if path == "" {
		if path, err = getDefaultConfigPath(); err != nil {
			return err
		}
	}
	data, err := loadConfig(path)
	if err != nil {
		return fmt.Errorf("Error loading config file %s: %w", path, err)
	}

	if err = c.LoadConfigData(data); err != nil {
		return fmt.Errorf("Error loading config file %s: %w", path, err)
	}
	return nil
}

// LoadConfigData loads TOML-formatted config data
func (c *Config) LoadConfigData(data []byte) error {
	tbl, err := parseConfig(data)
	if err != nil {
		return fmt.Errorf("Error parsing data: %s", err)
	}

	// Parse tags tables first:
	for _, tableName := range []string{"tags", "global_tags"} {
		if val, ok := tbl.Fields[tableName]; ok {
			subTable, ok := val.(*ast.Table)
			if !ok {
				return fmt.Errorf("invalid configuration, bad table name %q", tableName)
			}
			if err = c.toml.UnmarshalTable(subTable, c.Tags); err != nil {
				return fmt.Errorf("error parsing table name %q: %s", tableName, err)
			}
		}
	}

	// Parse agent table:
	if val, ok := tbl.Fields["agent"]; ok {
		subTable, ok := val.(*ast.Table)
		if !ok {
			return fmt.Errorf("invalid configuration, error parsing agent table")
		}
		if err = c.toml.UnmarshalTable(subTable, c.Agent); err != nil {
			return fmt.Errorf("error parsing [agent]: %w", err)
		}
	}

	if !c.Agent.OmitHostname {
		if c.Agent.Hostname == "" {
			hostname, err := os.Hostname()
			if err != nil {
				return err
			}

			c.Agent.Hostname = hostname
		}

		c.Tags["host"] = c.Agent.Hostname
	}

	// Set snmp agent translator default
	if c.Agent.SnmpTranslator == "" {
		c.Agent.SnmpTranslator = "netsnmp"
	}

	if len(c.UnusedFields) > 0 {
		return fmt.Errorf("line %d: configuration specified the fields %q, but they weren't used", tbl.Line, keys(c.UnusedFields))
	}

	// Parse all the rest of the plugins:
	for name, val := range tbl.Fields {
		subTable, ok := val.(*ast.Table)
		if !ok {
			return fmt.Errorf("invalid configuration, error parsing field %q as table", name)
		}

		switch name {
		case "agent", "global_tags", "tags":
		case "outputs":
			for pluginName, pluginVal := range subTable.Fields {
				switch pluginSubTable := pluginVal.(type) {
				// legacy [outputs.influxdb] support
				case *ast.Table:
					if err = c.addOutput(pluginName, pluginSubTable); err != nil {
						return fmt.Errorf("error parsing %s, %w", pluginName, err)
					}
				case []*ast.Table:
					for _, t := range pluginSubTable {
						if err = c.addOutput(pluginName, t); err != nil {
							return fmt.Errorf("error parsing %s array, %w", pluginName, err)
						}
					}
				default:
					return fmt.Errorf("unsupported config format: %s",
						pluginName)
				}
				if len(c.UnusedFields) > 0 {
					return fmt.Errorf("plugin %s.%s: line %d: configuration specified the fields %q, but they weren't used", name, pluginName, subTable.Line, keys(c.UnusedFields))
				}
			}
		case "inputs", "plugins":
			for pluginName, pluginVal := range subTable.Fields {
				switch pluginSubTable := pluginVal.(type) {
				// legacy [inputs.cpu] support
				case *ast.Table:
					if err = c.addInput(pluginName, pluginSubTable); err != nil {
						return fmt.Errorf("error parsing %s, %w", pluginName, err)
					}
				case []*ast.Table:
					for _, t := range pluginSubTable {
						if err = c.addInput(pluginName, t); err != nil {
							return fmt.Errorf("error parsing %s, %w", pluginName, err)
						}
					}
				default:
					return fmt.Errorf("Unsupported config format: %s",
						pluginName)
				}
				if len(c.UnusedFields) > 0 {
					return fmt.Errorf("plugin %s.%s: line %d: configuration specified the fields %q, but they weren't used", name, pluginName, subTable.Line, keys(c.UnusedFields))
				}
			}
		case "processors":
			for pluginName, pluginVal := range subTable.Fields {
				switch pluginSubTable := pluginVal.(type) {
				case []*ast.Table:
					for _, t := range pluginSubTable {
						if err = c.addProcessor(pluginName, t); err != nil {
							return fmt.Errorf("error parsing %s, %w", pluginName, err)
						}
					}
				default:
					return fmt.Errorf("Unsupported config format: %s",
						pluginName)
				}
				if len(c.UnusedFields) > 0 {
					return fmt.Errorf("plugin %s.%s: line %d: configuration specified the fields %q, but they weren't used", name, pluginName, subTable.Line, keys(c.UnusedFields))
				}
			}
		case "aggregators":
			for pluginName, pluginVal := range subTable.Fields {
				switch pluginSubTable := pluginVal.(type) {
				case []*ast.Table:
					for _, t := range pluginSubTable {
						if err = c.addAggregator(pluginName, t); err != nil {
							return fmt.Errorf("Error parsing %s, %s", pluginName, err)
						}
					}
				default:
					return fmt.Errorf("Unsupported config format: %s",
						pluginName)
				}
				if len(c.UnusedFields) > 0 {
					return fmt.Errorf("plugin %s.%s: line %d: configuration specified the fields %q, but they weren't used", name, pluginName, subTable.Line, keys(c.UnusedFields))
				}
			}
		// Assume it's an input input for legacy config file support if no other
		// identifiers are present
		default:
			if err = c.addInput(name, subTable); err != nil {
				return fmt.Errorf("Error parsing %s, %s", name, err)
			}
		}
	}

	if len(c.Processors) > 1 {
		sort.Sort(c.Processors)
	}

	return nil
}

// trimBOM trims the Byte-Order-Marks from the beginning of the file.
// this is for Windows compatibility only.
// see https://github.com/influxdata/telegraf/issues/1378
func trimBOM(f []byte) []byte {
	return bytes.TrimPrefix(f, []byte("\xef\xbb\xbf"))
}

// escapeEnv escapes a value for inserting into a TOML string.
func escapeEnv(value string) string {
	return envVarEscaper.Replace(value)
}

func loadConfig(config string) ([]byte, error) {
	if fetchURLRe.MatchString(config) {
		u, err := url.Parse(config)
		if err != nil {
			return nil, err
		}

		switch u.Scheme {
		case "https", "http":
			return fetchConfig(u)
		default:
			return nil, fmt.Errorf("scheme %q not supported", u.Scheme)
		}
	}

	// If it isn't a https scheme, try it as a file
	return os.ReadFile(config)
}

func fetchConfig(u *url.URL) ([]byte, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	if v, exists := os.LookupEnv("INFLUX_TOKEN"); exists {
		req.Header.Add("Authorization", "Token "+v)
	}
	req.Header.Add("Accept", "application/toml")
	req.Header.Set("User-Agent", internal.ProductToken())

	retries := 3
	for i := 0; i <= retries; i++ {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("Retry %d of %d failed connecting to HTTP config server %s", i, retries, err)
		}

		if resp.StatusCode != http.StatusOK {
			if i < retries {
				log.Printf("Error getting HTTP config.  Retry %d of %d in %s.  Status=%d", i, retries, httpLoadConfigRetryInterval, resp.StatusCode)
				time.Sleep(httpLoadConfigRetryInterval)
				continue
			}
			return nil, fmt.Errorf("Retry %d of %d failed to retrieve remote config: %s", i, retries, resp.Status)
		}
		defer resp.Body.Close()
		return io.ReadAll(resp.Body)
	}

	return nil, nil
}

// parseConfig loads a TOML configuration from a provided path and
// returns the AST produced from the TOML parser. When loading the file, it
// will find environment variables and replace them.
func parseConfig(contents []byte) (*ast.Table, error) {
	contents = trimBOM(contents)

	parameters := envVarRe.FindAllSubmatch(contents, -1)
	for _, parameter := range parameters {
		if len(parameter) != 3 {
			continue
		}

		var envVar []byte
		if parameter[1] != nil {
			envVar = parameter[1]
		} else if parameter[2] != nil {
			envVar = parameter[2]
		} else {
			continue
		}

		envVal, ok := os.LookupEnv(strings.TrimPrefix(string(envVar), "$"))
		if ok {
			envVal = escapeEnv(envVal)
			contents = bytes.Replace(contents, parameter[0], []byte(envVal), 1)
		}
	}

	return toml.Parse(contents)
}

func (c *Config) addAggregator(name string, table *ast.Table) error {
	creator, ok := aggregators.Aggregators[name]
	if !ok {
		// Handle removed, deprecated plugins
		if di, deprecated := aggregators.Deprecations[name]; deprecated {
			printHistoricPluginDeprecationNotice("aggregators", name, di)
			return fmt.Errorf("plugin deprecated")
		}
		return fmt.Errorf("Undefined but requested aggregator: %s", name)
	}
	aggregator := creator()

	conf, err := c.buildAggregator(name, table)
	if err != nil {
		return err
	}

	if err := c.toml.UnmarshalTable(table, aggregator); err != nil {
		return err
	}

	if err := c.printUserDeprecation("aggregators", name, aggregator); err != nil {
		return err
	}

	c.Aggregators = append(c.Aggregators, models.NewRunningAggregator(aggregator, conf))
	return nil
}

func (c *Config) probeParser(table *ast.Table) bool {
	var dataformat string
	c.getFieldString(table, "data_format", &dataformat)

	_, ok := parsers.Parsers[dataformat]
	return ok
}

func (c *Config) addParser(parentname string, table *ast.Table) (*models.RunningParser, error) {
	var dataformat string
	c.getFieldString(table, "data_format", &dataformat)

	creator, ok := parsers.Parsers[dataformat]
	if !ok {
		return nil, fmt.Errorf("Undefined but requested parser: %s", dataformat)
	}
	parser := creator(parentname)

	conf, err := c.buildParser(parentname, table)
	if err != nil {
		return nil, err
	}

	if err := c.toml.UnmarshalTable(table, parser); err != nil {
		return nil, err
	}

	running := models.NewRunningParser(parser, conf)
	c.Parsers = append(c.Parsers, running)

	return running, nil
}

func (c *Config) addProcessor(name string, table *ast.Table) error {
	creator, ok := processors.Processors[name]
	if !ok {
		// Handle removed, deprecated plugins
		if di, deprecated := processors.Deprecations[name]; deprecated {
			printHistoricPluginDeprecationNotice("processors", name, di)
			return fmt.Errorf("plugin deprecated")
		}
		return fmt.Errorf("Undefined but requested processor: %s", name)
	}

	processorConfig, err := c.buildProcessor(name, table)
	if err != nil {
		return err
	}

	rf, err := c.newRunningProcessor(creator, processorConfig, table)
	if err != nil {
		return err
	}
	c.Processors = append(c.Processors, rf)

	// save a copy for the aggregator
	rf, err = c.newRunningProcessor(creator, processorConfig, table)
	if err != nil {
		return err
	}
	c.AggProcessors = append(c.AggProcessors, rf)

	return nil
}

func (c *Config) newRunningProcessor(
	creator processors.StreamingCreator,
	processorConfig *models.ProcessorConfig,
	table *ast.Table,
) (*models.RunningProcessor, error) {
	processor := creator()

	if p, ok := processor.(unwrappable); ok {
		if err := c.toml.UnmarshalTable(table, p.Unwrap()); err != nil {
			return nil, err
		}
	} else {
		if err := c.toml.UnmarshalTable(table, processor); err != nil {
			return nil, err
		}
	}

	if err := c.printUserDeprecation("processors", processorConfig.Name, processor); err != nil {
		return nil, err
	}

	rf := models.NewRunningProcessor(processor, processorConfig)
	return rf, nil
}

func (c *Config) addOutput(name string, table *ast.Table) error {
	if len(c.OutputFilters) > 0 && !sliceContains(name, c.OutputFilters) {
		return nil
	}
	creator, ok := outputs.Outputs[name]
	if !ok {
		// Handle removed, deprecated plugins
		if di, deprecated := outputs.Deprecations[name]; deprecated {
			printHistoricPluginDeprecationNotice("outputs", name, di)
			return fmt.Errorf("plugin deprecated")
		}
		return fmt.Errorf("undefined but requested output: %s", name)
	}
	output := creator()

	// If the output has a SetSerializer function, then this means it can write
	// arbitrary types of output, so build the serializer and set it.
	if t, ok := output.(serializers.SerializerOutput); ok {
		serializer, err := c.buildSerializer(table)
		if err != nil {
			return err
		}
		t.SetSerializer(serializer)
	}

	outputConfig, err := c.buildOutput(name, table)
	if err != nil {
		return err
	}

	if err := c.toml.UnmarshalTable(table, output); err != nil {
		return err
	}

	if err := c.printUserDeprecation("outputs", name, output); err != nil {
		return err
	}

	if c, ok := interface{}(output).(interface{ TLSConfig() (*tls.Config, error) }); ok {
		if _, err := c.TLSConfig(); err != nil {
			return err
		}
	}

	ro := models.NewRunningOutput(output, outputConfig, c.Agent.MetricBatchSize, c.Agent.MetricBufferLimit)
	c.Outputs = append(c.Outputs, ro)
	return nil
}

func (c *Config) addInput(name string, table *ast.Table) error {
	if len(c.InputFilters) > 0 && !sliceContains(name, c.InputFilters) {
		return nil
	}

	// For inputs with parsers we need to compute the set of
	// options that is not covered by both, the parser and the input.
	// We achieve this by keeping a local book of missing entries
	// that counts the number of misses. In case we have a parser
	// for the input both need to miss the entry. We count the
	// missing entries at the end.
	missThreshold := 0
	missCount := make(map[string]int)
	c.setLocalMissingTomlFieldTracker(missCount)
	defer c.resetMissingTomlFieldTracker()

	creator, ok := inputs.Inputs[name]
	if !ok {
		// Handle removed, deprecated plugins
		if di, deprecated := inputs.Deprecations[name]; deprecated {
			printHistoricPluginDeprecationNotice("inputs", name, di)
			return fmt.Errorf("plugin deprecated")
		}

		return fmt.Errorf("Undefined but requested input: %s", name)
	}
	input := creator()

	// If the input has a SetParser or SetParserFunc function, it can accept
	// arbitrary data-formats, so build the requested parser and set it.
	if t, ok := input.(telegraf.ParserInput); ok {
		missThreshold = 1
		if parser, err := c.addParser(name, table); err == nil {
			t.SetParser(parser)
		} else {
			missThreshold = 0
			// Fallback to the old way of instantiating the parsers.
			config, err := c.getParserConfig(name, table)
			if err != nil {
				return err
			}
			parser, err := c.buildParserOld(name, config)
			if err != nil {
				return err
			}
			t.SetParser(parser)
		}
	}

	// Keep the old interface for backward compatibility
	if t, ok := input.(parsers.ParserInput); ok {
		// DEPRECATED: Please switch your plugin to telegraf.ParserInput.
		missThreshold = 1
		if parser, err := c.addParser(name, table); err == nil {
			t.SetParser(parser)
		} else {
			missThreshold = 0
			// Fallback to the old way of instantiating the parsers.
			config, err := c.getParserConfig(name, table)
			if err != nil {
				return err
			}
			parser, err := c.buildParserOld(name, config)
			if err != nil {
				return err
			}
			t.SetParser(parser)
		}
	}

	if t, ok := input.(telegraf.ParserFuncInput); ok {
		missThreshold = 1
		if c.probeParser(table) {
			t.SetParserFunc(func() (telegraf.Parser, error) {
				parser, err := c.addParser(name, table)
				if err != nil {
					return nil, err
				}
				err = parser.Init()
				return parser, err
			})
		} else {
			missThreshold = 0
			// Fallback to the old way
			config, err := c.getParserConfig(name, table)
			if err != nil {
				return err
			}
			t.SetParserFunc(func() (telegraf.Parser, error) {
				return c.buildParserOld(name, config)
			})
		}
	}

	if t, ok := input.(parsers.ParserFuncInput); ok {
		// DEPRECATED: Please switch your plugin to telegraf.ParserFuncInput.
		missThreshold = 1
		if c.probeParser(table) {
			t.SetParserFunc(func() (parsers.Parser, error) {
				parser, err := c.addParser(name, table)
				if err != nil {
					return nil, err
				}
				err = parser.Init()
				return parser, err
			})
		} else {
			missThreshold = 0
			// Fallback to the old way
			config, err := c.getParserConfig(name, table)
			if err != nil {
				return err
			}
			t.SetParserFunc(func() (parsers.Parser, error) {
				return c.buildParserOld(name, config)
			})
		}
	}

	pluginConfig, err := c.buildInput(name, table)
	if err != nil {
		return err
	}

	if err := c.toml.UnmarshalTable(table, input); err != nil {
		return err
	}

	if err := c.printUserDeprecation("inputs", name, input); err != nil {
		return err
	}

	if c, ok := interface{}(input).(interface{ TLSConfig() (*tls.Config, error) }); ok {
		if _, err := c.TLSConfig(); err != nil {
			return err
		}
	}

	rp := models.NewRunningInput(input, pluginConfig)
	rp.SetDefaultTags(c.Tags)
	c.Inputs = append(c.Inputs, rp)

	// Check the number of misses against the threshold
	for key, count := range missCount {
		if count <= missThreshold {
			continue
		}
		if err := c.missingTomlField(nil, key); err != nil {
			return err
		}
	}

	return nil
}

// buildAggregator parses Aggregator specific items from the ast.Table,
// builds the filter and returns a
// models.AggregatorConfig to be inserted into models.RunningAggregator
func (c *Config) buildAggregator(name string, tbl *ast.Table) (*models.AggregatorConfig, error) {
	conf := &models.AggregatorConfig{
		Name:   name,
		Delay:  time.Millisecond * 100,
		Period: time.Second * 30,
		Grace:  time.Second * 0,
	}

	c.getFieldDuration(tbl, "period", &conf.Period)
	c.getFieldDuration(tbl, "delay", &conf.Delay)
	c.getFieldDuration(tbl, "grace", &conf.Grace)
	c.getFieldBool(tbl, "drop_original", &conf.DropOriginal)
	c.getFieldString(tbl, "name_prefix", &conf.MeasurementPrefix)
	c.getFieldString(tbl, "name_suffix", &conf.MeasurementSuffix)
	c.getFieldString(tbl, "name_override", &conf.NameOverride)
	c.getFieldString(tbl, "alias", &conf.Alias)

	conf.Tags = make(map[string]string)
	if node, ok := tbl.Fields["tags"]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			if err := c.toml.UnmarshalTable(subtbl, conf.Tags); err != nil {
				return nil, fmt.Errorf("could not parse tags for input %s", name)
			}
		}
	}

	if c.hasErrs() {
		return nil, c.firstErr()
	}

	var err error
	conf.Filter, err = c.buildFilter(tbl)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

// buildParser parses Parser specific items from the ast.Table,
// builds the filter and returns a
// models.ParserConfig to be inserted into models.RunningParser
func (c *Config) buildParser(name string, tbl *ast.Table) (*models.ParserConfig, error) {
	var dataformat string
	c.getFieldString(tbl, "data_format", &dataformat)

	conf := &models.ParserConfig{
		Parent:     name,
		DataFormat: dataformat,
	}

	return conf, nil
}

// buildProcessor parses Processor specific items from the ast.Table,
// builds the filter and returns a
// models.ProcessorConfig to be inserted into models.RunningProcessor
func (c *Config) buildProcessor(name string, tbl *ast.Table) (*models.ProcessorConfig, error) {
	conf := &models.ProcessorConfig{Name: name}

	c.getFieldInt64(tbl, "order", &conf.Order)
	c.getFieldString(tbl, "alias", &conf.Alias)

	if c.hasErrs() {
		return nil, c.firstErr()
	}

	var err error
	conf.Filter, err = c.buildFilter(tbl)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

// buildFilter builds a Filter
// (tagpass/tagdrop/namepass/namedrop/fieldpass/fielddrop) to
// be inserted into the models.OutputConfig/models.InputConfig
// to be used for glob filtering on tags and measurements
func (c *Config) buildFilter(tbl *ast.Table) (models.Filter, error) {
	f := models.Filter{}

	c.getFieldStringSlice(tbl, "namepass", &f.NamePass)
	c.getFieldStringSlice(tbl, "namedrop", &f.NameDrop)

	c.getFieldStringSlice(tbl, "pass", &f.FieldPass)
	c.getFieldStringSlice(tbl, "fieldpass", &f.FieldPass)

	c.getFieldStringSlice(tbl, "drop", &f.FieldDrop)
	c.getFieldStringSlice(tbl, "fielddrop", &f.FieldDrop)

	c.getFieldTagFilter(tbl, "tagpass", &f.TagPass)
	c.getFieldTagFilter(tbl, "tagdrop", &f.TagDrop)

	c.getFieldStringSlice(tbl, "tagexclude", &f.TagExclude)
	c.getFieldStringSlice(tbl, "taginclude", &f.TagInclude)

	if c.hasErrs() {
		return f, c.firstErr()
	}

	if err := f.Compile(); err != nil {
		return f, err
	}

	return f, nil
}

// buildInput parses input specific items from the ast.Table,
// builds the filter and returns a
// models.InputConfig to be inserted into models.RunningInput
func (c *Config) buildInput(name string, tbl *ast.Table) (*models.InputConfig, error) {
	cp := &models.InputConfig{Name: name}
	c.getFieldDuration(tbl, "interval", &cp.Interval)
	c.getFieldDuration(tbl, "precision", &cp.Precision)
	c.getFieldDuration(tbl, "collection_jitter", &cp.CollectionJitter)
	c.getFieldDuration(tbl, "collection_offset", &cp.CollectionOffset)
	c.getFieldString(tbl, "name_prefix", &cp.MeasurementPrefix)
	c.getFieldString(tbl, "name_suffix", &cp.MeasurementSuffix)
	c.getFieldString(tbl, "name_override", &cp.NameOverride)
	c.getFieldString(tbl, "alias", &cp.Alias)

	cp.Tags = make(map[string]string)
	if node, ok := tbl.Fields["tags"]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			if err := c.toml.UnmarshalTable(subtbl, cp.Tags); err != nil {
				return nil, fmt.Errorf("could not parse tags for input %s", name)
			}
		}
	}

	if c.hasErrs() {
		return nil, c.firstErr()
	}

	var err error
	cp.Filter, err = c.buildFilter(tbl)
	if err != nil {
		return cp, err
	}
	return cp, nil
}

// buildParserOld grabs the necessary entries from the ast.Table for creating
// a parsers.Parser object, and creates it, which can then be added onto
// an Input object.
func (c *Config) buildParserOld(name string, config *parsers.Config) (telegraf.Parser, error) {
	parser, err := parsers.NewParser(config)
	if err != nil {
		return nil, err
	}
	logger := models.NewLogger("parsers", config.DataFormat, name)
	models.SetLoggerOnPlugin(parser, logger)
	if initializer, ok := parser.(telegraf.Initializer); ok {
		if err := initializer.Init(); err != nil {
			return nil, err
		}
	}

	return parser, nil
}

func (c *Config) getParserConfig(name string, tbl *ast.Table) (*parsers.Config, error) {
	pc := &parsers.Config{
		JSONStrict: true,
	}

	c.getFieldString(tbl, "data_format", &pc.DataFormat)

	// Legacy support, exec plugin originally parsed JSON by default.
	if name == "exec" && pc.DataFormat == "" {
		pc.DataFormat = "json"
	} else if pc.DataFormat == "" {
		pc.DataFormat = "influx"
	}

	c.getFieldString(tbl, "separator", &pc.Separator)

	c.getFieldStringSlice(tbl, "templates", &pc.Templates)
	c.getFieldStringSlice(tbl, "tag_keys", &pc.TagKeys)
	c.getFieldStringSlice(tbl, "json_string_fields", &pc.JSONStringFields)
	c.getFieldString(tbl, "json_name_key", &pc.JSONNameKey)
	c.getFieldString(tbl, "json_query", &pc.JSONQuery)
	c.getFieldString(tbl, "json_time_key", &pc.JSONTimeKey)
	c.getFieldString(tbl, "json_time_format", &pc.JSONTimeFormat)
	c.getFieldString(tbl, "json_timezone", &pc.JSONTimezone)
	c.getFieldBool(tbl, "json_strict", &pc.JSONStrict)
	c.getFieldString(tbl, "data_type", &pc.DataType)
	c.getFieldString(tbl, "collectd_auth_file", &pc.CollectdAuthFile)
	c.getFieldString(tbl, "collectd_security_level", &pc.CollectdSecurityLevel)
	c.getFieldString(tbl, "collectd_parse_multivalue", &pc.CollectdSplit)

	c.getFieldStringSlice(tbl, "collectd_typesdb", &pc.CollectdTypesDB)

	c.getFieldString(tbl, "dropwizard_metric_registry_path", &pc.DropwizardMetricRegistryPath)
	c.getFieldString(tbl, "dropwizard_time_path", &pc.DropwizardTimePath)
	c.getFieldString(tbl, "dropwizard_time_format", &pc.DropwizardTimeFormat)
	c.getFieldString(tbl, "dropwizard_tags_path", &pc.DropwizardTagsPath)
	c.getFieldStringMap(tbl, "dropwizard_tag_paths", &pc.DropwizardTagPathsMap)

	//for grok data_format
	c.getFieldStringSlice(tbl, "grok_named_patterns", &pc.GrokNamedPatterns)
	c.getFieldStringSlice(tbl, "grok_patterns", &pc.GrokPatterns)
	c.getFieldString(tbl, "grok_custom_patterns", &pc.GrokCustomPatterns)
	c.getFieldStringSlice(tbl, "grok_custom_pattern_files", &pc.GrokCustomPatternFiles)
	c.getFieldString(tbl, "grok_timezone", &pc.GrokTimezone)
	c.getFieldString(tbl, "grok_unique_timestamp", &pc.GrokUniqueTimestamp)

	c.getFieldStringSlice(tbl, "form_urlencoded_tag_keys", &pc.FormUrlencodedTagKeys)

	c.getFieldString(tbl, "value_field_name", &pc.ValueFieldName)

	// for influx parser
	c.getFieldString(tbl, "influx_parser_type", &pc.InfluxParserType)

	//for XPath parser family
	if choice.Contains(pc.DataFormat, []string{"xml", "xpath_json", "xpath_msgpack", "xpath_protobuf"}) {
		c.getFieldString(tbl, "xpath_protobuf_file", &pc.XPathProtobufFile)
		c.getFieldString(tbl, "xpath_protobuf_type", &pc.XPathProtobufType)
		c.getFieldStringSlice(tbl, "xpath_protobuf_import_paths", &pc.XPathProtobufImportPaths)
		c.getFieldBool(tbl, "xpath_print_document", &pc.XPathPrintDocument)

		// Determine the actual xpath configuration tables
		node, xpathOK := tbl.Fields["xpath"]
		if !xpathOK {
			// Add this for backward compatibility
			node, xpathOK = tbl.Fields[pc.DataFormat]
		}
		if xpathOK {
			if subtbls, ok := node.([]*ast.Table); ok {
				pc.XPathConfig = make([]parsers.XPathConfig, len(subtbls))
				for i, subtbl := range subtbls {
					subcfg := pc.XPathConfig[i]
					c.getFieldString(subtbl, "metric_name", &subcfg.MetricQuery)
					c.getFieldString(subtbl, "metric_selection", &subcfg.Selection)
					c.getFieldString(subtbl, "timestamp", &subcfg.Timestamp)
					c.getFieldString(subtbl, "timestamp_format", &subcfg.TimestampFmt)
					c.getFieldStringMap(subtbl, "tags", &subcfg.Tags)
					c.getFieldStringMap(subtbl, "fields", &subcfg.Fields)
					c.getFieldStringMap(subtbl, "fields_int", &subcfg.FieldsInt)
					c.getFieldString(subtbl, "field_selection", &subcfg.FieldSelection)
					c.getFieldBool(subtbl, "field_name_expansion", &subcfg.FieldNameExpand)
					c.getFieldString(subtbl, "field_name", &subcfg.FieldNameQuery)
					c.getFieldString(subtbl, "field_value", &subcfg.FieldValueQuery)
					c.getFieldString(subtbl, "tag_selection", &subcfg.TagSelection)
					c.getFieldBool(subtbl, "tag_name_expansion", &subcfg.TagNameExpand)
					c.getFieldString(subtbl, "tag_name", &subcfg.TagNameQuery)
					c.getFieldString(subtbl, "tag_value", &subcfg.TagValueQuery)
					pc.XPathConfig[i] = subcfg
				}
			}
		}
	}

	//for JSONPath parser
	if node, ok := tbl.Fields["json_v2"]; ok {
		if metricConfigs, ok := node.([]*ast.Table); ok {
			pc.JSONV2Config = make([]parsers.JSONV2Config, len(metricConfigs))
			for i, metricConfig := range metricConfigs {
				mc := pc.JSONV2Config[i]
				c.getFieldString(metricConfig, "measurement_name", &mc.MeasurementName)
				if mc.MeasurementName == "" {
					mc.MeasurementName = name
				}
				c.getFieldString(metricConfig, "measurement_name_path", &mc.MeasurementNamePath)
				c.getFieldString(metricConfig, "timestamp_path", &mc.TimestampPath)
				c.getFieldString(metricConfig, "timestamp_format", &mc.TimestampFormat)
				c.getFieldString(metricConfig, "timestamp_timezone", &mc.TimestampTimezone)

				mc.Fields = getFieldSubtable(c, metricConfig)
				mc.Tags = getTagSubtable(c, metricConfig)

				if objectconfigs, ok := metricConfig.Fields["object"]; ok {
					if objectconfigs, ok := objectconfigs.([]*ast.Table); ok {
						for _, objectConfig := range objectconfigs {
							var o json_v2.JSONObject
							c.getFieldString(objectConfig, "path", &o.Path)
							c.getFieldBool(objectConfig, "optional", &o.Optional)
							c.getFieldString(objectConfig, "timestamp_key", &o.TimestampKey)
							c.getFieldString(objectConfig, "timestamp_format", &o.TimestampFormat)
							c.getFieldString(objectConfig, "timestamp_timezone", &o.TimestampTimezone)
							c.getFieldBool(objectConfig, "disable_prepend_keys", &o.DisablePrependKeys)
							c.getFieldStringSlice(objectConfig, "included_keys", &o.IncludedKeys)
							c.getFieldStringSlice(objectConfig, "excluded_keys", &o.ExcludedKeys)
							c.getFieldStringSlice(objectConfig, "tags", &o.Tags)
							c.getFieldStringMap(objectConfig, "renames", &o.Renames)
							c.getFieldStringMap(objectConfig, "fields", &o.Fields)

							o.FieldPaths = getFieldSubtable(c, objectConfig)
							o.TagPaths = getTagSubtable(c, objectConfig)

							mc.JSONObjects = append(mc.JSONObjects, o)
						}
					}
				}

				pc.JSONV2Config[i] = mc
			}
		}
	}

	pc.MetricName = name

	if c.hasErrs() {
		return nil, c.firstErr()
	}

	return pc, nil
}

func getFieldSubtable(c *Config, metricConfig *ast.Table) []json_v2.DataSet {
	var fields []json_v2.DataSet

	if fieldConfigs, ok := metricConfig.Fields["field"]; ok {
		if fieldConfigs, ok := fieldConfigs.([]*ast.Table); ok {
			for _, fieldconfig := range fieldConfigs {
				var f json_v2.DataSet
				c.getFieldString(fieldconfig, "path", &f.Path)
				c.getFieldString(fieldconfig, "rename", &f.Rename)
				c.getFieldString(fieldconfig, "type", &f.Type)
				c.getFieldBool(fieldconfig, "optional", &f.Optional)
				fields = append(fields, f)
			}
		}
	}

	return fields
}

func getTagSubtable(c *Config, metricConfig *ast.Table) []json_v2.DataSet {
	var tags []json_v2.DataSet

	if fieldConfigs, ok := metricConfig.Fields["tag"]; ok {
		if fieldConfigs, ok := fieldConfigs.([]*ast.Table); ok {
			for _, fieldconfig := range fieldConfigs {
				var t json_v2.DataSet
				c.getFieldString(fieldconfig, "path", &t.Path)
				c.getFieldString(fieldconfig, "rename", &t.Rename)
				t.Type = "string"
				tags = append(tags, t)
				c.getFieldBool(fieldconfig, "optional", &t.Optional)
			}
		}
	}

	return tags
}

// buildSerializer grabs the necessary entries from the ast.Table for creating
// a serializers.Serializer object, and creates it, which can then be added onto
// an Output object.
func (c *Config) buildSerializer(tbl *ast.Table) (serializers.Serializer, error) {
	sc := &serializers.Config{TimestampUnits: 1 * time.Second}

	c.getFieldString(tbl, "data_format", &sc.DataFormat)

	if sc.DataFormat == "" {
		sc.DataFormat = "influx"
	}

	c.getFieldString(tbl, "prefix", &sc.Prefix)
	c.getFieldString(tbl, "template", &sc.Template)
	c.getFieldStringSlice(tbl, "templates", &sc.Templates)
	c.getFieldString(tbl, "carbon2_format", &sc.Carbon2Format)
	c.getFieldString(tbl, "carbon2_sanitize_replace_char", &sc.Carbon2SanitizeReplaceChar)
	c.getFieldInt(tbl, "influx_max_line_bytes", &sc.InfluxMaxLineBytes)

	c.getFieldBool(tbl, "influx_sort_fields", &sc.InfluxSortFields)
	c.getFieldBool(tbl, "influx_uint_support", &sc.InfluxUintSupport)
	c.getFieldBool(tbl, "graphite_tag_support", &sc.GraphiteTagSupport)
	c.getFieldString(tbl, "graphite_tag_sanitize_mode", &sc.GraphiteTagSanitizeMode)

	c.getFieldString(tbl, "graphite_separator", &sc.GraphiteSeparator)

	c.getFieldDuration(tbl, "json_timestamp_units", &sc.TimestampUnits)
	c.getFieldString(tbl, "json_timestamp_format", &sc.TimestampFormat)

	c.getFieldBool(tbl, "splunkmetric_hec_routing", &sc.HecRouting)
	c.getFieldBool(tbl, "splunkmetric_multimetric", &sc.SplunkmetricMultiMetric)

	c.getFieldStringSlice(tbl, "wavefront_source_override", &sc.WavefrontSourceOverride)
	c.getFieldBool(tbl, "wavefront_use_strict", &sc.WavefrontUseStrict)
	c.getFieldBool(tbl, "wavefront_disable_prefix_conversion", &sc.WavefrontDisablePrefixConversion)

	c.getFieldBool(tbl, "prometheus_export_timestamp", &sc.PrometheusExportTimestamp)
	c.getFieldBool(tbl, "prometheus_sort_metrics", &sc.PrometheusSortMetrics)
	c.getFieldBool(tbl, "prometheus_string_as_label", &sc.PrometheusStringAsLabel)

	if c.hasErrs() {
		return nil, c.firstErr()
	}

	return serializers.NewSerializer(sc)
}

// buildOutput parses output specific items from the ast.Table,
// builds the filter and returns an
// models.OutputConfig to be inserted into models.RunningInput
// Note: error exists in the return for future calls that might require error
func (c *Config) buildOutput(name string, tbl *ast.Table) (*models.OutputConfig, error) {
	filter, err := c.buildFilter(tbl)
	if err != nil {
		return nil, err
	}
	oc := &models.OutputConfig{
		Name:   name,
		Filter: filter,
	}

	// TODO: support FieldPass/FieldDrop on outputs

	c.getFieldDuration(tbl, "flush_interval", &oc.FlushInterval)
	c.getFieldDuration(tbl, "flush_jitter", &oc.FlushJitter)

	c.getFieldInt(tbl, "metric_buffer_limit", &oc.MetricBufferLimit)
	c.getFieldInt(tbl, "metric_batch_size", &oc.MetricBatchSize)
	c.getFieldString(tbl, "alias", &oc.Alias)
	c.getFieldString(tbl, "name_override", &oc.NameOverride)
	c.getFieldString(tbl, "name_suffix", &oc.NameSuffix)
	c.getFieldString(tbl, "name_prefix", &oc.NamePrefix)

	if c.hasErrs() {
		return nil, c.firstErr()
	}

	return oc, nil
}

func (c *Config) missingTomlField(_ reflect.Type, key string) error {
	switch key {
	case "alias", "carbon2_format", "carbon2_sanitize_replace_char", "collectd_auth_file",
		"collectd_parse_multivalue", "collectd_security_level", "collectd_typesdb", "collection_jitter",
		"collection_offset",
		"data_format", "data_type", "delay", "drop", "drop_original", "dropwizard_metric_registry_path",
		"dropwizard_tag_paths", "dropwizard_tags_path", "dropwizard_time_format", "dropwizard_time_path",
		"fielddrop", "fieldpass", "flush_interval", "flush_jitter", "form_urlencoded_tag_keys",
		"grace", "graphite_separator", "graphite_tag_sanitize_mode", "graphite_tag_support",
		"grok_custom_pattern_files", "grok_custom_patterns", "grok_named_patterns", "grok_patterns",
		"grok_timezone", "grok_unique_timestamp", "influx_max_line_bytes", "influx_parser_type", "influx_sort_fields",
		"influx_uint_support", "interval", "json_name_key", "json_query", "json_strict",
		"json_string_fields", "json_time_format", "json_time_key", "json_timestamp_format", "json_timestamp_units", "json_timezone", "json_v2",
		"lvm", "metric_batch_size", "metric_buffer_limit", "name_override", "name_prefix",
		"name_suffix", "namedrop", "namepass", "order", "pass", "period", "precision",
		"prefix", "prometheus_export_timestamp", "prometheus_ignore_timestamp", "prometheus_sort_metrics", "prometheus_string_as_label",
		"separator", "splunkmetric_hec_routing", "splunkmetric_multimetric", "tag_keys",
		"tagdrop", "tagexclude", "taginclude", "tagpass", "tags", "template", "templates",
		"value_field_name", "wavefront_source_override", "wavefront_use_strict", "wavefront_disable_prefix_conversion",
		"xml", "xpath", "xpath_json", "xpath_msgpack", "xpath_protobuf", "xpath_print_document",
		"xpath_protobuf_file", "xpath_protobuf_type", "xpath_protobuf_import_paths":

		// ignore fields that are common to all plugins.
	default:
		c.UnusedFields[key] = true
	}
	return nil
}

func (c *Config) setLocalMissingTomlFieldTracker(counter map[string]int) {
	f := func(_ reflect.Type, key string) error {
		if c, ok := counter[key]; ok {
			counter[key] = c + 1
		} else {
			counter[key] = 1
		}
		return nil
	}
	c.toml.MissingField = f
}

func (c *Config) resetMissingTomlFieldTracker() {
	c.toml.MissingField = c.missingTomlField
}

func (c *Config) getFieldString(tbl *ast.Table, fieldName string, target *string) {
	if node, ok := tbl.Fields[fieldName]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				*target = str.Value
			}
		}
	}
}

func (c *Config) getFieldDuration(tbl *ast.Table, fieldName string, target interface{}) {
	if node, ok := tbl.Fields[fieldName]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				d, err := time.ParseDuration(str.Value)
				if err != nil {
					c.addError(tbl, fmt.Errorf("error parsing duration: %w", err))
					return
				}
				targetVal := reflect.ValueOf(target).Elem()
				targetVal.Set(reflect.ValueOf(d))
			}
		}
	}
}

func (c *Config) getFieldBool(tbl *ast.Table, fieldName string, target *bool) {
	var err error
	if node, ok := tbl.Fields[fieldName]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			switch t := kv.Value.(type) {
			case *ast.Boolean:
				*target, err = t.Boolean()
				if err != nil {
					c.addError(tbl, fmt.Errorf("unknown boolean value type %q, expecting boolean", kv.Value))
					return
				}
			case *ast.String:
				*target, err = strconv.ParseBool(t.Value)
				if err != nil {
					c.addError(tbl, fmt.Errorf("unknown boolean value type %q, expecting boolean", kv.Value))
					return
				}
			default:
				c.addError(tbl, fmt.Errorf("unknown boolean value type %q, expecting boolean", kv.Value.Source()))
				return
			}
		}
	}
}

func (c *Config) getFieldInt(tbl *ast.Table, fieldName string, target *int) {
	if node, ok := tbl.Fields[fieldName]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if iAst, ok := kv.Value.(*ast.Integer); ok {
				i, err := iAst.Int()
				if err != nil {
					c.addError(tbl, fmt.Errorf("unexpected int type %q, expecting int", iAst.Value))
					return
				}
				*target = int(i)
			}
		}
	}
}

func (c *Config) getFieldInt64(tbl *ast.Table, fieldName string, target *int64) {
	if node, ok := tbl.Fields[fieldName]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if iAst, ok := kv.Value.(*ast.Integer); ok {
				i, err := iAst.Int()
				if err != nil {
					c.addError(tbl, fmt.Errorf("unexpected int type %q, expecting int", iAst.Value))
					return
				}
				*target = i
			}
		}
	}
}

func (c *Config) getFieldStringSlice(tbl *ast.Table, fieldName string, target *[]string) {
	if node, ok := tbl.Fields[fieldName]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			ary, ok := kv.Value.(*ast.Array)
			if !ok {
				c.addError(tbl, fmt.Errorf("found unexpected format while parsing %q, expecting string array/slice format", fieldName))
				return
			}
			for _, elem := range ary.Value {
				if str, ok := elem.(*ast.String); ok {
					*target = append(*target, str.Value)
				}
			}
		}
	}
}

func (c *Config) getFieldTagFilter(tbl *ast.Table, fieldName string, target *[]models.TagFilter) {
	if node, ok := tbl.Fields[fieldName]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			for name, val := range subtbl.Fields {
				if kv, ok := val.(*ast.KeyValue); ok {
					ary, ok := kv.Value.(*ast.Array)
					if !ok {
						c.addError(tbl, fmt.Errorf("found unexpected format while parsing %q, expecting string array/slice format on each entry", fieldName))
						return
					}

					tagFilter := models.TagFilter{Name: name}
					for _, elem := range ary.Value {
						if str, ok := elem.(*ast.String); ok {
							tagFilter.Filter = append(tagFilter.Filter, str.Value)
						}
					}
					*target = append(*target, tagFilter)
				}
			}
		}
	}
}

func (c *Config) getFieldStringMap(tbl *ast.Table, fieldName string, target *map[string]string) {
	*target = map[string]string{}
	if node, ok := tbl.Fields[fieldName]; ok {
		if subtbl, ok := node.(*ast.Table); ok {
			for name, val := range subtbl.Fields {
				if kv, ok := val.(*ast.KeyValue); ok {
					if str, ok := kv.Value.(*ast.String); ok {
						(*target)[name] = str.Value
					}
				}
			}
		}
	}
}

func keys(m map[string]bool) []string {
	result := []string{}
	for k := range m {
		result = append(result, k)
	}
	return result
}

func (c *Config) hasErrs() bool {
	return len(c.errs) > 0
}

func (c *Config) firstErr() error {
	if len(c.errs) == 0 {
		return nil
	}
	return c.errs[0]
}

func (c *Config) addError(tbl *ast.Table, err error) {
	c.errs = append(c.errs, fmt.Errorf("line %d:%d: %w", tbl.Line, tbl.Position, err))
}

// unwrappable lets you retrieve the original telegraf.Processor from the
// StreamingProcessor. This is necessary because the toml Unmarshaller won't
// look inside composed types.
type unwrappable interface {
	Unwrap() telegraf.Processor
}

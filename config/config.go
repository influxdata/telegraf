package config

import (
	"bytes"
	"crypto/tls"
	"errors"
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
	"sync"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/serializers"
)

var (
	// envVarRe is a regex to find environment variables in the config file
	envVarRe = regexp.MustCompile(`\${(\w+)}|\$(\w+)`)

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
	toml              *toml.Config
	errs              []error // config load errors.
	UnusedFields      map[string]bool
	unusedFieldsMutex *sync.Mutex

	Tags          map[string]string
	InputFilters  []string
	OutputFilters []string

	Agent       *AgentConfig
	Inputs      []*models.RunningInput
	Outputs     []*models.RunningOutput
	Aggregators []*models.RunningAggregator
	// Processors have a slice wrapper type because they need to be sorted
	Processors    models.RunningProcessors
	AggProcessors models.RunningProcessors
	// Parsers are created by their inputs during gather. Config doesn't keep track of them
	// like the other plugins because they need to be garbage collected (See issue #11809)

	Deprecations map[string][]int64
	version      *semver.Version
}

// NewConfig creates a new struct to hold the Telegraf config.
// For historical reasons, It holds the actual instances of the running plugins
// once the configuration is parsed.
func NewConfig() *Config {
	c := &Config{
		UnusedFields:      map[string]bool{},
		unusedFieldsMutex: &sync.Mutex{},

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
		Processors:    make([]*models.RunningProcessor, 0),
		AggProcessors: make([]*models.RunningProcessor, 0),
		InputFilters:  make([]string, 0),
		OutputFilters: make([]string, 0),
		Deprecations:  make(map[string][]int64),
	}

	// Handle unknown version
	version := internal.Version
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
	// By default, or when set to "0s", precision will be set to the same
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
	// This can be used to avoid many plugins querying constraint devices
	// at the same time by manually scheduling them in time.
	CollectionOffset Duration

	// FlushInterval is the Interval at which to flush data
	FlushInterval Duration

	// FlushJitter Jitters the flush interval by a random amount.
	// This is primarily to avoid large write spikes for users running a large
	// number of telegraf instances.
	// ie, a jitter of 5s and interval 10s means flushes will happen every 10-15s
	FlushJitter Duration

	// MetricBatchSize is the maximum number of metrics that is written to an
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

func sliceContains(name string, list []string) bool {
	for _, b := range list {
		if b == name {
			return true
		}
	}
	return false
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
//  1. $TELEGRAF_CONFIG_PATH
//  2. $HOME/.telegraf/telegraf.conf
//  3. /etc/telegraf/telegraf.conf
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
	return "", fmt.Errorf("no config file specified, and could not find one"+
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
	data, err := LoadConfigFile(path)
	if err != nil {
		return fmt.Errorf("error loading config file %s: %w", path, err)
	}

	if err = c.LoadConfigData(data); err != nil {
		return fmt.Errorf("error loading config file %s: %w", path, err)
	}
	return nil
}

// LoadConfigData loads TOML-formatted config data
func (c *Config) LoadConfigData(data []byte) error {
	tbl, err := parseConfig(data)
	if err != nil {
		return fmt.Errorf("error parsing data: %s", err)
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
					return fmt.Errorf("unsupported config format: %s",
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
					return fmt.Errorf("unsupported config format: %s",
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
							return fmt.Errorf("error parsing %s, %s", pluginName, err)
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
		// Assume it's an input for legacy config file support if no other
		// identifiers are present
		default:
			if err = c.addInput(name, subTable); err != nil {
				return fmt.Errorf("error parsing %s, %s", name, err)
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

func LoadConfigFile(config string) ([]byte, error) {
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
	buffer, err := os.ReadFile(config)
	if err != nil {
		return nil, err
	}

	mimeType := http.DetectContentType(buffer)
	if !strings.Contains(mimeType, "text/plain") {
		return nil, fmt.Errorf("provided config is not a TOML file: %s", config)
	}

	return buffer, nil
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
		body, err, retry := func() ([]byte, error, bool) {
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("retry %d of %d failed connecting to HTTP config server %s", i, retries, err), false
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				if i < retries {
					log.Printf("Error getting HTTP config.  Retry %d of %d in %s.  Status=%d", i, retries, httpLoadConfigRetryInterval, resp.StatusCode)
					return nil, nil, true
				}
				return nil, fmt.Errorf("retry %d of %d failed to retrieve remote config: %s", i, retries, resp.Status), false
			}
			body, err := io.ReadAll(resp.Body)
			return body, err, false
		}()

		if err != nil {
			return nil, err
		}

		if retry {
			time.Sleep(httpLoadConfigRetryInterval)
			continue
		}

		return body, err
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
		return fmt.Errorf("undefined but requested aggregator: %s", name)
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

func (c *Config) probeParser(parentcategory string, parentname string, table *ast.Table) bool {
	var dataformat string
	c.getFieldString(table, "data_format", &dataformat)
	if dataformat == "" {
		dataformat = setDefaultParser(parentcategory, parentname)
	}

	creator, ok := parsers.Parsers[dataformat]
	if !ok {
		return false
	}

	// Try to parse the options to detect if any of them is misspelled
	// We don't actually use the parser, so no need to check the error.
	parser := creator("")
	_ = c.toml.UnmarshalTable(table, parser)

	return true
}

func (c *Config) addParser(parentcategory, parentname string, table *ast.Table) (*models.RunningParser, error) {
	var dataformat string
	c.getFieldString(table, "data_format", &dataformat)
	if dataformat == "" {
		dataformat = setDefaultParser(parentcategory, parentname)
	}

	var influxParserType string
	c.getFieldString(table, "influx_parser_type", &influxParserType)
	if dataformat == "influx" && influxParserType == "upstream" {
		dataformat = "influx_upstream"
	}

	creator, ok := parsers.Parsers[dataformat]
	if !ok {
		return nil, fmt.Errorf("undefined but requested parser: %s", dataformat)
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
	err = running.Init()
	return running, err
}

func (c *Config) addProcessor(name string, table *ast.Table) error {
	creator, ok := processors.Processors[name]
	if !ok {
		// Handle removed, deprecated plugins
		if di, deprecated := processors.Deprecations[name]; deprecated {
			printHistoricPluginDeprecationNotice("processors", name, di)
			return fmt.Errorf("plugin deprecated")
		}
		return fmt.Errorf("undefined but requested processor: %s", name)
	}

	// For processors with parsers we need to compute the set of
	// options that is not covered by both, the parser and the processor.
	// We achieve this by keeping a local book of missing entries
	// that counts the number of misses. In case we have a parser
	// for the input both need to miss the entry. We count the
	// missing entries at the end.
	missCount := make(map[string]int)
	missCountThreshold := 0
	c.setLocalMissingTomlFieldTracker(missCount)
	defer c.resetMissingTomlFieldTracker()

	processorConfig, err := c.buildProcessor(name, table)
	if err != nil {
		return err
	}

	// Setup the processor running before the aggregators
	processorBefore, hasParser, err := c.setupProcessor(processorConfig.Name, creator, table)
	if err != nil {
		return err
	}
	rf := models.NewRunningProcessor(processorBefore, processorConfig)
	c.Processors = append(c.Processors, rf)

	// Setup another (new) processor instance running after the aggregator
	processorAfter, _, err := c.setupProcessor(processorConfig.Name, creator, table)
	if err != nil {
		return err
	}
	rf = models.NewRunningProcessor(processorAfter, processorConfig)
	c.AggProcessors = append(c.AggProcessors, rf)

	// Check the number of misses against the threshold
	if hasParser {
		missCountThreshold = 2
	}
	for key, count := range missCount {
		if count <= missCountThreshold {
			continue
		}
		if err := c.missingTomlField(nil, key); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) setupProcessor(name string, creator processors.StreamingCreator, table *ast.Table) (telegraf.StreamingProcessor, bool, error) {
	var hasParser bool

	streamingProcessor := creator()

	var processor interface{}
	if p, ok := streamingProcessor.(unwrappable); ok {
		processor = p.Unwrap()
	} else {
		processor = streamingProcessor
	}

	// If the (underlying) processor has a SetParser or SetParserFunc function,
	// it can accept arbitrary data-formats, so build the requested parser and
	// set it.
	if t, ok := processor.(telegraf.ParserPlugin); ok {
		parser, err := c.addParser("processors", name, table)
		if err != nil {
			return nil, true, fmt.Errorf("adding parser failed: %w", err)
		}
		t.SetParser(parser)
		hasParser = true
	}

	if t, ok := processor.(telegraf.ParserFuncPlugin); ok {
		if !c.probeParser("processors", name, table) {
			return nil, false, errors.New("parser not found")
		}
		t.SetParserFunc(func() (telegraf.Parser, error) {
			return c.addParser("processors", name, table)
		})
		hasParser = true
	}

	if err := c.toml.UnmarshalTable(table, processor); err != nil {
		return nil, hasParser, fmt.Errorf("unmarshalling failed: %w", err)
	}

	err := c.printUserDeprecation("processors", name, processor)
	return streamingProcessor, hasParser, err
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
	missCount := make(map[string]int)
	missCountThreshold := 0
	c.setLocalMissingTomlFieldTracker(missCount)
	defer c.resetMissingTomlFieldTracker()

	creator, ok := inputs.Inputs[name]
	if !ok {
		// Handle removed, deprecated plugins
		if di, deprecated := inputs.Deprecations[name]; deprecated {
			printHistoricPluginDeprecationNotice("inputs", name, di)
			return fmt.Errorf("plugin deprecated")
		}

		return fmt.Errorf("undefined but requested input: %s", name)
	}
	input := creator()

	// If the input has a SetParser or SetParserFunc function, it can accept
	// arbitrary data-formats, so build the requested parser and set it.
	if t, ok := input.(telegraf.ParserPlugin); ok {
		missCountThreshold = 1
		parser, err := c.addParser("inputs", name, table)
		if err != nil {
			return fmt.Errorf("adding parser failed: %w", err)
		}
		t.SetParser(parser)
	}

	// Keep the old interface for backward compatibility
	if t, ok := input.(parsers.ParserInput); ok {
		// DEPRECATED: Please switch your plugin to telegraf.ParserPlugin.
		missCountThreshold = 1
		parser, err := c.addParser("inputs", name, table)
		if err != nil {
			return fmt.Errorf("adding parser failed: %w", err)
		}
		t.SetParser(parser)
	}

	if t, ok := input.(telegraf.ParserFuncPlugin); ok {
		missCountThreshold = 1
		if !c.probeParser("inputs", name, table) {
			return errors.New("parser not found")
		}
		t.SetParserFunc(func() (telegraf.Parser, error) {
			return c.addParser("inputs", name, table)
		})
	}

	if t, ok := input.(parsers.ParserFuncInput); ok {
		// DEPRECATED: Please switch your plugin to telegraf.ParserFuncPlugin.
		missCountThreshold = 1
		if !c.probeParser("inputs", name, table) {
			return errors.New("parser not found")
		}
		t.SetParserFunc(func() (parsers.Parser, error) {
			return c.addParser("inputs", name, table)
		})
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
		if count <= missCountThreshold {
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

	c.getFieldTagFilter(tbl, "tagpass", &f.TagPassFilters)
	c.getFieldTagFilter(tbl, "tagdrop", &f.TagDropFilters)

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
	c.getFieldBool(tbl, "csv_column_prefix", &sc.CSVPrefix)
	c.getFieldBool(tbl, "csv_header", &sc.CSVHeader)
	c.getFieldString(tbl, "csv_separator", &sc.CSVSeparator)
	c.getFieldString(tbl, "csv_timestamp_format", &sc.TimestampFormat)
	c.getFieldInt(tbl, "influx_max_line_bytes", &sc.InfluxMaxLineBytes)
	c.getFieldBool(tbl, "influx_sort_fields", &sc.InfluxSortFields)
	c.getFieldBool(tbl, "influx_uint_support", &sc.InfluxUintSupport)
	c.getFieldBool(tbl, "graphite_tag_support", &sc.GraphiteTagSupport)
	c.getFieldString(tbl, "graphite_tag_sanitize_mode", &sc.GraphiteTagSanitizeMode)

	c.getFieldString(tbl, "graphite_separator", &sc.GraphiteSeparator)

	c.getFieldDuration(tbl, "json_timestamp_units", &sc.TimestampUnits)
	c.getFieldString(tbl, "json_timestamp_format", &sc.TimestampFormat)
	c.getFieldString(tbl, "json_transformation", &sc.Transformation)

	c.getFieldBool(tbl, "splunkmetric_hec_routing", &sc.HecRouting)
	c.getFieldBool(tbl, "splunkmetric_multimetric", &sc.SplunkmetricMultiMetric)

	c.getFieldStringSlice(tbl, "wavefront_source_override", &sc.WavefrontSourceOverride)
	c.getFieldBool(tbl, "wavefront_use_strict", &sc.WavefrontUseStrict)
	c.getFieldBool(tbl, "wavefront_disable_prefix_conversion", &sc.WavefrontDisablePrefixConversion)

	c.getFieldBool(tbl, "prometheus_export_timestamp", &sc.PrometheusExportTimestamp)
	c.getFieldBool(tbl, "prometheus_sort_metrics", &sc.PrometheusSortMetrics)
	c.getFieldBool(tbl, "prometheus_string_as_label", &sc.PrometheusStringAsLabel)
	c.getFieldBool(tbl, "prometheus_compact_encoding", &sc.PrometheusCompactEncoding)

	if c.hasErrs() {
		return nil, c.firstErr()
	}

	return serializers.NewSerializer(sc)
}

// buildOutput parses output specific items from the ast.Table,
// builds the filter and returns a
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
	// General options to ignore
	case "alias",
		"collection_jitter", "collection_offset",
		"data_format", "delay", "drop", "drop_original",
		"fielddrop", "fieldpass", "flush_interval", "flush_jitter",
		"grace",
		"interval",
		"lvm", // What is this used for?
		"metric_batch_size", "metric_buffer_limit",
		"name_override", "name_prefix", "name_suffix", "namedrop", "namepass",
		"order",
		"pass", "period", "precision",
		"tagdrop", "tagexclude", "taginclude", "tagpass", "tags":

	// Parser options to ignore
	case "data_type", "influx_parser_type":

	// Serializer options to ignore
	case "prefix", "template", "templates",
		"carbon2_format", "carbon2_sanitize_replace_char",
		"csv_column_prefix", "csv_header", "csv_separator", "csv_timestamp_format",
		"graphite_tag_sanitize_mode", "graphite_tag_support", "graphite_separator",
		"influx_max_line_bytes", "influx_sort_fields", "influx_uint_support",
		"json_timestamp_format", "json_timestamp_units", "json_transformation",
		"prometheus_export_timestamp", "prometheus_sort_metrics", "prometheus_string_as_label",
		"prometheus_compact_encoding",
		"splunkmetric_hec_routing", "splunkmetric_multimetric",
		"wavefront_disable_prefix_conversion", "wavefront_source_override", "wavefront_use_strict":
	default:
		c.unusedFieldsMutex.Lock()
		c.UnusedFields[key] = true
		c.unusedFieldsMutex.Unlock()
	}
	return nil
}

func (c *Config) setLocalMissingTomlFieldTracker(counter map[string]int) {
	f := func(t reflect.Type, key string) error {
		// Check if we are in a root element that might share options among
		// each other. Those root elements are plugins of all types.
		// All other elements are subtables of their respective plugin and
		// should just be hit once anyway. Therefore, we mark them with a
		// high number to handle them correctly later.
		pt := reflect.PtrTo(t)
		root := pt.Implements(reflect.TypeOf((*telegraf.Input)(nil)).Elem())
		root = root || pt.Implements(reflect.TypeOf((*telegraf.ServiceInput)(nil)).Elem())
		root = root || pt.Implements(reflect.TypeOf((*telegraf.Output)(nil)).Elem())
		root = root || pt.Implements(reflect.TypeOf((*telegraf.Aggregator)(nil)).Elem())
		root = root || pt.Implements(reflect.TypeOf((*telegraf.Processor)(nil)).Elem())
		root = root || pt.Implements(reflect.TypeOf((*telegraf.Parser)(nil)).Elem())

		c, ok := counter[key]
		if !root {
			counter[key] = 100
		} else if !ok {
			counter[key] = 1
		} else {
			counter[key] = c + 1
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
							tagFilter.Values = append(tagFilter.Values, str.Value)
						}
					}
					*target = append(*target, tagFilter)
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

func setDefaultParser(category string, name string) string {
	// Legacy support, exec plugin originally parsed JSON by default.
	if category == "inputs" && name == "exec" {
		return "json"
	}

	return "influx"
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

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

	"github.com/compose-spec/compose-go/template"
	"github.com/compose-spec/compose-go/utils"
	"github.com/coreos/go-semver/semver"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/persister"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/csv"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/secretstores"
	"github.com/influxdata/telegraf/plugins/serializers"
)

var (
	httpLoadConfigRetryInterval = 10 * time.Second

	// fetchURLRe is a regex to determine whether the requested file should
	// be fetched from a remote or read from the filesystem.
	fetchURLRe = regexp.MustCompile(`^\w+://`)

	// oldVarRe is a regex to reproduce pre v1.27.0 environment variable
	// replacement behavior
	oldVarRe = regexp.MustCompile(`\$(?i:(?P<named>[_a-z][_a-z0-9]*)|{(?:(?P<braced>[_a-z][_a-z0-9]*(?::?[-+?](.*))?)}|(?P<invalid>)))`)
	// OldEnvVarReplacement is a switch to allow going back to pre v1.27.0
	// environment variable replacement behavior
	OldEnvVarReplacement = false

	// Password specified via command-line
	Password Secret
)

// Config specifies the URL/user/password for the database that telegraf
// will be logging to, as well as all the plugins that the user has
// specified
type Config struct {
	toml              *toml.Config
	errs              []error // config load errors.
	UnusedFields      map[string]bool
	unusedFieldsMutex *sync.Mutex

	Tags               map[string]string
	InputFilters       []string
	OutputFilters      []string
	SecretStoreFilters []string

	SecretStores map[string]telegraf.SecretStore

	Agent       *AgentConfig
	Inputs      []*models.RunningInput
	Outputs     []*models.RunningOutput
	Aggregators []*models.RunningAggregator
	// Processors have a slice wrapper type because they need to be sorted
	Processors        models.RunningProcessors
	AggProcessors     models.RunningProcessors
	fileProcessors    OrderedPlugins
	fileAggProcessors OrderedPlugins

	// Parsers are created by their inputs during gather. Config doesn't keep track of them
	// like the other plugins because they need to be garbage collected (See issue #11809)

	Deprecations map[string][]int64
	version      *semver.Version

	Persister *persister.Persister

	NumberSecrets uint64
}

// Ordered plugins used to keep the order in which they appear in a file
type OrderedPlugin struct {
	Line   int
	plugin any
}
type OrderedPlugins []*OrderedPlugin

func (op OrderedPlugins) Len() int           { return len(op) }
func (op OrderedPlugins) Swap(i, j int)      { op[i], op[j] = op[j], op[i] }
func (op OrderedPlugins) Less(i, j int) bool { return op[i].Line < op[j].Line }

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

		Tags:               make(map[string]string),
		Inputs:             make([]*models.RunningInput, 0),
		Outputs:            make([]*models.RunningOutput, 0),
		Processors:         make([]*models.RunningProcessor, 0),
		AggProcessors:      make([]*models.RunningProcessor, 0),
		SecretStores:       make(map[string]telegraf.SecretStore),
		fileProcessors:     make([]*OrderedPlugin, 0),
		fileAggProcessors:  make([]*OrderedPlugin, 0),
		InputFilters:       make([]string, 0),
		OutputFilters:      make([]string, 0),
		SecretStoreFilters: make([]string, 0),
		Deprecations:       make(map[string][]int64),
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
	FlushBufferWhenFull bool `toml:"flush_buffer_when_full" deprecated:"0.13.0;1.30.0;option is ignored"`

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

	// Name of the file to load the state of plugins from and store the state to.
	// If uncommented and not empty, this file will be used to save the state of
	// stateful plugins on termination of Telegraf. If the file exists on start,
	// the state in the file will be restored for the plugins.
	Statefile string `toml:"statefile"`

	// Flag to always keep tags explicitly defined in the plugin itself and
	// ensure those tags always pass filtering.
	AlwaysIncludeLocalTags bool `toml:"always_include_local_tags"`

	// Flag to always keep tags explicitly defined in the global tags section
	// and ensure those tags always pass filtering.
	AlwaysIncludeGlobalTags bool `toml:"always_include_global_tags"`
}

// InputNames returns a list of strings of the configured inputs.
func (c *Config) InputNames() []string {
	name := make([]string, 0, len(c.Inputs))
	for _, input := range c.Inputs {
		name = append(name, input.Config.Name)
	}
	return PluginNameCounts(name)
}

// AggregatorNames returns a list of strings of the configured aggregators.
func (c *Config) AggregatorNames() []string {
	name := make([]string, 0, len(c.Aggregators))
	for _, aggregator := range c.Aggregators {
		name = append(name, aggregator.Config.Name)
	}
	return PluginNameCounts(name)
}

// ProcessorNames returns a list of strings of the configured processors.
func (c *Config) ProcessorNames() []string {
	name := make([]string, 0, len(c.Processors))
	for _, processor := range c.Processors {
		name = append(name, processor.Config.Name)
	}
	return PluginNameCounts(name)
}

// OutputNames returns a list of strings of the configured outputs.
func (c *Config) OutputNames() []string {
	name := make([]string, 0, len(c.Outputs))
	for _, output := range c.Outputs {
		name = append(name, output.Config.Name)
	}
	return PluginNameCounts(name)
}

// SecretstoreNames returns a list of strings of the configured secret-stores.
func (c *Config) SecretstoreNames() []string {
	names := make([]string, 0, len(c.SecretStores))
	for name := range c.SecretStores {
		names = append(names, name)
	}
	return PluginNameCounts(names)
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
	tags := make([]string, 0, len(c.Tags))
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

// WalkDirectory collects all toml files that need to be loaded
func WalkDirectory(path string) ([]string, error) {
	var files []string
	walkfn := func(thispath string, info os.FileInfo, _ error) error {
		if info == nil {
			log.Printf("W! Telegraf is not permitted to read %s", thispath)
			return nil
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), "..") {
				// skip Kubernetes mounts, preventing loading the same config twice
				return filepath.SkipDir
			}

			return nil
		}
		name := info.Name()
		if len(name) < 6 || name[len(name)-5:] != ".conf" {
			return nil
		}
		files = append(files, thispath)
		return nil
	}
	return files, filepath.Walk(path, walkfn)
}

// Try to find a default config file at these locations (in order):
//  1. $TELEGRAF_CONFIG_PATH
//  2. $HOME/.telegraf/telegraf.conf
//  3. /etc/telegraf/telegraf.conf and /etc/telegraf/telegraf.d/*.conf
func GetDefaultConfigPath() ([]string, error) {
	envfile := os.Getenv("TELEGRAF_CONFIG_PATH")
	homefile := os.ExpandEnv("${HOME}/.telegraf/telegraf.conf")
	etcfile := "/etc/telegraf/telegraf.conf"
	etcfolder := "/etc/telegraf/telegraf.d"

	if runtime.GOOS == "windows" {
		programFiles := os.Getenv("ProgramFiles")
		if programFiles == "" { // Should never happen
			programFiles = `C:\Program Files`
		}
		etcfile = programFiles + `\Telegraf\telegraf.conf`
		etcfolder = programFiles + `\Telegraf\telegraf.d\`
	}

	for _, path := range []string{envfile, homefile} {
		if isURL(path) {
			return []string{path}, nil
		}
		if _, err := os.Stat(path); err == nil {
			return []string{path}, nil
		}
	}

	// At this point we need to check if the files under /etc/telegraf are
	// populated and return them all.
	confFiles := []string{}
	if _, err := os.Stat(etcfile); err == nil {
		confFiles = append(confFiles, etcfile)
	}
	if _, err := os.Stat(etcfolder); err == nil {
		files, err := WalkDirectory(etcfolder)
		if err != nil {
			log.Printf("W! unable walk %q: %s", etcfolder, err)
		}
		confFiles = append(confFiles, files...)
	}
	if len(confFiles) > 0 {
		return confFiles, nil
	}

	// if we got here, we didn't find a file in a default location
	return nil, fmt.Errorf("no config file specified, and could not find one"+
		" in $TELEGRAF_CONFIG_PATH, %s, %s, or %s/*.conf", homefile, etcfile, etcfolder)
}

// isURL checks if string is valid url
func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// LoadConfig loads the given config files and applies it to c
func (c *Config) LoadConfig(path string) error {
	var err error
	paths := []string{}

	if path == "" {
		if paths, err = GetDefaultConfigPath(); err != nil {
			return err
		}
	} else {
		paths = append(paths, path)
	}

	for _, path := range paths {
		if !c.Agent.Quiet {
			log.Printf("I! Loading config: %s", path)
		}

		data, _, err := LoadConfigFile(path)
		if err != nil {
			return fmt.Errorf("error loading config file %s: %w", path, err)
		}

		if err = c.LoadConfigData(data); err != nil {
			return fmt.Errorf("error loading config file %s: %w", path, err)
		}
	}

	return nil
}

func (c *Config) LoadAll(configFiles ...string) error {
	for _, fConfig := range configFiles {
		if err := c.LoadConfig(fConfig); err != nil {
			return err
		}
	}

	// Sort the processors according to their `order` setting while
	// using a stable sort to keep the file loading / file position order.
	sort.Stable(c.Processors)
	sort.Stable(c.AggProcessors)

	// Set snmp agent translator default
	if c.Agent.SnmpTranslator == "" {
		c.Agent.SnmpTranslator = "netsnmp"
	}

	// Check if there is enough lockable memory for the secret
	c.NumberSecrets = uint64(secretCount.Load())

	// Let's link all secrets to their secret-stores
	return c.LinkSecrets()
}

// LoadConfigData loads TOML-formatted config data
func (c *Config) LoadConfigData(data []byte) error {
	tbl, err := parseConfig(data)
	if err != nil {
		return fmt.Errorf("error parsing data: %w", err)
	}

	// Parse tags tables first:
	for _, tableName := range []string{"tags", "global_tags"} {
		if val, ok := tbl.Fields[tableName]; ok {
			subTable, ok := val.(*ast.Table)
			if !ok {
				return fmt.Errorf("invalid configuration, bad table name %q", tableName)
			}
			if err = c.toml.UnmarshalTable(subTable, c.Tags); err != nil {
				return fmt.Errorf("error parsing table name %q: %w", tableName, err)
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

	// Warn when explicitly setting the old snmp translator
	if c.Agent.SnmpTranslator == "netsnmp" {
		models.PrintOptionValueDeprecationNotice(telegraf.Warn, "agent", "snmp_translator", "netsnmp", telegraf.DeprecationInfo{
			Since:     "1.25.0",
			RemovalIn: "2.0.0",
			Notice:    "Use 'gosmi' instead",
		})
	}

	// Setup the persister if requested
	if c.Agent.Statefile != "" {
		c.Persister = &persister.Persister{
			Filename: c.Agent.Statefile,
		}
	}

	if len(c.UnusedFields) > 0 {
		return fmt.Errorf("line %d: configuration specified the fields %q, but they weren't used", tbl.Line, keys(c.UnusedFields))
	}

	// Initialize the file-sorting slices
	c.fileProcessors = make(OrderedPlugins, 0)
	c.fileAggProcessors = make(OrderedPlugins, 0)

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
					return fmt.Errorf("plugin %s.%s: line %d: configuration specified the fields %q, but they weren't used",
						name, pluginName, subTable.Line, keys(c.UnusedFields))
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
					return fmt.Errorf("plugin %s.%s: line %d: configuration specified the fields %q, but they weren't used",
						name, pluginName, subTable.Line, keys(c.UnusedFields))
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
					return fmt.Errorf(
						"plugin %s.%s: line %d: configuration specified the fields %q, but they weren't used",
						name,
						pluginName,
						subTable.Line,
						keys(c.UnusedFields),
					)
				}
			}
		case "aggregators":
			for pluginName, pluginVal := range subTable.Fields {
				switch pluginSubTable := pluginVal.(type) {
				case []*ast.Table:
					for _, t := range pluginSubTable {
						if err = c.addAggregator(pluginName, t); err != nil {
							return fmt.Errorf("error parsing %s, %w", pluginName, err)
						}
					}
				default:
					return fmt.Errorf("unsupported config format: %s",
						pluginName)
				}
				if len(c.UnusedFields) > 0 {
					return fmt.Errorf("plugin %s.%s: line %d: configuration specified the fields %q, but they weren't used",
						name, pluginName, subTable.Line, keys(c.UnusedFields))
				}
			}
		case "secretstores":
			for pluginName, pluginVal := range subTable.Fields {
				switch pluginSubTable := pluginVal.(type) {
				case []*ast.Table:
					for _, t := range pluginSubTable {
						if err = c.addSecretStore(pluginName, t); err != nil {
							return fmt.Errorf("error parsing %s, %w", pluginName, err)
						}
					}
				default:
					return fmt.Errorf("unsupported config format: %s", pluginName)
				}
				if len(c.UnusedFields) > 0 {
					msg := "plugin %s.%s: line %d: configuration specified the fields %q, but they weren't used"
					return fmt.Errorf(msg, name, pluginName, subTable.Line, keys(c.UnusedFields))
				}
			}

		// Assume it's an input for legacy config file support if no other
		// identifiers are present
		default:
			if err = c.addInput(name, subTable); err != nil {
				return fmt.Errorf("error parsing %s, %w", name, err)
			}
		}
	}

	// Sort the processor according to the order they appeared in this file
	// In a later stage, we sort them using the `order` option.
	sort.Sort(c.fileProcessors)
	for _, op := range c.fileProcessors {
		c.Processors = append(c.Processors, op.plugin.(*models.RunningProcessor))
	}

	sort.Sort(c.fileAggProcessors)
	for _, op := range c.fileAggProcessors {
		c.AggProcessors = append(c.AggProcessors, op.plugin.(*models.RunningProcessor))
	}

	return nil
}

// trimBOM trims the Byte-Order-Marks from the beginning of the file.
// this is for Windows compatibility only.
// see https://github.com/influxdata/telegraf/issues/1378
func trimBOM(f []byte) []byte {
	return bytes.TrimPrefix(f, []byte("\xef\xbb\xbf"))
}

// LoadConfigFile loads the content of a configuration file and returns it
// together with a flag denoting if the file is from a remote location such
// as a web server.
func LoadConfigFile(config string) ([]byte, bool, error) {
	if fetchURLRe.MatchString(config) {
		u, err := url.Parse(config)
		if err != nil {
			return nil, true, err
		}

		switch u.Scheme {
		case "https", "http":
			data, err := fetchConfig(u)
			return data, true, err
		default:
			return nil, true, fmt.Errorf("scheme %q not supported", u.Scheme)
		}
	}

	// If it isn't a https scheme, try it as a file
	buffer, err := os.ReadFile(config)
	if err != nil {
		return nil, false, err
	}

	mimeType := http.DetectContentType(buffer)
	if !strings.Contains(mimeType, "text/plain") {
		return nil, false, fmt.Errorf("provided config is not a TOML file: %s", config)
	}

	return buffer, false, nil
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
				return nil, fmt.Errorf("retry %d of %d failed connecting to HTTP config server: %w", i, retries, err), false
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
	var err error
	contents, err = removeComments(contents)
	if err != nil {
		return nil, err
	}
	outputBytes, err := substituteEnvironment(contents, OldEnvVarReplacement)
	if err != nil {
		return nil, err
	}
	return toml.Parse(outputBytes)
}

func removeComments(contents []byte) ([]byte, error) {
	tomlReader := bytes.NewReader(contents)

	// Initialize variables for tracking state
	var inQuote, inComment, escaped bool
	var quoteChar byte

	// Initialize buffer for modified TOML data
	var output bytes.Buffer

	buf := make([]byte, 1)
	// Iterate over each character in the file
	for {
		_, err := tomlReader.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		char := buf[0]

		// Toggle the escaped state at backslash to we have true every odd occurrence.
		if char == '\\' {
			escaped = !escaped
		}

		if inComment {
			// If we're currently in a comment, check if this character ends the comment
			if char == '\n' {
				// End of line, comment is finished
				inComment = false
				_, _ = output.WriteRune('\n')
			}
		} else if inQuote {
			// If we're currently in a quote, check if this character ends the quote
			if char == quoteChar && !escaped {
				// End of quote, we're no longer in a quote
				inQuote = false
			}
			output.WriteByte(char)
		} else {
			// Not in a comment or a quote
			if (char == '"' || char == '\'') && !escaped {
				// Start of quote
				inQuote = true
				quoteChar = char
				output.WriteByte(char)
			} else if char == '#' && !escaped {
				// Start of comment
				inComment = true
			} else {
				// Not a comment or a quote, just output the character
				output.WriteByte(char)
			}
		}

		// Reset escaping if any other character occurred
		if char != '\\' {
			escaped = false
		}
	}
	return output.Bytes(), nil
}

func substituteEnvironment(contents []byte, oldReplacementBehavior bool) ([]byte, error) {
	options := []template.Option{
		template.WithReplacementFunction(func(s string, m template.Mapping, cfg *template.Config) (string, error) {
			result, applied, err := template.DefaultReplacementAppliedFunc(s, m, cfg)
			if err == nil && !applied {
				// Keep undeclared environment-variable patterns to reproduce
				// pre-v1.27 behavior
				return s, nil
			}
			if err != nil && strings.HasPrefix(err.Error(), "Invalid template:") {
				// Keep invalid template patterns to ignore regexp substitutions
				// like ${1}
				return s, nil
			}
			return result, err
		}),
		template.WithoutLogging,
	}
	if oldReplacementBehavior {
		options = append(options, template.WithPattern(oldVarRe))
	}

	envMap := utils.GetAsEqualsMap(os.Environ())
	retVal, err := template.SubstituteWithOptions(string(contents), func(k string) (string, bool) {
		if v, ok := envMap[k]; ok {
			return v, ok
		}
		return "", false
	}, options...)
	return []byte(retVal), err
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

func (c *Config) addSecretStore(name string, table *ast.Table) error {
	if len(c.SecretStoreFilters) > 0 && !sliceContains(name, c.SecretStoreFilters) {
		return nil
	}

	var storeid string
	c.getFieldString(table, "id", &storeid)
	if storeid == "" {
		return fmt.Errorf("%q secret-store without ID", name)
	}
	if !secretStorePattern.MatchString(storeid) {
		return fmt.Errorf("invalid secret-store ID %q, must only contain letters, numbers or underscore", storeid)
	}

	creator, ok := secretstores.SecretStores[name]
	if !ok {
		// Handle removed, deprecated plugins
		if di, deprecated := secretstores.Deprecations[name]; deprecated {
			printHistoricPluginDeprecationNotice("secretstores", name, di)
			return errors.New("plugin deprecated")
		}
		return fmt.Errorf("undefined but requested secretstores: %s", name)
	}
	store := creator(storeid)

	if err := c.toml.UnmarshalTable(table, store); err != nil {
		return err
	}

	if err := c.printUserDeprecation("secretstores", name, store); err != nil {
		return err
	}

	logger := models.NewLogger("secretstores", name, "")
	models.SetLoggerOnPlugin(store, logger)

	if err := store.Init(); err != nil {
		return fmt.Errorf("error initializing secret-store %q: %w", storeid, err)
	}

	if _, found := c.SecretStores[storeid]; found {
		return fmt.Errorf("duplicate ID %q for secretstore %q", storeid, name)
	}
	c.SecretStores[storeid] = store
	return nil
}

func (c *Config) LinkSecrets() error {
	for _, s := range unlinkedSecrets {
		resolvers := make(map[string]telegraf.ResolveFunc)
		for _, ref := range s.GetUnlinked() {
			// Split the reference and lookup the resolver
			storeid, key := splitLink(ref)
			store, found := c.SecretStores[storeid]
			if !found {
				return fmt.Errorf("unknown secret-store for %q", ref)
			}
			resolver, err := store.GetResolver(key)
			if err != nil {
				return fmt.Errorf("retrieving resolver for %q failed: %w", ref, err)
			}
			resolvers[ref] = resolver
		}
		// Inject the resolver list into the secret
		if err := s.Link(resolvers); err != nil {
			return fmt.Errorf("retrieving resolver failed: %w", err)
		}
	}
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

	// Handle reset-mode of CSV parsers to stay backward compatible (see issue #12022)
	if dataformat == "csv" && parentcategory == "inputs" {
		if parentname == "exec" {
			csvParser := parser.(*csv.Parser)
			csvParser.ResetMode = "always"
		}
	}

	if err := c.toml.UnmarshalTable(table, parser); err != nil {
		return nil, err
	}

	conf := &models.ParserConfig{
		Parent:     parentname,
		DataFormat: dataformat,
	}
	running := models.NewRunningParser(parser, conf)
	err := running.Init()
	return running, err
}

func (c *Config) addSerializer(parentname string, table *ast.Table) (*models.RunningSerializer, error) {
	var dataformat string
	c.getFieldString(table, "data_format", &dataformat)
	if dataformat == "" {
		dataformat = "influx"
	}

	creator, ok := serializers.Serializers[dataformat]
	if !ok {
		return nil, fmt.Errorf("undefined but requested serializer: %s", dataformat)
	}
	serializer := creator()

	if err := c.toml.UnmarshalTable(table, serializer); err != nil {
		return nil, err
	}

	conf := &models.SerializerConfig{
		Parent:     parentname,
		DataFormat: dataformat,
	}
	running := models.NewRunningSerializer(serializer, conf)
	err := running.Init()
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

	// Setup the processor running before the aggregators
	processorBeforeConfig, err := c.buildProcessor("processors", name, table)
	if err != nil {
		return err
	}
	processorBefore, count, err := c.setupProcessor(processorBeforeConfig.Name, creator, table)
	if err != nil {
		return err
	}
	rf := models.NewRunningProcessor(processorBefore, processorBeforeConfig)
	c.fileProcessors = append(c.fileProcessors, &OrderedPlugin{table.Line, rf})

	// Setup another (new) processor instance running after the aggregator
	processorAfterConfig, err := c.buildProcessor("aggprocessors", name, table)
	if err != nil {
		return err
	}
	processorAfter, _, err := c.setupProcessor(processorAfterConfig.Name, creator, table)
	if err != nil {
		return err
	}
	rf = models.NewRunningProcessor(processorAfter, processorAfterConfig)
	c.fileAggProcessors = append(c.fileAggProcessors, &OrderedPlugin{table.Line, rf})

	// Check the number of misses against the threshold. We need to double
	// the count as the processor setup is executed twice.
	missCountThreshold = 2 * count
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

func (c *Config) setupProcessor(name string, creator processors.StreamingCreator, table *ast.Table) (telegraf.StreamingProcessor, int, error) {
	var optionTestCount int

	streamingProcessor := creator()

	var processor interface{}
	if p, ok := streamingProcessor.(processors.HasUnwrap); ok {
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
			return nil, 0, fmt.Errorf("adding parser failed: %w", err)
		}
		t.SetParser(parser)
		optionTestCount++
	}

	if t, ok := processor.(telegraf.ParserFuncPlugin); ok {
		if !c.probeParser("processors", name, table) {
			return nil, 0, errors.New("parser not found")
		}
		t.SetParserFunc(func() (telegraf.Parser, error) {
			return c.addParser("processors", name, table)
		})
		optionTestCount++
	}

	// If the (underlying) processor has a SetSerializer function it can accept
	// arbitrary data-formats, so build the requested serializer and set it.
	if t, ok := processor.(telegraf.SerializerPlugin); ok {
		serializer, err := c.addSerializer(name, table)
		if err != nil {
			return nil, 0, fmt.Errorf("adding serializer failed: %w", err)
		}
		t.SetSerializer(serializer)
		optionTestCount++
	}

	if err := c.toml.UnmarshalTable(table, processor); err != nil {
		return nil, 0, fmt.Errorf("unmarshalling failed: %w", err)
	}

	err := c.printUserDeprecation("processors", name, processor)
	return streamingProcessor, optionTestCount, err
}

func (c *Config) addOutput(name string, table *ast.Table) error {
	if len(c.OutputFilters) > 0 && !sliceContains(name, c.OutputFilters) {
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
	if t, ok := output.(telegraf.SerializerPlugin); ok {
		missThreshold = 1
		serializer, err := c.addSerializer(name, table)
		if err != nil {
			return err
		}
		t.SetSerializer(serializer)
	} else if t, ok := output.(serializers.SerializerOutput); ok {
		// Keep the old interface for backward compatibility
		// DEPRECATED: Please switch your plugin to telegraf.Serializers
		missThreshold = 1
		serializer, err := c.addSerializer(name, table)
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

	// Check the number of misses against the threshold
	for key, count := range missCount {
		if count <= missThreshold {
			continue
		}
		if err := c.missingTomlField(nil, key); err != nil {
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

	if t, ok := input.(telegraf.ParserFuncPlugin); ok {
		missCountThreshold = 1
		if !c.probeParser("inputs", name, table) {
			return errors.New("parser not found")
		}
		t.SetParserFunc(func() (telegraf.Parser, error) {
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

	// Check the number of misses against the threshold
	for key, count := range missCount {
		if count <= missCountThreshold {
			continue
		}
		if err := c.missingTomlField(nil, key); err != nil {
			return err
		}
	}

	rp := models.NewRunningInput(input, pluginConfig)
	rp.SetDefaultTags(c.Tags)
	c.Inputs = append(c.Inputs, rp)

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

	// Generate an ID for the plugin
	conf.ID, err = generatePluginID("aggregators."+name, tbl)
	return conf, err
}

// buildProcessor parses Processor specific items from the ast.Table,
// builds the filter and returns a
// models.ProcessorConfig to be inserted into models.RunningProcessor
func (c *Config) buildProcessor(category, name string, tbl *ast.Table) (*models.ProcessorConfig, error) {
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

	// Generate an ID for the plugin
	conf.ID, err = generatePluginID(category+"."+name, tbl)
	return conf, err
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

	c.getFieldString(tbl, "metricpass", &f.MetricPass)

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
	cp := &models.InputConfig{
		Name:                    name,
		AlwaysIncludeLocalTags:  c.Agent.AlwaysIncludeLocalTags,
		AlwaysIncludeGlobalTags: c.Agent.AlwaysIncludeGlobalTags,
	}
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

	// Generate an ID for the plugin
	cp.ID, err = generatePluginID("inputs."+name, tbl)
	return cp, err
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

	// Generate an ID for the plugin
	oc.ID, err = generatePluginID("outputs."+name, tbl)
	return oc, err
}

func (c *Config) missingTomlField(_ reflect.Type, key string) error {
	switch key {
	// General options to ignore
	case "alias", "always_include_local_tags",
		"collection_jitter", "collection_offset",
		"data_format", "delay", "drop", "drop_original",
		"fielddrop", "fieldpass", "flush_interval", "flush_jitter",
		"grace",
		"interval",
		"lvm", // What is this used for?
		"metric_batch_size", "metric_buffer_limit", "metricpass",
		"name_override", "name_prefix", "name_suffix", "namedrop", "namepass",
		"order",
		"pass", "period", "precision",
		"tagdrop", "tagexclude", "taginclude", "tagpass", "tags":

	// Secret-store options to ignore
	case "id":

	// Parser and serializer options to ignore
	case "data_type", "influx_parser_type":

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
		root = root || pt.Implements(reflect.TypeOf((*telegraf.StreamingProcessor)(nil)).Elem())
		root = root || pt.Implements(reflect.TypeOf((*telegraf.Parser)(nil)).Elem())
		root = root || pt.Implements(reflect.TypeOf((*telegraf.Serializer)(nil)).Elem())

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
			} else {
				c.addError(tbl, fmt.Errorf("found unexpected format while parsing %q, expecting int", fieldName))
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

// +build !solaris
// +build go1.10

package logparser

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	taillog "github.com/sgtsquiggs/tail/logger"
	"github.com/sgtsquiggs/tail/logline"
	"github.com/sgtsquiggs/tail/tailer"
	"github.com/sgtsquiggs/tail/watcher"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	// Parsers
)

const (
	defaultPollInterval = time.Millisecond * 250
	defaultWatchMethod  = "inotify"
)

// LogParser in the primary interface for the plugin
type GrokConfig struct {
	MeasurementName    string `toml:"measurement"`
	Patterns           []string
	NamedPatterns      []string
	CustomPatterns     string
	CustomPatternFiles []string
	Timezone           string
	UniqueTimestamp    string
}

type logEntry struct {
	path string
	line string
}

// LogParserPlugin is the primary struct to implement the interface for logparser plugin
type LogParserPlugin struct {
	Files         []string
	FromBeginning bool
	WatchMethod   string
	PollInterval  internal.Duration

	watcher watcher.Watcher
	tailer  *tailer.Tailer
	lines   chan *logline.LogLine
	files   map[string]bool
	wg      sync.WaitGroup
	acc     telegraf.Accumulator

	sync.Mutex

	GrokParser parsers.Parser
	GrokConfig GrokConfig `toml:"grok"`
}

const sampleConfig = `
  ## Log files to parse.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   /var/log/**.log     -> recursively find all .log files in /var/log
  ##   /var/log/*/*.log    -> find all .log files with a parent dir in /var/log
  ##   /var/log/apache.log -> only tail the apache log file
  files = ["/var/log/apache/access.log"]

  ## Read files that currently exist from the beginning. Files that are created
  ## while telegraf is running (and that match the "files" globs) will always
  ## be read from the beginning.
  from_beginning = false

  ## Method used to watch for file updates.  Can be either "inotify" or "poll".
  # watch_method = "inotify"

  ## Poll interval. Used when watch_method is set to "poll".
  # poll_interval = "250ms"

  ## Parse logstash-style "grok" patterns:
  [inputs.logparser.grok]
    ## This is a list of patterns to check the given log file(s) for.
    ## Note that adding patterns here increases processing time. The most
    ## efficient configuration is to have one pattern per logparser.
    ## Other common built-in patterns are:
    ##   %{COMMON_LOG_FORMAT}   (plain apache & nginx access logs)
    ##   %{COMBINED_LOG_FORMAT} (access logs + referrer & agent)
    patterns = ["%{COMBINED_LOG_FORMAT}"]

    ## Name of the outputted measurement name.
    measurement = "apache_access_log"

    ## Full path(s) to custom pattern files.
    custom_pattern_files = []

    ## Custom patterns can also be defined here. Put one pattern per line.
    custom_patterns = '''
    '''

    ## Timezone allows you to provide an override for timestamps that
    ## don't already include an offset
    ## e.g. 04/06/2016 12:41:45 data one two 5.43µs
    ##
    ## Default: "" which renders UTC
    ## Options are as follows:
    ##   1. Local             -- interpret based on machine localtime
    ##   2. "Canada/Eastern"  -- Unix TZ values like those found in https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
    ##   3. UTC               -- or blank/unspecified, will return timestamp in UTC
    # timezone = "Canada/Eastern"

	## When set to "disable", timestamp will not incremented if there is a
	## duplicate.
    # unique_timestamp = "auto"
`

// SampleConfig returns the sample configuration for the plugin
func (l *LogParserPlugin) SampleConfig() string {
	return sampleConfig
}

// Description returns the human readable description for the plugin
func (l *LogParserPlugin) Description() string {
	return "Stream and parse log file(s)."
}

// Gather is the primary function to collect the metrics for the plugin
func (l *LogParserPlugin) Gather(acc telegraf.Accumulator) error {
	l.Lock()
	defer l.Unlock()

	// always start from the beginning of files that appear while we're running
	return l.tailNewFiles(true)
}

// Start kicks off collection of stats for the plugin
func (l *LogParserPlugin) Start(acc telegraf.Accumulator) error {
	l.Lock()
	defer l.Unlock()

	l.acc = acc
	l.lines = make(chan *logline.LogLine)
	l.files = make(map[string]bool)

	var poll bool
	if l.WatchMethod == "poll" {
		poll = true
	}

	var err error
	l.watcher, err = watcher.NewLogWatcher(l.PollInterval.Duration, !poll, watcher.Logger(taillog.DiscardingLogger))
	if err != nil {
		defer close(l.lines)
		return err
	}

	opts := []tailer.Option{tailer.Logger(taillog.DiscardingLogger)}
	if l.FromBeginning {
		opts = append(opts, tailer.OneShot)
	}

	l.tailer, err = tailer.New(l.lines, l.watcher, opts...)
	if err != nil {
		defer l.watcher.Close()
		defer close(l.lines)
		return err
	}

	mName := "logparser"
	if l.GrokConfig.MeasurementName != "" {
		mName = l.GrokConfig.MeasurementName
	}

	// Looks for fields which implement LogParser interface
	config := &parsers.Config{
		MetricName:             mName,
		GrokPatterns:           l.GrokConfig.Patterns,
		GrokNamedPatterns:      l.GrokConfig.NamedPatterns,
		GrokCustomPatterns:     l.GrokConfig.CustomPatterns,
		GrokCustomPatternFiles: l.GrokConfig.CustomPatternFiles,
		GrokTimezone:           l.GrokConfig.Timezone,
		GrokUniqueTimestamp:    l.GrokConfig.UniqueTimestamp,
		DataFormat:             "grok",
	}

	l.GrokParser, err = parsers.NewParser(config)
	if err != nil {
		return err
	}

	l.wg.Add(1)
	go l.parse()

	return l.tailNewFiles(l.FromBeginning)
}

// check the globs against files on disk, and start tailing any new files.
// Assumes l's lock is held!
func (l *LogParserPlugin) tailNewFiles(fromBeginning bool) error {

	for _, filepath := range l.Files {
		g, err := globpath.Compile(filepath)
		if err != nil {
			l.acc.AddError(fmt.Errorf("E! Error Glob %s failed to compile, %s", filepath, err))
		}
		for _, file := range g.Match() {
			if _, ok := l.files[file]; ok {
				// we're already tailing this file
				continue
			}

			err := l.tailer.TailPath(file)
			if err != nil {
				l.acc.AddError(err)
				continue
			}

			log.Printf("D! [inputs.logparser] tail added for file: %v", file)

			// create a goroutine for each "tailer"
			l.files[file] = true
		}
	}

	return nil
}

// parse is launched as a goroutine to watch the l.lines channel.
// when a line is available, parse parses it and adds the metric(s) to the
// accumulator.
func (l *LogParserPlugin) parse() {
	defer l.wg.Done()

	var m telegraf.Metric
	var err error
	for line := range l.lines {
		// Fix up files with Windows line endings.
		line.Line = strings.TrimRight(line.Line, "\r")
		if line.Line == "" || line.Line == "\n" {
			continue
		}

		m, err = l.GrokParser.ParseLine(line.Line)
		if err == nil {
			if m != nil {
				tags := m.Tags()
				tags["path"] = line.Filename
				l.acc.AddFields(m.Name(), m.Fields(), tags, m.Time())
			}
		} else {
			log.Println("E! Error parsing log line: " + err.Error())
		}
	}
}

// Stop will end the metrics collection process on file tailers
func (l *LogParserPlugin) Stop() {
	l.Lock()
	defer l.Unlock()

	err := l.tailer.Close()
	if err != nil {
		l.acc.AddError(fmt.Errorf("E! Error closing tailer, Error: %s\n", err))
	} else {
		for file := range l.files {
			log.Printf("D! [inputs.logparser] tail removed for file: %v", file)
		}
	}

	l.wg.Wait()
}

func init() {
	inputs.Add("logparser", func() telegraf.Input {
		return &LogParserPlugin{
			WatchMethod:  defaultWatchMethod,
			PollInterval: internal.Duration{Duration: defaultPollInterval},
		}
	})
}

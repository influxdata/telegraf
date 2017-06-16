package logparser

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/influxdata/tail"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/inputs"

	// Parsers
	"github.com/influxdata/telegraf/plugins/inputs/logparser/grok"
)

type LogParser interface {
	ParseLine(line string) (telegraf.Metric, error)
	Compile() error
}

type LogParserPlugin struct {
	Files         []string
	FromBeginning bool

	tailers map[string]*tail.Tail
	lines   chan string
	done    chan struct{}
	wg      sync.WaitGroup
	acc     telegraf.Accumulator
	parsers []LogParser

	sync.Mutex

	GrokParser *grok.Parser `toml:"grok"`
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

  ## Parse logstash-style "grok" patterns:
  ##   Telegraf built-in parsing patterns: https://goo.gl/dkay10
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
`

func (l *LogParserPlugin) SampleConfig() string {
	return sampleConfig
}

func (l *LogParserPlugin) Description() string {
	return "Stream and parse log file(s)."
}

func (l *LogParserPlugin) Gather(acc telegraf.Accumulator) error {
	l.Lock()
	defer l.Unlock()

	// always start from the beginning of files that appear while we're running
	return l.tailNewfiles(true)
}

func (l *LogParserPlugin) Start(acc telegraf.Accumulator) error {
	l.Lock()
	defer l.Unlock()

	l.acc = acc
	l.lines = make(chan string, 1000)
	l.done = make(chan struct{})
	l.tailers = make(map[string]*tail.Tail)

	// Looks for fields which implement LogParser interface
	l.parsers = []LogParser{}
	s := reflect.ValueOf(l).Elem()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		if !f.CanInterface() {
			continue
		}

		if lpPlugin, ok := f.Interface().(LogParser); ok {
			if reflect.ValueOf(lpPlugin).IsNil() {
				continue
			}
			l.parsers = append(l.parsers, lpPlugin)
		}
	}

	if len(l.parsers) == 0 {
		return fmt.Errorf("ERROR: logparser input plugin: no parser defined.")
	}

	// compile log parser patterns:
	for _, parser := range l.parsers {
		if err := parser.Compile(); err != nil {
			return err
		}
	}

	l.wg.Add(1)
	go l.parser()

	return l.tailNewfiles(l.FromBeginning)
}

// check the globs against files on disk, and start tailing any new files.
// Assumes l's lock is held!
func (l *LogParserPlugin) tailNewfiles(fromBeginning bool) error {
	var seek tail.SeekInfo
	if !fromBeginning {
		seek.Whence = 2
		seek.Offset = 0
	}

	// Create a "tailer" for each file
	for _, filepath := range l.Files {
		g, err := globpath.Compile(filepath)
		if err != nil {
			log.Printf("E! Error Glob %s failed to compile, %s", filepath, err)
			continue
		}
		files := g.Match()

		for file, _ := range files {
			if _, ok := l.tailers[file]; ok {
				// we're already tailing this file
				continue
			}

			tailer, err := tail.TailFile(file,
				tail.Config{
					ReOpen:    true,
					Follow:    true,
					Location:  &seek,
					MustExist: true,
				})
			l.acc.AddError(err)

			// create a goroutine for each "tailer"
			l.wg.Add(1)
			go l.receiver(tailer)
			l.tailers[file] = tailer
		}
	}

	return nil
}

// receiver is launched as a goroutine to continuously watch a tailed logfile
// for changes and send any log lines down the l.lines channel.
func (l *LogParserPlugin) receiver(tailer *tail.Tail) {
	defer l.wg.Done()

	var line *tail.Line
	for line = range tailer.Lines {

		if line.Err != nil {
			log.Printf("E! Error tailing file %s, Error: %s\n",
				tailer.Filename, line.Err)
			continue
		}

		// Fix up files with Windows line endings.
		text := strings.TrimRight(line.Text, "\r")

		select {
		case <-l.done:
		case l.lines <- text:
		}
	}
}

// parser is launched as a goroutine to watch the l.lines channel.
// when a line is available, parser parses it and adds the metric(s) to the
// accumulator.
func (l *LogParserPlugin) parser() {
	defer l.wg.Done()

	var m telegraf.Metric
	var err error
	var line string
	for {
		select {
		case <-l.done:
			return
		case line = <-l.lines:
			if line == "" || line == "\n" {
				continue
			}
		}

		for _, parser := range l.parsers {
			m, err = parser.ParseLine(line)
			if err == nil {
				if m != nil {
					l.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
				}
			} else {
				log.Println("E! Error parsing log line: " + err.Error())
			}
		}
	}
}

func (l *LogParserPlugin) Stop() {
	l.Lock()
	defer l.Unlock()

	for _, t := range l.tailers {
		err := t.Stop()
		if err != nil {
			log.Printf("E! Error stopping tail on file %s\n", t.Filename)
		}
		t.Cleanup()
	}
	close(l.done)
	l.wg.Wait()
}

func init() {
	inputs.Add("logparser", func() telegraf.Input {
		return &LogParserPlugin{}
	})
}

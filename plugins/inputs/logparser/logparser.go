package logparser

import (
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/hpcloud/tail"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/errchan"
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

	tailers []*tail.Tail
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
  ## Read file from beginning.
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
	return nil
}

func (l *LogParserPlugin) Start(acc telegraf.Accumulator) error {
	l.Lock()
	defer l.Unlock()

	l.acc = acc
	l.lines = make(chan string, 1000)
	l.done = make(chan struct{})

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
	errChan := errchan.New(len(l.parsers))
	for _, parser := range l.parsers {
		if err := parser.Compile(); err != nil {
			errChan.C <- err
		}
	}
	if err := errChan.Error(); err != nil {
		return err
	}

	var seek tail.SeekInfo
	if !l.FromBeginning {
		seek.Whence = 2
		seek.Offset = 0
	}

	l.wg.Add(1)
	go l.parser()

	// Create a "tailer" for each file
	for _, filepath := range l.Files {
		g, err := globpath.Compile(filepath)
		if err != nil {
			log.Printf("E! Error Glob %s failed to compile, %s", filepath, err)
			continue
		}
		files := g.Match()
		errChan = errchan.New(len(files))
		for file, _ := range files {
			tailer, err := tail.TailFile(file,
				tail.Config{
					ReOpen:    true,
					Follow:    true,
					Location:  &seek,
					MustExist: true,
				})
			errChan.C <- err

			// create a goroutine for each "tailer"
			l.wg.Add(1)
			go l.receiver(tailer)
			l.tailers = append(l.tailers, tailer)
		}
	}

	return errChan.Error()
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

		select {
		case <-l.done:
		case l.lines <- line.Text:
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

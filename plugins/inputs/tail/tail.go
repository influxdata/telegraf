// +build !solaris
// +build go1.10

package tail

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
)

const (
	defaultPollInterval = time.Millisecond * 250
	defaultWatchMethod  = "inotify"
)

type Tail struct {
	Files         []string
	FromBeginning bool
	WatchMethod   string
	PollInterval  internal.Duration

	watcher    watcher.Watcher
	tailer     *tailer.Tailer
	lines      chan *logline.LogLine
	files      map[string]bool
	parserFunc parsers.ParserFunc
	wg         sync.WaitGroup
	acc        telegraf.Accumulator

	sync.Mutex
}

const sampleConfig = `
  ## files to tail.
  ## These accept standard unix glob matching rules, but with the addition of
  ## ** as a "super asterisk". ie:
  ##   "/var/log/**.log"  -> recursively find all .log files in /var/log
  ##   "/var/log/*/*.log" -> find all .log files with a parent dir in /var/log
  ##   "/var/log/apache.log" -> just tail the apache log file
  ##
  ## See https://github.com/gobwas/glob for more examples
  ##
  files = ["/var/mymetrics.out"]
  ## Read file from beginning.
  from_beginning = false

  ## Method used to watch for file updates.  Can be either "inotify" or "poll".
  # watch_method = "inotify"

  ## Poll interval. Used when watch_method is set to "poll".
  # poll_interval = "250ms"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (t *Tail) SampleConfig() string {
	return sampleConfig
}

func (t *Tail) Description() string {
	return "Stream a log file, like the tail -f command"
}

func (t *Tail) Gather(acc telegraf.Accumulator) error {
	t.Lock()
	defer t.Unlock()

	return t.tailNewFiles(true)
}

func (t *Tail) Start(acc telegraf.Accumulator) error {
	t.Lock()
	defer t.Unlock()

	t.acc = acc
	t.lines = make(chan *logline.LogLine)
	t.files = make(map[string]bool)

	var poll bool
	if t.WatchMethod == "poll" {
		poll = true
	}

	var err error
	t.watcher, err = watcher.NewLogWatcher(t.PollInterval.Duration, !poll, watcher.Logger(taillog.DiscardingLogger))
	if err != nil {
		defer close(t.lines)
		return err
	}

	opts := []tailer.Option{tailer.Logger(taillog.DiscardingLogger)}
	if t.FromBeginning {
		opts = append(opts, tailer.OneShot)
	}

	t.tailer, err = tailer.New(t.lines, t.watcher, opts...)
	if err != nil {
		defer t.watcher.Close()
		defer close(t.lines)
		return err
	}

	parser, err := t.parserFunc()
	if err != nil {
		defer t.tailer.Close()
		return err
	}
	t.wg.Add(1)
	go t.parse(parser)

	return t.tailNewFiles(t.FromBeginning)
}

func (t *Tail) tailNewFiles(fromBeginning bool) error {

	for _, filepath := range t.Files {
		g, err := globpath.Compile(filepath)
		if err != nil {
			t.acc.AddError(fmt.Errorf("E! Error Glob %s failed to compile, %s", filepath, err))
		}
		for _, file := range g.Match() {
			if _, ok := t.files[file]; ok {
				// we're already tailing this file
				continue
			}

			err := t.tailer.TailPath(file)
			if err != nil {
				t.acc.AddError(err)
				continue
			}

			log.Printf("D! [inputs.tail] tail added for file: %v", file)

			// create a goroutine for each "tailer"
			t.files[file] = true
		}
	}

	return nil
}

// parse is launched as a goroutine to watch the l.lines channel.
// when a line is available, parse parses it and adds the metric(s) to the
// accumulator.
func (t *Tail) parse(parser parsers.Parser) {
	defer t.wg.Done()

	var firstLine = true
	var metrics []telegraf.Metric
	var m telegraf.Metric
	var err error
	for line := range t.lines {
		line.Line = strings.TrimRight(line.Line, "\r")

		if firstLine {
			metrics, err = parser.Parse([]byte(line.Line))
			if err == nil {
				if len(metrics) == 0 {
					firstLine = false
					continue
				} else {
					m = metrics[0]
				}
			}
			firstLine = false
		} else {
			m, err = parser.ParseLine(line.Line)
		}

		if err == nil {
			if m != nil {
				tags := m.Tags()
				tags["path"] = line.Filename
				t.acc.AddFields(m.Name(), m.Fields(), tags, m.Time())
			}
		} else {
			t.acc.AddError(fmt.Errorf("E! Malformed log line in %s: [%s], Error: %s\n",
				line.Filename, line.Line, err))
		}
	}
}

func (t *Tail) Stop() {
	t.Lock()
	defer t.Unlock()

	err := t.tailer.Close()
	if err != nil {
		t.acc.AddError(fmt.Errorf("E! Error closing tailer, Error: %s\n", err))
	} else {
		for file := range t.files {
			log.Printf("D! [inputs.logparser] tail removed for file: %v", file)
		}
	}

	t.wg.Wait()
}

func (t *Tail) SetParserFunc(fn parsers.ParserFunc) {
	t.parserFunc = fn
}

func init() {
	inputs.Add("tail", func() telegraf.Input {
		return &Tail{
			FromBeginning: false,
			PollInterval:  internal.Duration{Duration: defaultPollInterval},
		}
	})
}

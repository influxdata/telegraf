// +build !solaris

package tail

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/dimchansky/utfbom"
	"github.com/influxdata/tail"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/common/encoding"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/csv"
)

const (
	defaultWatchMethod         = "inotify"
	defaultMaxUndeliveredLines = 1000
)

var (
	offsets      = make(map[string]int64)
	offsetsMutex = new(sync.Mutex)
)

type empty struct{}
type semaphore chan empty

type Tail struct {
	Files               []string `toml:"files"`
	FromBeginning       bool     `toml:"from_beginning"`
	Pipe                bool     `toml:"pipe"`
	WatchMethod         string   `toml:"watch_method"`
	MaxUndeliveredLines int      `toml:"max_undelivered_lines"`
	CharacterEncoding   string   `toml:"character_encoding"`

	Log        telegraf.Logger `toml:"-"`
	tailers    map[string]*tail.Tail
	offsets    map[string]int64
	parserFunc parsers.ParserFunc
	wg         sync.WaitGroup

	acc telegraf.TrackingAccumulator

	MultilineConfig MultilineConfig `toml:"multiline"`
	multiline       *Multiline

	ctx     context.Context
	cancel  context.CancelFunc
	sem     semaphore
	decoder *encoding.Decoder
}

func NewTail() *Tail {
	offsetsMutex.Lock()
	offsetsCopy := make(map[string]int64, len(offsets))
	for k, v := range offsets {
		offsetsCopy[k] = v
	}
	offsetsMutex.Unlock()

	return &Tail{
		FromBeginning:       false,
		MaxUndeliveredLines: 1000,
		offsets:             offsetsCopy,
	}
}

const sampleConfig = `
  ## File names or a pattern to tail.
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
  # from_beginning = false

  ## Whether file is a named pipe
  # pipe = false

  ## Method used to watch for file updates.  Can be either "inotify" or "poll".
  # watch_method = "inotify"

  ## Maximum lines of the file to process that have not yet be written by the
  ## output.  For best throughput set based on the number of metrics on each
  ## line and the size of the output's metric_batch_size.
  # max_undelivered_lines = 1000

  ## Character encoding to use when interpreting the file contents.  Invalid
  ## characters are replaced using the unicode replacement character.  When set
  ## to the empty string the data is not decoded to text.
  ##   ex: character_encoding = "utf-8"
  ##       character_encoding = "utf-16le"
  ##       character_encoding = "utf-16be"
  ##       character_encoding = ""
  # character_encoding = ""

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## multiline parser/codec
  ## https://www.elastic.co/guide/en/logstash/2.4/plugins-filters-multiline.html
  #[inputs.tail.multiline]
    ## The pattern should be a regexp which matches what you believe to be an
	## indicator that the field is part of an event consisting of multiple lines of log data.
    #pattern = "^\s"

    ## This field must be either "previous" or "next".
	## If a line matches the pattern, "previous" indicates that it belongs to the previous line,
	## whereas "next" indicates that the line belongs to the next one.
    #match_which_line = "previous"

    ## The invert_match field can be true or false (defaults to false).
    ## If true, a message not matching the pattern will constitute a match of the multiline
	## filter and the what will be applied. (vice-versa is also true)
    #invert_match = false

    ## After the specified timeout, this plugin sends a multiline event even if no new pattern
	## is found to start a new event. The default timeout is 5s.
    #timeout = 5s
`

func (t *Tail) SampleConfig() string {
	return sampleConfig
}

func (t *Tail) Description() string {
	return "Parse the new lines appended to a file"
}

func (t *Tail) Init() error {
	if t.MaxUndeliveredLines == 0 {
		return errors.New("max_undelivered_lines must be positive")
	}
	t.sem = make(semaphore, t.MaxUndeliveredLines)

	var err error
	t.decoder, err = encoding.NewDecoder(t.CharacterEncoding)
	return err
}

func (t *Tail) Gather(acc telegraf.Accumulator) error {
	return t.tailNewFiles(true)
}

func (t *Tail) Start(acc telegraf.Accumulator) error {
	t.acc = acc.WithTracking(t.MaxUndeliveredLines)

	t.ctx, t.cancel = context.WithCancel(context.Background())

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for {
			select {
			case <-t.ctx.Done():
				return
			case <-t.acc.Delivered():
				<-t.sem
			}
		}
	}()

	var err error
	t.multiline, err = t.MultilineConfig.NewMultiline()

	if err != nil {
		return err
	}

	t.tailers = make(map[string]*tail.Tail)

	err = t.tailNewFiles(t.FromBeginning)

	// clear offsets
	t.offsets = make(map[string]int64)
	// assumption that once Start is called, all parallel plugins have already been initialized
	offsetsMutex.Lock()
	offsets = make(map[string]int64)
	offsetsMutex.Unlock()

	return err
}

func (t *Tail) tailNewFiles(fromBeginning bool) error {
	var poll bool
	if t.WatchMethod == "poll" {
		poll = true
	}

	// Create a "tailer" for each file
	for _, filepath := range t.Files {
		g, err := globpath.Compile(filepath)
		if err != nil {
			t.Log.Errorf("Glob %q failed to compile: %s", filepath, err.Error())
		}
		for _, file := range g.Match() {
			if _, ok := t.tailers[file]; ok {
				// we're already tailing this file
				continue
			}

			var seek *tail.SeekInfo
			if !t.Pipe && !fromBeginning {
				if offset, ok := t.offsets[file]; ok {
					t.Log.Debugf("Using offset %d for %q", offset, file)
					seek = &tail.SeekInfo{
						Whence: 0,
						Offset: offset,
					}
				} else {
					seek = &tail.SeekInfo{
						Whence: 2,
						Offset: 0,
					}
				}
			}

			tailer, err := tail.TailFile(file,
				tail.Config{
					ReOpen:    true,
					Follow:    true,
					Location:  seek,
					MustExist: true,
					Poll:      poll,
					Pipe:      t.Pipe,
					Logger:    tail.DiscardingLogger,
					OpenReaderFunc: func(rd io.Reader) io.Reader {
						r, _ := utfbom.Skip(t.decoder.Reader(rd))
						return r
					},
				})

			if err != nil {
				t.Log.Debugf("Failed to open file (%s): %v", file, err)
				continue
			}

			t.Log.Debugf("Tail added for %q", file)

			parser, err := t.parserFunc()
			if err != nil {
				t.Log.Errorf("Creating parser: %s", err.Error())
				continue
			}

			// create a goroutine for each "tailer"
			t.wg.Add(1)

			go func() {
				defer t.wg.Done()
				t.receiver(parser, tailer)

				t.Log.Debugf("Tail removed for %q", tailer.Filename)

				if err := tailer.Err(); err != nil {
					t.Log.Errorf("Tailing %q: %s", tailer.Filename, err.Error())
				}
			}()

			t.tailers[tailer.Filename] = tailer
		}
	}
	return nil
}

// ParseLine parses a line of text.
func parseLine(parser parsers.Parser, line string, firstLine bool) ([]telegraf.Metric, error) {
	switch parser.(type) {
	case *csv.Parser:
		// The csv parser parses headers in Parse and skips them in ParseLine.
		// As a temporary solution call Parse only when getting the first
		// line from the file.
		if firstLine {
			return parser.Parse([]byte(line))
		} else {
			m, err := parser.ParseLine(line)
			if err != nil {
				return nil, err
			}

			if m != nil {
				return []telegraf.Metric{m}, nil
			}
			return []telegraf.Metric{}, nil
		}
	default:
		return parser.Parse([]byte(line))
	}
}

// Receiver is launched as a goroutine to continuously watch a tailed logfile
// for changes, parse any incoming msgs, and add to the accumulator.
func (t *Tail) receiver(parser parsers.Parser, tailer *tail.Tail) {
	var firstLine = true

	// holds the individual lines of multi-line log entries.
	var buffer bytes.Buffer

	var timer *time.Timer
	var timeout <-chan time.Time

	// The multiline mode requires a timer in order to flush the multiline buffer
	// if no new lines are incoming.
	if t.multiline.IsEnabled() {
		timer = time.NewTimer(t.MultilineConfig.Timeout.Duration)
		timeout = timer.C
	}

	channelOpen := true
	tailerOpen := true
	var line *tail.Line

	for {
		line = nil

		if timer != nil {
			timer.Reset(t.MultilineConfig.Timeout.Duration)
		}

		select {
		case <-t.ctx.Done():
			channelOpen = false
		case line, tailerOpen = <-tailer.Lines:
			if !tailerOpen {
				channelOpen = false
			}
		case <-timeout:
		}

		var text string

		if line != nil {
			// Fix up files with Windows line endings.
			text = strings.TrimRight(line.Text, "\r")

			if t.multiline.IsEnabled() {
				if text = t.multiline.ProcessLine(text, &buffer); text == "" {
					continue
				}
			}
		}
		if line == nil || !channelOpen || !tailerOpen {
			if text += t.multiline.Flush(&buffer); text == "" {
				if !channelOpen {
					return
				}

				continue
			}
		}

		if line != nil && line.Err != nil {
			t.Log.Errorf("Tailing %q: %s", tailer.Filename, line.Err.Error())
			continue
		}

		metrics, err := parseLine(parser, text, firstLine)
		if err != nil {
			t.Log.Errorf("Malformed log line in %q: [%q]: %s",
				tailer.Filename, text, err.Error())
			continue
		}
		firstLine = false

		for _, metric := range metrics {
			metric.AddTag("path", tailer.Filename)
		}

		// try writing out metric first without blocking
		select {
		case t.sem <- empty{}:
			t.acc.AddTrackingMetricGroup(metrics)
			if t.ctx.Err() != nil {
				return // exit!
			}
			continue // next loop
		default:
			// no room. switch to blocking write.
		}

		// Block until plugin is stopping or room is available to add metrics.
		select {
		case <-t.ctx.Done():
			return
		case t.sem <- empty{}:
			t.acc.AddTrackingMetricGroup(metrics)
		}
	}
}

func (t *Tail) Stop() {
	for _, tailer := range t.tailers {
		if !t.Pipe && !t.FromBeginning {
			// store offset for resume
			offset, err := tailer.Tell()
			if err == nil {
				t.Log.Debugf("Recording offset %d for %q", offset, tailer.Filename)
			} else {
				t.Log.Errorf("Recording offset for %q: %s", tailer.Filename, err.Error())
			}
		}
		err := tailer.Stop()
		if err != nil {
			t.Log.Errorf("Stopping tail on %q: %s", tailer.Filename, err.Error())
		}
	}

	t.cancel()
	t.wg.Wait()

	// persist offsets
	offsetsMutex.Lock()
	for k, v := range t.offsets {
		offsets[k] = v
	}
	offsetsMutex.Unlock()
}

func (t *Tail) SetParserFunc(fn parsers.ParserFunc) {
	t.parserFunc = fn
}

func init() {
	inputs.Add("tail", func() telegraf.Input {
		return NewTail()
	})
}

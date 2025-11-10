//go:generate ../../../tools/readme_config_includer/generator
//go:build !solaris

package tail

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dimchansky/utfbom"
	"github.com/influxdata/tail"
	"github.com/pborman/ansi"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/common/encoding"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

//go:embed sample.conf
var sampleConfig string

var (
	once sync.Once

	offsets      = make(map[string]int64)
	offsetsMutex = new(sync.Mutex)
)

type Tail struct {
	Files               []string `toml:"files"`
	FromBeginning       bool     `toml:"from_beginning" deprecated:"1.34.0;1.40.0;use 'initial_read_offset' with value 'beginning' instead"`
	InitialReadOffset   string   `toml:"initial_read_offset"`
	Pipe                bool     `toml:"pipe"`
	WatchMethod         string   `toml:"watch_method"`
	MaxUndeliveredLines int      `toml:"max_undelivered_lines"`
	CharacterEncoding   string   `toml:"character_encoding"`
	PathTag             string   `toml:"path_tag"`

	Filters      []string `toml:"filters"`
	filterColors bool

	Log          telegraf.Logger `toml:"-"`
	tailers      map[string]*tail.Tail
	tailersMutex sync.RWMutex
	offsets      map[string]int64
	parserFunc   telegraf.ParserFunc
	wg           sync.WaitGroup

	acc telegraf.TrackingAccumulator

	MultilineConfig multilineConfig `toml:"multiline"`
	multiline       *multiline

	ctx     context.Context
	cancel  context.CancelFunc
	sem     semaphore
	decoder *encoding.Decoder

	nomatch map[string]bool
}

type empty struct{}
type semaphore chan empty

func (*Tail) SampleConfig() string {
	return sampleConfig
}

func (t *Tail) SetParserFunc(fn telegraf.ParserFunc) {
	t.parserFunc = fn
}

func (t *Tail) Init() error {
	// Backward compatibility setting
	if t.InitialReadOffset == "" {
		if t.FromBeginning {
			t.InitialReadOffset = "beginning"
		} else {
			t.InitialReadOffset = "saved-or-end"
		}
	}

	// Check settings
	switch t.InitialReadOffset {
	case "":
		t.InitialReadOffset = "saved-or-end"
	case "beginning", "end", "saved-or-end", "saved-or-beginning":
	default:
		return fmt.Errorf("invalid 'initial_read_offset' setting %q", t.InitialReadOffset)
	}

	if t.MaxUndeliveredLines == 0 {
		return errors.New("max_undelivered_lines must be positive")
	}
	t.sem = make(semaphore, t.MaxUndeliveredLines)

	for _, filter := range t.Filters {
		if filter == "ansi_color" {
			t.filterColors = true
		}
	}

	// init offsets
	t.offsets = make(map[string]int64)

	dec, err := encoding.NewDecoder(t.CharacterEncoding)
	if err != nil {
		return fmt.Errorf("creating decoder failed: %w", err)
	}
	t.decoder = dec

	// Initialize the map to keep track of patterns that did not produce any
	// matching files to warn about potential permission issues only once.
	t.nomatch = make(map[string]bool)

	return nil
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
	t.multiline, err = t.MultilineConfig.newMultiline()

	if err != nil {
		return err
	}

	t.tailers = make(map[string]*tail.Tail)

	err = t.tailNewFiles()
	if err != nil {
		return err
	}

	// assumption that once Start is called, all parallel plugins have already been initialized
	offsetsMutex.Lock()
	offsets = make(map[string]int64)
	offsetsMutex.Unlock()

	return err
}

func (t *Tail) getSeekInfo(file string) (*tail.SeekInfo, error) {
	// Pipes do not support seeking
	if t.Pipe {
		return nil, nil
	}

	// Determine the actual position for continuing
	switch t.InitialReadOffset {
	case "beginning":
		return &tail.SeekInfo{Whence: 0, Offset: 0}, nil
	case "end":
		return &tail.SeekInfo{Whence: 2, Offset: 0}, nil
	case "", "saved-or-end":
		if offset, ok := t.offsets[file]; ok {
			t.Log.Debugf("Using offset %d for %q", offset, file)
			return &tail.SeekInfo{Whence: 0, Offset: offset}, nil
		}
		return &tail.SeekInfo{Whence: 2, Offset: 0}, nil
	case "saved-or-beginning":
		if offset, ok := t.offsets[file]; ok {
			t.Log.Debugf("Using offset %d for %q", offset, file)
			return &tail.SeekInfo{Whence: 0, Offset: offset}, nil
		}
		return &tail.SeekInfo{Whence: 0, Offset: 0}, nil
	default:
		return nil, errors.New("invalid 'initial_read_offset' setting")
	}
}

func (t *Tail) GetState() interface{} {
	return t.offsets
}

func (t *Tail) SetState(state interface{}) error {
	offsetsState, ok := state.(map[string]int64)
	if !ok {
		return errors.New("state has to be of type 'map[string]int64'")
	}
	for k, v := range offsetsState {
		t.offsets[k] = v
	}
	return nil
}

func (t *Tail) Gather(_ telegraf.Accumulator) error {
	return t.tailNewFiles()
}

func (t *Tail) Stop() {
	t.tailersMutex.Lock()
	defer t.tailersMutex.Unlock()

	for filename, tailer := range t.tailers {
		if !t.Pipe {
			// store offset for resume
			offset, err := tailer.Tell()
			if err == nil {
				t.Log.Debugf("Recording offset %d for %q", offset, tailer.Filename)
				t.offsets[tailer.Filename] = offset
			} else {
				t.Log.Errorf("Recording offset for %q: %v", tailer.Filename, err)
			}
		}
		err := tailer.Stop()
		if err != nil {
			t.Log.Errorf("Stopping tail on %q: %v", tailer.Filename, err)
		}

		// Explicitly delete the tailer from the map to avoid memory leaks
		delete(t.tailers, filename)
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

func (t *Tail) tailNewFiles() error {
	var poll bool
	if t.WatchMethod == "poll" {
		poll = true
	}

	// Track files that we're currently processing
	currentFiles := make(map[string]bool)

	// Create a "tailer" for each file
	for _, filepath := range t.Files {
		g, err := globpath.Compile(filepath)
		if err != nil {
			t.Log.Errorf("Glob %q failed to compile: %v", filepath, err)
			continue
		}

		// Work around an issue in the doublestar library that requires read
		// permissions on the directory to glob files (see
		// https://github.com/bmatcuk/doublestar/issues/103).
		// However, if the given file does not require globbing we can try to
		// keep the file directly.
		matches := g.Match()
		if len(matches) == 0 {
			if _, err := os.Lstat(filepath); err == nil {
				matches = append(matches, filepath)
			} else if !t.nomatch[filepath] && strings.ContainsAny(filepath, "*?[") {
				// Read permissions are required to expand wildcards so find all
				// directories followed by wildcards and check if they are readable
				parts := strings.Split(filepath, string(os.PathSeparator))
				for i, e := range parts {
					if e == "" || !strings.ContainsAny(e, "*?[") {
						continue
					}
					partialPath := strings.Join(parts[:i], string(os.PathSeparator))
					if stat, err := os.Stat(partialPath); err != nil || !stat.IsDir() {
						break
					}
					if _, err := os.ReadDir(partialPath); errors.Is(err, os.ErrPermission) {
						t.Log.Warnf(
							"Directory %q is not readable but is followed by wildcards,"+
								"make sure you set read permissions or globbing will not work!", partialPath,
						)
						break
					}
				}

				t.nomatch[filepath] = true
			}
		}

		for _, file := range matches {
			// Mark this file as currently being processed
			currentFiles[file] = true

			// Check if we're already tailing this file
			t.tailersMutex.RLock()
			_, alreadyTailing := t.tailers[file]
			t.tailersMutex.RUnlock()

			if alreadyTailing {
				// already tailing this file
				continue
			}

			seek, err := t.getSeekInfo(file)
			if err != nil {
				return err
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
				t.Log.Errorf("Creating parser: %v", err)
				continue
			}

			// create a goroutine for each "tailer"
			t.wg.Add(1)

			// Store the tailer in the map before starting the goroutine
			t.tailersMutex.Lock()
			t.tailers[tailer.Filename] = tailer
			t.tailersMutex.Unlock()

			go func(tl *tail.Tail) {
				defer t.wg.Done()
				t.receiver(parser, tl)

				t.Log.Debugf("Tail removed for %q", tl.Filename)

				if err := tl.Err(); err != nil {
					if strings.HasSuffix(err.Error(), "permission denied") {
						t.Log.Errorf("Deleting tailer for %q due to: %v", tl.Filename, err)
						t.tailersMutex.Lock()
						delete(t.tailers, tl.Filename)
						t.tailersMutex.Unlock()
					} else {
						t.Log.Errorf("Tailing %q: %v", tl.Filename, err)
					}
				}
			}(tailer)
		}
	}

	// Clean up tailers for files that are no longer being monitored
	return t.cleanupUnusedTailers(currentFiles)
}

// cleanupUnusedTailers stops and removes tailers for files that are no longer being monitored.
// It uses defer to ensure the mutex is always unlocked, even if errors occur.
func (t *Tail) cleanupUnusedTailers(currentFiles map[string]bool) error {
	t.tailersMutex.Lock()
	defer t.tailersMutex.Unlock()

	for file, tailer := range t.tailers {
		if !currentFiles[file] {
			// This file is no longer in our glob pattern matches
			// We need to stop tailing it and remove it from our list
			t.Log.Debugf("Removing tailer for %q as it's no longer in the glob pattern", file)

			// Stop the tailer first to avoid race conditions
			// This ensures the tail goroutine is no longer running when we call Tell()
			if err := tailer.Stop(); err != nil {
				t.Log.Errorf("Stopping tail on %q: %v", tailer.Filename, err)
			}

			// Now it's safe to get and save the offset since the tailer is stopped
			if !t.Pipe {
				offset, err := tailer.Tell()
				if err == nil {
					t.Log.Debugf("Recording offset %d for %q", offset, tailer.Filename)
					t.offsets[tailer.Filename] = offset
				} else {
					// This can happen if the file was already removed or closed
					t.Log.Debugf("Could not get offset for %q: %v", tailer.Filename, err)
				}
			}

			// Remove from our map
			delete(t.tailers, file)
		}
	}

	return nil
}

func parseLine(parser telegraf.Parser, line string) ([]telegraf.Metric, error) {
	m, err := parser.Parse([]byte(line))
	if err != nil {
		if errors.Is(err, parsers.ErrEOF) {
			return nil, nil
		}
		return nil, err
	}
	return m, err
}

// receiver is launched as a goroutine to continuously watch a tailed logfile
// for changes, parse any incoming messages, and add to the accumulator.
func (t *Tail) receiver(parser telegraf.Parser, tailer *tail.Tail) {
	// holds the individual lines of multi-line log entries.
	var buffer bytes.Buffer

	var timer *time.Timer
	var timeout <-chan time.Time

	// The multiline mode requires a timer in order to flush the multiline buffer
	// if no new lines are incoming.
	if t.multiline.isEnabled() {
		timer = time.NewTimer(time.Duration(*t.MultilineConfig.Timeout))
		timeout = timer.C
	}

	channelOpen := true
	tailerOpen := true
	var line *tail.Line

	for {
		line = nil

		if timer != nil {
			timer.Reset(time.Duration(*t.MultilineConfig.Timeout))
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

			if t.multiline.isEnabled() {
				if text = t.multiline.processLine(text, &buffer); text == "" {
					continue
				}
			}
		}
		if line == nil || !channelOpen || !tailerOpen {
			if text += flush(&buffer); text == "" {
				if !channelOpen {
					return
				}

				continue
			}
		}

		if line != nil && line.Err != nil {
			t.Log.Errorf("Tailing %q: %v", tailer.Filename, line.Err)
			continue
		}

		if t.filterColors {
			out, err := ansi.Strip([]byte(text))
			if err != nil {
				t.Log.Errorf("Cannot strip ansi colors from %s: %s", text, err)
			}
			text = string(out)
		}

		metrics, err := parseLine(parser, text)
		if err != nil {
			t.Log.Errorf("Malformed log line in %q: [%q]: %v",
				tailer.Filename, text, err)
			continue
		}
		if len(metrics) == 0 {
			once.Do(func() {
				t.Log.Debug(internal.NoMetricsCreatedMsg)
			})
		}
		if t.PathTag != "" {
			for _, metric := range metrics {
				metric.AddTag(t.PathTag, tailer.Filename)
			}
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
		// Tail is trying to close so drain the sem to allow the receiver
		// to exit. This condition is hit when the tailer may have hit the
		// maximum undelivered lines and is trying to close.
		case <-tailer.Dying():
			<-t.sem
		case t.sem <- empty{}:
			t.acc.AddTrackingMetricGroup(metrics)
		}
	}
}

func newTail() *Tail {
	offsetsMutex.Lock()
	offsetsCopy := make(map[string]int64, len(offsets))
	for k, v := range offsets {
		offsetsCopy[k] = v
	}
	offsetsMutex.Unlock()

	return &Tail{
		MaxUndeliveredLines: 1000,
		offsets:             offsetsCopy,
		PathTag:             "path",
	}
}

func init() {
	inputs.Add("tail", func() telegraf.Input {
		return newTail()
	})
}

package directory_monitor

import (
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/djherbis/times"
	"golang.org/x/sync/semaphore"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/csv"
	"github.com/influxdata/telegraf/selfstat"
)

var (
	defaultFilesToMonitor             = []string{}
	defaultFilesToIgnore              = []string{}
	defaultMaxBufferedMetrics         = 10000
	defaultDirectoryDurationThreshold = config.Duration(0 * time.Millisecond)
	defaultFileQueueSize              = 100000
)

type DirectoryMonitor struct {
	Directory         string `toml:"directory"`
	FinishedDirectory string `toml:"finished_directory"`
	ErrorDirectory    string `toml:"error_directory"`
	FileTag           string `toml:"file_tag"`

	FilesToMonitor             []string        `toml:"files_to_monitor"`
	FilesToIgnore              []string        `toml:"files_to_ignore"`
	MaxBufferedMetrics         int             `toml:"max_buffered_metrics"`
	DirectoryDurationThreshold config.Duration `toml:"directory_duration_threshold"`
	Log                        telegraf.Logger `toml:"-"`
	FileQueueSize              int             `toml:"file_queue_size"`

	filesInUse          sync.Map
	cancel              context.CancelFunc
	context             context.Context
	parserFunc          parsers.ParserFunc
	filesProcessed      selfstat.Stat
	filesDropped        selfstat.Stat
	waitGroup           *sync.WaitGroup
	acc                 telegraf.TrackingAccumulator
	sem                 *semaphore.Weighted
	fileRegexesToMatch  []*regexp.Regexp
	fileRegexesToIgnore []*regexp.Regexp
	filesToProcess      chan string
}

func (monitor *DirectoryMonitor) Gather(_ telegraf.Accumulator) error {
	// Get all files sitting in the directory.
	files, err := os.ReadDir(monitor.Directory)
	if err != nil {
		return fmt.Errorf("unable to monitor the targeted directory: %w", err)
	}

	for _, file := range files {
		filePath := monitor.Directory + "/" + file.Name()

		// We've been cancelled via Stop().
		if monitor.context.Err() != nil {
			//nolint:nilerr // context cancelation is not an error
			return nil
		}

		stat, err := times.Stat(filePath)
		if err != nil {
			continue
		}

		timeThresholdExceeded := time.Since(stat.AccessTime()) >= time.Duration(monitor.DirectoryDurationThreshold)

		// If file is decaying, process it.
		if timeThresholdExceeded {
			monitor.processFile(file)
		}
	}

	return nil
}

func (monitor *DirectoryMonitor) Start(acc telegraf.Accumulator) error {
	// Use tracking to determine when more metrics can be added without overflowing the outputs.
	monitor.acc = acc.WithTracking(monitor.MaxBufferedMetrics)
	go func() {
		for range monitor.acc.Delivered() {
			monitor.sem.Release(1)
		}
	}()

	// Monitor the files channel and read what they receive.
	monitor.waitGroup.Add(1)
	go func() {
		monitor.Monitor()
		monitor.waitGroup.Done()
	}()

	return nil
}

func (monitor *DirectoryMonitor) Stop() {
	// Before stopping, wrap up all file-reading routines.
	monitor.cancel()
	close(monitor.filesToProcess)
	monitor.Log.Warnf("Exiting the Directory Monitor plugin. Waiting to quit until all current files are finished.")
	monitor.waitGroup.Wait()
}

func (monitor *DirectoryMonitor) Monitor() {
	for filePath := range monitor.filesToProcess {
		if monitor.context.Err() != nil {
			return
		}

		// Prevent goroutines from taking the same file as another.
		if _, exists := monitor.filesInUse.LoadOrStore(filePath, true); exists {
			continue
		}

		monitor.read(filePath)

		// We've finished reading the file and moved it away, delete it from files in use.
		monitor.filesInUse.Delete(filePath)
	}
}

func (monitor *DirectoryMonitor) processFile(file os.DirEntry) {
	if file.IsDir() {
		return
	}

	filePath := monitor.Directory + "/" + file.Name()

	// File must be configured to be monitored, if any configuration...
	if !monitor.isMonitoredFile(file.Name()) {
		return
	}

	// ...and should not be configured to be ignored.
	if monitor.isIgnoredFile(file.Name()) {
		return
	}

	select {
	case monitor.filesToProcess <- filePath:
	default:
	}
}

func (monitor *DirectoryMonitor) read(filePath string) {
	// Open, read, and parse the contents of the file.
	err := monitor.ingestFile(filePath)
	if _, isPathError := err.(*os.PathError); isPathError {
		return
	}

	// Handle a file read error. We don't halt execution but do document, log, and move the problematic file.
	if err != nil {
		monitor.Log.Errorf("Error while reading file: '" + filePath + "'. " + err.Error())
		monitor.filesDropped.Incr(1)
		if monitor.ErrorDirectory != "" {
			monitor.moveFile(filePath, monitor.ErrorDirectory)
		}
		return
	}

	// File is finished, move it to the 'finished' directory.
	monitor.moveFile(filePath, monitor.FinishedDirectory)
	monitor.filesProcessed.Incr(1)
}

func (monitor *DirectoryMonitor) ingestFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	parser, err := monitor.parserFunc()
	if err != nil {
		return fmt.Errorf("E! Creating parser: %s", err.Error())
	}

	// Handle gzipped files.
	var reader io.Reader
	if filepath.Ext(filePath) == ".gz" {
		reader, err = gzip.NewReader(file)
		if err != nil {
			return err
		}
	} else {
		reader = file
	}

	return monitor.parseFile(parser, reader, file.Name())
}

func (monitor *DirectoryMonitor) parseFile(parser parsers.Parser, reader io.Reader, fileName string) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		metrics, err := monitor.parseLine(parser, scanner.Bytes())
		if err != nil {
			return err
		}

		if monitor.FileTag != "" {
			for _, m := range metrics {
				m.AddTag(monitor.FileTag, filepath.Base(fileName))
			}
		}

		if err := monitor.sendMetrics(metrics); err != nil {
			return err
		}
	}

	return nil
}

func (monitor *DirectoryMonitor) parseLine(parser parsers.Parser, line []byte) ([]telegraf.Metric, error) {
	switch parser.(type) {
	case *csv.Parser:
		m, err := parser.Parse(line)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, nil
			}
			return nil, err
		}
		return m, err
	default:
		return parser.Parse(line)
	}
}

func (monitor *DirectoryMonitor) sendMetrics(metrics []telegraf.Metric) error {
	// Report the metrics for the file.
	for _, m := range metrics {
		// Block until metric can be written.
		if err := monitor.sem.Acquire(monitor.context, 1); err != nil {
			return err
		}
		monitor.acc.AddTrackingMetricGroup([]telegraf.Metric{m})
	}
	return nil
}

func (monitor *DirectoryMonitor) moveFile(filePath string, directory string) {
	err := os.Rename(filePath, directory+"/"+filepath.Base(filePath))

	if err != nil {
		monitor.Log.Errorf("Error while moving file '" + filePath + "' to another directory. Error: " + err.Error())
	}
}

func (monitor *DirectoryMonitor) isMonitoredFile(fileName string) bool {
	if len(monitor.fileRegexesToMatch) == 0 {
		return true
	}

	// Only monitor matching files.
	for _, regex := range monitor.fileRegexesToMatch {
		if regex.MatchString(fileName) {
			return true
		}
	}

	return false
}

func (monitor *DirectoryMonitor) isIgnoredFile(fileName string) bool {
	// Skip files that are set to be ignored.
	for _, regex := range monitor.fileRegexesToIgnore {
		if regex.MatchString(fileName) {
			return true
		}
	}

	return false
}

func (monitor *DirectoryMonitor) SetParserFunc(fn parsers.ParserFunc) {
	monitor.parserFunc = fn
}

func (monitor *DirectoryMonitor) Init() error {
	if monitor.Directory == "" || monitor.FinishedDirectory == "" {
		return errors.New("missing one of the following required config options: directory, finished_directory")
	}

	if monitor.FileQueueSize <= 0 {
		return errors.New("file queue size needs to be more than 0")
	}

	// Finished directory can be created if not exists for convenience.
	if _, err := os.Stat(monitor.FinishedDirectory); os.IsNotExist(err) {
		err = os.Mkdir(monitor.FinishedDirectory, 0777)
		if err != nil {
			return err
		}
	}

	monitor.filesDropped = selfstat.Register("directory_monitor", "files_dropped", map[string]string{})
	monitor.filesProcessed = selfstat.Register("directory_monitor", "files_processed", map[string]string{})

	// If an error directory should be used but has not been configured yet, create one ourselves.
	if monitor.ErrorDirectory != "" {
		if _, err := os.Stat(monitor.ErrorDirectory); os.IsNotExist(err) {
			err := os.Mkdir(monitor.ErrorDirectory, 0777)
			if err != nil {
				return err
			}
		}
	}

	monitor.waitGroup = &sync.WaitGroup{}
	monitor.sem = semaphore.NewWeighted(int64(monitor.MaxBufferedMetrics))
	monitor.context, monitor.cancel = context.WithCancel(context.Background())
	monitor.filesToProcess = make(chan string, monitor.FileQueueSize)

	// Establish file matching / exclusion regexes.
	for _, matcher := range monitor.FilesToMonitor {
		regex, err := regexp.Compile(matcher)
		if err != nil {
			return err
		}
		monitor.fileRegexesToMatch = append(monitor.fileRegexesToMatch, regex)
	}

	for _, matcher := range monitor.FilesToIgnore {
		regex, err := regexp.Compile(matcher)
		if err != nil {
			return err
		}
		monitor.fileRegexesToIgnore = append(monitor.fileRegexesToIgnore, regex)
	}

	return nil
}

func init() {
	inputs.Add("directory_monitor", func() telegraf.Input {
		return &DirectoryMonitor{
			FilesToMonitor:             defaultFilesToMonitor,
			FilesToIgnore:              defaultFilesToIgnore,
			MaxBufferedMetrics:         defaultMaxBufferedMetrics,
			DirectoryDurationThreshold: defaultDirectoryDurationThreshold,
			FileQueueSize:              defaultFileQueueSize,
		}
	})
}

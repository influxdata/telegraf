package request_aggregates

import (
	"fmt"
	"github.com/hpcloud/tail"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"log"
	"regexp"
	"sync"
	"time"
)

type RequestAggregates struct {
	File                 string
	TimestampPosition    int
	TimestampFormat      string
	TimePosition         int
	TimePercentiles      []float32
	TimeWindowSize       internal.Duration
	TimeWindows          int
	ResultPosition       int
	ResultSuccessRegex   string
	ThroughputWindowSize internal.Duration
	ThroughputWindows    int

	isTimestampEpoch      bool
	successRegexp         *regexp.Regexp
	tailer                *tail.Tail
	timeWindowSlice       []Window
	throughputWindowSlice []Window
	timeTimer             *time.Timer
	throughputTimer       *time.Timer
	stopTimeChan          chan bool
	stopThroughputChan    chan bool
	timeMutex             sync.Mutex
	throughputMutex       sync.Mutex
	wg                    sync.WaitGroup
	sync.Mutex
}

func NewRequestAggregates() *RequestAggregates {
	return &RequestAggregates{
		TimeWindows:       2,
		ThroughputWindows: 10}
}

const sampleConfig = `
  # File to monitor.
  file = "/var/server/access.csv"
  # Position of the timestamp of the request in every line
  timestamp_position = 0
  # Format of the timestamp (any layout accepted by Go Time.Parse or s/ms/us/ns for epoch time)
  timestamp_format = "ms"
  # Position of the time value to calculate in the log file (starting from 0)
  time_position = 1
  # Window to consider for time percentiles
  time_window_size = "60s"
  # Windows to keep in memory before flushing in order to avoid requests coming in after a window is shut.
  # If the CSV file is sorted by timestamp, this can be set to 1
  time_windows = 5
  # List of percentiles to calculate
  time_percentiles = [90.0, 95.0, 99.0, 99.99]
  # Position of the result column (success or failure)
  result_position = 3
  # Regular expression used to determine if the result is successful or not (if empty only request_aggregates_all
  # time series) will be generated
  result_success_regex = ".*true.*"
  # Time window to calculate throughput counters
  throughput_window_size = "1s"
  # Number of windows to keep in memory for throughput calculation
  throughput_windows = 300
  # List of tags and their values to add to every data point
  [inputs.aggregates.tags]
    name = "myserver"
`

func (ra *RequestAggregates) SampleConfig() string {
	return sampleConfig
}

func (ra *RequestAggregates) Description() string {
	return "Generates a set of aggregate values for a requests and their response times."
}

func (ra *RequestAggregates) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (ra *RequestAggregates) Start(acc telegraf.Accumulator) error {
	ra.Lock()
	defer ra.Unlock()

	err := ra.validateConfig()
	if err != nil {
		return err
	}

	// Create tailer
	ra.tailer, err = tail.TailFile(ra.File, tail.Config{
		Follow:   true,
		ReOpen:   true,
		Location: &tail.SeekInfo{Whence: 2, Offset: 0}})
	if err != nil {
		return fmt.Errorf("ERROR tailing file %s, Error: %s", ra.File, err)
	}

	// Create first time window and start go routine to manage them
	now := time.Now()
	ra.timeWindowSlice = append(ra.timeWindowSlice, &TimeWindow{
		StartTime: now, EndTime: now.Add(ra.TimeWindowSize.Duration),
		OnlyTotal: ra.successRegexp == nil, Percentiles: ra.TimePercentiles})
	ra.timeTimer = time.NewTimer(ra.TimeWindowSize.Duration)
	ra.stopTimeChan = make(chan bool, 1)
	ra.wg.Add(1)
	go ra.manageTimeWindows(acc)

	// Create first throughput window and start go routine to manage them
	ra.throughputWindowSlice = append(ra.throughputWindowSlice, &ThroughputWindow{
		StartTime: now, EndTime: now.Add(ra.ThroughputWindowSize.Duration)})
	ra.throughputTimer = time.NewTimer(ra.ThroughputWindowSize.Duration)
	ra.stopThroughputChan = make(chan bool, 1)
	ra.wg.Add(1)
	go ra.manageThroughputWindows(acc)

	// Start go routine to tail the file and put requests in windows
	ra.wg.Add(1)
	go ra.gatherFromFile(ra.tailer, acc)

	return nil
}

func (ra *RequestAggregates) Stop() {
	ra.Lock()
	defer ra.Unlock()

	err := ra.tailer.Stop()
	if err != nil {
		log.Printf("ERROR: could not stop tail on file %s\n", ra.File)
	}
	ra.tailer.Cleanup()

	ra.timeTimer.Stop()
	ra.stopTimeChan <- true
	ra.throughputTimer.Stop()
	ra.stopThroughputChan <- true

	ra.wg.Wait()
}

// Validates the configuration in the struct
func (ra *RequestAggregates) validateConfig() error {
	var err error

	// Compile regex to identify success
	if ra.ResultSuccessRegex != "" {
		ra.successRegexp, err = regexp.Compile(ra.ResultSuccessRegex)
		if err != nil {
			return fmt.Errorf("ERROR: success regexp is not valid, Error: %s", err)
		}
	}
	// Check if timestamp format is valid
	switch ra.TimestampFormat {
	case "s", "ms", "us", "ns":
		ra.isTimestampEpoch = true
		break
	default:
		if time.Now().Format(ra.TimestampFormat) == ra.TimestampFormat {
			return fmt.Errorf("ERROR: incorrect timestamp format")
		}
	}
	// Check percentiles are valid
	for _, percentile := range ra.TimePercentiles {
		if percentile <= 0 || percentile >= 100 {
			return fmt.Errorf("ERROR: percentiles must be numbers between 0 and 100 (not inclusive)")
		}
	}
	//Check duration of windows
	if ra.TimeWindowSize.Duration <= time.Duration(0) || ra.ThroughputWindowSize.Duration <= time.Duration(0) {
		return fmt.Errorf("ERROR: windows need to be a positive duration")
	}
	// Check number of windows
	if ra.TimeWindows <= 0 || ra.ThroughputWindows <= 0 {
		return fmt.Errorf("ERROR: at least one window is required")
	}

	return nil
}

// Executed as a go routine, tails a given file and puts the parsed requests into their respective windows.
func (ra *RequestAggregates) gatherFromFile(tailer *tail.Tail, acc telegraf.Accumulator) {
	defer ra.wg.Done()

	requestParser := &RequestParser{
		TimestampPosition: ra.TimestampPosition,
		TimestampFormat:   ra.TimestampFormat,
		IsTimeEpoch:       ra.isTimestampEpoch,
		TimePosition:      ra.TimePosition,
		ResultPosition:    ra.ResultPosition,
		SuccessRegexp:     ra.successRegexp}

	var err error
	var line *tail.Line
	var request *Request
	for line = range tailer.Lines {
		// Parse and validate line
		if line.Err != nil {
			log.Printf("ERROR: could not tail file %s, Error: %s\n", tailer.Filename, err)
			continue
		}
		request, err = requestParser.ParseLine(line.Text)
		if err != nil {
			log.Printf("ERROR: malformed line in %s: [%s], Error: %s\n", tailer.Filename, line.Text, err)
			continue
		}

		// Wait until the window is created (it is possible that the line is read before the time ticks)
		for ra.timeWindowSlice[len(ra.timeWindowSlice)-1].End().Before(request.Timestamp) {
			time.Sleep(time.Millisecond * 10)
		}
		// Add request to time window
		ra.timeMutex.Lock()
		err = addToWindow(ra.timeWindowSlice, request)
		if err != nil {
			log.Printf("ERROR: could not find a time window, Request: %v, Error %s\n", request, err)
		}
		ra.timeMutex.Unlock()

		// Wait until the window is created (it is possible that the line is read before the time ticks)
		for ra.throughputWindowSlice[len(ra.throughputWindowSlice)-1].End().Before(request.Timestamp) {
			time.Sleep(time.Millisecond * 10)
		}
		// Add request to throughput window
		ra.throughputMutex.Lock()
		err = addToWindow(ra.throughputWindowSlice, request)
		if err != nil {
			log.Printf("ERROR: could not find a throughput window, Request: %v, Error %s\n", request, err)
		}
		ra.throughputMutex.Unlock()
	}
}

// Executed as a go routine, manages the windows related to time measures, creating new ones and flushing old ones
func (ra *RequestAggregates) manageTimeWindows(acc telegraf.Accumulator) {
	defer ra.wg.Done()
	onlyTotal := ra.successRegexp == nil
	for {
		select {
		// If the timer is triggered
		case <-ra.timeTimer.C:
			ra.timeMutex.Lock()
			// Create new window with the start time of the last one's end time
			startTime := ra.timeWindowSlice[len(ra.timeWindowSlice)-1].End()
			endTime := startTime.Add(ra.TimeWindowSize.Duration)
			ra.timeWindowSlice = append(ra.timeWindowSlice, &TimeWindow{
				StartTime: startTime, EndTime: endTime,
				OnlyTotal: onlyTotal, Percentiles: ra.TimePercentiles})
			// Flush oldest one if necessary
			if len(ra.timeWindowSlice) > ra.TimeWindows {
				ra.timeWindowSlice = flushWindow(ra.timeWindowSlice, acc)
			}
			ra.timeMutex.Unlock()
			// Reset time till the end of the window
			ra.timeTimer.Reset(endTime.Sub(time.Now()))
		// If the stop signal is received
		case <-ra.stopTimeChan:
			ra.timeMutex.Lock()
			ra.timeWindowSlice = flushAllWindows(ra.timeWindowSlice, acc)
			ra.timeMutex.Unlock()
			return
		}
	}
}

// Executed as a go routine, manages the windows related to throughput measures, creating new ones and flushing old ones
func (ra *RequestAggregates) manageThroughputWindows(acc telegraf.Accumulator) {
	defer ra.wg.Done()
	for {
		select {
		// If the timer is triggered
		case <-ra.throughputTimer.C:
			ra.throughputMutex.Lock()
			// Create new window with the start time of the last one's end time
			startTime := ra.throughputWindowSlice[len(ra.throughputWindowSlice)-1].End()
			endTime := startTime.Add(ra.ThroughputWindowSize.Duration)
			ra.throughputWindowSlice = append(ra.throughputWindowSlice, &ThroughputWindow{
				StartTime: startTime, EndTime: endTime})
			// Flush oldest one if necessary
			if len(ra.throughputWindowSlice) > ra.ThroughputWindows {
				ra.throughputWindowSlice = flushWindow(ra.throughputWindowSlice, acc)
			}
			ra.throughputMutex.Unlock()
			ra.throughputTimer.Reset(endTime.Sub(time.Now()))
		// If the stop signal is received
		case <-ra.stopThroughputChan:
			ra.throughputMutex.Lock()
			ra.throughputWindowSlice = flushAllWindows(ra.throughputWindowSlice, acc)
			ra.throughputMutex.Unlock()
			return
		}
	}
}

// Removes the window at the front of the slice of windows and flushes its aggregated metrics to the accumulator
func flushWindow(windows []Window, acc telegraf.Accumulator) []Window {
	if len(windows) > 0 {
		var window Window
		window, windows = windows[0], windows[1:]
		metrics, err := window.Aggregate()
		if err != nil {
			log.Printf("ERROR: could not flush window, Error: %s\n", err)
		}
		for _, metric := range metrics {
			acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
		}
	}
	return windows
}

// Flushes all windows ot the accumulator
func flushAllWindows(windows []Window, acc telegraf.Accumulator) []Window {
	for len(windows) > 0 {
		windows = flushWindow(windows, acc)
	}
	return windows
}

// Adds a request to a window, returns and error if it could not be added
func addToWindow(windows []Window, request *Request) error {
	if len(windows) == 0 {
		return fmt.Errorf("ERROR: no windows found")
	}
	first := windows[len(windows)-1]
	if first.End().Before(request.Timestamp) {
		return fmt.Errorf("ERROR: request is newer than any window")
	}
	last := windows[0]
	if last.Start().After(request.Timestamp) {
		return fmt.Errorf("ERROR: request is older than any window, try adding more windows")
	}
	for i := range windows {
		window := windows[i]
		if (window.Start().Before(request.Timestamp) || window.Start().Equal(request.Timestamp)) &&
			window.End().After(request.Timestamp) {
			return window.Add(request)
		}
	}
	return fmt.Errorf("ERROR: no window could be found")
}

func init() {
	inputs.Add("request_aggregates", func() telegraf.Input {
		return NewRequestAggregates()
	})
}

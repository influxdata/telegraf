package request_aggregates

import (
	"fmt"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
	"time"
)

// This tests the start/stop and gather functionality
func TestNewRequestAggregates(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	tmpfile.WriteString(fmt.Sprintf("%v", time.Now().UnixNano()) + ",123\n")
	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	windowSize := internal.Duration{Duration: time.Millisecond * 100}
	acc := &testutil.Accumulator{}
	ra := &RequestAggregates{
		File:                 tmpfile.Name(),
		TimestampFormat:      "ns",
		TimeWindowSize:       windowSize,
		TimeWindows:          1,
		ThroughputWindowSize: windowSize,
		ThroughputWindows:    10,
		TimestampPosition:    0,
		TimePosition:         1}
	// When we start
	ra.Start(acc)
	// A tailer is created
	ra.Lock()
	require.NotNil(t, ra.tailer)
	require.Equal(t, tmpfile.Name(), ra.tailer.Filename)
	ra.Unlock()
	// The windows are initialised
	ra.timeMutex.Lock()
	require.Equal(t, 1, len(ra.timeWindowSlice))
	ra.timeMutex.Unlock()
	ra.throughputMutex.Lock()
	require.Equal(t, 1, len(ra.throughputWindowSlice))
	ra.throughputMutex.Unlock()
	acc.Lock()
	require.Equal(t, 0, len(acc.Metrics))
	acc.Unlock()

	// When we put a request in the file
	time.Sleep(time.Millisecond * 30)
	tmpfile.WriteString(fmt.Sprintf("%v", time.Now().UnixNano()) + ",456\n")
	time.Sleep(time.Millisecond * 30)
	// One metric is stored in the windows
	ra.throughputMutex.Lock()
	require.Equal(t, int64(1), ra.throughputWindowSlice[0].(*ThroughputWindow).RequestsTotal)
	ra.throughputMutex.Unlock()
	ra.timeMutex.Lock()
	require.Equal(t, 1, len(ra.timeWindowSlice[0].(*TimeWindow).TimesTotal))
	require.Equal(t, float64(456), ra.timeWindowSlice[0].(*TimeWindow).TimesTotal[0])
	ra.timeMutex.Unlock()

	// After the first window is expired
	time.Sleep(windowSize.Duration)
	// One of the windows has flushed one metric
	ra.timeMutex.Lock()
	require.Equal(t, 1, len(ra.timeWindowSlice))
	ra.timeMutex.Unlock()
	ra.throughputMutex.Lock()
	require.Equal(t, 2, len(ra.throughputWindowSlice))
	ra.throughputMutex.Unlock()
	acc.Lock()
	require.Equal(t, 1, len(acc.Metrics))
	acc.Unlock()

	// When we stop
	ra.Stop()
	// All the metrics should have been flushed
	ra.timeMutex.Lock()
	require.Equal(t, 0, len(ra.timeWindowSlice))
	ra.timeMutex.Unlock()
	ra.throughputMutex.Lock()
	require.Equal(t, 0, len(ra.throughputWindowSlice))
	ra.throughputMutex.Unlock()
	acc.Lock()
	require.Equal(t, 4, len(acc.Metrics))
	acc.Unlock()
}

func TestRequestAggregates_validateConfig(t *testing.T) {
	// Empty config
	ra := &RequestAggregates{}
	require.Error(t, ra.validateConfig())
	// Minimum config
	ra = &RequestAggregates{
		TimestampFormat:      "ms",
		TimeWindowSize:       internal.Duration{Duration: time.Millisecond * 10},
		TimeWindows:          2,
		ThroughputWindowSize: internal.Duration{Duration: time.Millisecond * 10},
		ThroughputWindows:    10}
	require.NoError(t, ra.validateConfig())
	// Regexp for success
	ra.ResultSuccessRegex = "*success.*"
	require.Error(t, ra.validateConfig())
	ra.ResultSuccessRegex = ".*success.*"
	require.NoError(t, ra.validateConfig())
	// Time format
	ra.TimestampFormat = "thisisnotavalidformat"
	require.Error(t, ra.validateConfig())
	ra.TimestampFormat = ""
	require.Error(t, ra.validateConfig())
	ra.TimestampFormat = "Mon Jan _2 15:04:05 2006"
	require.NoError(t, ra.validateConfig())
	// Percentiles
	ra.TimePercentiles = []float32{80, 90, 100}
	require.Error(t, ra.validateConfig())
	ra.TimePercentiles = []float32{0, 90, 99}
	require.Error(t, ra.validateConfig())
	ra.TimePercentiles = []float32{80, 90, 99}
	require.NoError(t, ra.validateConfig())
	// Window size
	ra.TimeWindowSize = internal.Duration{Duration: time.Duration(0)}
	require.Error(t, ra.validateConfig())
	ra.TimeWindowSize = internal.Duration{Duration: time.Duration(-1)}
	require.Error(t, ra.validateConfig())
	ra.TimeWindowSize = internal.Duration{Duration: time.Duration(1)}
	require.NoError(t, ra.validateConfig())
	ra.ThroughputWindowSize = internal.Duration{Duration: time.Duration(0)}
	require.Error(t, ra.validateConfig())
	ra.ThroughputWindowSize = internal.Duration{Duration: time.Duration(-1)}
	require.Error(t, ra.validateConfig())
	ra.ThroughputWindowSize = internal.Duration{Duration: time.Duration(1)}
	require.NoError(t, ra.validateConfig())
	// Number of windows
	ra.TimeWindows = 0
	require.Error(t, ra.validateConfig())
	ra.TimeWindows = -1
	require.Error(t, ra.validateConfig())
	ra.TimeWindows = 1
	require.NoError(t, ra.validateConfig())
	ra.ThroughputWindows = 0
	require.Error(t, ra.validateConfig())
	ra.ThroughputWindows = -1
	require.Error(t, ra.validateConfig())
	ra.ThroughputWindows = 1
	require.NoError(t, ra.validateConfig())
}

func TestRequestAggregates_manageTimeWindows_OnlyTotal(t *testing.T) {
	windowSize := internal.Duration{Duration: time.Millisecond * 100}
	acc := &testutil.Accumulator{}
	now := time.Now()
	ra := &RequestAggregates{
		TimeWindows:     2,
		TimeWindowSize:  windowSize,
		TimePercentiles: []float32{70, 80, 90},
		timeTimer:       time.NewTimer(windowSize.Duration),
		stopTimeChan:    make(chan bool, 1)}

	// Add first window and start routine
	ra.timeWindowSlice = append(ra.timeWindowSlice, &TimeWindow{
		StartTime: now, EndTime: now.Add(windowSize.Duration), OnlyTotal: true, Percentiles: ra.TimePercentiles})
	ra.wg.Add(1)
	go ra.manageTimeWindows(acc)

	// Check values at different points
	time.Sleep(time.Millisecond * 30)
	ra.timeMutex.Lock()
	require.Equal(t, 1, len(ra.timeWindowSlice))
	ra.timeMutex.Unlock()
	acc.Lock()
	require.Equal(t, 0, len(acc.Metrics))
	acc.Unlock()
	time.Sleep(windowSize.Duration)
	ra.timeMutex.Lock()
	require.Equal(t, 2, len(ra.timeWindowSlice))
	ra.timeMutex.Unlock()
	acc.Lock()
	require.Equal(t, 0, len(acc.Metrics))
	acc.Unlock()
	time.Sleep(windowSize.Duration)
	ra.timeMutex.Lock()
	require.Equal(t, 2, len(ra.timeWindowSlice))
	ra.timeMutex.Unlock()
	acc.Lock()
	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, now.Add(windowSize.Duration), acc.Metrics[0].Time)
	acc.Unlock()

	// Stop and wait for the process to finish
	ra.timeTimer.Stop()
	ra.stopTimeChan <- true
	ra.wg.Wait()

	// Check that all metrics were flushed
	ra.timeMutex.Lock()
	require.Equal(t, 0, len(ra.timeWindowSlice))
	ra.timeMutex.Unlock()
	acc.Lock()
	require.Equal(t, 3, len(acc.Metrics))
	require.Equal(t, now.Add(windowSize.Duration).Add(windowSize.Duration), acc.Metrics[1].Time)
	require.Equal(t, now.Add(windowSize.Duration).Add(windowSize.Duration).Add(windowSize.Duration), acc.Metrics[2].Time)
	acc.Unlock()
}

func TestRequestAggregates_manageTimeWindows_All(t *testing.T) {
	windowSize := internal.Duration{Duration: time.Millisecond * 100}
	acc := &testutil.Accumulator{}
	now := time.Now()
	ra := &RequestAggregates{
		TimeWindows:     2,
		TimeWindowSize:  windowSize,
		TimePercentiles: []float32{70, 80, 90},
		successRegexp:   regexp.MustCompile(".*success.*"),
		timeTimer:       time.NewTimer(windowSize.Duration),
		stopTimeChan:    make(chan bool, 1)}

	// Add first window and start routine
	ra.timeWindowSlice = append(ra.timeWindowSlice, &TimeWindow{
		StartTime: now, EndTime: now.Add(windowSize.Duration), OnlyTotal: false, Percentiles: ra.TimePercentiles})
	ra.wg.Add(1)
	go ra.manageTimeWindows(acc)

	// Check values at different points
	time.Sleep(time.Millisecond * 30)
	ra.timeMutex.Lock()
	require.Equal(t, 1, len(ra.timeWindowSlice))
	ra.timeMutex.Unlock()
	acc.Lock()
	require.Equal(t, 0, len(acc.Metrics))
	acc.Unlock()
	time.Sleep(windowSize.Duration)
	ra.timeMutex.Lock()
	require.Equal(t, 2, len(ra.timeWindowSlice))
	ra.timeMutex.Unlock()
	acc.Lock()
	require.Equal(t, 0, len(acc.Metrics))
	acc.Unlock()
	time.Sleep(windowSize.Duration)
	ra.timeMutex.Lock()
	require.Equal(t, 2, len(ra.timeWindowSlice))
	ra.timeMutex.Unlock()
	acc.Lock()
	require.Equal(t, 3, len(acc.Metrics))
	require.Equal(t, now.Add(windowSize.Duration), acc.Metrics[0].Time)
	require.Equal(t, now.Add(windowSize.Duration), acc.Metrics[1].Time)
	require.Equal(t, now.Add(windowSize.Duration), acc.Metrics[2].Time)
	acc.Unlock()

	// Stop and wait for the process to finish
	ra.timeTimer.Stop()
	ra.stopTimeChan <- true
	ra.wg.Wait()

	// Check that all metrics were flushed
	ra.timeMutex.Lock()
	require.Equal(t, 0, len(ra.timeWindowSlice))
	ra.timeMutex.Unlock()
	acc.Lock()
	require.Equal(t, 9, len(acc.Metrics))
	require.Equal(t, now.Add(windowSize.Duration).Add(windowSize.Duration), acc.Metrics[3].Time)
	require.Equal(t, now.Add(windowSize.Duration).Add(windowSize.Duration), acc.Metrics[4].Time)
	require.Equal(t, now.Add(windowSize.Duration).Add(windowSize.Duration), acc.Metrics[5].Time)
	require.Equal(t, now.Add(windowSize.Duration).Add(windowSize.Duration).Add(windowSize.Duration), acc.Metrics[6].Time)
	require.Equal(t, now.Add(windowSize.Duration).Add(windowSize.Duration).Add(windowSize.Duration), acc.Metrics[7].Time)
	require.Equal(t, now.Add(windowSize.Duration).Add(windowSize.Duration).Add(windowSize.Duration), acc.Metrics[8].Time)
	acc.Unlock()
}

func TestRequestAggregates_manageThroughputWindows(t *testing.T) {
	windowSize := internal.Duration{Duration: time.Millisecond * 100}
	acc := &testutil.Accumulator{}
	now := time.Now()
	ra := &RequestAggregates{
		ThroughputWindows:    2,
		ThroughputWindowSize: windowSize,
		throughputTimer:      time.NewTimer(windowSize.Duration),
		stopThroughputChan:   make(chan bool, 1)}

	// Add first window and start routine
	ra.throughputWindowSlice = append(ra.throughputWindowSlice, &ThroughputWindow{
		StartTime: now, EndTime: now.Add(windowSize.Duration)})
	ra.wg.Add(1)
	go ra.manageThroughputWindows(acc)

	// Check values at different points
	time.Sleep(time.Millisecond * 30)
	ra.throughputMutex.Lock()
	require.Equal(t, 1, len(ra.throughputWindowSlice))
	ra.throughputMutex.Unlock()
	acc.Lock()
	require.Equal(t, 0, len(acc.Metrics))
	acc.Unlock()
	time.Sleep(windowSize.Duration)
	ra.throughputMutex.Lock()
	require.Equal(t, 2, len(ra.throughputWindowSlice))
	ra.throughputMutex.Unlock()
	acc.Lock()
	require.Equal(t, 0, len(acc.Metrics))
	acc.Unlock()
	time.Sleep(windowSize.Duration)
	ra.throughputMutex.Lock()
	require.Equal(t, 2, len(ra.throughputWindowSlice))
	ra.throughputMutex.Unlock()
	acc.Lock()
	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, now.Add(windowSize.Duration), acc.Metrics[0].Time)
	acc.Unlock()

	// Stop and wait for the process to finish
	ra.throughputTimer.Stop()
	ra.stopThroughputChan <- true
	ra.wg.Wait()

	// Check that all metrics were flushed
	ra.throughputMutex.Lock()
	require.Equal(t, 0, len(ra.throughputWindowSlice))
	ra.throughputMutex.Unlock()
	acc.Lock()
	require.Equal(t, 3, len(acc.Metrics))
	require.Equal(t, now.Add(windowSize.Duration).Add(windowSize.Duration), acc.Metrics[1].Time)
	require.Equal(t, now.Add(windowSize.Duration).Add(windowSize.Duration).Add(windowSize.Duration), acc.Metrics[2].Time)
	acc.Unlock()
}

func TestRequestAggregates_flushWindow(t *testing.T) {
	acc := &testutil.Accumulator{}
	now := time.Now()
	windows := []Window{&ThroughputWindow{StartTime: now, EndTime: now.Add(time.Duration(60))}}
	windows = flushWindow(windows, acc)
	require.Equal(t, 0, len(windows))
	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, MeasurementThroughput, acc.Metrics[0].Measurement)
}

func TestRequestAggregates_flushAllWindows(t *testing.T) {
	acc := &testutil.Accumulator{}
	now := time.Now()
	windows := []Window{&ThroughputWindow{StartTime: now, EndTime: now.Add(time.Duration(60))},
		&ThroughputWindow{StartTime: now.Add(time.Duration(60)), EndTime: now.Add(time.Duration(120))},
		&ThroughputWindow{StartTime: now.Add(time.Duration(120)), EndTime: now.Add(time.Duration(180))}}
	windows = flushAllWindows(windows, acc)
	require.Equal(t, 0, len(windows))
	require.Equal(t, 3, len(acc.Metrics))
}

func TestRequestAggregates_addToWindow(t *testing.T) {
	now := time.Now()
	var windows []Window
	// Error if there are no windows (not added)
	err := addToWindow(windows, &Request{Timestamp: now.Add(time.Duration(30))})
	require.Error(t, err)
	// Okay when one window
	firstWindow := &ThroughputWindow{StartTime: now, EndTime: now.Add(time.Duration(60))}
	windows = append(windows, firstWindow)
	err = addToWindow(windows, &Request{Timestamp: now.Add(time.Duration(30))})
	require.NoError(t, err)
	require.Equal(t, int64(1), firstWindow.RequestsTotal)
	// Okay when timestamp equal to start of window
	err = addToWindow(windows, &Request{Timestamp: now})
	require.NoError(t, err)
	require.Equal(t, int64(2), firstWindow.RequestsTotal)
	// Error when timestamp equal to end of window
	err = addToWindow(windows, &Request{Timestamp: now.Add(time.Duration(60))})
	require.Error(t, err)
	// Okay with more windows
	middleWindow := &ThroughputWindow{StartTime: now.Add(time.Duration(60)), EndTime: now.Add(time.Duration(120))}
	lastWindow := &ThroughputWindow{StartTime: now.Add(time.Duration(120)), EndTime: now.Add(time.Duration(180))}
	windows = append(windows, middleWindow)
	windows = append(windows, lastWindow)
	err = addToWindow(windows, &Request{Timestamp: now.Add(time.Duration(90))})
	require.NoError(t, err)
	require.Equal(t, int64(1), middleWindow.RequestsTotal)
	err = addToWindow(windows, &Request{Timestamp: now.Add(time.Duration(150))})
	require.NoError(t, err)
	require.Equal(t, int64(1), lastWindow.RequestsTotal)
	// Error when later than last window
	err = addToWindow(windows, &Request{Timestamp: now.Add(time.Duration(220))})
	require.Error(t, err)
	// Error when before first window
	err = addToWindow(windows, &Request{Timestamp: now.Add(time.Duration(-20))})
	require.Error(t, err)
}

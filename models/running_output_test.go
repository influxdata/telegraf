package models

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/selfstat"
	"github.com/influxdata/telegraf/testutil"
)

var first5 = []telegraf.Metric{
	testutil.TestMetric(101, "metric1"),
	testutil.TestMetric(101, "metric2"),
	testutil.TestMetric(101, "metric3"),
	testutil.TestMetric(101, "metric4"),
	testutil.TestMetric(101, "metric5"),
}

var next5 = []telegraf.Metric{
	testutil.TestMetric(101, "metric6"),
	testutil.TestMetric(101, "metric7"),
	testutil.TestMetric(101, "metric8"),
	testutil.TestMetric(101, "metric9"),
	testutil.TestMetric(101, "metric10"),
}

// Test that NameDrop filters ger properly applied.
func TestRunningOutputDropFilter(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			NameDrop: []string{"metric1", "metric2"},
		},
	}
	require.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 8)
}

// Test that NameDrop filters without a match do nothing.
func TestRunningOutputPassFilter(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			NameDrop: []string{"metric1000", "foo*"},
		},
	}
	require.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 10)
}

// Test that tags are properly included
func TestRunningOutputTagIncludeNoMatch(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			TagInclude: []string{"nothing*"},
		},
	}
	require.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Empty(t, m.Metrics()[0].Tags())
}

// Test that tags are properly excluded
func TestRunningOutputTagExcludeMatch(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			TagExclude: []string{"tag*"},
		},
	}
	require.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Empty(t, m.Metrics()[0].Tags())
}

// Test that tags are properly Excluded
func TestRunningOutputTagExcludeNoMatch(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			TagExclude: []string{"nothing*"},
		},
	}
	require.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Len(t, m.Metrics()[0].Tags(), 1)
}

// Test that tags are properly included
func TestRunningOutputTagIncludeMatch(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{
			TagInclude: []string{"tag*"},
		},
	}
	require.NoError(t, conf.Filter.Compile())

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Len(t, m.Metrics()[0].Tags(), 1)
}

// Test that measurement name overriding correctly
func TestRunningOutputNameOverride(t *testing.T) {
	conf := &OutputConfig{
		NameOverride: "new_metric_name",
	}

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Equal(t, "new_metric_name", m.Metrics()[0].Name())
}

// Test that measurement name prefix is added correctly
func TestRunningOutputNamePrefix(t *testing.T) {
	conf := &OutputConfig{
		NamePrefix: "prefix_",
	}

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Equal(t, "prefix_metric1", m.Metrics()[0].Name())
}

// Test that measurement name suffix is added correctly
func TestRunningOutputNameSuffix(t *testing.T) {
	conf := &OutputConfig{
		NameSuffix: "_suffix",
	}

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	ro.AddMetric(testutil.TestMetric(101, "metric1"))
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 1)
	require.Equal(t, "metric1_suffix", m.Metrics()[0].Name())
}

// Test that we can write metrics with simple default setup.
func TestRunningOutputDefault(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	require.Empty(t, m.Metrics())

	err := ro.Write()
	require.NoError(t, err)
	require.Len(t, m.Metrics(), 10)
}

func TestRunningOutputWriteFail(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{batchAcceptSize: -1}
	ro := NewRunningOutput(m, conf, 4, 12)

	// Fill buffer to limit twice
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// manual write fails
	err := ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	m.batchAcceptSize = 0
	err = ro.Write()
	require.NoError(t, err)

	require.Len(t, m.Metrics(), 10)
}

// Verify that the order of points is preserved during write failure.
func TestRunningOutputWriteFailOrder(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{batchAcceptSize: -1}
	ro := NewRunningOutput(m, conf, 100, 1000)

	// add 5 metrics
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// Write fails
	err := ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	m.batchAcceptSize = 0

	// add 5 more metrics
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	err = ro.Write()
	require.NoError(t, err)

	// Verify that 10 metrics were written
	require.Len(t, m.Metrics(), 10)
	// Verify that they are in order
	expected := append(first5, next5...)
	require.Equal(t, expected, m.Metrics())
}

// Verify that the order of points is preserved during many write failures.
func TestRunningOutputWriteFailOrder2(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{batchAcceptSize: -1}
	ro := NewRunningOutput(m, conf, 5, 100)

	// add 5 metrics
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	// Write fails
	err := ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// add 5 metrics
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	// Write fails
	err = ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// add 5 metrics
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	// Write fails
	err = ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// add 5 metrics
	for _, metric := range next5 {
		ro.AddMetric(metric)
	}
	// Write fails
	err = ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	m.batchAcceptSize = 0
	err = ro.Write()
	require.NoError(t, err)

	// Verify that 20 metrics were written
	require.Len(t, m.Metrics(), 20)
	// Verify that they are in order
	expected := append(first5, next5...)
	expected = append(expected, first5...)
	expected = append(expected, next5...)
	require.Equal(t, expected, m.Metrics())
}

// Verify that the order of points is preserved when there is a remainder
// of points for the batch.
func TestRunningOutputWriteFailOrder3(t *testing.T) {
	conf := &OutputConfig{
		Filter: Filter{},
	}

	m := &mockOutput{batchAcceptSize: -1}
	ro := NewRunningOutput(m, conf, 5, 1000)

	// add 5 metrics
	for _, metric := range first5 {
		ro.AddMetric(metric)
	}
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// Write fails
	err := ro.Write()
	require.Error(t, err)
	// no successful flush yet
	require.Empty(t, m.Metrics())

	// add and attempt to write a single metric:
	ro.AddMetric(next5[0])
	err = ro.Write()
	require.Error(t, err)

	// unset fail and write metrics
	m.batchAcceptSize = 0

	err = ro.Write()
	require.NoError(t, err)

	// Verify that 6 metrics were written
	require.Len(t, m.Metrics(), 6)
	// Verify that they are in order
	expected := []telegraf.Metric{first5[0], first5[1], first5[2], first5[3], first5[4], next5[0]}
	require.Equal(t, expected, m.Metrics())
}

func TestRunningOutputBufferFullyDrained(t *testing.T) {
	// Setup output with a post-write hook to be able to block write until
	// we added more metrics
	conf := &OutputConfig{
		Filter: Filter{},
	}
	var shouldBlock atomic.Bool
	shouldBlock.Store(true)
	addMore := make(chan bool)
	defer func() { close(addMore) }()
	waitForAddedMetrics := make(chan bool)
	defer func() { close(waitForAddedMetrics) }()
	plugin := &mockOutput{
		batchAcceptSize: 0,
		postWriteHook: func([]telegraf.Metric) error {
			// Wait for the first full write and block until the test code
			// added the new metrics
			if shouldBlock.CompareAndSwap(true, false) {
				addMore <- true
				<-waitForAddedMetrics
			}
			return nil
		},
	}
	const batchSize = 5
	ro := NewRunningOutput(plugin, conf, batchSize, 100)

	// Create a multiple of batch size many metrics beyond the batch size
	const totalMetrics = 10 * batchSize
	inputs := make([]telegraf.Metric, 0, totalMetrics)
	for i := range totalMetrics {
		inputs = append(inputs, testutil.TestMetric(i, "test"))
	}

	// Setup a event based writing loop similar to what the agent code does.
	// Remember the first write will block to allow us adding more metrics.
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var wg sync.WaitGroup
	var modelWriteErr error
	wg.Add(1)
	go func(cctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-cctx.Done():
				return
			case <-ro.BatchReady:
				if modelWriteErr = ro.Write(); modelWriteErr != nil {
					return
				}
			}
		}
	}(ctx)

	// Add a few metrics, i.e. more than batch size
	for _, m := range inputs[:20] {
		ro.AddMetric(m)
	}

	// Wait for entering the actual output write and add the remaining metrics.
	// Afterwards unblock the writer.
	<-addMore
	for _, m := range inputs[20:] {
		ro.AddMetric(m)
	}
	waitForAddedMetrics <- true

	// Wait for writing to finish and stop the write loop
	require.Eventually(t, func() bool {
		return len(plugin.Metrics()) >= len(inputs)
	}, 3*time.Second, 100*time.Millisecond)
	cancel()
	wg.Wait()

	// Check for writing errors and make sure all metrics were written,
	// including the ones added while writing took place
	require.NoError(t, modelWriteErr)
	require.Equal(t, 10, int(plugin.writes.Load()))
	require.Len(t, plugin.Metrics(), totalMetrics)
}

func TestRunningOutputBufferImmediateRestartOnContinuousWrite(t *testing.T) {
	// Setup output with a post-write hook to be able to block write until
	// we added more metrics
	conf := &OutputConfig{
		Filter: Filter{},
	}

	const batchSize = 5
	var shouldBlock atomic.Bool
	shouldBlock.Store(true)
	addMore := make(chan bool)
	defer func() { close(addMore) }()
	waitForAddedMetrics := make(chan bool)
	defer func() { close(waitForAddedMetrics) }()
	plugin := &mockOutput{
		batchAcceptSize: 0,
		preWriteHook: func(ms []telegraf.Metric) error {
			// Wait for the first non-full write and block until the test code
			// added the new metrics
			if len(ms) < batchSize && shouldBlock.CompareAndSwap(true, false) {
				addMore <- true
				<-waitForAddedMetrics
			}
			return nil
		},
	}
	ro := NewRunningOutput(plugin, conf, batchSize, 100)

	// Create a multiple of batch size many metrics beyond the batch size
	const totalMetrics = 10 * batchSize
	inputs := make([]telegraf.Metric, 0, totalMetrics)
	for i := range totalMetrics {
		inputs = append(inputs, testutil.TestMetric(i, "test"))
	}

	// Add a few metrics but not a multiple of the batch size
	for _, m := range inputs[:19] {
		ro.AddMetric(m)
	}

	// Start writing and add new metrics as soon as the last non-full batch is
	// written. At this time add the remaining metrics to check if we are
	// immediately getting a new write signal.
	var modelWriteErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		modelWriteErr = ro.Write()
	}()

	// Wait for the writer to see the non-full batch, add the remaining metrics
	// and unblock the writer
	<-addMore
	for _, m := range inputs[19:] {
		ro.AddMetric(m)
	}
	waitForAddedMetrics <- true

	// Wait for writing to finish
	wg.Wait()
	require.NoError(t, modelWriteErr)

	// Check for the new-batch-available trigger
	require.Eventually(t, func() bool {
		select {
		case <-ro.BatchReady:
			return true
		default:
			return false
		}
	}, 3*time.Second, 100*time.Millisecond)

	// Trigger the requested write and make sure all metrics were written,
	// including the ones added while writing took place
	require.NoError(t, ro.Write())
	require.Len(t, plugin.Metrics(), totalMetrics)
}

func TestRunningOutputNoRetriggerOnError(t *testing.T) {
	// Setup output with a post-write hook to be able to block write until
	// we added more metrics
	conf := &OutputConfig{
		Filter: Filter{},
	}

	plugin := &mockOutput{
		batchAcceptSize: 0,
		preWriteHook: func([]telegraf.Metric) error {
			// In this test we are handling a failing output
			return errors.New("writing failed")
		},
	}
	const batchSize = 5
	ro := NewRunningOutput(plugin, conf, batchSize, 100)

	// Create a multiple of batch size many metrics beyond the batch size
	const totalMetrics = 10 * batchSize
	inputs := make([]telegraf.Metric, 0, totalMetrics)
	for i := range totalMetrics {
		inputs = append(inputs, testutil.TestMetric(i, "test"))
	}

	// Add the metrics
	for _, m := range inputs {
		ro.AddMetric(m)
	}

	// Setup a event based writing loop similar to what the agent code does.
	// Remember the first write will block to allow us adding more metrics.
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	var errCount atomic.Uint32
	var wg sync.WaitGroup
	wg.Add(1)
	go func(cctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-cctx.Done():
				return
			case <-ro.BatchReady:
				if err := ro.Write(); err != nil {
					errCount.Add(1)
				}
			}
		}
	}(ctx)

	// Wait for the trigger loop to exit. This should happen latest after the
	// defined timeout.
	wg.Wait()

	// Check for writing errors and make sure all metrics were written,
	// including the ones added while writing took place
	require.Equal(t, errCount.Load(), plugin.writes.Load())
	require.Equal(t, 1, int(plugin.writes.Load()))
	require.Equal(t, totalMetrics, ro.buffer.Len())
}

func TestRunningOutputNoRetriggerOnSuccessfulPartialWriteError(t *testing.T) {
	// Setup output with a post-write hook to be able to block write until
	// we added more metrics
	conf := &OutputConfig{
		Filter: Filter{},
	}

	plugin := &mockOutput{
		batchAcceptSize: 0,
		preWriteHook: func(m []telegraf.Metric) error {
			// In this test we are handling a failing output
			drop := make([]int, 0, len(m)-1)
			for i := range len(m) - 1 {
				drop = append(drop, i+1)
			}
			return &internal.PartialWriteError{
				Err:           errors.New("writing failed"),
				MetricsAccept: []int{0},
				MetricsReject: drop,
			}
		},
	}
	const batchSize = 5
	ro := NewRunningOutput(plugin, conf, batchSize, 100)

	// Create a multiple of batch size many metrics beyond the batch size
	const batchCount = 10
	const totalMetrics = batchCount * batchSize
	inputs := make([]telegraf.Metric, 0, totalMetrics)
	for i := range totalMetrics {
		inputs = append(inputs, testutil.TestMetric(i, "test"))
	}

	// Add the metrics
	for _, m := range inputs {
		ro.AddMetric(m)
	}

	// Setup a event based writing loop similar to what the agent code does.
	// Remember the first write will block to allow us adding more metrics.
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var triggerCount atomic.Uint32
	var errCount atomic.Uint32
	var wg sync.WaitGroup
	wg.Add(1)
	go func(cctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-cctx.Done():
				return
			case <-ro.BatchReady:
				triggerCount.Add(1)
				if err := ro.Write(); err != nil {
					errCount.Add(1)
				}
			}
		}
	}(ctx)

	// Wait for the trigger loop to exit. This should happen latest after the
	// defined timeout.
	require.Eventually(t, func() bool { return ro.buffer.Len() == 0 }, 3*time.Second, 100*time.Millisecond)
	cancel()
	wg.Wait()

	// Check for writing errors and make sure all metrics were written,
	// including the ones added while writing took place
	require.Equal(t, errCount.Load(), plugin.writes.Load())
	require.Equal(t, triggerCount.Load(), plugin.writes.Load())
	require.Equal(t, batchCount, int(plugin.writes.Load()))
}

func TestRunningOutputNoRetriggerOnUnsuccessfulPartialWriteError(t *testing.T) {
	// Setup output with a post-write hook to be able to block write until
	// we added more metrics
	conf := &OutputConfig{
		Filter: Filter{},
	}

	plugin := &mockOutput{
		batchAcceptSize: 0,
		preWriteHook: func([]telegraf.Metric) error {
			return &internal.PartialWriteError{
				Err: errors.New("unable to connect"),
			}
		},
	}
	const batchSize = 5
	ro := NewRunningOutput(plugin, conf, batchSize, 100)

	// Create a multiple of batch size many metrics beyond the batch size
	const batchCount = 10
	const totalMetrics = batchCount * batchSize
	inputs := make([]telegraf.Metric, 0, totalMetrics)
	for i := range totalMetrics {
		inputs = append(inputs, testutil.TestMetric(i, "test"))
	}

	// Add the metrics
	for _, m := range inputs {
		ro.AddMetric(m)
	}

	// Setup a event based writing loop similar to what the agent code does.
	// Remember the first write will block to allow us adding more metrics.
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	var triggerCount atomic.Uint32
	var errCount atomic.Uint32
	var wg sync.WaitGroup
	wg.Add(1)
	go func(cctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-cctx.Done():
				return
			case <-ro.BatchReady:
				triggerCount.Add(1)
				if err := ro.Write(); err != nil {
					errCount.Add(1)
				}
			}
		}
	}(ctx)

	// Wait for the trigger loop to exit. This should happen latest after the
	// defined timeout.
	wg.Wait()
	cancel()

	// Check for writing errors and make sure all metrics were written,
	// including the ones added while writing took place
	require.Equal(t, 1, int(triggerCount.Load()))
}

func TestRunningOutputInternalMetrics(t *testing.T) {
	_ = NewRunningOutput(
		&mockOutput{},
		&OutputConfig{
			Filter: Filter{},
			Name:   "test_name",
			Alias:  "test_alias",
		},
		5,
		10)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"internal_write",
			map[string]string{
				"_id":    "",
				"output": "test_name",
				"alias":  "test_alias",
			},
			map[string]interface{}{
				"buffer_limit":     10,
				"buffer_size":      0,
				"errors":           0,
				"metrics_added":    0,
				"metrics_rejected": 0,
				"metrics_dropped":  0,
				"metrics_filtered": 0,
				"metrics_written":  0,
				"write_time_ns":    0,
				"startup_errors":   0,
			},
			time.Unix(0, 0),
		),
	}

	var actual []telegraf.Metric
	for _, m := range selfstat.Metrics() {
		output, _ := m.GetTag("output")
		if m.Name() == "internal_write" && output == "test_name" {
			actual = append(actual, m)
		}
	}
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
}

func TestRunningOutputStartupBehaviorInvalid(t *testing.T) {
	ro := NewRunningOutput(
		&mockOutput{},
		&OutputConfig{
			Filter:               Filter{},
			Name:                 "test_name",
			Alias:                "test_alias",
			StartupErrorBehavior: "foo",
		},
		5, 10,
	)
	require.ErrorContains(t, ro.Init(), "invalid 'startup_error_behavior'")
}

func TestRunningOutputRetryableStartupBehaviorDefault(t *testing.T) {
	serr := &internal.StartupError{
		Err:   errors.New("retryable err"),
		Retry: true,
	}
	ro := NewRunningOutput(
		&mockOutput{
			startupErrorCount: 1,
			startupError:      serr,
		},
		&OutputConfig{
			Filter: Filter{},
			Name:   "test_name",
			Alias:  "test_alias",
		},
		5, 10,
	)
	require.NoError(t, ro.Init())

	// If Connect() fails, the agent will stop
	require.ErrorIs(t, ro.Connect(), serr)
	require.False(t, ro.started)
}

func TestRunningOutputRetryableStartupBehaviorError(t *testing.T) {
	serr := &internal.StartupError{
		Err:   errors.New("retryable err"),
		Retry: true,
	}
	ro := NewRunningOutput(
		&mockOutput{
			startupErrorCount: 1,
			startupError:      serr,
		},
		&OutputConfig{
			Filter:               Filter{},
			Name:                 "test_name",
			Alias:                "test_alias",
			StartupErrorBehavior: "error",
		},
		5, 10,
	)
	require.NoError(t, ro.Init())

	// If Connect() fails, the agent will stop
	require.ErrorIs(t, ro.Connect(), serr)
	require.False(t, ro.started)
}

func TestRunningOutputRetryableStartupBehaviorRetry(t *testing.T) {
	serr := &internal.StartupError{
		Err:   errors.New("retryable err"),
		Retry: true,
	}
	mo := &mockOutput{
		startupErrorCount: 2,
		startupError:      serr,
	}
	ro := NewRunningOutput(
		mo,
		&OutputConfig{
			Filter:               Filter{},
			Name:                 "test_name",
			Alias:                "test_alias",
			StartupErrorBehavior: "retry",
		},
		5, 10,
	)
	require.NoError(t, ro.Init())

	// For retry, Connect() should succeed even though there is an error but
	// should return an error on Write() until we successfully connect.
	require.NotErrorIs(t, ro.Connect(), serr)
	require.False(t, ro.started)

	ro.AddMetric(testutil.TestMetric(1))
	require.ErrorIs(t, ro.Write(), internal.ErrNotConnected)
	require.False(t, ro.started)

	ro.AddMetric(testutil.TestMetric(2))
	require.NoError(t, ro.Write())
	require.True(t, ro.started)
	require.Equal(t, 1, int(mo.writes.Load()))

	ro.AddMetric(testutil.TestMetric(3))
	require.NoError(t, ro.Write())
	require.True(t, ro.started)
	require.Equal(t, 2, int(mo.writes.Load()))
}

func TestRunningOutputRetryableStartupBehaviorIgnore(t *testing.T) {
	serr := &internal.StartupError{
		Err:   errors.New("retryable err"),
		Retry: true,
	}
	mo := &mockOutput{
		startupErrorCount: 2,
		startupError:      serr,
	}
	ro := NewRunningOutput(
		mo,
		&OutputConfig{
			Filter:               Filter{},
			Name:                 "test_name",
			Alias:                "test_alias",
			StartupErrorBehavior: "ignore",
		},
		5, 10,
	)
	require.NoError(t, ro.Init())

	// For ignore, Connect() should return a fatal error if connection fails.
	// This will force the agent to remove the plugin.
	var fatalErr *internal.FatalError
	require.ErrorAs(t, ro.Connect(), &fatalErr)
	require.ErrorIs(t, fatalErr, serr)
	require.False(t, ro.started)
}

func TestRunningOutputNonRetryableStartupBehaviorDefault(t *testing.T) {
	serr := &internal.StartupError{
		Err:   errors.New("non-retryable err"),
		Retry: false,
	}

	for _, behavior := range []string{"", "error", "retry", "ignore"} {
		t.Run(behavior, func(t *testing.T) {
			mo := &mockOutput{
				startupErrorCount: 2,
				startupError:      serr,
			}
			ro := NewRunningOutput(
				mo,
				&OutputConfig{
					Filter:               Filter{},
					Name:                 "test_name",
					Alias:                "test_alias",
					StartupErrorBehavior: behavior,
				},
				5, 10,
			)
			require.NoError(t, ro.Init())

			// Non-retryable error should pass through and in turn the agent
			// will stop and exit.
			require.ErrorIs(t, ro.Connect(), serr)
			require.False(t, ro.started)
		})
	}
}

func TestRunningOutputUntypedStartupBehaviorIgnore(t *testing.T) {
	serr := errors.New("untyped err")

	for _, behavior := range []string{"", "error", "retry", "ignore"} {
		t.Run(behavior, func(t *testing.T) {
			mo := &mockOutput{
				startupErrorCount: 2,
				startupError:      serr,
			}
			ro := NewRunningOutput(
				mo,
				&OutputConfig{
					Filter:               Filter{},
					Name:                 "test_name",
					Alias:                "test_alias",
					StartupErrorBehavior: behavior,
				},
				5, 10,
			)
			require.NoError(t, ro.Init())

			// Untyped error should pass through and in turn the agent will
			// stop and exit.
			require.ErrorIs(t, ro.Connect(), serr)
			require.False(t, ro.started)
		})
	}
}

func TestRunningOutputPartiallyStarted(t *testing.T) {
	serr := &internal.StartupError{
		Err:     errors.New("partial err"),
		Retry:   true,
		Partial: true,
	}
	mo := &mockOutput{
		startupErrorCount: 2,
		startupError:      serr,
	}
	ro := NewRunningOutput(
		mo,
		&OutputConfig{
			Filter:               Filter{},
			Name:                 "test_name",
			Alias:                "test_alias",
			StartupErrorBehavior: "retry",
		},
		5, 10,
	)
	require.NoError(t, ro.Init())

	// For retry, Connect() should succeed even though there is an error but
	// should return an error on Write() until we successfully connect.
	require.NotErrorIs(t, ro.Connect(), serr)
	require.False(t, ro.started)

	ro.AddMetric(testutil.TestMetric(1))
	require.NoError(t, ro.Write())
	require.False(t, ro.started)
	require.Equal(t, 1, int(mo.writes.Load()))

	ro.AddMetric(testutil.TestMetric(2))
	require.NoError(t, ro.Write())
	require.True(t, ro.started)
	require.Equal(t, 2, int(mo.writes.Load()))

	ro.AddMetric(testutil.TestMetric(3))
	require.NoError(t, ro.Write())
	require.True(t, ro.started)
	require.Equal(t, 3, int(mo.writes.Load()))
}

func TestRunningOutputWritePartialSuccess(t *testing.T) {
	plugin := &mockOutput{
		batchAcceptSize: 4,
	}
	model := NewRunningOutput(plugin, &OutputConfig{}, 5, 10)
	require.NoError(t, model.Init())
	require.NoError(t, model.Connect())
	defer model.Close()

	// Fill buffer completely
	for _, metric := range first5 {
		model.AddMetric(metric)
	}
	for _, metric := range next5 {
		model.AddMetric(metric)
	}

	// We no not expect any successful flush yet
	require.Empty(t, plugin.Metrics())
	require.Equal(t, 10, model.buffer.Len())

	// Write to the output. This should only partially succeed with the first
	// few metrics removed from buffer
	require.ErrorIs(t, model.Write(), internal.ErrSizeLimitReached)
	require.Len(t, plugin.metrics, 4)
	require.Equal(t, 6, model.buffer.Len())

	// The next write should remove the next metrics from the buffer
	require.ErrorIs(t, model.Write(), internal.ErrSizeLimitReached)
	require.Len(t, plugin.metrics, 8)
	require.Equal(t, 2, model.buffer.Len())

	// The last write should succeed straight away and all metrics should have
	// been received by the output
	require.NoError(t, model.Write())
	testutil.RequireMetricsEqual(t, append(first5, next5...), plugin.metrics)
	require.Zero(t, model.buffer.Len())
}

func TestRunningOutputWritePartialSuccessAndLoss(t *testing.T) {
	lost := 0
	plugin := &mockOutput{
		batchAcceptSize:  4,
		metricFatalIndex: &lost,
	}
	model := NewRunningOutput(plugin, &OutputConfig{}, 5, 10)
	require.NoError(t, model.Init())
	require.NoError(t, model.Connect())
	defer model.Close()

	// Fill buffer completely
	for _, metric := range first5 {
		model.AddMetric(metric)
	}
	for _, metric := range next5 {
		model.AddMetric(metric)
	}
	expected := []telegraf.Metric{
		/* fatal, */ first5[1], first5[2], first5[3],
		/* fatal, */ next5[0], next5[1], next5[2],
		next5[3], next5[4],
	}

	// We no not expect any successful flush yet
	require.Empty(t, plugin.Metrics())
	require.Equal(t, 10, model.buffer.Len())

	// Write to the output. This should only partially succeed with the first
	// few metrics removed from buffer
	require.ErrorIs(t, model.Write(), internal.ErrSizeLimitReached)
	require.Len(t, plugin.metrics, 3)
	require.Equal(t, 6, model.buffer.Len())

	// The next write should remove the next metrics from the buffer
	require.ErrorIs(t, model.Write(), internal.ErrSizeLimitReached)
	require.Len(t, plugin.metrics, 6)
	require.Equal(t, 2, model.buffer.Len())

	// The last write should succeed straight away and all metrics should have
	// been received by the output
	require.NoError(t, model.Write())
	testutil.RequireMetricsEqual(t, expected, plugin.metrics)
	require.Zero(t, model.buffer.Len())
}

func TestRunningOutputWriteBatchPartialSuccess(t *testing.T) {
	plugin := &mockOutput{
		batchAcceptSize: 4,
	}
	model := NewRunningOutput(plugin, &OutputConfig{}, 5, 10)
	require.NoError(t, model.Init())
	require.NoError(t, model.Connect())
	defer model.Close()

	// Fill buffer completely
	for _, metric := range first5 {
		model.AddMetric(metric)
	}
	for _, metric := range next5 {
		model.AddMetric(metric)
	}

	// We no not expect any successful flush yet
	require.Empty(t, plugin.Metrics())
	require.Equal(t, 10, model.buffer.Len())

	// Write to the output. This should only partially succeed with the first
	// few metrics removed from buffer
	require.ErrorIs(t, model.WriteBatch(), internal.ErrSizeLimitReached)
	require.Len(t, plugin.metrics, 4)
	require.Equal(t, 6, model.buffer.Len())

	// The next write should remove the next metrics from the buffer
	require.ErrorIs(t, model.WriteBatch(), internal.ErrSizeLimitReached)
	require.Len(t, plugin.metrics, 8)
	require.Equal(t, 2, model.buffer.Len())

	// The last write should succeed straight away and all metrics should have
	// been received by the output
	require.NoError(t, model.WriteBatch())
	testutil.RequireMetricsEqual(t, append(first5, next5...), plugin.metrics)
	require.Zero(t, model.buffer.Len())
}

func TestRunningOutputWriteBatchPartialSuccessAndLoss(t *testing.T) {
	lost := 0
	plugin := &mockOutput{
		batchAcceptSize:  4,
		metricFatalIndex: &lost,
	}
	model := NewRunningOutput(plugin, &OutputConfig{}, 5, 10)
	require.NoError(t, model.Init())
	require.NoError(t, model.Connect())
	defer model.Close()

	// Fill buffer completely
	for _, metric := range first5 {
		model.AddMetric(metric)
	}
	for _, metric := range next5 {
		model.AddMetric(metric)
	}
	expected := []telegraf.Metric{
		/* fatal, */ first5[1], first5[2], first5[3],
		/* fatal, */ next5[0], next5[1], next5[2],
		next5[3], next5[4],
	}

	// We no not expect any successful flush yet
	require.Empty(t, plugin.Metrics())
	require.Equal(t, 10, model.buffer.Len())

	// Write to the output. This should only partially succeed with the first
	// few metrics removed from buffer
	require.ErrorIs(t, model.WriteBatch(), internal.ErrSizeLimitReached)
	require.Len(t, plugin.metrics, 3)
	require.Equal(t, 6, model.buffer.Len())

	// The next write should remove the next metrics from the buffer
	require.ErrorIs(t, model.WriteBatch(), internal.ErrSizeLimitReached)
	require.Len(t, plugin.metrics, 6)
	require.Equal(t, 2, model.buffer.Len())

	// The last write should succeed straight away and all metrics should have
	// been received by the output
	require.NoError(t, model.WriteBatch())
	testutil.RequireMetricsEqual(t, expected, plugin.metrics)
	require.Zero(t, model.buffer.Len())
}

// Benchmark adding metrics.
func BenchmarkRunningOutputAddWrite(b *testing.B) {
	conf := &OutputConfig{
		Filter: Filter{},
	}
	m := &perfOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	for n := 0; n < b.N; n++ {
		ro.AddMetric(testutil.TestMetric(101, "metric1"))
		ro.Write() //nolint:errcheck // skip checking err for benchmark tests
	}
}

// Benchmark adding metrics.
func BenchmarkRunningOutputAddWriteEvery100(b *testing.B) {
	conf := &OutputConfig{
		Filter: Filter{},
	}
	m := &perfOutput{}
	ro := NewRunningOutput(m, conf, 1000, 10000)

	for n := 0; n < b.N; n++ {
		ro.AddMetric(testutil.TestMetric(101, "metric1"))
		if n%100 == 0 {
			ro.Write() //nolint:errcheck // skip checking err for benchmark tests
		}
	}
}

// Benchmark adding metrics.
func BenchmarkRunningOutputAddFailWrites(b *testing.B) {
	conf := &OutputConfig{
		Filter: Filter{},
	}
	m := &perfOutput{failWrite: true}
	ro := NewRunningOutput(m, conf, 1000, 10000)
	for n := 0; n < b.N; n++ {
		ro.AddMetric(testutil.TestMetric(101, "metric1"))
	}
}

type mockOutput struct {
	sync.Mutex

	metrics []telegraf.Metric

	// Failing output simulation
	batchAcceptSize  int
	metricFatalIndex *int

	// Startup error simulation
	startupError      error
	startupErrorCount int
	writes            atomic.Uint32

	// Utility for getting notified about writes and also to manipulate
	// the write behavior
	preWriteHook  func([]telegraf.Metric) error
	postWriteHook func([]telegraf.Metric) error
}

func (m *mockOutput) Connect() error {
	if m.startupErrorCount == 0 {
		return nil
	}
	if m.startupErrorCount > 0 {
		m.startupErrorCount--
	}
	return m.startupError
}

func (*mockOutput) Close() error {
	return nil
}

func (*mockOutput) SampleConfig() string {
	return ""
}

func (m *mockOutput) Write(metrics []telegraf.Metric) error {
	m.writes.Add(1)

	m.Lock()
	defer m.Unlock()

	// Execute hook if any
	if m.preWriteHook != nil {
		if err := m.preWriteHook(metrics); err != nil {
			return err
		}
	}

	// Simulate a failed write
	if m.batchAcceptSize < 0 {
		return errors.New("failed write")
	}

	// Simulate a successful write
	var resultErr error
	if m.batchAcceptSize == 0 || len(metrics) <= m.batchAcceptSize {
		m.metrics = append(m.metrics, metrics...)
	} else {
		// Simulate a partially successful write
		werr := &internal.PartialWriteError{Err: internal.ErrSizeLimitReached}
		for i, x := range metrics {
			if m.metricFatalIndex != nil && i == *m.metricFatalIndex {
				werr.MetricsReject = append(werr.MetricsReject, i)
			} else if i < m.batchAcceptSize {
				m.metrics = append(m.metrics, x)
				werr.MetricsAccept = append(werr.MetricsAccept, i)
			}
		}
		resultErr = werr
	}

	// Execute hook if any
	if m.postWriteHook != nil {
		if err := m.postWriteHook(metrics); err != nil {
			return err
		}
	}

	return resultErr
}

func (m *mockOutput) Metrics() []telegraf.Metric {
	m.Lock()
	defer m.Unlock()
	return m.metrics
}

type perfOutput struct {
	// if true, mock write failure
	failWrite bool
}

func (*perfOutput) Connect() error {
	return nil
}

func (*perfOutput) Close() error {
	return nil
}

func (*perfOutput) SampleConfig() string {
	return ""
}

func (m *perfOutput) Write([]telegraf.Metric) error {
	if m.failWrite {
		return errors.New("failed write")
	}
	return nil
}

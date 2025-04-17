package shim

import (
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestOutputShim(t *testing.T) {
	o := &testOutput{}

	stdinReader, stdinWriter := io.Pipe()

	s := New()
	s.stdin = stdinReader
	require.NoError(t, s.AddOutput(o))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := s.RunOutput(); err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())

	m := metric.New("thing",
		map[string]string{
			"a": "b",
		},
		map[string]interface{}{
			"v": 1,
		},
		time.Now(),
	)
	b, err := serializer.Serialize(m)
	require.NoError(t, err)
	_, err = stdinWriter.Write(b)
	require.NoError(t, err)
	require.NoError(t, stdinWriter.Close())

	wg.Wait()

	require.Len(t, o.MetricsWritten, 1)
	testutil.RequireMetricEqual(t, m, o.MetricsWritten[0])
}

func TestOutputShimWithBatchSize(t *testing.T) {
	o := &testOutput{}

	stdinReader, stdinWriter := io.Pipe()

	// Setup a shim with a batch size but no timeout
	s := New()
	s.stdin = stdinReader
	s.BatchSize = 5
	s.BatchTimeout = 0
	require.NoError(t, s.AddOutput(o))

	// Start the output processing
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := s.RunOutput(); err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	// Serialize the test metric
	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())
	m := metric.New("thing",
		map[string]string{
			"a": "b",
		},
		map[string]interface{}{
			"v": 1,
		},
		time.Now(),
	)
	payload, err := serializer.Serialize(m)
	require.NoError(t, err)

	// Write a few more metrics than the batch-size and check that we only get
	// a full batch before closing the input stream.
	expected := make([]telegraf.Metric, 0, s.BatchSize+3)
	for range cap(expected) {
		_, err := stdinWriter.Write(payload)
		require.NoError(t, err)
		expected = append(expected, m)
	}

	// Wait for the metrics to arrive
	require.Eventually(t, func() bool {
		return o.Count.Load() >= uint32(s.BatchSize)
	}, 3*time.Second, 100*time.Millisecond)
	testutil.RequireMetricsEqual(t, expected[:s.BatchSize], o.MetricsWritten)

	// Closing the input should force the remaining metrics to be written
	require.NoError(t, stdinWriter.Close())
	wg.Wait()
	testutil.RequireMetricsEqual(t, expected, o.MetricsWritten)
}

func TestOutputShimWithFlushTimeout(t *testing.T) {
	o := &testOutput{}

	stdinReader, stdinWriter := io.Pipe()

	// Setup a shim with a batch size and a short timeout
	s := New()
	s.stdin = stdinReader
	s.BatchSize = 5
	s.BatchTimeout = 500 * time.Millisecond
	require.NoError(t, s.AddOutput(o))

	// Start the output processing
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := s.RunOutput(); err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	// Serialize the test metric
	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())
	m := metric.New("thing",
		map[string]string{
			"a": "b",
		},
		map[string]interface{}{
			"v": 1,
		},
		time.Now(),
	)
	payload, err := serializer.Serialize(m)
	require.NoError(t, err)

	// Write less metrics than the batch-size and check if the flush timeout
	// triggers..
	expected := make([]telegraf.Metric, 0, s.BatchSize-1)
	for range cap(expected) {
		_, err := stdinWriter.Write(payload)
		require.NoError(t, err)
		expected = append(expected, m)
	}
	// Wait for the batch to be flushed
	require.Eventually(t, func() bool {
		return o.Count.Load() >= uint32(len(expected))
	}, 3*time.Second, 100*time.Millisecond)

	testutil.RequireMetricsEqual(t, expected, o.MetricsWritten)

	// Closing the input should not change anything
	require.NoError(t, stdinWriter.Close())
	wg.Wait()
	testutil.RequireMetricsEqual(t, expected, o.MetricsWritten)
}

type testOutput struct {
	MetricsWritten []telegraf.Metric
	Count          atomic.Uint32
}

func (*testOutput) Connect() error {
	return nil
}
func (*testOutput) Close() error {
	return nil
}
func (o *testOutput) Write(metrics []telegraf.Metric) error {
	o.MetricsWritten = append(o.MetricsWritten, metrics...)
	o.Count.Store(uint32(len(o.MetricsWritten)))
	return nil
}

func (*testOutput) SampleConfig() string {
	return ""
}

package shim

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestOutputShim(t *testing.T) {
	o := &testOutput{}

	stdinReader, stdinWriter := io.Pipe()

	s := New()
	s.stdin = stdinReader
	err := s.AddOutput(o)
	require.NoError(t, err)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		err := s.RunOutput()
		require.NoError(t, err)
		wg.Done()
	}()

	serializer, _ := serializers.NewInfluxSerializer()

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
	err = stdinWriter.Close()
	require.NoError(t, err)

	wg.Wait()

	require.Len(t, o.MetricsWritten, 1)
	mOut := o.MetricsWritten[0]

	testutil.RequireMetricEqual(t, m, mOut)
}

type testOutput struct {
	MetricsWritten []telegraf.Metric
}

func (o *testOutput) Connect() error {
	return nil
}
func (o *testOutput) Close() error {
	return nil
}
func (o *testOutput) Write(metrics []telegraf.Metric) error {
	o.MetricsWritten = append(o.MetricsWritten, metrics...)
	return nil
}

func (o *testOutput) SampleConfig() string {
	return ""
}

func (o *testOutput) Description() string {
	return ""
}

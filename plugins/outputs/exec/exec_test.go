package exec

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	parsers_influx "github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

var now = time.Date(2020, 6, 30, 16, 16, 0, 0, time.UTC)

type MockRunner struct {
	runs []int
}

// Run runs the command.
func (c *MockRunner) Run(_ time.Duration, _, _ []string, buffer io.Reader) error {
	parser := parsers_influx.NewStreamParser(buffer)
	numMetrics := 0

	for {
		_, err := parser.Next()
		if err != nil {
			if errors.Is(err, parsers_influx.EOF) {
				break // stream ended
			}
			continue
		}
		numMetrics++
	}

	c.runs = append(c.runs, numMetrics)
	return nil
}

func TestExternalOutputBatch(t *testing.T) {
	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())

	runner := MockRunner{}

	e := &Exec{
		UseBatchFormat: true,
		serializer:     serializer,
		Log:            testutil.Logger{},
		runner:         &runner,
	}

	m := metric.New(
		"cpu",
		map[string]string{"name": "cpu1"},
		map[string]interface{}{"idle": 50, "sys": 30},
		now,
	)

	require.NoError(t, e.Connect())
	require.NoError(t, e.Write([]telegraf.Metric{m, m}))
	// Make sure it executed the command once, with 2 metrics
	require.Equal(t, []int{2}, runner.runs)
	require.NoError(t, e.Close())
}

func TestExternalOutputNoBatch(t *testing.T) {
	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())
	runner := MockRunner{}

	e := &Exec{
		UseBatchFormat: false,
		serializer:     serializer,
		Log:            testutil.Logger{},
		runner:         &runner,
	}

	m := metric.New(
		"cpu",
		map[string]string{"name": "cpu1"},
		map[string]interface{}{"idle": 50, "sys": 30},
		now,
	)

	require.NoError(t, e.Connect())
	require.NoError(t, e.Write([]telegraf.Metric{m, m}))
	// Make sure it executed the command twice, both with a single metric
	require.Equal(t, []int{1, 1}, runner.runs)
	require.NoError(t, e.Close())
}

func TestExec(t *testing.T) {
	t.Skip("Skipping test due to OS/executable dependencies and race condition when ran as part of a test-all")

	tests := []struct {
		name    string
		command []string
		err     bool
		metrics []telegraf.Metric
	}{
		{
			name:    "test success",
			command: []string{"tee"},
			err:     false,
			metrics: testutil.MockMetrics(),
		},
		{
			name:    "test doesn't accept stdin",
			command: []string{"sleep", "5s"},
			err:     true,
			metrics: testutil.MockMetrics(),
		},
		{
			name:    "test command not found",
			command: []string{"/no/exist", "-h"},
			err:     true,
			metrics: testutil.MockMetrics(),
		},
		{
			name:    "test no metrics output",
			command: []string{"tee"},
			err:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Exec{
				Command: tt.command,
				Timeout: config.Duration(time.Second),
				runner:  &CommandRunner{},
			}

			s := &influx.Serializer{}
			require.NoError(t, s.Init())
			e.SetSerializer(s)

			require.NoError(t, e.Connect())
			require.Equal(t, tt.err, e.Write(tt.metrics) != nil)
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name string
		buf  *bytes.Buffer
		len  int
	}{
		{
			name: "long out",
			buf:  bytes.NewBufferString(strings.Repeat("a", maxStderrBytes+100)),
			len:  maxStderrBytes + len("..."),
		},
		{
			name: "multiline out",
			buf:  bytes.NewBufferString("hola\ngato\n"),
			len:  len("hola") + len("..."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := truncate(*tt.buf)
			require.Len(t, s, tt.len)
		})
	}
}

func TestExecDocs(t *testing.T) {
	e := &Exec{}
	e.SampleConfig()
	require.NoError(t, e.Close())

	e = &Exec{runner: &CommandRunner{}}
	require.NoError(t, e.Close())
}

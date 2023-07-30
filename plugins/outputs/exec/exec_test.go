package exec

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/influxdata/telegraf/metric"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	influxParser "github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

var now = time.Date(2020, 6, 30, 16, 16, 0, 0, time.UTC)

func TestExternalOutputBatch(t *testing.T) {
	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())

	exe, err := os.Executable()
	require.NoError(t, err)

	e := &Exec{
		Command:        []string{exe, "-testoutput"},
		Environment:    []string{"PLUGINS_OUTPUTS_EXEC_MODE=application", "METRIC_NAME=cpu"},
		Timeout:        3000000000,
		UseBatchFormat: true,
		serializer:     serializer,
		Log:            testutil.Logger{},
	}

	require.NoError(t, e.Init())

	m := metric.New(
		"cpu",
		map[string]string{"name": "cpu1"},
		map[string]interface{}{"idle": 50, "sys": 30},
		now,
	)

	require.NoError(t, e.Connect())
	require.NoError(t, e.Write([]telegraf.Metric{m, m}))
	// Make sure it executed the command once, with 2 metrics
	require.Equal(t, e.outBuffer.String(), "2\n")
	require.NoError(t, e.Close())
}

func TestExternalOutputNoBatch(t *testing.T) {
	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())

	exe, err := os.Executable()
	require.NoError(t, err)

	e := &Exec{
		Command:        []string{exe, "-testoutput"},
		Environment:    []string{"PLUGINS_OUTPUTS_EXEC_MODE=application", "METRIC_NAME=cpu"},
		Timeout:        3000000000,
		UseBatchFormat: false,
		serializer:     serializer,
		Log:            testutil.Logger{},
	}

	require.NoError(t, e.Init())

	m := metric.New(
		"cpu",
		map[string]string{"name": "cpu1"},
		map[string]interface{}{"idle": 50, "sys": 30},
		now,
	)

	require.NoError(t, e.Connect())
	require.NoError(t, e.Write([]telegraf.Metric{m, m}))
	// Make sure it executed the command twice, both with a single metric
	require.Equal(t, e.outBuffer.String(), "1\n1\n")
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
			metrics: []telegraf.Metric{},
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
	c := CommandRunner{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := c.truncate(*tt.buf)
			require.Equal(t, tt.len, len(s))
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

var testoutput = flag.Bool("testoutput", false,
	"if true, act like line input program instead of test")

func TestMain(m *testing.M) {
	flag.Parse()
	runMode := os.Getenv("PLUGINS_OUTPUTS_EXEC_MODE")
	if *testoutput && runMode == "application" {
		runOutputConsumerProgram()
		os.Exit(0)
	}
	code := m.Run()
	os.Exit(code)
}

func runOutputConsumerProgram() {
	metricName := os.Getenv("METRIC_NAME")
	parser := influxParser.NewStreamParser(os.Stdin)
	numMetrics := 0

	for {
		m, err := parser.Next()
		if err != nil {
			if errors.Is(err, influxParser.EOF) {
				break // stream ended
			}
			var parseErr *influxParser.ParseError
			if errors.As(err, &parseErr) {
				fmt.Fprintf(os.Stderr, "parse ERR %v\n", parseErr)
				//nolint:revive // error code is important for this "test"
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "ERR %v\n", err)
			//nolint:revive // error code is important for this "test"
			os.Exit(1)
		}
		numMetrics++

		expected := testutil.MustMetric(metricName,
			map[string]string{"name": "cpu1"},
			map[string]interface{}{"idle": 50, "sys": 30},
			now,
		)

		if !testutil.MetricEqual(expected, m) {
			fmt.Fprintf(os.Stderr, "metric doesn't match expected\n")
			//nolint:revive // error code is important for this "test"
			os.Exit(1)
		}
	}
	fmt.Fprintf(os.Stdout, "%d\n", numMetrics)
	os.Exit(0)
}

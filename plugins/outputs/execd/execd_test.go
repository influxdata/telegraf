package execd

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

var now = time.Date(2020, 6, 30, 16, 16, 0, 0, time.UTC)

func TestExternalOutputWorks(t *testing.T) {
	influxSerializer, err := serializers.NewInfluxSerializer()
	require.NoError(t, err)

	exe, err := os.Executable()
	require.NoError(t, err)

	e := &Execd{
		Command:      []string{exe, "-testoutput"},
		RestartDelay: config.Duration(5 * time.Second),
		serializer:   influxSerializer,
		Log:          testutil.Logger{},
	}

	require.NoError(t, e.Init())

	wg := &sync.WaitGroup{}
	wg.Add(1)
	e.process.ReadStderrFn = func(rstderr io.Reader) {
		scanner := bufio.NewScanner(rstderr)

		for scanner.Scan() {
			t.Errorf("stderr: %q", scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			if !strings.HasSuffix(err.Error(), "already closed") {
				t.Errorf("error reading stderr: %v", err)
			}
		}
		wg.Done()
	}

	m := metric.New(
		"cpu",
		map[string]string{"name": "cpu1"},
		map[string]interface{}{"idle": 50, "sys": 30},
		now,
	)

	require.NoError(t, e.Connect())
	require.NoError(t, e.Write([]telegraf.Metric{m}))
	require.NoError(t, e.Close())
	wg.Wait()
}

var testoutput = flag.Bool("testoutput", false,
	"if true, act like line input program instead of test")

func TestMain(m *testing.M) {
	flag.Parse()
	if *testoutput {
		runOutputConsumerProgram()
		os.Exit(0)
	}
	code := m.Run()
	os.Exit(code)
}

func runOutputConsumerProgram() {
	parser := influx.NewStreamParser(os.Stdin)

	for {
		m, err := parser.Next()
		if err != nil {
			if err == influx.EOF {
				return // stream ended
			}
			if parseErr, isParseError := err.(*influx.ParseError); isParseError {
				//nolint:errcheck,revive // Test will fail anyway
				fmt.Fprintf(os.Stderr, "parse ERR %v\n", parseErr)
				//nolint:revive // error code is important for this "test"
				os.Exit(1)
			}
			//nolint:errcheck,revive // Test will fail anyway
			fmt.Fprintf(os.Stderr, "ERR %v\n", err)
			//nolint:revive // error code is important for this "test"
			os.Exit(1)
		}

		expected := testutil.MustMetric("cpu",
			map[string]string{"name": "cpu1"},
			map[string]interface{}{"idle": 50, "sys": 30},
			now,
		)

		if !testutil.MetricEqual(expected, m) {
			//nolint:errcheck,revive // Test will fail anyway
			fmt.Fprintf(os.Stderr, "metric doesn't match expected\n")
			//nolint:revive // error code is important for this "test"
			os.Exit(1)
		}
	}
}

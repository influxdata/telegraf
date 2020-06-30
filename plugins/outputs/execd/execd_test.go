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

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
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

	m, err := metric.New(
		"cpu",
		map[string]string{"name": "cpu1"},
		map[string]interface{}{"idle": 50, "sys": 30},
		now,
	)
	require.NoError(t, err)

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
		metric, err := parser.Next()
		if err != nil {
			if err == influx.EOF {
				return // stream ended
			}
			if parseErr, isParseError := err.(*influx.ParseError); isParseError {
				fmt.Fprintf(os.Stderr, "parse ERR %v\n", parseErr)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "ERR %v\n", err)
			os.Exit(1)
		}

		expected := testutil.MustMetric("cpu",
			map[string]string{"name": "cpu1"},
			map[string]interface{}{"idle": 50, "sys": 30},
			now,
		)

		if !testutil.MetricEqual(expected, metric) {
			fmt.Fprintf(os.Stderr, "metric doesn't match expected\n")
			os.Exit(1)
		}
	}
}

package execd

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

func TestExternalProcessorWorks(t *testing.T) {
	e := New()
	e.Log = testutil.Logger{}

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	e.SetParser(parser)

	exe, err := os.Executable()
	require.NoError(t, err)
	t.Log(exe)
	e.Command = []string{exe, "-countmultiplier"}
	e.Environment = []string{"PLUGINS_PROCESSORS_EXECD_MODE=application", "FIELD_NAME=count"}
	e.RestartDelay = config.Duration(5 * time.Second)

	acc := &testutil.Accumulator{}

	require.NoError(t, e.Start(acc))

	now := time.Now()
	orig := now
	for i := 0; i < 10; i++ {
		m := metric.New("test",
			map[string]string{
				"city": "Toronto",
			},
			map[string]interface{}{
				"population": 6000000,
				"count":      1,
			},
			now)
		now = now.Add(1)

		require.NoError(t, e.Add(m, acc))
	}

	acc.Wait(1)
	e.Stop()
	acc.Wait(9)

	metrics := acc.GetTelegrafMetrics()
	m := metrics[0]

	expected := testutil.MustMetric("test",
		map[string]string{
			"city": "Toronto",
		},
		map[string]interface{}{
			"population": 6000000,
			"count":      2,
		},
		orig,
	)
	testutil.RequireMetricEqual(t, expected, m)

	metricTime := m.Time().UnixNano()

	// make sure the other 9 are ordered properly
	for i := 0; i < 9; i++ {
		m = metrics[i+1]
		require.EqualValues(t, metricTime+1, m.Time().UnixNano())
		metricTime = m.Time().UnixNano()
	}
}

func TestParseLinesWithNewLines(t *testing.T) {
	e := New()
	e.Log = testutil.Logger{}

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	e.SetParser(parser)

	exe, err := os.Executable()
	require.NoError(t, err)
	t.Log(exe)
	e.Command = []string{exe, "-countmultiplier"}
	e.Environment = []string{"PLUGINS_PROCESSORS_EXECD_MODE=application", "FIELD_NAME=count"}
	e.RestartDelay = config.Duration(5 * time.Second)

	acc := &testutil.Accumulator{}

	require.NoError(t, e.Start(acc))

	now := time.Now()
	orig := now

	m := metric.New("test",
		map[string]string{
			"author": "Mr. Gopher",
		},
		map[string]interface{}{
			"phrase": "Gophers are amazing creatures.\nAbsolutely amazing.",
			"count":  3,
		},
		now)

	require.NoError(t, e.Add(m, acc))

	acc.Wait(1)
	e.Stop()

	processedMetric := acc.GetTelegrafMetrics()[0]

	expectedMetric := testutil.MustMetric("test",
		map[string]string{
			"author": "Mr. Gopher",
		},
		map[string]interface{}{
			"phrase": "Gophers are amazing creatures.\nAbsolutely amazing.",
			"count":  6,
		},
		orig,
	)

	testutil.RequireMetricEqual(t, expectedMetric, processedMetric)
}

var countmultiplier = flag.Bool("countmultiplier", false,
	"if true, act like line input program instead of test")

func TestMain(m *testing.M) {
	flag.Parse()
	runMode := os.Getenv("PLUGINS_PROCESSORS_EXECD_MODE")
	if *countmultiplier && runMode == "application" {
		runCountMultiplierProgram()
		os.Exit(0)
	}
	code := m.Run()
	os.Exit(code)
}

func runCountMultiplierProgram() {
	fieldName := os.Getenv("FIELD_NAME")
	parser := influx.NewStreamParser(os.Stdin)
	serializer := serializers.NewInfluxSerializer()

	for {
		m, err := parser.Next()
		if err != nil {
			if err == influx.EOF {
				return // stream ended
			}
			if parseErr, isParseError := err.(*influx.ParseError); isParseError {
				fmt.Fprintf(os.Stderr, "parse ERR %v\n", parseErr)
				//nolint:revive // os.Exit called intentionally
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "ERR %v\n", err)
			//nolint:revive // os.Exit called intentionally
			os.Exit(1)
		}

		c, found := m.GetField(fieldName)
		if !found {
			fmt.Fprintf(os.Stderr, "metric has no %s field\n", fieldName)
			//nolint:revive // os.Exit called intentionally
			os.Exit(1)
		}
		switch t := c.(type) {
		case float64:
			t *= 2
			m.AddField(fieldName, t)
		case int64:
			t *= 2
			m.AddField(fieldName, t)
		default:
			fmt.Fprintf(os.Stderr, "%s is not an unknown type, it's a %T\n", fieldName, c)
			//nolint:revive // os.Exit called intentionally
			os.Exit(1)
		}
		b, err := serializer.Serialize(m)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR %v\n", err)
			//nolint:revive // os.Exit called intentionally
			os.Exit(1)
		}
		fmt.Fprint(os.Stdout, string(b))
	}
}

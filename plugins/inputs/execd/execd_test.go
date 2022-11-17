package execd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/parsers/prometheus"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

func TestSettingConfigWorks(t *testing.T) {
	cfg := `
	[[inputs.execd]]
		command = ["a", "b", "c"]
		environment = ["d=e", "f=1"]
		restart_delay = "1m"
		signal = "SIGHUP"
	`
	conf := config.NewConfig()
	require.NoError(t, conf.LoadConfigData([]byte(cfg)))

	require.Len(t, conf.Inputs, 1)
	inp, ok := conf.Inputs[0].Input.(*Execd)
	require.True(t, ok)
	require.EqualValues(t, []string{"a", "b", "c"}, inp.Command)
	require.EqualValues(t, []string{"d=e", "f=1"}, inp.Environment)
	require.EqualValues(t, 1*time.Minute, inp.RestartDelay)
	require.EqualValues(t, "SIGHUP", inp.Signal)
}

func TestExternalInputWorks(t *testing.T) {
	influxParser := models.NewRunningParser(&influx.Parser{}, &models.ParserConfig{})
	require.NoError(t, influxParser.Init())

	exe, err := os.Executable()
	require.NoError(t, err)

	e := &Execd{
		Command:      []string{exe, "-counter"},
		Environment:  []string{"PLUGINS_INPUTS_EXECD_MODE=application", "METRIC_NAME=counter"},
		RestartDelay: config.Duration(5 * time.Second),
		Signal:       "STDIN",
		Log:          testutil.Logger{},
	}
	e.SetParser(influxParser)

	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)
	acc := agent.NewAccumulator(&TestMetricMaker{}, metrics)

	require.NoError(t, e.Start(acc))
	require.NoError(t, e.Gather(acc))

	// grab a metric and make sure it's a thing
	m := readChanWithTimeout(t, metrics, 10*time.Second)

	e.Stop()

	require.Equal(t, "counter", m.Name())
	val, ok := m.GetField("count")
	require.True(t, ok)
	require.EqualValues(t, 0, val)
}

func TestParsesLinesContainingNewline(t *testing.T) {
	parser := models.NewRunningParser(&influx.Parser{}, &models.ParserConfig{})
	require.NoError(t, parser.Init())

	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)
	acc := agent.NewAccumulator(&TestMetricMaker{}, metrics)

	e := &Execd{
		RestartDelay: config.Duration(5 * time.Second),
		Signal:       "STDIN",
		acc:          acc,
		Log:          testutil.Logger{},
	}
	e.SetParser(parser)

	cases := []struct {
		Name  string
		Value string
	}{
		{
			Name:  "no-newline",
			Value: "my message",
		}, {
			Name:  "newline",
			Value: "my\nmessage",
		},
	}

	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			line := fmt.Sprintf("event message=\"%v\" 1587128639239000000", test.Value)

			e.outputReader(strings.NewReader(line))

			m := readChanWithTimeout(t, metrics, 1*time.Second)

			require.Equal(t, "event", m.Name())
			val, ok := m.GetField("message")
			require.True(t, ok)
			require.Equal(t, test.Value, val)
		})
	}
}

func TestParsesPrometheus(t *testing.T) {
	parser := models.NewRunningParser(&prometheus.Parser{}, &models.ParserConfig{})
	require.NoError(t, parser.Init())

	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)

	var acc testutil.Accumulator

	e := &Execd{
		RestartDelay: config.Duration(5 * time.Second),
		Signal:       "STDIN",
		acc:          &acc,
		Log:          testutil.Logger{},
	}
	e.SetParser(parser)

	lines := `# HELP This is just a test metric.
# TYPE test summary
test{handler="execd",quantile="0.5"} 42.0
`
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus",
			map[string]string{"handler": "execd", "quantile": "0.5"},
			map[string]interface{}{"test": float64(42.0)},
			time.Unix(0, 0),
		),
	}

	e.outputReader(strings.NewReader(lines))
	check := func() bool { return acc.NMetrics() == uint64(len(expected)) }
	require.Eventually(t, check, 1*time.Second, 100*time.Millisecond)
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
}

func readChanWithTimeout(t *testing.T, metrics chan telegraf.Metric, timeout time.Duration) telegraf.Metric {
	to := time.NewTimer(timeout)
	defer to.Stop()
	select {
	case m := <-metrics:
		return m
	case <-to.C:
		require.FailNow(t, "timeout waiting for metric")
	}
	return nil
}

type TestMetricMaker struct{}

func (tm *TestMetricMaker) Name() string {
	return "TestPlugin"
}

func (tm *TestMetricMaker) LogName() string {
	return tm.Name()
}

func (tm *TestMetricMaker) MakeMetric(aMetric telegraf.Metric) telegraf.Metric {
	return aMetric
}

func (tm *TestMetricMaker) Log() telegraf.Logger {
	return models.NewLogger("TestPlugin", "test", "")
}

var counter = flag.Bool("counter", false,
	"if true, act like line input program instead of test")

func TestMain(m *testing.M) {
	flag.Parse()
	runMode := os.Getenv("PLUGINS_INPUTS_EXECD_MODE")
	if *counter && runMode == "application" {
		if err := runCounterProgram(); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	code := m.Run()
	os.Exit(code)
}

func runCounterProgram() error {
	envMetricName := os.Getenv("METRIC_NAME")
	i := 0
	serializer := serializers.NewInfluxSerializer()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		m := metric.New(envMetricName,
			map[string]string{},
			map[string]interface{}{
				"count": i,
			},
			time.Now(),
		)
		i++

		b, err := serializer.Serialize(m)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR %v\n", err)
			return err
		}
		if _, err := fmt.Fprint(os.Stdout, string(b)); err != nil {
			return err
		}
	}
	return nil
}

package collectd

import (
	"context"
	"testing"
	"time"

	"collectd.org/api"
	"collectd.org/network"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

type AuthMap struct {
	Passwd map[string]string
}

func (p *AuthMap) Password(user string) (string, error) {
	return p.Passwd[user], nil
}

type metricData struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

type testCase struct {
	vl       []api.ValueList
	expected []metricData
}

var singleMetric = testCase{
	[]api.ValueList{
		{
			Identifier: api.Identifier{
				Host:           "xyzzy",
				Plugin:         "cpu",
				PluginInstance: "1",
				Type:           "cpu",
				TypeInstance:   "user",
			},
			Values: []api.Value{
				api.Counter(42),
			},
			DSNames: []string(nil),
		},
	},
	[]metricData{
		{
			"cpu_value",
			map[string]string{
				"type_instance": "user",
				"host":          "xyzzy",
				"instance":      "1",
				"type":          "cpu",
			},
			map[string]interface{}{
				"value": float64(42),
			},
		},
	},
}

var multiMetric = testCase{
	[]api.ValueList{
		{
			Identifier: api.Identifier{
				Host:           "xyzzy",
				Plugin:         "cpu",
				PluginInstance: "0",
				Type:           "cpu",
				TypeInstance:   "user",
			},
			Values: []api.Value{
				api.Derive(42),
				api.Gauge(42),
			},
			DSNames: []string{"t1", "t2"},
		},
	},
	[]metricData{
		{
			"cpu_0",
			map[string]string{
				"type_instance": "user",
				"host":          "xyzzy",
				"instance":      "0",
				"type":          "cpu",
			},
			map[string]interface{}{
				"value": float64(42),
			},
		},
		{
			"cpu_1",
			map[string]string{
				"type_instance": "user",
				"host":          "xyzzy",
				"instance":      "0",
				"type":          "cpu",
			},
			map[string]interface{}{
				"value": float64(42),
			},
		},
	},
}

func TestNewCollectdParser(t *testing.T) {
	parser := Parser{
		ParseMultiValue: "join",
	}
	require.NoError(t, parser.Init())
	require.Equal(t, network.None, parser.popts.SecurityLevel)
	require.NotNil(t, parser.popts.PasswordLookup)
	require.Nil(t, parser.popts.TypesDB)
}

func TestParse(t *testing.T) {
	cases := []testCase{singleMetric, multiMetric}

	for _, tc := range cases {
		buf, err := writeValueList(t.Context(), tc.vl)
		require.NoError(t, err)
		bytes, err := buf.Bytes()
		require.NoError(t, err)

		parser := &Parser{}
		require.NoError(t, parser.Init())
		metrics, err := parser.Parse(bytes)
		require.NoError(t, err)

		assertEqualMetrics(t, tc.expected, metrics)
	}
}

func TestParseMultiValueSplit(t *testing.T) {
	buf, err := writeValueList(t.Context(), multiMetric.vl)
	require.NoError(t, err)
	bytes, err := buf.Bytes()
	require.NoError(t, err)

	parser := &Parser{ParseMultiValue: "split"}
	require.NoError(t, parser.Init())
	metrics, err := parser.Parse(bytes)
	require.NoError(t, err)

	require.Len(t, metrics, 2)
}

func TestParseMultiValueJoin(t *testing.T) {
	buf, err := writeValueList(t.Context(), multiMetric.vl)
	require.NoError(t, err)
	bytes, err := buf.Bytes()
	require.NoError(t, err)

	parser := &Parser{ParseMultiValue: "join"}
	require.NoError(t, parser.Init())
	metrics, err := parser.Parse(bytes)
	require.NoError(t, err)

	require.Len(t, metrics, 1)
}

func TestParse_DefaultTags(t *testing.T) {
	buf, err := writeValueList(t.Context(), singleMetric.vl)
	require.NoError(t, err)
	bytes, err := buf.Bytes()
	require.NoError(t, err)

	parser := &Parser{}
	require.NoError(t, parser.Init())
	parser.SetDefaultTags(map[string]string{
		"foo": "bar",
	})
	require.NoError(t, err)
	metrics, err := parser.Parse(bytes)
	require.NoError(t, err)

	require.Equal(t, "bar", metrics[0].Tags()["foo"])
}

func TestParse_SignSecurityLevel(t *testing.T) {
	parser := &Parser{
		SecurityLevel: "sign",
		AuthFile:      "testdata/authfile",
	}
	require.NoError(t, parser.Init())

	// Signed data
	buf, err := writeValueList(t.Context(), singleMetric.vl)
	require.NoError(t, err)
	buf.Sign("user0", "bar")
	bytes, err := buf.Bytes()
	require.NoError(t, err)

	metrics, err := parser.Parse(bytes)
	require.NoError(t, err)
	assertEqualMetrics(t, singleMetric.expected, metrics)

	// Encrypted data
	buf, err = writeValueList(t.Context(), singleMetric.vl)
	require.NoError(t, err)
	buf.Encrypt("user0", "bar")
	bytes, err = buf.Bytes()
	require.NoError(t, err)

	metrics, err = parser.Parse(bytes)
	require.NoError(t, err)
	assertEqualMetrics(t, singleMetric.expected, metrics)

	// Plain text data skipped
	buf, err = writeValueList(t.Context(), singleMetric.vl)
	require.NoError(t, err)
	bytes, err = buf.Bytes()
	require.NoError(t, err)

	metrics, err = parser.Parse(bytes)
	require.NoError(t, err)
	require.Empty(t, metrics)

	// Wrong password error
	buf, err = writeValueList(t.Context(), singleMetric.vl)
	require.NoError(t, err)
	buf.Sign("x", "y")
	bytes, err = buf.Bytes()
	require.NoError(t, err)

	_, err = parser.Parse(bytes)
	require.Error(t, err)
}

func TestParse_EncryptSecurityLevel(t *testing.T) {
	parser := &Parser{
		SecurityLevel: "encrypt",
		AuthFile:      "testdata/authfile",
	}
	require.NoError(t, parser.Init())

	// Signed data skipped
	buf, err := writeValueList(t.Context(), singleMetric.vl)
	require.NoError(t, err)
	buf.Sign("user0", "bar")
	bytes, err := buf.Bytes()
	require.NoError(t, err)

	metrics, err := parser.Parse(bytes)
	require.NoError(t, err)
	require.Empty(t, metrics)

	// Encrypted data
	buf, err = writeValueList(t.Context(), singleMetric.vl)
	require.NoError(t, err)
	buf.Encrypt("user0", "bar")
	bytes, err = buf.Bytes()
	require.NoError(t, err)

	metrics, err = parser.Parse(bytes)
	require.NoError(t, err)
	assertEqualMetrics(t, singleMetric.expected, metrics)

	// Plain text data skipped
	buf, err = writeValueList(t.Context(), singleMetric.vl)
	require.NoError(t, err)
	bytes, err = buf.Bytes()
	require.NoError(t, err)

	metrics, err = parser.Parse(bytes)
	require.NoError(t, err)
	require.Empty(t, metrics)

	// Wrong password error
	buf, err = writeValueList(t.Context(), singleMetric.vl)
	require.NoError(t, err)
	buf.Sign("x", "y")
	bytes, err = buf.Bytes()
	require.NoError(t, err)

	_, err = parser.Parse(bytes)
	require.Error(t, err)
}

func TestParseLine(t *testing.T) {
	buf, err := writeValueList(t.Context(), singleMetric.vl)
	require.NoError(t, err)
	bytes, err := buf.Bytes()
	require.NoError(t, err)
	parser := Parser{
		ParseMultiValue: "split",
	}
	require.NoError(t, parser.Init())
	m, err := parser.ParseLine(string(bytes))
	require.NoError(t, err)

	assertEqualMetrics(t, singleMetric.expected, []telegraf.Metric{m})
}

func writeValueList(testContext context.Context, valueLists []api.ValueList) (*network.Buffer, error) {
	buffer := network.NewBuffer(0)

	for i := range valueLists {
		err := buffer.Write(testContext, &valueLists[i])
		if err != nil {
			return nil, err
		}
	}

	return buffer, nil
}

func assertEqualMetrics(t *testing.T, expected []metricData, received []telegraf.Metric) {
	require.Equal(t, len(expected), len(received))
	for i, m := range received {
		require.Equal(t, expected[i].name, m.Name())
		require.Equal(t, expected[i].tags, m.Tags())
		require.Equal(t, expected[i].fields, m.Fields())
	}
}

var benchmarkData = []api.ValueList{
	{
		Identifier: api.Identifier{
			Host:           "xyzzy",
			Plugin:         "cpu",
			PluginInstance: "1",
			Type:           "cpu",
			TypeInstance:   "user",
		},
		Values: []api.Value{
			api.Counter(4),
		},
		DSNames: []string(nil),
	},
	{
		Identifier: api.Identifier{
			Host:           "xyzzy",
			Plugin:         "cpu",
			PluginInstance: "2",
			Type:           "cpu",
			TypeInstance:   "user",
		},
		Values: []api.Value{
			api.Counter(5),
		},
		DSNames: []string(nil),
	},
}

func TestBenchmarkData(t *testing.T) {
	expected := []telegraf.Metric{
		metric.New(
			"cpu_value",
			map[string]string{
				"host":          "xyzzy",
				"instance":      "1",
				"type":          "cpu",
				"type_instance": "user",
			},
			map[string]interface{}{
				"value": 4.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"cpu_value",
			map[string]string{
				"host":          "xyzzy",
				"instance":      "2",
				"type":          "cpu",
				"type_instance": "user",
			},
			map[string]interface{}{
				"value": 5.0,
			},
			time.Unix(0, 0),
		),
	}

	buf, err := writeValueList(t.Context(), benchmarkData)
	require.NoError(t, err)
	bytes, err := buf.Bytes()
	require.NoError(t, err)

	parser := &Parser{}
	require.NoError(t, parser.Init())
	actual, err := parser.Parse(bytes)
	require.NoError(t, err)

	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func BenchmarkParsing(b *testing.B) {
	buf, err := writeValueList(b.Context(), benchmarkData)
	require.NoError(b, err)
	bytes, err := buf.Bytes()
	require.NoError(b, err)

	parser := &Parser{}
	require.NoError(b, parser.Init())

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		parser.Parse(bytes)
	}
}

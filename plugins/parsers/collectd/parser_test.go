package collectd

import (
	"context"
	"testing"

	"collectd.org/api"
	"collectd.org/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
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
	parser, err := NewCollectdParser("", "", []string{}, "join")
	require.Nil(t, err)
	require.Equal(t, parser.popts.SecurityLevel, network.None)
	require.NotNil(t, parser.popts.PasswordLookup)
	require.Nil(t, parser.popts.TypesDB)
}

func TestParse(t *testing.T) {
	cases := []testCase{singleMetric, multiMetric}

	for _, tc := range cases {
		buf, err := writeValueList(tc.vl)
		require.Nil(t, err)
		bytes, err := buf.Bytes()
		require.Nil(t, err)

		parser := &CollectdParser{}
		require.Nil(t, err)
		metrics, err := parser.Parse(bytes)
		require.Nil(t, err)

		assertEqualMetrics(t, tc.expected, metrics)
	}
}

func TestParseMultiValueSplit(t *testing.T) {
	buf, err := writeValueList(multiMetric.vl)
	require.Nil(t, err)
	bytes, err := buf.Bytes()
	require.Nil(t, err)

	parser := &CollectdParser{ParseMultiValue: "split"}
	metrics, err := parser.Parse(bytes)
	require.Nil(t, err)

	assert.Equal(t, 2, len(metrics))
}

func TestParse_DefaultTags(t *testing.T) {
	buf, err := writeValueList(singleMetric.vl)
	require.Nil(t, err)
	bytes, err := buf.Bytes()
	require.Nil(t, err)

	parser := &CollectdParser{}
	parser.SetDefaultTags(map[string]string{
		"foo": "bar",
	})
	require.Nil(t, err)
	metrics, err := parser.Parse(bytes)
	require.Nil(t, err)

	require.Equal(t, "bar", metrics[0].Tags()["foo"])
}

func TestParse_SignSecurityLevel(t *testing.T) {
	parser := &CollectdParser{}
	popts := &network.ParseOpts{
		SecurityLevel: network.Sign,
		PasswordLookup: &AuthMap{
			map[string]string{
				"user0": "bar",
			},
		},
	}
	parser.SetParseOpts(popts)

	// Signed data
	buf, err := writeValueList(singleMetric.vl)
	require.Nil(t, err)
	buf.Sign("user0", "bar")
	bytes, err := buf.Bytes()
	require.Nil(t, err)

	metrics, err := parser.Parse(bytes)
	require.Nil(t, err)
	assertEqualMetrics(t, singleMetric.expected, metrics)

	// Encrypted data
	buf, err = writeValueList(singleMetric.vl)
	require.Nil(t, err)
	buf.Encrypt("user0", "bar")
	bytes, err = buf.Bytes()
	require.Nil(t, err)

	metrics, err = parser.Parse(bytes)
	require.Nil(t, err)
	assertEqualMetrics(t, singleMetric.expected, metrics)

	// Plain text data skipped
	buf, err = writeValueList(singleMetric.vl)
	require.Nil(t, err)
	bytes, err = buf.Bytes()
	require.Nil(t, err)

	metrics, err = parser.Parse(bytes)
	require.Nil(t, err)
	require.Equal(t, []telegraf.Metric{}, metrics)

	// Wrong password error
	buf, err = writeValueList(singleMetric.vl)
	require.Nil(t, err)
	buf.Sign("x", "y")
	bytes, err = buf.Bytes()
	require.Nil(t, err)

	metrics, err = parser.Parse(bytes)
	require.NotNil(t, err)
}

func TestParse_EncryptSecurityLevel(t *testing.T) {
	parser := &CollectdParser{}
	popts := &network.ParseOpts{
		SecurityLevel: network.Encrypt,
		PasswordLookup: &AuthMap{
			map[string]string{
				"user0": "bar",
			},
		},
	}
	parser.SetParseOpts(popts)

	// Signed data skipped
	buf, err := writeValueList(singleMetric.vl)
	require.Nil(t, err)
	buf.Sign("user0", "bar")
	bytes, err := buf.Bytes()
	require.Nil(t, err)

	metrics, err := parser.Parse(bytes)
	require.Nil(t, err)
	require.Equal(t, []telegraf.Metric{}, metrics)

	// Encrypted data
	buf, err = writeValueList(singleMetric.vl)
	require.Nil(t, err)
	buf.Encrypt("user0", "bar")
	bytes, err = buf.Bytes()
	require.Nil(t, err)

	metrics, err = parser.Parse(bytes)
	require.Nil(t, err)
	assertEqualMetrics(t, singleMetric.expected, metrics)

	// Plain text data skipped
	buf, err = writeValueList(singleMetric.vl)
	require.Nil(t, err)
	bytes, err = buf.Bytes()
	require.Nil(t, err)

	metrics, err = parser.Parse(bytes)
	require.Nil(t, err)
	require.Equal(t, []telegraf.Metric{}, metrics)

	// Wrong password error
	buf, err = writeValueList(singleMetric.vl)
	require.Nil(t, err)
	buf.Sign("x", "y")
	bytes, err = buf.Bytes()
	require.Nil(t, err)

	metrics, err = parser.Parse(bytes)
	require.NotNil(t, err)
}

func TestParseLine(t *testing.T) {
	buf, err := writeValueList(singleMetric.vl)
	require.Nil(t, err)
	bytes, err := buf.Bytes()
	require.Nil(t, err)

	parser, err := NewCollectdParser("", "", []string{}, "split")
	require.Nil(t, err)
	metric, err := parser.ParseLine(string(bytes))
	require.Nil(t, err)

	assertEqualMetrics(t, singleMetric.expected, []telegraf.Metric{metric})
}

func writeValueList(valueLists []api.ValueList) (*network.Buffer, error) {
	buffer := network.NewBuffer(0)

	ctx := context.Background()
	for _, vl := range valueLists {
		err := buffer.Write(ctx, &vl)
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

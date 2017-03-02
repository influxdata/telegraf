package collectd

import (
	"context"
	"testing"

	"collectd.org/api"
	"collectd.org/network"
	"github.com/stretchr/testify/require"
)

type metricData struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

type testData struct {
	vl       []api.ValueList
	expected []metricData
}

var parseData = testData{
	[]api.ValueList{
		api.ValueList{
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
			DSNames: []string(nil),
		},
		api.ValueList{
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
		metricData{
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
		metricData{
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
		metricData{
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

func TestParse(t *testing.T) {
	require := require.New(t)

	td := parseData

	bytes, err := serializeValueList(td.vl)
	require.Nil(err)

	parser := CollectdParser{}
	metrics, err := parser.Parse(bytes)
	require.Nil(err)

	require.Equal(len(td.expected), len(metrics))

	for i, m := range metrics {
		require.Equal(td.expected[i].name, m.Name())
		require.Equal(td.expected[i].tags, m.Tags())
		require.Equal(td.expected[i].fields, m.Fields())
	}
}

func TestParseLine_MultipleMetrics(t *testing.T) {
	require := require.New(t)

	bytes, err := serializeValueList(parseData.vl)
	require.Nil(err)

	parser := CollectdParser{}
	_, err = parser.ParseLine(string(bytes))

	require.NotNil(err)
}

func TestParseLine(t *testing.T) {
	require := require.New(t)

	vl := []api.ValueList{
		api.ValueList{
			Identifier: api.Identifier{
				Host:           "xyzzy",
				Plugin:         "cpu",
				PluginInstance: "0",
				Type:           "cpu",
				TypeInstance:   "user",
			},
			Values: []api.Value{
				api.Derive(42),
			},
			DSNames: []string(nil),
		},
	}

	expected := metricData{
		"cpu_value",
		map[string]string{
			"type_instance": "user",
			"host":          "xyzzy",
			"instance":      "0",
			"type":          "cpu",
		},
		map[string]interface{}{
			"value": float64(42),
		},
	}

	bytes, err := serializeValueList(vl)
	require.Nil(err)

	parser := CollectdParser{}
	metric, err := parser.ParseLine(string(bytes))
	require.Nil(err)

	require.Equal(expected.name, metric.Name())
	require.Equal(expected.tags, metric.Tags())
	require.Equal(expected.fields, metric.Fields())
}

func serializeValueList(valueLists []api.ValueList) ([]byte, error) {
	buffer := network.NewBuffer(0)
	ctx := context.Background()

	for _, vl := range valueLists {
		err := buffer.Write(ctx, &vl)
		if err != nil {
			return nil, err
		}
	}

	bytes, err := buffer.Bytes()
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

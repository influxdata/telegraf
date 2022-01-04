package wavefront

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func TestParse(t *testing.T) {
	parser := NewWavefrontParser(nil)

	parsedMetrics, err := parser.Parse([]byte("test.metric 1"))
	require.NoError(t, err)
	testMetric := metric.New("test.metric", map[string]string{}, map[string]interface{}{"value": 1.}, time.Unix(0, 0))
	require.Equal(t, parsedMetrics[0].Name(), testMetric.Name())
	require.Equal(t, parsedMetrics[0].Fields(), testMetric.Fields())

	parsedMetrics, err = parser.Parse([]byte("\u2206test.delta 1 1530939936"))
	require.NoError(t, err)
	testMetric = metric.New("\u2206test.delta", map[string]string{},
		map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\u0394test.delta 1 1530939936"))
	require.NoError(t, err)
	testMetric = metric.New("\u0394test.delta", map[string]string{},
		map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\u0394test.delta 1.234 1530939936 source=\"mysource\" tag2=value2"))
	require.NoError(t, err)
	testMetric = metric.New("\u0394test.delta", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.234}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("test.metric 1 1530939936"))
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("test.metric 1 1530939936 source=mysource"))
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" 1.1234 1530939936 source=\"mysource\""))
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" 1.1234 1530939936 \"source\"=\"mysource\" tag2=value2"))
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" -1.1234 1530939936 \"source\"=\"mysource\" tag2=value2"))
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": -1.1234}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" 1.1234e04 1530939936 \"source\"=\"mysource\" tag2=value2"))
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234e04}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" 1.1234e-04 1530939936 \"source\"=\"mysource\" tag2=value2"))
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234e-04}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("test.metric		 1.1234      1530939936 	source=\"mysource\"    tag2=value2     "))
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)
}

func TestParseLine(t *testing.T) {
	parser := NewWavefrontParser(nil)

	parsedMetric, err := parser.ParseLine("test.metric 1")
	require.NoError(t, err)
	testMetric := metric.New("test.metric", map[string]string{}, map[string]interface{}{"value": 1.}, time.Unix(0, 0))
	require.Equal(t, parsedMetric.Name(), testMetric.Name())
	require.Equal(t, parsedMetric.Fields(), testMetric.Fields())

	parsedMetric, err = parser.ParseLine("test.metric 1 1530939936")
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetric, testMetric)

	parsedMetric, err = parser.ParseLine("test.metric 1 1530939936 source=mysource")
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetric, testMetric)

	parsedMetric, err = parser.ParseLine("\"test.metric\" 1.1234 1530939936 source=\"mysource\"")
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetric, testMetric)

	parsedMetric, err = parser.ParseLine("\"test.metric\" 1.1234 1530939936 \"source\"=\"mysource\" tag2=value2")
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetric, testMetric)

	parsedMetric, err = parser.ParseLine("test.metric		 1.1234      1530939936 	source=\"mysource\"    tag2=value2     ")
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetric, testMetric)
}

func TestParseMultiple(t *testing.T) {
	parser := NewWavefrontParser(nil)

	parsedMetrics, err := parser.Parse([]byte("test.metric 1\ntest.metric2 2 1530939936"))
	require.NoError(t, err)
	testMetric1 := metric.New("test.metric", map[string]string{}, map[string]interface{}{"value": 1.}, time.Unix(0, 0))
	testMetric2 := metric.New("test.metric2", map[string]string{}, map[string]interface{}{"value": 2.}, time.Unix(1530939936, 0))
	testMetrics := []telegraf.Metric{testMetric1, testMetric2}
	require.Equal(t, parsedMetrics[0].Name(), testMetrics[0].Name())
	require.Equal(t, parsedMetrics[0].Fields(), testMetrics[0].Fields())
	require.EqualValues(t, parsedMetrics[1], testMetrics[1])

	parsedMetrics, err = parser.Parse([]byte("test.metric 1 1530939936 source=mysource\n\"test.metric\" 1.1234 1530939936 source=\"mysource\""))
	require.NoError(t, err)
	testMetric1 = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	testMetric2 = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	testMetrics = []telegraf.Metric{testMetric1, testMetric2}
	require.EqualValues(t, parsedMetrics, testMetrics)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" 1.1234 1530939936 \"source\"=\"mysource\" tag2=value2\ntest.metric		 1.1234      1530939936 	source=\"mysource\"    tag2=value2     "))
	require.NoError(t, err)
	testMetric1 = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	testMetric2 = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	testMetrics = []telegraf.Metric{testMetric1, testMetric2}
	require.EqualValues(t, parsedMetrics, testMetrics)

	parsedMetrics, err = parser.Parse([]byte("test.metric 1 1530939936 source=mysource\n\"test.metric\" 1.1234 1530939936 source=\"mysource\"\ntest.metric3 333 1530939936 tagit=valueit"))
	require.NoError(t, err)
	testMetric1 = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	testMetric2 = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	testMetric3 := metric.New("test.metric3", map[string]string{"tagit": "valueit"}, map[string]interface{}{"value": 333.}, time.Unix(1530939936, 0))
	testMetrics = []telegraf.Metric{testMetric1, testMetric2, testMetric3}
	require.EqualValues(t, parsedMetrics, testMetrics)
}

func TestParseSpecial(t *testing.T) {
	parser := NewWavefrontParser(nil)

	parsedMetric, err := parser.ParseLine("\"test.metric\" 1 1530939936")
	require.NoError(t, err)
	testMetric := metric.New("test.metric", map[string]string{}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetric, testMetric)

	parsedMetric, err = parser.ParseLine("test.metric 1 1530939936 tag1=\"val\\\"ue1\"")
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"tag1": "val\\\"ue1"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetric, testMetric)
}

func TestParseInvalid(t *testing.T) {
	parser := NewWavefrontParser(nil)

	_, err := parser.Parse([]byte("test.metric"))
	require.Error(t, err)

	_, err = parser.Parse([]byte("test.metric string"))
	require.Error(t, err)

	_, err = parser.Parse([]byte("test.metric 1 string"))
	require.Error(t, err)

	_, err = parser.Parse([]byte("test.\u2206delta 1"))
	require.Error(t, err)

	_, err = parser.Parse([]byte("test.metric 1 1530939936 tag_no_pair"))
	require.Error(t, err)

	_, err = parser.Parse([]byte("test.metric 1 1530939936 tag_broken_value=\""))
	require.Error(t, err)

	_, err = parser.Parse([]byte("\"test.metric 1 1530939936"))
	require.Error(t, err)

	_, err = parser.Parse([]byte("test.metric 1 1530939936 tag1=val\\\"ue1"))
	require.Error(t, err)

	_, err = parser.Parse([]byte("\"test.metric\" -1.12-34 1530939936 \"source\"=\"mysource\" tag2=value2"))
	require.Error(t, err)
}

func TestParseDefaultTags(t *testing.T) {
	parser := NewWavefrontParser(map[string]string{"myDefault": "value1", "another": "test2"})

	parsedMetrics, err := parser.Parse([]byte("test.metric 1 1530939936"))
	require.NoError(t, err)
	testMetric := metric.New("test.metric", map[string]string{"myDefault": "value1", "another": "test2"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("test.metric 1 1530939936 source=mysource"))
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"myDefault": "value1", "another": "test2", "source": "mysource"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" 1.1234 1530939936 another=\"test3\""))
	require.NoError(t, err)
	testMetric = metric.New("test.metric", map[string]string{"myDefault": "value1", "another": "test2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	require.EqualValues(t, parsedMetrics[0], testMetric)
}

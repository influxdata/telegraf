package wavefront

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	parser := NewWavefrontParser(nil)

	parsedMetrics, err := parser.Parse([]byte("test.metric 1"))
	assert.NoError(t, err)
	testMetric, err := metric.New("test.metric", map[string]string{}, map[string]interface{}{"value": 1.}, time.Unix(0, 0))
	assert.NoError(t, err)
	assert.Equal(t, parsedMetrics[0].Name(), testMetric.Name())
	assert.Equal(t, parsedMetrics[0].Fields(), testMetric.Fields())

	parsedMetrics, err = parser.Parse([]byte("\u2206test.delta 1 1530939936"))
	assert.NoError(t, err)
	testMetric, err = metric.New("\u2206test.delta", map[string]string{},
		map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\u0394test.delta 1 1530939936"))
	assert.NoError(t, err)
	testMetric, err = metric.New("\u0394test.delta", map[string]string{},
		map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\u0394test.delta 1.234 1530939936 source=\"mysource\" tag2=value2"))
	assert.NoError(t, err)
	testMetric, err = metric.New("\u0394test.delta", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.234}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("test.metric 1 1530939936"))
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("test.metric 1 1530939936 source=mysource"))
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" 1.1234 1530939936 source=\"mysource\""))
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" 1.1234 1530939936 \"source\"=\"mysource\" tag2=value2"))
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" -1.1234 1530939936 \"source\"=\"mysource\" tag2=value2"))
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": -1.1234}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" 1.1234e04 1530939936 \"source\"=\"mysource\" tag2=value2"))
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234e04}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" 1.1234e-04 1530939936 \"source\"=\"mysource\" tag2=value2"))
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234e-04}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("test.metric		 1.1234      1530939936 	source=\"mysource\"    tag2=value2     "))
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

}

func TestParseLine(t *testing.T) {
	parser := NewWavefrontParser(nil)

	parsedMetric, err := parser.ParseLine("test.metric 1")
	assert.NoError(t, err)
	testMetric, err := metric.New("test.metric", map[string]string{}, map[string]interface{}{"value": 1.}, time.Unix(0, 0))
	assert.NoError(t, err)
	assert.Equal(t, parsedMetric.Name(), testMetric.Name())
	assert.Equal(t, parsedMetric.Fields(), testMetric.Fields())

	parsedMetric, err = parser.ParseLine("test.metric 1 1530939936")
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetric, testMetric)

	parsedMetric, err = parser.ParseLine("test.metric 1 1530939936 source=mysource")
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetric, testMetric)

	parsedMetric, err = parser.ParseLine("\"test.metric\" 1.1234 1530939936 source=\"mysource\"")
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetric, testMetric)

	parsedMetric, err = parser.ParseLine("\"test.metric\" 1.1234 1530939936 \"source\"=\"mysource\" tag2=value2")
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetric, testMetric)

	parsedMetric, err = parser.ParseLine("test.metric		 1.1234      1530939936 	source=\"mysource\"    tag2=value2     ")
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetric, testMetric)
}

func TestParseMultiple(t *testing.T) {
	parser := NewWavefrontParser(nil)

	parsedMetrics, err := parser.Parse([]byte("test.metric 1\ntest.metric2 2 1530939936"))
	assert.NoError(t, err)
	testMetric1, err := metric.New("test.metric", map[string]string{}, map[string]interface{}{"value": 1.}, time.Unix(0, 0))
	assert.NoError(t, err)
	testMetric2, err := metric.New("test.metric2", map[string]string{}, map[string]interface{}{"value": 2.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	testMetrics := []telegraf.Metric{testMetric1, testMetric2}
	assert.Equal(t, parsedMetrics[0].Name(), testMetrics[0].Name())
	assert.Equal(t, parsedMetrics[0].Fields(), testMetrics[0].Fields())
	assert.EqualValues(t, parsedMetrics[1], testMetrics[1])

	parsedMetrics, err = parser.Parse([]byte("test.metric 1 1530939936 source=mysource\n\"test.metric\" 1.1234 1530939936 source=\"mysource\""))
	assert.NoError(t, err)
	testMetric1, err = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	testMetric2, err = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	testMetrics = []telegraf.Metric{testMetric1, testMetric2}
	assert.EqualValues(t, parsedMetrics, testMetrics)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" 1.1234 1530939936 \"source\"=\"mysource\" tag2=value2\ntest.metric		 1.1234      1530939936 	source=\"mysource\"    tag2=value2     "))
	assert.NoError(t, err)
	testMetric1, err = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	testMetric2, err = metric.New("test.metric", map[string]string{"source": "mysource", "tag2": "value2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	testMetrics = []telegraf.Metric{testMetric1, testMetric2}
	assert.EqualValues(t, parsedMetrics, testMetrics)

	parsedMetrics, err = parser.Parse([]byte("test.metric 1 1530939936 source=mysource\n\"test.metric\" 1.1234 1530939936 source=\"mysource\"\ntest.metric3 333 1530939936 tagit=valueit"))
	assert.NoError(t, err)
	testMetric1, err = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	testMetric2, err = metric.New("test.metric", map[string]string{"source": "mysource"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	testMetric3, err := metric.New("test.metric3", map[string]string{"tagit": "valueit"}, map[string]interface{}{"value": 333.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	testMetrics = []telegraf.Metric{testMetric1, testMetric2, testMetric3}
	assert.EqualValues(t, parsedMetrics, testMetrics)

}

func TestParseSpecial(t *testing.T) {
	parser := NewWavefrontParser(nil)

	parsedMetric, err := parser.ParseLine("\"test.metric\" 1 1530939936")
	assert.NoError(t, err)
	testMetric, err := metric.New("test.metric", map[string]string{}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetric, testMetric)

	parsedMetric, err = parser.ParseLine("test.metric 1 1530939936 tag1=\"val\\\"ue1\"")
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"tag1": "val\\\"ue1"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetric, testMetric)

}

func TestParseInvalid(t *testing.T) {
	parser := NewWavefrontParser(nil)

	_, err := parser.Parse([]byte("test.metric"))
	assert.Error(t, err)

	_, err = parser.Parse([]byte("test.metric string"))
	assert.Error(t, err)

	_, err = parser.Parse([]byte("test.metric 1 string"))
	assert.Error(t, err)

	_, err = parser.Parse([]byte("test.\u2206delta 1"))
	assert.Error(t, err)

	_, err = parser.Parse([]byte("test.metric 1 1530939936 tag_no_pair"))
	assert.Error(t, err)

	_, err = parser.Parse([]byte("test.metric 1 1530939936 tag_broken_value=\""))
	assert.Error(t, err)

	_, err = parser.Parse([]byte("\"test.metric 1 1530939936"))
	assert.Error(t, err)

	_, err = parser.Parse([]byte("test.metric 1 1530939936 tag1=val\\\"ue1"))
	assert.Error(t, err)

	_, err = parser.Parse([]byte("\"test.metric\" -1.12-34 1530939936 \"source\"=\"mysource\" tag2=value2"))
	assert.Error(t, err)

}

func TestParseDefaultTags(t *testing.T) {
	parser := NewWavefrontParser(map[string]string{"myDefault": "value1", "another": "test2"})

	parsedMetrics, err := parser.Parse([]byte("test.metric 1 1530939936"))
	assert.NoError(t, err)
	testMetric, err := metric.New("test.metric", map[string]string{"myDefault": "value1", "another": "test2"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("test.metric 1 1530939936 source=mysource"))
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"myDefault": "value1", "another": "test2", "source": "mysource"}, map[string]interface{}{"value": 1.}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

	parsedMetrics, err = parser.Parse([]byte("\"test.metric\" 1.1234 1530939936 another=\"test3\""))
	assert.NoError(t, err)
	testMetric, err = metric.New("test.metric", map[string]string{"myDefault": "value1", "another": "test2"}, map[string]interface{}{"value": 1.1234}, time.Unix(1530939936, 0))
	assert.NoError(t, err)
	assert.EqualValues(t, parsedMetrics[0], testMetric)

}

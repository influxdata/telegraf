package csv

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
)

var DefaultTime = func() time.Time {
	return time.Unix(3600, 0)
}

func TestBasicCSV(t *testing.T) {
	p := &Parser{
		ColumnNames: []string{"first", "second", "third"},
		TagColumns:  []string{"third"},
		TimeFunc:    DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)

	_, err = p.ParseLine("1.4,true,hi")
	require.NoError(t, err)
}

func TestHeaderConcatenationCSV(t *testing.T) {
	p := &Parser{
		HeaderRowCount:    2,
		MeasurementColumn: "3",
		TimeFunc:          DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)
	testCSV := `first,second
1,2,3
3.4,70,test_name`

	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, "test_name", metrics[0].Name())
}

func TestHeaderOverride(t *testing.T) {
	p := &Parser{
		HeaderRowCount:    1,
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
		TimeFunc:          DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)
	testCSV := `line1,line2,line3
3.4,70,test_name`
	expectedFields := map[string]interface{}{
		"first":  3.4,
		"second": int64(70),
	}
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, "test_name", metrics[0].Name())
	require.Equal(t, expectedFields, metrics[0].Fields())

	testCSVRows := []string{"line1,line2,line3\r\n", "3.4,70,test_name\r\n"}

	p = &Parser{
		HeaderRowCount:    1,
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
		TimeFunc:          DefaultTime,
	}
	err = p.Init()
	require.NoError(t, err)
	metrics, err = p.Parse([]byte(testCSVRows[0]))
	require.NoError(t, err)
	require.Empty(t, metrics)
	m, err := p.ParseLine(testCSVRows[1])
	require.NoError(t, err)
	require.Equal(t, "test_name", m.Name())
	require.Equal(t, expectedFields, m.Fields())
}

func TestTimestamp(t *testing.T) {
	p := &Parser{
		HeaderRowCount:    1,
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
		TimestampColumn:   "first",
		TimestampFormat:   "02/01/06 03:04:05 PM",
		TimeFunc:          DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `line1,line2,line3
23/05/09 04:05:06 PM,70,test_name
07/11/09 04:05:06 PM,80,test_name2`
	metrics, err := p.Parse([]byte(testCSV))

	require.NoError(t, err)
	require.Equal(t, int64(1243094706000000000), metrics[0].Time().UnixNano())
	require.Equal(t, int64(1257609906000000000), metrics[1].Time().UnixNano())
}

func TestTimestampYYYYMMDDHHmm(t *testing.T) {
	p := &Parser{
		HeaderRowCount:    1,
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
		TimestampColumn:   "first",
		TimestampFormat:   "200601021504",
		TimeFunc:          DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `line1,line2,line3
200905231605,70,test_name
200907111605,80,test_name2`
	metrics, err := p.Parse([]byte(testCSV))

	require.NoError(t, err)
	require.Equal(t, int64(1243094700000000000), metrics[0].Time().UnixNano())
	require.Equal(t, int64(1247328300000000000), metrics[1].Time().UnixNano())
}
func TestTimestampError(t *testing.T) {
	p := &Parser{
		HeaderRowCount:    1,
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
		TimestampColumn:   "first",
		TimeFunc:          DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)
	testCSV := `line1,line2,line3
23/05/09 04:05:06 PM,70,test_name
07/11/09 04:05:06 PM,80,test_name2`
	_, err = p.Parse([]byte(testCSV))
	require.Equal(t, errors.New("timestamp format must be specified"), err)
}

func TestTimestampUnixFormat(t *testing.T) {
	p := &Parser{
		HeaderRowCount:    1,
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
		TimestampColumn:   "first",
		TimestampFormat:   "unix",
		TimeFunc:          DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)
	testCSV := `line1,line2,line3
1243094706,70,test_name
1257609906,80,test_name2`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, int64(1243094706000000000), metrics[0].Time().UnixNano())
	require.Equal(t, int64(1257609906000000000), metrics[1].Time().UnixNano())
}

func TestTimestampUnixMSFormat(t *testing.T) {
	p := &Parser{
		HeaderRowCount:    1,
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
		TimestampColumn:   "first",
		TimestampFormat:   "unix_ms",
		TimeFunc:          DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)
	testCSV := `line1,line2,line3
1243094706123,70,test_name
1257609906123,80,test_name2`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, int64(1243094706123000000), metrics[0].Time().UnixNano())
	require.Equal(t, int64(1257609906123000000), metrics[1].Time().UnixNano())
}

func TestQuotedCharacter(t *testing.T) {
	p := &Parser{
		HeaderRowCount:    1,
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
		TimeFunc:          DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `line1,line2,line3
"3,4",70,test_name`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, "3,4", metrics[0].Fields()["first"])
}

func TestDelimiter(t *testing.T) {
	p := &Parser{
		HeaderRowCount:    1,
		Delimiter:         "%",
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
		TimeFunc:          DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `line1%line2%line3
3,4%70%test_name`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, "3,4", metrics[0].Fields()["first"])
}

func TestNullDelimiter(t *testing.T) {
	p := &Parser{
		HeaderRowCount: 0,
		Delimiter:      "\u0000",
		ColumnNames:    []string{"first", "second", "third"},
		TimeFunc:       DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := strings.Join([]string{"3.4", "70", "test_name"}, "\u0000")
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.InDelta(t, float64(3.4), metrics[0].Fields()["first"], testutil.DefaultDelta)
	require.Equal(t, int64(70), metrics[0].Fields()["second"])
	require.Equal(t, "test_name", metrics[0].Fields()["third"])
}

func TestValueConversion(t *testing.T) {
	p := &Parser{
		HeaderRowCount: 0,
		Delimiter:      ",",
		ColumnNames:    []string{"first", "second", "third", "fourth"},
		MetricName:     "test_value",
		TimeFunc:       DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)
	testCSV := `3.3,4,true,hello`

	expectedTags := make(map[string]string)
	expectedFields := map[string]interface{}{
		"first":  3.3,
		"second": 4,
		"third":  true,
		"fourth": "hello",
	}

	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)

	expectedMetric := metric.New("test_value", expectedTags, expectedFields, time.Unix(0, 0))
	returnedMetric := metric.New(metrics[0].Name(), metrics[0].Tags(), metrics[0].Fields(), time.Unix(0, 0))

	// deep equal fields
	require.Equal(t, expectedMetric.Fields(), returnedMetric.Fields())

	// Test explicit type conversion.
	p.ColumnTypes = []string{"float", "int", "bool", "string"}

	metrics, err = p.Parse([]byte(testCSV))
	require.NoError(t, err)

	returnedMetric = metric.New(metrics[0].Name(), metrics[0].Tags(), metrics[0].Fields(), time.Unix(0, 0))

	// deep equal fields
	require.Equal(t, expectedMetric.Fields(), returnedMetric.Fields())
}

func TestSkipComment(t *testing.T) {
	p := &Parser{
		HeaderRowCount: 0,
		Comment:        "#",
		ColumnNames:    []string{"first", "second", "third", "fourth"},
		MetricName:     "test_value",
		TimeFunc:       DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)
	testCSV := `#3.3,4,true,hello
4,9.9,true,name_this`

	expectedFields := map[string]interface{}{
		"first":  int64(4),
		"second": 9.9,
		"third":  true,
		"fourth": "name_this",
	}

	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, expectedFields, metrics[0].Fields())
}

func TestTrimSpace(t *testing.T) {
	p := &Parser{
		HeaderRowCount: 0,
		TrimSpace:      true,
		ColumnNames:    []string{"first", "second", "third", "fourth"},
		MetricName:     "test_value",
		TimeFunc:       DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)
	testCSV := ` 3.3, 4,    true,hello`

	expectedFields := map[string]interface{}{
		"first":  3.3,
		"second": int64(4),
		"third":  true,
		"fourth": "hello",
	}

	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, expectedFields, metrics[0].Fields())

	p = &Parser{
		HeaderRowCount: 2,
		TrimSpace:      true,
		TimeFunc:       DefaultTime,
	}
	err = p.Init()
	require.NoError(t, err)
	testCSV = "   col  ,  col  ,col\n" +
		"  1  ,  2  ,3\n" +
		"  test  space  ,  80  ,test_name"

	metrics, err = p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{"col1": "test  space", "col2": int64(80), "col3": "test_name"}, metrics[0].Fields())
}

func TestTrimSpaceDelimitedBySpace(t *testing.T) {
	p := &Parser{
		Delimiter:      " ",
		HeaderRowCount: 1,
		TrimSpace:      true,
		TimeFunc:       DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `   first   second   third   fourth
abcdefgh        0       2    false
  abcdef      3.3       4     true
       f        0       2    false`

	expectedFields := map[string]interface{}{
		"first":  "abcdef",
		"second": 3.3,
		"third":  int64(4),
		"fourth": true,
	}

	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, expectedFields, metrics[1].Fields())
}

func TestSkipRows(t *testing.T) {
	p := &Parser{
		HeaderRowCount:    1,
		SkipRows:          1,
		TagColumns:        []string{"line1"},
		MeasurementColumn: "line3",
		TimeFunc:          DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `garbage nonsense
line1,line2,line3
hello,80,test_name2`

	expectedFields := map[string]interface{}{
		"line2": int64(80),
	}
	expectedTags := map[string]string{
		"line1": "hello",
	}
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, "test_name2", metrics[0].Name())
	require.Equal(t, expectedFields, metrics[0].Fields())
	require.Equal(t, expectedTags, metrics[0].Tags())

	p = &Parser{
		HeaderRowCount:    1,
		SkipRows:          1,
		TagColumns:        []string{"line1"},
		MeasurementColumn: "line3",
		TimeFunc:          DefaultTime,
	}
	err = p.Init()
	require.NoError(t, err)
	testCSVRows := []string{"garbage nonsense\r\n", "line1,line2,line3\r\n", "hello,80,test_name2\r\n"}

	metrics, err = p.Parse([]byte(testCSVRows[0]))
	require.ErrorIs(t, err, parsers.ErrEOF)
	require.Nil(t, metrics)
	m, err := p.ParseLine(testCSVRows[1])
	require.NoError(t, err)
	require.Nil(t, m)
	m, err = p.ParseLine(testCSVRows[2])
	require.NoError(t, err)
	require.Equal(t, "test_name2", m.Name())
	require.Equal(t, expectedFields, m.Fields())
	require.Equal(t, expectedTags, m.Tags())
}

func TestSkipColumns(t *testing.T) {
	p := &Parser{
		SkipColumns: 1,
		ColumnNames: []string{"line1", "line2"},
		TimeFunc:    DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)
	testCSV := `hello,80,test_name`

	expectedFields := map[string]interface{}{
		"line1": int64(80),
		"line2": "test_name",
	}
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, expectedFields, metrics[0].Fields())
}

func TestSkipColumnsWithHeader(t *testing.T) {
	p := &Parser{
		SkipColumns:    1,
		HeaderRowCount: 2,
		TimeFunc:       DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `col,col,col
1,2,3
trash,80,test_name`

	// we should expect an error if we try to get col1
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{"col2": int64(80), "col3": "test_name"}, metrics[0].Fields())
}

func TestMultiHeader(t *testing.T) {
	p := &Parser{
		HeaderRowCount: 2,
		TimeFunc:       DefaultTime,
	}
	require.NoError(t, p.Init())
	testCSV := `col,col
1,2
80,test_name`

	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{"col1": int64(80), "col2": "test_name"}, metrics[0].Fields())

	testCSVRows := []string{"col,col\r\n", "1,2\r\n", "80,test_name\r\n"}

	p = &Parser{
		HeaderRowCount: 2,
		TimeFunc:       DefaultTime,
	}
	err = p.Init()
	require.NoError(t, err)

	metrics, err = p.Parse([]byte(testCSVRows[0]))
	require.ErrorIs(t, err, parsers.ErrEOF)
	require.Nil(t, metrics)
	m, err := p.ParseLine(testCSVRows[1])
	require.NoError(t, err)
	require.Nil(t, m)
	m, err = p.ParseLine(testCSVRows[2])
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{"col1": int64(80), "col2": "test_name"}, m.Fields())
}

func TestParseStream(t *testing.T) {
	p := &Parser{
		MetricName:     "csv",
		HeaderRowCount: 1,
		TimeFunc:       DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)

	csvHeader := "a,b,c"
	csvBody := "1,2,3"

	metrics, err := p.Parse([]byte(csvHeader))
	require.NoError(t, err)
	require.Empty(t, metrics)
	m, err := p.ParseLine(csvBody)
	require.NoError(t, err)
	testutil.RequireMetricEqual(t,
		testutil.MustMetric(
			"csv",
			map[string]string{},
			map[string]interface{}{
				"a": int64(1),
				"b": int64(2),
				"c": int64(3),
			},
			DefaultTime(),
		), m)
}

func TestParseLineMultiMetricErrorMessage(t *testing.T) {
	p := &Parser{
		MetricName:     "csv",
		HeaderRowCount: 1,
		TimeFunc:       DefaultTime,
	}
	require.NoError(t, p.Init())

	csvHeader := "a,b,c"
	csvOneRow := "1,2,3"
	csvTwoRows := "4,5,6\n7,8,9"

	metrics, err := p.Parse([]byte(csvHeader))
	require.NoError(t, err)
	require.Empty(t, metrics)
	m, err := p.ParseLine(csvOneRow)
	require.NoError(t, err)
	testutil.RequireMetricEqual(t,
		testutil.MustMetric(
			"csv",
			map[string]string{},
			map[string]interface{}{
				"a": int64(1),
				"b": int64(2),
				"c": int64(3),
			},
			DefaultTime(),
		), m)
	m, err = p.ParseLine(csvTwoRows)
	require.Errorf(t, err, "expected 1 metric found 2")
	require.Nil(t, m)
	metrics, err = p.Parse([]byte(csvTwoRows))
	require.NoError(t, err)
	require.Len(t, metrics, 2)
}

func TestTimestampUnixFloatPrecision(t *testing.T) {
	p := &Parser{
		MetricName:      "csv",
		ColumnNames:     []string{"time", "value"},
		TimestampColumn: "time",
		TimestampFormat: "unix",
		TimeFunc:        DefaultTime,
	}
	err := p.Init()
	require.NoError(t, err)

	data := `1551129661.95456123352050781250,42`

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"csv",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(1551129661, 954561233),
		),
	}

	metrics, err := p.Parse([]byte(data))
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, metrics)
}

func TestSkipMeasurementColumn(t *testing.T) {
	p := &Parser{
		MetricName:      "csv",
		HeaderRowCount:  1,
		TimestampColumn: "timestamp",
		TimestampFormat: "unix",
		TimeFunc:        DefaultTime,
		TrimSpace:       true,
	}
	err := p.Init()
	require.NoError(t, err)

	data := `id,value,timestamp
		1,5,1551129661.954561233`

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"csv",
			map[string]string{},
			map[string]interface{}{
				"id":    1,
				"value": 5,
			},
			time.Unix(1551129661, 954561233),
		),
	}

	metrics, err := p.Parse([]byte(data))
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, metrics)
}

func TestSkipTimestampColumn(t *testing.T) {
	p := &Parser{
		MetricName:      "csv",
		HeaderRowCount:  1,
		TimestampColumn: "timestamp",
		TimestampFormat: "unix",
		TimeFunc:        DefaultTime,
		TrimSpace:       true,
	}
	err := p.Init()
	require.NoError(t, err)

	data := `id,value,timestamp
		1,5,1551129661.954561233`

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"csv",
			map[string]string{},
			map[string]interface{}{
				"id":    1,
				"value": 5,
			},
			time.Unix(1551129661, 954561233),
		),
	}

	metrics, err := p.Parse([]byte(data))
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, metrics)
}

func TestTimestampTimezone(t *testing.T) {
	p := &Parser{
		HeaderRowCount:    1,
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
		TimestampColumn:   "first",
		TimestampFormat:   "02/01/06 03:04:05 PM",
		TimeFunc:          DefaultTime,
		Timezone:          "Asia/Jakarta",
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `line1,line2,line3
23/05/09 11:05:06 PM,70,test_name
07/11/09 11:05:06 PM,80,test_name2`
	metrics, err := p.Parse([]byte(testCSV))

	require.NoError(t, err)
	require.Equal(t, int64(1243094706000000000), metrics[0].Time().UnixNano())
	require.Equal(t, int64(1257609906000000000), metrics[1].Time().UnixNano())
}

func TestEmptyMeasurementName(t *testing.T) {
	p := &Parser{
		MetricName:        "csv",
		HeaderRowCount:    1,
		ColumnNames:       []string{"", "b"},
		MeasurementColumn: "",
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `,b
1,2`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric("csv",
			map[string]string{},
			map[string]interface{}{
				"b": 2,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime())
}

func TestNumericMeasurementName(t *testing.T) {
	p := &Parser{
		MetricName:        "csv",
		HeaderRowCount:    1,
		ColumnNames:       []string{"a", "b"},
		MeasurementColumn: "a",
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `a,b
1,2`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric("1",
			map[string]string{},
			map[string]interface{}{
				"b": 2,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime())
}

func TestStaticMeasurementName(t *testing.T) {
	p := &Parser{
		MetricName:     "csv",
		HeaderRowCount: 1,
		ColumnNames:    []string{"a", "b"},
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `a,b
1,2`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric("csv",
			map[string]string{},
			map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime())
}

func TestSkipEmptyStringValue(t *testing.T) {
	p := &Parser{
		MetricName:     "csv",
		HeaderRowCount: 1,
		ColumnNames:    []string{"a", "b"},
		SkipValues:     []string{""},
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `a,b
1,""`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric("csv",
			map[string]string{},
			map[string]interface{}{
				"a": 1,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime())
}

func TestSkipSpecifiedStringValue(t *testing.T) {
	p := &Parser{
		MetricName:     "csv",
		HeaderRowCount: 1,
		ColumnNames:    []string{"a", "b"},
		SkipValues:     []string{"MM"},
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `a,b
1,MM`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric("csv",
			map[string]string{},
			map[string]interface{}{
				"a": 1,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime())
}

func TestSkipErrorOnCorruptedCSVLine(t *testing.T) {
	p := &Parser{
		HeaderRowCount:  1,
		TimestampColumn: "date",
		TimestampFormat: "02/01/06 03:04:05 PM",
		TimeFunc:        DefaultTime,
		SkipErrors:      true,
		Log:             testutil.Logger{},
	}
	err := p.Init()
	require.NoError(t, err)

	testCSV := `date,a,b
23/05/09 11:05:06 PM,1,2
corrupted_line
07/11/09 04:06:07 PM,3,4`

	expectedFields0 := map[string]interface{}{
		"a": int64(1),
		"b": int64(2),
	}

	expectedFields1 := map[string]interface{}{
		"a": int64(3),
		"b": int64(4),
	}

	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, expectedFields0, metrics[0].Fields())
	require.Equal(t, expectedFields1, metrics[1].Fields())
}

func TestParseMetadataSeparators(t *testing.T) {
	p := &Parser{
		ColumnNames:  []string{"a", "b"},
		MetadataRows: 0,
	}
	err := p.Init()
	require.NoError(t, err)
	p = &Parser{
		ColumnNames:  []string{"a", "b"},
		MetadataRows: 1,
	}
	err = p.Init()
	require.Error(t, err)
	require.Equal(t, "initializing separators failed: "+
		"csv_metadata_separators required when specifying csv_metadata_rows", err.Error())
	p = &Parser{
		ColumnNames:        []string{"a", "b"},
		MetadataRows:       1,
		MetadataSeparators: []string{",", "=", ",", ":", "=", ":="},
	}
	err = p.Init()
	require.NoError(t, err)
	require.Len(t, p.metadataSeparatorList, 4)
	require.Empty(t, p.MetadataTrimSet)
	require.Equal(t, metadataPattern{":=", ",", "=", ":"}, p.metadataSeparatorList)
	p = &Parser{
		ColumnNames:        []string{"a", "b"},
		MetadataRows:       1,
		MetadataSeparators: []string{",", ":", "=", ":="},
		MetadataTrimSet:    " #'",
	}
	err = p.Init()
	require.NoError(t, err)
	require.Len(t, p.metadataSeparatorList, 4)
	require.Len(t, p.MetadataTrimSet, 3)
	require.Equal(t, metadataPattern{":=", ",", ":", "="}, p.metadataSeparatorList)
}

func TestParseMetadataRow(t *testing.T) {
	p := &Parser{
		ColumnNames:        []string{"a", "b"},
		MetadataRows:       5,
		MetadataSeparators: []string{":=", ",", ":", "="},
	}
	err := p.Init()
	require.NoError(t, err)
	require.Empty(t, p.metadataTags)
	m := p.parseMetadataRow("# this is a not matching string")
	require.Nil(t, m)
	m = p.parseMetadataRow("# key1 : value1 \r\n")
	require.Equal(t, map[string]string{"# key1 ": " value1 "}, m)
	m = p.parseMetadataRow("key2=1234\n")
	require.Equal(t, map[string]string{"key2": "1234"}, m)
	m = p.parseMetadataRow(" file created : 2021-10-08T12:34:18+10:00 \r\n")
	require.Equal(t, map[string]string{" file created ": " 2021-10-08T12:34:18+10:00 "}, m)
	m = p.parseMetadataRow("file created: 2021-10-08T12:34:18\t\r\r\n")
	require.Equal(t, map[string]string{"file created": " 2021-10-08T12:34:18\t"}, m)
	p = &Parser{
		ColumnNames:        []string{"a", "b"},
		MetadataRows:       5,
		MetadataSeparators: []string{":=", ",", ":", "="},
		MetadataTrimSet:    " #'",
	}
	err = p.Init()
	require.NoError(t, err)
	require.Empty(t, p.metadataTags)
	m = p.parseMetadataRow("# this is a not matching string")
	require.Nil(t, m)
	m = p.parseMetadataRow("# key1 : value1 \r\n")
	require.Equal(t, map[string]string{"key1": "value1"}, m)
	m = p.parseMetadataRow("key2=1234\n")
	require.Equal(t, map[string]string{"key2": "1234"}, m)
	m = p.parseMetadataRow(" file created : 2021-10-08T12:34:18+10:00 \r\n")
	require.Equal(t, map[string]string{"file created": "2021-10-08T12:34:18+10:00"}, m)
	m = p.parseMetadataRow("file created: '2021-10-08T12:34:18'\r\n")
	require.Equal(t, map[string]string{"file created": "2021-10-08T12:34:18"}, m)
}

func TestParseCSVFileWithMetadata(t *testing.T) {
	p := &Parser{
		HeaderRowCount:     1,
		SkipRows:           2,
		MetadataRows:       4,
		Comment:            "#",
		TagColumns:         []string{"type"},
		MetadataSeparators: []string{":", "="},
		MetadataTrimSet:    " #",
	}
	err := p.Init()
	require.NoError(t, err)
	testCSV := `garbage nonsense that needs be skipped

# version= 1.0

    invalid meta data that can be ignored.
file created: 2021-10-08T12:34:18+10:00
timestamp,type,name,status
2020-11-23T08:19:27+10:00,Reader,R002,1
#2020-11-04T13:23:04+10:00,Reader,R031,0
2020-11-04T13:29:47+10:00,Coordinator,C001,0`
	expectedFields := []map[string]interface{}{
		{
			"name":      "R002",
			"status":    int64(1),
			"timestamp": "2020-11-23T08:19:27+10:00",
		},
		{
			"name":      "C001",
			"status":    int64(0),
			"timestamp": "2020-11-04T13:29:47+10:00",
		},
	}
	expectedTags := []map[string]string{
		{
			"file created": "2021-10-08T12:34:18+10:00",
			"test":         "tag",
			"type":         "Reader",
			"version":      "1.0",
		},
		{
			"file created": "2021-10-08T12:34:18+10:00",
			"test":         "tag",
			"type":         "Coordinator",
			"version":      "1.0",
		},
	}
	// Set default Tags
	p.SetDefaultTags(map[string]string{"test": "tag"})
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	for i, m := range metrics {
		require.Equal(t, expectedFields[i], m.Fields())
		require.Equal(t, expectedTags[i], m.Tags())
	}

	p = &Parser{
		HeaderRowCount:     1,
		SkipRows:           2,
		MetadataRows:       4,
		Comment:            "#",
		TagColumns:         []string{"type", "version"},
		MetadataSeparators: []string{":", "="},
		MetadataTrimSet:    " #",
	}
	err = p.Init()
	require.NoError(t, err)
	testCSVRows := []string{
		"garbage nonsense that needs be skipped",
		"",
		"# version= 1.0\r\n",
		"",
		"    invalid meta data that can be ignored.\r\n",
		"file created: 2021-10-08T12:34:18+10:00",
		"timestamp,type,name,status\n",
		"2020-11-23T08:19:27+10:00,Reader,R002,1\r\n",
		"#2020-11-04T13:23:04+10:00,Reader,R031,0\n",
		"2020-11-04T13:29:47+10:00,Coordinator,C001,0",
	}

	// Set default Tags
	p.SetDefaultTags(map[string]string{"test": "tag"})
	rowIndex := 0
	for ; rowIndex < 6; rowIndex++ {
		m, err := p.ParseLine(testCSVRows[rowIndex])
		require.ErrorIs(t, err, parsers.ErrEOF)
		require.Nil(t, m)
	}
	m, err := p.ParseLine(testCSVRows[rowIndex])
	require.NoError(t, err)
	require.Nil(t, m)
	rowIndex++
	m, err = p.ParseLine(testCSVRows[rowIndex])
	require.NoError(t, err)
	require.Equal(t, expectedFields[0], m.Fields())
	require.Equal(t, expectedTags[0], m.Tags())
	rowIndex++
	m, err = p.ParseLine(testCSVRows[rowIndex])
	require.NoError(t, err)
	require.Nil(t, m)
	rowIndex++
	m, err = p.ParseLine(testCSVRows[rowIndex])
	require.NoError(t, err)
	require.Equal(t, expectedFields[1], m.Fields())
	require.Equal(t, expectedTags[1], m.Tags())
}

func TestOverwriteDefaultTagsAndMetaDataTags(t *testing.T) {
	csv := []byte(`second=orange
fourth=plain
1.4,apple,hi
`)
	defaultTags := map[string]string{"third": "bye", "fourth": "car"}

	tests := []struct {
		name         string
		tagOverwrite bool
		expectedTags map[string]string
	}{
		{
			name:         "Don't overwrite tags",
			tagOverwrite: false,
			expectedTags: map[string]string{"second": "orange", "third": "bye", "fourth": "car"},
		},
		{
			name:         "Overwrite tags",
			tagOverwrite: true,
			expectedTags: map[string]string{"second": "apple", "third": "hi", "fourth": "plain"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				ColumnNames:        []string{"first", "second", "third"},
				TagColumns:         []string{"second", "third"},
				TagOverwrite:       tt.tagOverwrite,
				MetadataRows:       2,
				MetadataSeparators: []string{"="},
			}

			require.NoError(t, p.Init())
			p.SetDefaultTags(defaultTags)

			metrics, err := p.Parse(csv)
			require.NoError(t, err)
			require.Len(t, metrics, 1)
			require.EqualValues(t, tt.expectedTags, metrics[0].Tags())
		})
	}
}

func TestParseCSVResetModeInvalid(t *testing.T) {
	p := &Parser{
		HeaderRowCount: 1,
		ResetMode:      "garbage",
	}
	require.Error(t, p.Init(), `unknown reset mode "garbage"`)
}

func TestParseCSVResetModeNone(t *testing.T) {
	testCSV := `garbage nonsense that needs be skipped

# version= 1.0

    invalid meta data that can be ignored.
file created: 2021-10-08T12:34:18+10:00
timestamp,type,name,status
2020-11-23T08:19:27+00:00,Reader,R002,1
#2020-11-04T13:23:04+00:00,Reader,R031,0
2020-11-04T13:29:47+00:00,Coordinator,C001,0`

	expected := []telegraf.Metric{
		metric.New(
			"",
			map[string]string{
				"file created": "2021-10-08T12:34:18+10:00",
				"test":         "tag",
				"type":         "Reader",
				"version":      "1.0",
			},
			map[string]interface{}{
				"name":   "R002",
				"status": int64(1),
			},
			time.Date(2020, 11, 23, 8, 19, 27, 0, time.UTC),
		),
		metric.New(
			"",
			map[string]string{
				"file created": "2021-10-08T12:34:18+10:00",
				"test":         "tag",
				"type":         "Coordinator",
				"version":      "1.0",
			},
			map[string]interface{}{
				"name":   "C001",
				"status": int64(0),
			},
			time.Date(2020, 11, 4, 13, 29, 47, 0, time.UTC),
		),
	}

	p := &Parser{
		HeaderRowCount:     1,
		SkipRows:           2,
		MetadataRows:       4,
		Comment:            "#",
		TagColumns:         []string{"type"},
		MetadataSeparators: []string{":", "="},
		MetadataTrimSet:    " #",
		TimestampColumn:    "timestamp",
		TimestampFormat:    "2006-01-02T15:04:05Z07:00",
		ResetMode:          "none",
	}
	require.NoError(t, p.Init())
	// Set default Tags
	p.SetDefaultTags(map[string]string{"test": "tag"})

	// Do the parsing the first time
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, metrics)

	// Parsing another data line should work when not resetting
	additionalCSV := "2021-12-01T19:01:00+00:00,Reader,R009,5\r\n"
	additionalExpected := []telegraf.Metric{
		metric.New(
			"",
			map[string]string{
				"file created": "2021-10-08T12:34:18+10:00",
				"test":         "tag",
				"type":         "Reader",
				"version":      "1.0",
			},
			map[string]interface{}{
				"name":   "R009",
				"status": int64(5),
			},
			time.Date(2021, 12, 1, 19, 1, 0, 0, time.UTC),
		),
	}
	metrics, err = p.Parse([]byte(additionalCSV))
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, additionalExpected, metrics)

	// This should fail when not resetting but reading again due to the header etc
	_, err = p.Parse([]byte(testCSV))
	require.Error(
		t,
		err,
		`parsing time "garbage nonsense that needs be skipped" as "2006-01-02T15:04:05Z07:00": cannot parse "garbage nonsense that needs be skipped" as "2006"`,
	)
}

func TestParseCSVLinewiseResetModeNone(t *testing.T) {
	testCSV := []string{
		"garbage nonsense that needs be skipped",
		"",
		"# version= 1.0\r\n",
		"",
		"    invalid meta data that can be ignored.\r\n",
		"file created: 2021-10-08T12:34:18+10:00",
		"timestamp,type,name,status\n",
		"2020-11-23T08:19:27+00:00,Reader,R002,1\r\n",
		"#2020-11-04T13:23:04+00:00,Reader,R031,0\n",
		"2020-11-04T13:29:47+00:00,Coordinator,C001,0",
	}

	expected := []telegraf.Metric{
		metric.New(
			"",
			map[string]string{
				"file created": "2021-10-08T12:34:18+10:00",
				"test":         "tag",
				"type":         "Reader",
				"version":      "1.0",
			},
			map[string]interface{}{
				"name":   "R002",
				"status": int64(1),
			},
			time.Date(2020, 11, 23, 8, 19, 27, 0, time.UTC),
		),
		metric.New(
			"",
			map[string]string{
				"file created": "2021-10-08T12:34:18+10:00",
				"test":         "tag",
				"type":         "Coordinator",
				"version":      "1.0",
			},
			map[string]interface{}{
				"name":   "C001",
				"status": int64(0),
			},
			time.Date(2020, 11, 4, 13, 29, 47, 0, time.UTC),
		),
	}

	p := &Parser{
		HeaderRowCount:     1,
		SkipRows:           2,
		MetadataRows:       4,
		Comment:            "#",
		TagColumns:         []string{"type"},
		MetadataSeparators: []string{":", "="},
		MetadataTrimSet:    " #",
		TimestampColumn:    "timestamp",
		TimestampFormat:    "2006-01-02T15:04:05Z07:00",
		ResetMode:          "none",
	}
	require.NoError(t, p.Init())

	// Set default Tags
	p.SetDefaultTags(map[string]string{"test": "tag"})

	// Do the parsing the first time
	var metrics []telegraf.Metric
	for i, r := range testCSV {
		m, err := p.ParseLine(r)
		// Header lines should return "not enough data"
		if i < p.SkipRows+p.MetadataRows {
			require.ErrorIs(t, err, parsers.ErrEOF)
			require.Nil(t, m)
			continue
		}
		require.NoErrorf(t, err, "failed in row %d", i)
		if m != nil {
			metrics = append(metrics, m)
		}
	}
	testutil.RequireMetricsEqual(t, expected, metrics)

	// Parsing another data line should work when not resetting
	additionalCSV := "2021-12-01T19:01:00+00:00,Reader,R009,5\r\n"
	additionalExpected := metric.New(
		"",
		map[string]string{
			"file created": "2021-10-08T12:34:18+10:00",
			"test":         "tag",
			"type":         "Reader",
			"version":      "1.0",
		},
		map[string]interface{}{
			"name":   "R009",
			"status": int64(5),
		},
		time.Date(2021, 12, 1, 19, 1, 0, 0, time.UTC),
	)
	m, err := p.ParseLine(additionalCSV)
	require.NoError(t, err)
	testutil.RequireMetricEqual(t, additionalExpected, m)

	// This should fail when not resetting but reading again due to the header etc
	_, err = p.ParseLine(testCSV[0])
	require.Error(
		t,
		err,
		`parsing time "garbage nonsense that needs be skipped" as "2006-01-02T15:04:05Z07:00": cannot parse "garbage nonsense that needs be skipped" as "2006"`,
	)
}

func TestParseCSVResetModeAlways(t *testing.T) {
	testCSV := `garbage nonsense that needs be skipped

# version= 1.0

    invalid meta data that can be ignored.
file created: 2021-10-08T12:34:18+10:00
timestamp,type,name,status
2020-11-23T08:19:27+00:00,Reader,R002,1
#2020-11-04T13:23:04+00:00,Reader,R031,0
2020-11-04T13:29:47+00:00,Coordinator,C001,0`

	expected := []telegraf.Metric{
		metric.New(
			"",
			map[string]string{
				"file created": "2021-10-08T12:34:18+10:00",
				"test":         "tag",
				"type":         "Reader",
				"version":      "1.0",
			},
			map[string]interface{}{
				"name":   "R002",
				"status": int64(1),
			},
			time.Date(2020, 11, 23, 8, 19, 27, 0, time.UTC),
		),
		metric.New(
			"",
			map[string]string{
				"file created": "2021-10-08T12:34:18+10:00",
				"test":         "tag",
				"type":         "Coordinator",
				"version":      "1.0",
			},
			map[string]interface{}{
				"name":   "C001",
				"status": int64(0),
			},
			time.Date(2020, 11, 4, 13, 29, 47, 0, time.UTC),
		),
	}

	p := &Parser{
		HeaderRowCount:     1,
		SkipRows:           2,
		MetadataRows:       4,
		Comment:            "#",
		TagColumns:         []string{"type", "category"},
		MetadataSeparators: []string{":", "="},
		MetadataTrimSet:    " #",
		TimestampColumn:    "timestamp",
		TimestampFormat:    "2006-01-02T15:04:05Z07:00",
		ResetMode:          "always",
	}
	require.NoError(t, p.Init())
	// Set default Tags
	p.SetDefaultTags(map[string]string{"test": "tag"})

	// Do the parsing the first time
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, metrics)

	// Parsing another data line should fail as it is interpreted as header
	additionalCSV := "2021-12-01T19:01:00+00:00,Reader,R009,5\r\n"
	metrics, err = p.Parse([]byte(additionalCSV))
	require.ErrorIs(t, err, parsers.ErrEOF)
	require.Nil(t, metrics)

	// Prepare a second CSV with different column names
	testCSV = `garbage nonsense that needs be skipped

# version= 1.0

    invalid meta data that can be ignored.
file created: 2021-10-08T12:34:18+10:00
timestamp,category,id,flag
2020-11-23T08:19:27+00:00,Reader,R002,1
#2020-11-04T13:23:04+00:00,Reader,R031,0
2020-11-04T13:29:47+00:00,Coordinator,C001,0`

	expected = []telegraf.Metric{
		metric.New(
			"",
			map[string]string{
				"file created": "2021-10-08T12:34:18+10:00",
				"test":         "tag",
				"category":     "Reader",
				"version":      "1.0",
			},
			map[string]interface{}{
				"id":   "R002",
				"flag": int64(1),
			},
			time.Date(2020, 11, 23, 8, 19, 27, 0, time.UTC),
		),
		metric.New(
			"",
			map[string]string{
				"file created": "2021-10-08T12:34:18+10:00",
				"test":         "tag",
				"category":     "Coordinator",
				"version":      "1.0",
			},
			map[string]interface{}{
				"id":   "C001",
				"flag": int64(0),
			},
			time.Date(2020, 11, 4, 13, 29, 47, 0, time.UTC),
		),
	}

	// This should work as the parser is reset
	metrics, err = p.Parse([]byte(testCSV))
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, metrics)
}

func TestParseCSVLinewiseResetModeAlways(t *testing.T) {
	testCSV := []string{
		"garbage nonsense that needs be skipped",
		"",
		"# version= 1.0\r\n",
		"",
		"    invalid meta data that can be ignored.\r\n",
		"file created: 2021-10-08T12:34:18+10:00",
		"timestamp,type,name,status\n",
		"2020-11-23T08:19:27+00:00,Reader,R002,1\r\n",
		"#2020-11-04T13:23:04+00:00,Reader,R031,0\n",
		"2020-11-04T13:29:47+00:00,Coordinator,C001,0",
	}

	expected := []telegraf.Metric{
		metric.New(
			"",
			map[string]string{
				"file created": "2021-10-08T12:34:18+10:00",
				"test":         "tag",
				"type":         "Reader",
				"version":      "1.0",
			},
			map[string]interface{}{
				"name":   "R002",
				"status": int64(1),
			},
			time.Date(2020, 11, 23, 8, 19, 27, 0, time.UTC),
		),
		metric.New(
			"",
			map[string]string{
				"file created": "2021-10-08T12:34:18+10:00",
				"test":         "tag",
				"type":         "Coordinator",
				"version":      "1.0",
			},
			map[string]interface{}{
				"name":   "C001",
				"status": int64(0),
			},
			time.Date(2020, 11, 4, 13, 29, 47, 0, time.UTC),
		),
	}

	p := &Parser{
		HeaderRowCount:     1,
		SkipRows:           2,
		MetadataRows:       4,
		Comment:            "#",
		TagColumns:         []string{"type"},
		MetadataSeparators: []string{":", "="},
		MetadataTrimSet:    " #",
		TimestampColumn:    "timestamp",
		TimestampFormat:    "2006-01-02T15:04:05Z07:00",
		ResetMode:          "always",
	}
	require.NoError(t, p.Init())

	// Set default Tags
	p.SetDefaultTags(map[string]string{"test": "tag"})

	// Do the parsing the first time
	var metrics []telegraf.Metric
	for i, r := range testCSV {
		m, err := p.ParseLine(r)
		// Header lines should return "not enough data"
		if i < p.SkipRows+p.MetadataRows {
			require.ErrorIs(t, err, parsers.ErrEOF)
			require.Nil(t, m)
			continue
		}
		require.NoErrorf(t, err, "failed in row %d", i)
		if m != nil {
			metrics = append(metrics, m)
		}
	}
	testutil.RequireMetricsEqual(t, expected, metrics)

	// Parsing another data line should work in line-wise parsing as
	// reset-mode "always" is ignored.
	additionalCSV := "2021-12-01T19:01:00+00:00,Reader,R009,5\r\n"
	additionalExpected := metric.New(
		"",
		map[string]string{
			"file created": "2021-10-08T12:34:18+10:00",
			"test":         "tag",
			"type":         "Reader",
			"version":      "1.0",
		},
		map[string]interface{}{
			"name":   "R009",
			"status": int64(5),
		},
		time.Date(2021, 12, 1, 19, 1, 0, 0, time.UTC),
	)
	m, err := p.ParseLine(additionalCSV)
	require.NoError(t, err)
	testutil.RequireMetricEqual(t, additionalExpected, m)

	// This should fail as reset-mode "always" is ignored in line-wise parsing
	_, err = p.ParseLine(testCSV[0])
	require.Error(
		t,
		err,
		`parsing time "garbage nonsense that needs be skipped" as "2006-01-02T15:04:05Z07:00": cannot parse "garbage nonsense that needs be skipped" as "2006"`,
	)
}

const benchmarkData = `tags_host,tags_platform,tags_sdkver,value,timestamp
myhost,python,3.11.5,5,1653643420
myhost,python,3.11.4,4,1653643420
`

func TestBenchmarkData(t *testing.T) {
	plugin := &Parser{
		MetricName:      "benchmark",
		HeaderRowCount:  1,
		TimestampColumn: "timestamp",
		TimestampFormat: "unix",
		TagColumns:      []string{"tags_host", "tags_platform", "tags_sdkver"},
	}
	require.NoError(t, plugin.Init())

	expected := []telegraf.Metric{
		metric.New(
			"benchmark",
			map[string]string{
				"tags_host":     "myhost",
				"tags_platform": "python",
				"tags_sdkver":   "3.11.5",
			},
			map[string]interface{}{
				"value": 5,
			},
			time.Unix(1653643420, 0),
		),
		metric.New(
			"benchmark",
			map[string]string{
				"tags_host":     "myhost",
				"tags_platform": "python",
				"tags_sdkver":   "3.11.4",
			},
			map[string]interface{}{
				"value": 4,
			},
			time.Unix(1653643420, 0),
		),
	}

	actual, err := plugin.Parse([]byte(benchmarkData))
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
}

func TestConcurrentParsing(t *testing.T) {
	// Test concurrent access to ensure the parser is thread-safe
	plugin := &Parser{
		MetricName:      "benchmark",
		HeaderRowCount:  1,
		TimestampColumn: "timestamp",
		TimestampFormat: "unix",
		TagColumns:      []string{"tags_host", "tags_platform", "tags_sdkver"},
		ResetMode:       "always", // Reset parser state on each Parse() call
	}
	require.NoError(t, plugin.Init())

	const numGoroutines = 5
	const numIterations = 10

	// Use the same test data for all goroutines to ensure consistency
	testData := benchmarkData

	// WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Start multiple goroutines to parse concurrently
	for i := range numGoroutines {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := range numIterations {
				metrics, err := plugin.Parse([]byte(testData))
				if err != nil {
					t.Logf("goroutine %d iteration %d: %v", goroutineID, j, err)
					t.Fail()
					return
				}

				// Verify we got the expected number of metrics
				if len(metrics) != 2 {
					t.Logf("goroutine %d iteration %d: expected 2 metrics, got %d", goroutineID, j, len(metrics))
					t.Fail()
					return
				}

				// Verify basic metric structure
				for k, metric := range metrics {
					if metric.Name() != "benchmark" {
						t.Logf("goroutine %d iteration %d metric %d: expected name 'benchmark', got '%s'", goroutineID, j, k, metric.Name())
						t.Fail()
						return
					}
					if len(metric.Tags()) != 3 {
						t.Logf("goroutine %d iteration %d metric %d: expected 3 tags, got %d", goroutineID, j, k, len(metric.Tags()))
						t.Fail()
						return
					}
					if len(metric.Fields()) != 1 {
						t.Logf("goroutine %d iteration %d metric %d: expected 1 field, got %d", goroutineID, j, k, len(metric.Fields()))
						t.Fail()
						return
					}
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for any errors
	require.False(t, t.Failed(), "Concurrent parsing failed with errors")
}

func TestConcurrentParseLineWithReset(t *testing.T) {
	// Test concurrent access with ParseLine which has more complex state management
	plugin := &Parser{
		MetricName:      "benchmark",
		HeaderRowCount:  1,
		TimestampColumn: "timestamp",
		TimestampFormat: "unix",
		TagColumns:      []string{"tags_host", "tags_platform", "tags_sdkver"},
		ResetMode:       "always",
	}
	require.NoError(t, plugin.Init())

	const numGoroutines = 5
	const numIterations = 10

	// Test data - each goroutine will use the same CSV format
	testCSV := `tags_host,tags_platform,tags_sdkver,value,timestamp
myhost,python,3.11.5,5,1653643420`

	// WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Start multiple goroutines to parse concurrently
	for i := range numGoroutines {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := range numIterations {
				// Parse complete CSV (with header)
				metrics, err := plugin.Parse([]byte(testCSV))
				if err != nil {
					t.Logf("goroutine %d iteration %d: %v", goroutineID, j, err)
					t.Fail()
					return
				}

				if len(metrics) != 1 {
					t.Logf("goroutine %d iteration %d: expected 1 metric, got %d", goroutineID, j, len(metrics))
					t.Fail()
					return
				}

				// Verify basic metric structure
				metric := metrics[0]
				if metric.Name() != "benchmark" {
					t.Logf("goroutine %d iteration %d: expected name 'benchmark', got '%s'", goroutineID, j, metric.Name())
					t.Fail()
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for any errors
	require.False(t, t.Failed(), "Concurrent parsing failed with errors")
}

func TestConcurrentParsingStress(t *testing.T) {
	// Stress test with more goroutines and iterations to ensure robust thread safety
	plugin := &Parser{
		MetricName:      "stress_test",
		HeaderRowCount:  1,
		TimestampColumn: "timestamp",
		TimestampFormat: "unix",
		TagColumns:      []string{"host", "service"},
		ResetMode:       "always",
	}
	require.NoError(t, plugin.Init())

	const numGoroutines = 20
	const numIterations = 100

	// Different CSV data to test various parsing scenarios
	testDataVariants := []string{
		`host,service,value,timestamp
server1,web,100,1653643420
server1,db,200,1653643421`,
		`host,service,value,timestamp
server2,api,150,1653643430
server2,cache,50,1653643431`,
		`host,service,value,timestamp
server3,worker,75,1653643440`,
	}

	// WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Start multiple goroutines to parse concurrently
	for i := range numGoroutines {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := range numIterations {
				// Use different data variants to test different scenarios
				testData := testDataVariants[j%len(testDataVariants)]

				metrics, err := plugin.Parse([]byte(testData))
				if err != nil {
					t.Logf("goroutine %d iteration %d: %v", goroutineID, j, err)
					t.Fail()
					return
				}

				// Verify we got some metrics (different variants have different counts)
				if len(metrics) == 0 {
					t.Logf("goroutine %d iteration %d: expected at least 1 metric, got 0", goroutineID, j)
					t.Fail()
					return
				}

				// Verify all metrics have the correct name and required fields
				for k, metric := range metrics {
					if metric.Name() != "stress_test" {
						t.Logf("goroutine %d iteration %d metric %d: expected name 'stress_test', got '%s'", goroutineID, j, k, metric.Name())
						t.Fail()
						return
					}

					// Should have host and service tags
					tags := metric.Tags()
					if _, exists := tags["host"]; !exists {
						t.Logf("goroutine %d iteration %d metric %d: missing 'host' tag", goroutineID, j, k)
						t.Fail()
						return
					}
					if _, exists := tags["service"]; !exists {
						t.Logf("goroutine %d iteration %d metric %d: missing 'service' tag", goroutineID, j, k)
						t.Fail()
						return
					}

					// Should have value field
					fields := metric.Fields()
					if _, exists := fields["value"]; !exists {
						t.Logf("goroutine %d iteration %d metric %d: missing 'value' field", goroutineID, j, k)
						t.Fail()
						return
					}
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check for any errors
	require.False(t, t.Failed(), "Concurrent parsing failed with errors")
}

func BenchmarkParsing(b *testing.B) {
	plugin := &Parser{
		MetricName:      "benchmark",
		HeaderRowCount:  1,
		TimestampColumn: "timestamp",
		TimestampFormat: "unix",
		TagColumns:      []string{"tags_host", "tags_platform", "tags_sdkver"},
	}
	require.NoError(b, plugin.Init())

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse([]byte(benchmarkData))
	}
}

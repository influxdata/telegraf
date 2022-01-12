package csv

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
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
	require.Equal(t, []telegraf.Metric{}, metrics)
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
	require.Equal(t, metrics[0].Time().UnixNano(), int64(1243094706000000000))
	require.Equal(t, metrics[1].Time().UnixNano(), int64(1257609906000000000))
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
	require.Equal(t, metrics[0].Time().UnixNano(), int64(1243094700000000000))
	require.Equal(t, metrics[1].Time().UnixNano(), int64(1247328300000000000))
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
	require.Equal(t, fmt.Errorf("timestamp format must be specified"), err)
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
	require.Equal(t, metrics[0].Time().UnixNano(), int64(1243094706000000000))
	require.Equal(t, metrics[1].Time().UnixNano(), int64(1257609906000000000))
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
	require.Equal(t, metrics[0].Time().UnixNano(), int64(1243094706123000000))
	require.Equal(t, metrics[1].Time().UnixNano(), int64(1257609906123000000))
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

	//deep equal fields
	require.Equal(t, expectedMetric.Fields(), returnedMetric.Fields())

	// Test explicit type conversion.
	p.ColumnTypes = []string{"float", "int", "bool", "string"}

	metrics, err = p.Parse([]byte(testCSV))
	require.NoError(t, err)

	returnedMetric = metric.New(metrics[0].Name(), metrics[0].Tags(), metrics[0].Fields(), time.Unix(0, 0))

	//deep equal fields
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
	require.Error(t, io.EOF, err)
	require.Error(t, err)
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
	require.Error(t, io.EOF, err)
	require.Error(t, err)
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
	require.Len(t, metrics, 0)
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
	require.Len(t, metrics, 0)
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
	require.Equal(t, metrics[0].Time().UnixNano(), int64(1243094706000000000))
	require.Equal(t, metrics[1].Time().UnixNano(), int64(1257609906000000000))
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

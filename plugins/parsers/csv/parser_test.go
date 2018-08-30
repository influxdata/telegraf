package csv

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

func TestBasicCSV(t *testing.T) {
	p := Parser{
		ColumnNames: []string{"first", "second", "third"},
		TagColumns:  []string{"third"},
	}

	_, err := p.ParseLine("1.4,true,hi")
	require.NoError(t, err)
}

func TestHeaderConcatenationCSV(t *testing.T) {
	p := Parser{
		HeaderRowCount:    2,
		MeasurementColumn: "3",
	}
	testCSV := `first,second
1,2,3
3.4,70,test_name`

	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, "test_name", metrics[0].Name())
}

func TestHeaderOverride(t *testing.T) {
	p := Parser{
		HeaderRowCount:    1,
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
	}
	testCSV := `line1,line2,line3
3.4,70,test_name`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, "test_name", metrics[0].Name())
}

func TestTimestamp(t *testing.T) {
	p := Parser{
		HeaderRowCount:    1,
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
		TimestampColumn:   "first",
		TimestampFormat:   "02/01/06 03:04:05 PM",
	}
	testCSV := `line1,line2,line3
23/05/09 04:05:06 PM,70,test_name
07/11/09 04:05:06 PM,80,test_name2`
	metrics, err := p.Parse([]byte(testCSV))

	require.NoError(t, err)
	require.Equal(t, metrics[0].Time().UnixNano(), int64(1243094706000000000))
	require.Equal(t, metrics[1].Time().UnixNano(), int64(1257609906000000000))
}

func TestTimestampError(t *testing.T) {
	p := Parser{
		HeaderRowCount:    1,
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
		TimestampColumn:   "first",
	}
	testCSV := `line1,line2,line3
23/05/09 04:05:06 PM,70,test_name
07/11/09 04:05:06 PM,80,test_name2`
	_, err := p.Parse([]byte(testCSV))
	require.Equal(t, fmt.Errorf("timestamp format must be specified"), err)
}

func TestQuotedCharacter(t *testing.T) {
	p := Parser{
		HeaderRowCount:    1,
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
	}

	testCSV := `line1,line2,line3
"3,4",70,test_name`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, "3,4", metrics[0].Fields()["first"])
}

func TestDelimiter(t *testing.T) {
	p := Parser{
		HeaderRowCount:    1,
		Delimiter:         "%",
		ColumnNames:       []string{"first", "second", "third"},
		MeasurementColumn: "third",
	}

	testCSV := `line1%line2%line3
3,4%70%test_name`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, "3,4", metrics[0].Fields()["first"])
}

func TestValueConversion(t *testing.T) {
	p := Parser{
		HeaderRowCount: 0,
		Delimiter:      ",",
		ColumnNames:    []string{"first", "second", "third", "fourth"},
		MetricName:     "test_value",
	}
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

	expectedMetric, err1 := metric.New("test_value", expectedTags, expectedFields, time.Unix(0, 0))
	returnedMetric, err2 := metric.New(metrics[0].Name(), metrics[0].Tags(), metrics[0].Fields(), time.Unix(0, 0))
	require.NoError(t, err1)
	require.NoError(t, err2)

	//deep equal fields
	require.Equal(t, expectedMetric.Fields(), returnedMetric.Fields())
}

func TestSkipComment(t *testing.T) {
	p := Parser{
		HeaderRowCount: 0,
		Comment:        "#",
		ColumnNames:    []string{"first", "second", "third", "fourth"},
		MetricName:     "test_value",
	}
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
	p := Parser{
		HeaderRowCount: 0,
		TrimSpace:      true,
		ColumnNames:    []string{"first", "second", "third", "fourth"},
		MetricName:     "test_value",
	}
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
}

func TestSkipRows(t *testing.T) {
	p := Parser{
		HeaderRowCount:    1,
		SkipRows:          1,
		TagColumns:        []string{"line1"},
		MeasurementColumn: "line3",
	}
	testCSV := `garbage nonsense
line1,line2,line3
hello,80,test_name2`

	expectedFields := map[string]interface{}{
		"line2": int64(80),
		"line3": "test_name2",
	}
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, expectedFields, metrics[0].Fields())
}

func TestSkipColumns(t *testing.T) {
	p := Parser{
		SkipColumns: 1,
		ColumnNames: []string{"line1", "line2"},
	}
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
	p := Parser{
		SkipColumns:    1,
		HeaderRowCount: 2,
	}
	testCSV := `col,col,col
	1,2,3
	trash,80,test_name`

	// we should expect an error if we try to get col1
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{"col2": int64(80), "col3": "test_name"}, metrics[0].Fields())
}

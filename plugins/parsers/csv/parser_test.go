package csv

import (
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

func TestBasicCSV(t *testing.T) {
	p := CSVParser{
		DataColumns:  []string{"first", "second", "third"},
		FieldColumns: []string{"first", "second"},
		TagColumns:   []string{"third"},
	}

	_, err := p.ParseLine("1.4,true,hi")
	require.NoError(t, err)
}

func TestHeaderCSV(t *testing.T) {
	p := CSVParser{
		Header:       true,
		FieldColumns: []string{"first", "second"},
		NameColumn:   "third",
	}
	testCSV := `first,second,third
3.4,70,test_name`

	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, "test_name", metrics[0].Name())
}

func TestHeaderOverride(t *testing.T) {
	p := CSVParser{
		Header:       true,
		DataColumns:  []string{"first", "second", "third"},
		FieldColumns: []string{"first", "second"},
		NameColumn:   "third",
	}
	testCSV := `line1,line2,line3
3.4,70,test_name`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, "test_name", metrics[0].Name())
}

func TestTimestamp(t *testing.T) {
	p := CSVParser{
		Header:          true,
		DataColumns:     []string{"first", "second", "third"},
		FieldColumns:    []string{"second"},
		NameColumn:      "third",
		TimestampColumn: "first",
		TimestampFormat: "02/01/06 03:04:05 PM",
	}
	testCSV := `line1,line2,line3
23/05/09 04:05:06 PM,70,test_name
07/11/09 04:05:06 PM,80,test_name2`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.NotEqual(t, metrics[1].Time(), metrics[0].Time())
}

func TestTimestampError(t *testing.T) {
	p := CSVParser{
		Header:          true,
		DataColumns:     []string{"first", "second", "third"},
		FieldColumns:    []string{"second"},
		NameColumn:      "third",
		TimestampColumn: "first",
	}
	testCSV := `line1,line2,line3
23/05/09 04:05:06 PM,70,test_name
07/11/09 04:05:06 PM,80,test_name2`
	_, err := p.Parse([]byte(testCSV))
	require.Equal(t, fmt.Errorf("timestamp format must be specified"), err)
}

func TestQuotedCharacter(t *testing.T) {
	p := CSVParser{
		Header:       true,
		DataColumns:  []string{"first", "second", "third"},
		FieldColumns: []string{"second", "first"},
		NameColumn:   "third",
	}

	testCSV := `line1,line2,line3
"3,4",70,test_name`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, "3,4", metrics[0].Fields()["first"])
}

func TestDelimiter(t *testing.T) {
	p := CSVParser{
		Header:       true,
		Delimiter:    "%",
		DataColumns:  []string{"first", "second", "third"},
		FieldColumns: []string{"second", "first"},
		NameColumn:   "third",
	}

	testCSV := `line1%line2%line3
3,4%70%test_name`
	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)
	require.Equal(t, "3,4", metrics[0].Fields()["first"])
}

func TestValueConversion(t *testing.T) {
	p := CSVParser{
		Header:       false,
		Delimiter:    ",",
		DataColumns:  []string{"first", "second", "third", "fourth"},
		FieldColumns: []string{"second", "first", "third", "fourth"},
		MetricName:   "test_value",
	}
	testCSV := `3.3,4,true,hello`

	expectedFields := map[string]interface{}{
		"first":  3.3,
		"second": 4,
		"third":  true,
		"fourth": "hello",
	}

	metrics, err := p.Parse([]byte(testCSV))
	require.NoError(t, err)

	expectedTags := make(map[string]string)
	goodMetric, err1 := metric.New("test_value", expectedTags, expectedFields, time.Unix(0, 0))
	returnedMetric, err2 := metric.New(metrics[0].Name(), metrics[0].Tags(), metrics[0].Fields(), time.Unix(0, 0))
	require.NoError(t, err1)
	require.NoError(t, err2)

	//deep equal fields
	for k := range goodMetric.Fields() {
		log.Printf("expected field: %v, %T", goodMetric.Fields()[k], goodMetric.Fields()[k])
		log.Printf("returned field: %v, %T", returnedMetric.Fields()[k], returnedMetric.Fields()[k])
	}
	require.True(t, reflect.DeepEqual(goodMetric.Fields(), returnedMetric.Fields()))
}

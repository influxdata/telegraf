package metric

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
)

// vars for making sure that the compiler doesnt optimize out the benchmarks:
var (
	s      string
	I      interface{}
	tags   map[string]string
	fields map[string]interface{}
)

func BenchmarkNewMetric(b *testing.B) {
	var mt telegraf.Metric
	for n := 0; n < b.N; n++ {
		mt, _ = New("test_metric",
			map[string]string{
				"test_tag_1": "tag_value_1",
				"test_tag_2": "tag_value_2",
				"test_tag_3": "tag_value_3",
			},
			map[string]interface{}{
				"string_field": "string",
				"int_field":    int64(1000),
				"float_field":  float64(2.1),
			},
			time.Now(),
		)
	}
	s = string(mt.String())
}

func BenchmarkTags(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var mt, _ = New("test_metric",
			map[string]string{
				"test_tag_1": "tag_value_1",
				"test_tag_2": "tag_value_2",
				"test_tag_3": "tag_value_3",
			},
			map[string]interface{}{
				"string_field": "string",
				"int_field":    int64(1000),
				"float_field":  float64(2.1),
			},
			time.Now(),
		)
		tags = mt.Tags()
	}
	s = fmt.Sprint(tags)
}

func BenchmarkFields(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var mt, _ = New("test_metric",
			map[string]string{
				"test_tag_1": "tag_value_1",
				"test_tag_2": "tag_value_2",
				"test_tag_3": "tag_value_3",
			},
			map[string]interface{}{
				"string_field": "string",
				"int_field":    int64(1000),
				"float_field":  float64(2.1),
			},
			time.Now(),
		)
		fields = mt.Fields()
	}
	s = fmt.Sprint(fields)
}

func BenchmarkString(b *testing.B) {
	mt, _ := New("test_metric",
		map[string]string{
			"test_tag_1": "tag_value_1",
			"test_tag_2": "tag_value_2",
			"test_tag_3": "tag_value_3",
		},
		map[string]interface{}{
			"string_field": "string",
			"int_field":    int64(1000),
			"float_field":  float64(2.1),
		},
		time.Now(),
	)
	var S string
	for n := 0; n < b.N; n++ {
		S = mt.String()
	}
	s = S
}

func BenchmarkSerialize(b *testing.B) {
	mt, _ := New("test_metric",
		map[string]string{
			"test_tag_1": "tag_value_1",
			"test_tag_2": "tag_value_2",
			"test_tag_3": "tag_value_3",
		},
		map[string]interface{}{
			"string_field": "string",
			"int_field":    int64(1000),
			"float_field":  float64(2.1),
		},
		time.Now(),
	)
	var B []byte
	for n := 0; n < b.N; n++ {
		B = mt.Serialize()
	}
	s = string(B)
}

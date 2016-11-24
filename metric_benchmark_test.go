package telegraf

import (
	"fmt"
	"testing"
	"time"
)

// vars for making sure that the compiler doesnt optimize out the benchmarks:
var (
	s      string
	I      interface{}
	tags   map[string]string
	fields map[string]interface{}
)

func BenchmarkNewMetric(b *testing.B) {
	var mt Metric
	for n := 0; n < b.N; n++ {
		mt, _ = NewMetric("test_metric",
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

func BenchmarkNewMetricAndInspect(b *testing.B) {
	var mt Metric
	for n := 0; n < b.N; n++ {
		mt, _ = NewMetric("test_metric",
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
		for k, v := range mt.Fields() {
			s = k
			I = v
		}
	}
	s = mt.String()
}

func BenchmarkTags(b *testing.B) {
	var mt, _ = NewMetric("test_metric",
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
	for n := 0; n < b.N; n++ {
		tags = mt.Tags()
	}
	s = fmt.Sprint(tags)
}

func BenchmarkFields(b *testing.B) {
	var mt, _ = NewMetric("test_metric",
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
	for n := 0; n < b.N; n++ {
		fields = mt.Fields()
	}
	s = fmt.Sprint(fields)
}

func BenchmarkSerializeMetric(b *testing.B) {
	mt, _ := NewMetric("test_metric",
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

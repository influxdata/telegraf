package statsd

import (
	"fmt"
	"reflect"
	"testing"
)

func newStatsdParser() (*Parser, error) {
	p, err := NewParser(
		"_",
		[]string{
			"servers.* .host.measurement*",
			"users.* .measurement*",
			"template.* .measurement.measurement.tag_x.tag_y.tag_z",
		},
		nil)
	return p, err
}

func BenchmarkCount(b *testing.B) {
	p, err := newStatsdParser()

	if err != nil {
		b.Fatalf("unexpected error creating parser, got %v", err)
	}

	for i := 0; i < b.N; i++ {
		_, err := p.Parse([]byte("servers.localhost.cpu.load:11|c"))

		if err != nil {
			b.Fatalf("Parse fail: %v", err)
		}
	}
}

func BenchmarkTiming(b *testing.B) {
	p, err := newStatsdParser()

	if err != nil {
		b.Fatalf("unexpected error creating parser, got %v", err)
	}

	for i := 0; i < b.N; i++ {
		_, err := p.Parse([]byte("servers.localhost.cpu.load:11|ms"))

		if err != nil {
			b.Fatalf("Parse fail: %v", err)
		}
	}
}

func BenchmarkSet(b *testing.B) {
	p, err := newStatsdParser()

	if err != nil {
		b.Fatalf("unexpected error creating parser, got %v", err)
	}

	for i := 0; i < b.N; i++ {
		_, err := p.Parse([]byte(fmt.Sprintf("users.active:%v|s", i)))

		if err != nil {
			b.Fatalf("Parse fail: %v", err)
		}
	}
}

func TestParse(t *testing.T) {
	type metric struct {
		measurement string
		tags        map[string]string
		value       interface{}
	}

	var tests = []struct {
		test     string
		input    string
		template []string
		metrics  []metric
	}{
		{
			test:  "counter",
			input: "cpu.a.b.foo.bar:50|c",
			template: []string{
				"cpu.* .measurement.measurement.tag_foo.tag_bar",
			},
			metrics: []metric{
				{
					measurement: "a_b",
					tags: map[string]string{
						"tag_foo":     "foo",
						"tag_bar":     "bar",
						"metric_type": "counter",
					},
					value: int64(50),
				},
			},
		},
		{
			test:  "counter with samplerate",
			input: "cpu.a.b.foo.bar:50|c|@0.25",
			template: []string{
				"cpu.* .measurement.measurement.tag_foo.tag_bar",
			},
			metrics: []metric{
				{
					measurement: "a_b",
					tags: map[string]string{
						"tag_foo":     "foo",
						"tag_bar":     "bar",
						"metric_type": "counter",
					},
					value: int64(200),
				},
			},
		},
		{
			test:  "gauge",
			input: "cpu.a.b.foo.bar:50|g",
			template: []string{
				"cpu.* .measurement.measurement.tag_foo.tag_bar",
			},
			metrics: []metric{
				{
					measurement: "a_b",
					tags: map[string]string{
						"tag_foo":     "foo",
						"tag_bar":     "bar",
						"metric_type": "gauge",
					},
					value: float64(50),
				},
			},
		},
		{
			test:  "gauge with additive",
			input: "cpu.a.b.foo.bar:+50|g\ncpu.a.b.foo.bar:-50|g",
			template: []string{
				"cpu.* .measurement.measurement.tag_foo.tag_bar",
			},
			metrics: []metric{
				{
					measurement: "a_b",
					tags: map[string]string{
						"tag_foo":     "foo",
						"tag_bar":     "bar",
						"metric_type": "gauge",
						"operation":   "additive",
					},
					value: float64(50),
				}, {
					measurement: "a_b",
					tags: map[string]string{
						"tag_foo":     "foo",
						"tag_bar":     "bar",
						"metric_type": "gauge",
						"operation":   "additive",
					},
					value: float64(-50),
				},
			},
		},
		{
			test:  "timing",
			input: "cpu.a.b.foo.bar:50|ms",
			template: []string{
				"cpu.* .measurement.measurement.tag_foo.tag_bar",
			},
			metrics: []metric{
				{
					measurement: "a_b",
					tags: map[string]string{
						"tag_foo":     "foo",
						"tag_bar":     "bar",
						"metric_type": "timing",
					},
					value: float64(50.0),
				},
			},
		},
		{
			test:  "timing with samplerate",
			input: "cpu.a.b.foo.bar:50|ms|@0.5",
			template: []string{
				"cpu.* .measurement.measurement.tag_foo.tag_bar",
			},
			metrics: []metric{
				{
					measurement: "a_b",
					tags: map[string]string{
						"tag_foo":     "foo",
						"tag_bar":     "bar",
						"metric_type": "timing",
					},
					value: float64(50.0),
				},
				{
					measurement: "a_b",
					tags: map[string]string{
						"tag_foo":     "foo",
						"tag_bar":     "bar",
						"metric_type": "timing",
					},
					value: float64(50.0),
				},
			},
		},
		{
			test:  "histogram",
			input: "cpu.a.b.foo.bar:50|h",
			template: []string{
				"cpu.* .measurement.measurement.tag_foo.tag_bar",
			},
			metrics: []metric{
				{
					measurement: "a_b",
					tags: map[string]string{
						"tag_foo":     "foo",
						"tag_bar":     "bar",
						"metric_type": "histogram",
					},
					value: float64(50.0),
				},
			},
		},
		{
			test:  "set",
			input: "cpu.a.b.foo.bar:user|s",
			template: []string{
				"cpu.* .measurement.measurement.tag_foo.tag_bar",
			},
			metrics: []metric{
				{
					measurement: "a_b",
					tags: map[string]string{
						"tag_foo":     "foo",
						"tag_bar":     "bar",
						"metric_type": "set",
					},
					value: "user",
				},
			},
		},
		{
			test:  "multi metric in one line",
			input: "cpu.a.b.foo.bar:11|c:12|c:13|ms",
			template: []string{
				"cpu.* .measurement.measurement.tag_foo.tag_bar",
			},
			metrics: []metric{
				{
					measurement: "a_b",
					tags: map[string]string{
						"tag_foo":     "foo",
						"tag_bar":     "bar",
						"metric_type": "counter",
					},
					value: int64(11),
				},
				{
					measurement: "a_b",
					tags: map[string]string{
						"tag_foo":     "foo",
						"tag_bar":     "bar",
						"metric_type": "counter",
					},
					value: int64(12),
				},
				{
					measurement: "a_b",
					tags: map[string]string{
						"tag_foo":     "foo",
						"tag_bar":     "bar",
						"metric_type": "timing",
					},
					value: float64(13),
				},
			},
		},
	}

	for _, test := range tests {
		t.Logf("TestParser %s\n", test.test)
		p, err := newStatsdParser()
		if err != nil {
			t.Fatalf(test.test+" unexpected error creating parser: %v\n", err)
		}
		p.Templates = test.template

		metrics, err := p.Parse([]byte(test.input))
		if err != nil {
			t.Fatalf(test.test+" unexpected error parsing test: %v\n", err)
		}

		if len(metrics) != len(test.metrics) {
			t.Fatalf(test.test+" metrics len mismatch. expected %d, got %d\n",
				len(test.metrics), len(metrics))
		}
		for i := range test.metrics {
			actual := metrics[i]
			expected := test.metrics[i]
			if actual.Name() != expected.measurement {
				t.Fatalf(test.test+" name parse fail. expected %v, got %v", expected.measurement, actual.Name())
			}

			if !reflect.DeepEqual(actual.Tags(), expected.tags) {
				t.Fatalf(test.test+" tag parse fail. expected %+v, got %+v", expected.tags, actual.Tags())
			}

			var valueIsEqual bool
			switch actual.Tags()["metric_type"] {
			case "set":
				l := actual.Fields()["value"].(string)
				r := expected.value.(string)
				valueIsEqual = l == r
			case "counter":
				l := actual.Fields()["value"].(int64)
				r := expected.value.(int64)
				valueIsEqual = l == r
			default:
				l := actual.Fields()["value"].(float64)
				r := expected.value.(float64)
				valueIsEqual = l == r
			}

			if !valueIsEqual {
				t.Fatalf(test.test+" value parse fail. expected %v, got %v", expected.value, actual.Fields())
			}
		}
	}
}

func TestParseInvalid(t *testing.T) {
	var tests = []string{
		"i.dont.have.a.pipe:45g",
		"i.dont.have.a.colon45|c",
		"invalid.metric.type:45|e",
		"invalid.plus.minus.non.gauge:+10|s",
		"invalid.plus.minus.non.gauge:+10|ms",
		"invalid.plus.minus.non.gauge:+10|h",
		"invalid.value:foobar|c",
		"invalid.value:d11|c",
		"invalid.value:1d1|c",
		"invalid.sample.with.set:1|s|@0.1",
		"invalid.sample.with.gauge:1|g|@0.1",
	}

	for _, test := range tests {
		t.Logf("TestParseInvalid %s\n", test)

		p, err := newStatsdParser()
		if err != nil {
			t.Fatalf(test+" unexpected error creating parser: %v", err)
		}

		_, err = p.Parse([]byte(test))
		if err != nil {
			t.Fatalf("Parsing %s should have resulted in a error\n", test)
		}
	}
}

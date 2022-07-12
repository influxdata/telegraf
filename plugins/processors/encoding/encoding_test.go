package encoding

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestEncoding(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		compression string
		encoding    string
		result      string
		err         string
	}{
		{
			name:        "gzip empty actually isn't empty",
			value:       "",
			compression: "gzip",
			encoding:    "base64",
			result:      "H4sIAAAAAAAA/wEAAP//AAAAAAAAAAA",
			err:         "",
		},
		{
			name:        "base64 empty remains empty",
			value:       "",
			compression: "",
			encoding:    "base64",
			result:      "",
			err:         "",
		},
		{
			name:        "gzip works with encoding",
			value:       "this is a test",
			compression: "gzip",
			encoding:    "base64",
			result:      "H4sIAAAAAAAA/yrJyCxWyCxWSFQoSS0uAQQAAP//6uceDQ4AAAA",
			err:         "",
		},
		{
			name:        "encoding works without compression",
			value:       "this is a test",
			compression: "",
			encoding:    "base64",
			result:      "dGhpcyBpcyBhIHRlc3Q",
			err:         "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := encode(testCase.value, testCase.compression, testCase.encoding)
			if testCase.err != "" {
				assert.EqualError(t, err, testCase.err)
			}
			assert.Equal(t, testCase.result, result)
		})
	}
}

func TestDecoding(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		compression string
		encoding    string
		result      string
		err         string
	}{
		{
			name:        "gzip unzips empty string",
			value:       "H4sIAAAAAAAA/wEAAP//AAAAAAAAAAA",
			compression: "gzip",
			encoding:    "base64",
			result:      "",
			err:         "",
		},
		{
			name:        "base64 empty remains empty",
			value:       "",
			compression: "",
			encoding:    "base64",
			result:      "",
			err:         "",
		},
		{
			name:        "gzip works with encoding",
			value:       "H4sIAAAAAAAA/yrJyCxWyCxWSFQoSS0uAQQAAP//6uceDQ4AAAA",
			compression: "gzip",
			encoding:    "base64",
			result:      "this is a test",
			err:         "",
		},
		{
			name:        "encoding works without compression",
			value:       "dGhpcyBpcyBhIHRlc3Q",
			compression: "",
			encoding:    "base64",
			result:      "this is a test",
			err:         "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := decode(testCase.value, testCase.compression, testCase.encoding)
			if testCase.err != "" {
				assert.EqualError(t, err, testCase.err)
			}
			assert.Equal(t, testCase.result, result)
		})
	}
}

func TestInitialization(t *testing.T) {
	tests := []struct {
		name       string
		err        string
		manipulate func(*Encoding)
		validate   func(*Encoding) error
	}{
		{
			name:       "valid combination works",
			manipulate: func(e *Encoding) {},
		},
		{
			name:       "invalid compression causes error",
			manipulate: func(e *Encoding) { e.Compression = "invalid" },
			err:        "'invalid' is not a supported compression algorithm. It must be 'gzip' or '' to skip.",
		},
		{
			name:       "empty compression works",
			manipulate: func(e *Encoding) { e.Compression = "" },
		},
		{
			name:       "empty encoding causes error",
			manipulate: func(e *Encoding) { e.Encoding = "" },
			err:        "'encoding' is required for the encoding processor.",
		},
		{
			name:       "invalid encoding causes error",
			manipulate: func(e *Encoding) { e.Encoding = "invalid" },
			err:        "'invalid' is not a supported encoding. It must be 'base64'.",
		},
		{
			name:       "encode operation works",
			manipulate: func(e *Encoding) { e.Operation = "encode" },
		},
		{
			name:       "decode operation works",
			manipulate: func(e *Encoding) { e.Operation = "decode" },
		},
		{
			name:       "invalid operation causes error",
			manipulate: func(e *Encoding) { e.Operation = "invalid" },
			err:        "'invalid' is not a supported operation. It must be one of 'encode' or 'decode'.",
		},
		{
			name:       "assigns dest field to source field if unspecified",
			manipulate: func(e *Encoding) { e.DestField = "" },
			validate: func(e *Encoding) error {
				if e.DestField != "sample" {
					return fmt.Errorf("the dest field should be set to the source field if it's missing")
				}
				if e.RemoveOriginal {
					return fmt.Errorf("remove original shouldn't be true if the dest field is the source field")
				}
				return nil
			},
		},
		{
			name: "no remove original if dest and source fields are the same",
			manipulate: func(e *Encoding) {
				e.DestField = e.Field
				e.RemoveOriginal = true
			},
			validate: func(e *Encoding) error {
				if e.RemoveOriginal {
					return fmt.Errorf("remove original shouldn't be true if the dest field is the source field")
				}
				return nil
			},
		},
		{
			name: "allows removal of original field",
			manipulate: func(e *Encoding) {
				e.RemoveOriginal = true
			},
			validate: func(e *Encoding) error {
				if !e.RemoveOriginal {
					return fmt.Errorf("remove of the original should be possible")
				}
				return nil
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := &Encoding{
				Field:          "sample",
				RemoveOriginal: false,
				DestField:      "new-sample",
				Operation:      "encode",
				Compression:    "gzip",
				Encoding:       "base64",
			}

			if testCase.manipulate != nil {
				testCase.manipulate(e)
			}

			err := e.Init()
			if testCase.err != "" {
				assert.EqualError(t, err, testCase.err)
			} else {
				assert.NoError(t, err)
			}

			if testCase.validate != nil {
				assert.NoError(t, testCase.validate(e))
			}
		})
	}
}

var start time.Time = time.Now()

func m() *metricBuilder {
	return &metricBuilder{
		measurement: "a",
		tags:        make(map[string]string),
		fields:      make(map[string]interface{}),
		timestamp:   start,
	}
}

type metricBuilder struct {
	measurement string
	tags        map[string]string
	fields      map[string]interface{}
	timestamp   time.Time
}

func (m *metricBuilder) tag(tagKey, tagValue string) *metricBuilder {
	m.tags[tagKey] = tagValue
	return m
}

func (m *metricBuilder) field(fieldKey string, fieldValue interface{}) *metricBuilder {
	m.fields[fieldKey] = fieldValue
	return m
}

func (m *metricBuilder) build() telegraf.Metric {
	return metric.New(m.measurement, m.tags, m.fields, m.timestamp)
}

func TestEncodingMetrics(t *testing.T) {
	tests := []struct {
		name      string
		metrics   []telegraf.Metric
		encoded   []telegraf.Metric
		configure func(e *Encoding)
	}{
		{
			name: "compresses by compression",
			metrics: []telegraf.Metric{
				m().field("sample", "test value").build(),
			},
			encoded: []telegraf.Metric{
				m().field("sample", "H4sIAAAAAAAA/ypJLS5RKEvMKU0FBAAA//8rQd3sCgAAAA").build(),
			},
		},
		{
			name: "skips compression if missing",
			configure: func(e *Encoding) {
				e.Compression = ""
			},
			metrics: []telegraf.Metric{
				m().field("sample", "test value").build(),
			},
			encoded: []telegraf.Metric{
				m().field("sample", "dGVzdCB2YWx1ZQ").build(),
			},
		},
		{
			name: "compresses by field",
			configure: func(e *Encoding) {
				e.Compression = ""
				e.CompressionField = "compression"
			},
			metrics: []telegraf.Metric{
				m().field("sample", "test value").field("compression", "gzip").build(),
			},
			encoded: []telegraf.Metric{
				m().
					field("compression", "gzip").
					field("sample", "H4sIAAAAAAAA/ypJLS5RKEvMKU0FBAAA//8rQd3sCgAAAA").
					build(),
			},
		},
		{
			name: "compresses by tag",
			configure: func(e *Encoding) {
				e.Compression = ""
				e.CompressionTag = "compression"
			},
			metrics: []telegraf.Metric{
				m().field("sample", "test value").tag("compression", "gzip").build(),
			},
			encoded: []telegraf.Metric{
				m().
					tag("compression", "gzip").
					field("sample", "H4sIAAAAAAAA/ypJLS5RKEvMKU0FBAAA//8rQd3sCgAAAA").
					build(),
			},
		},
		{
			name: "compresses by tag over field",
			configure: func(e *Encoding) {
				e.Compression = ""
				e.CompressionTag = "compression"
				e.CompressionField = "compression"
			},
			metrics: []telegraf.Metric{
				m().
					tag("compression", "").
					field("compression", "gzip").
					field("sample", "test value").build(),
			},
			encoded: []telegraf.Metric{
				m().
					tag("compression", "").
					field("compression", "gzip").
					field("sample", "dGVzdCB2YWx1ZQ").build(),
			},
		},
		{
			name: "compresses by field over default",
			configure: func(e *Encoding) {
				e.Compression = "gzip"
				e.CompressionField = "compression"
			},
			metrics: []telegraf.Metric{
				m().
					field("compression", "").
					field("sample", "test value").build(),
			},
			encoded: []telegraf.Metric{
				m().
					field("compression", "").
					field("sample", "dGVzdCB2YWx1ZQ").build(),
			},
		},
		{
			name: "removes original",
			configure: func(e *Encoding) {
				e.DestField = "target"
				e.RemoveOriginal = true
				e.Compression = ""
			},
			metrics: []telegraf.Metric{
				m().field("sample", "test value").build(),
			},
			encoded: []telegraf.Metric{
				m().field("target", "dGVzdCB2YWx1ZQ").build(),
			},
		},
		{
			name: "leaves original",
			configure: func(e *Encoding) {
				e.DestField = "target"
				e.RemoveOriginal = false
				e.Compression = ""
			},
			metrics: []telegraf.Metric{
				m().field("sample", "test value").build(),
			},
			encoded: []telegraf.Metric{
				m().field("sample", "test value").field("target", "dGVzdCB2YWx1ZQ").build(),
			},
		},
		{
			name: "leaves metric with non-string field",
			metrics: []telegraf.Metric{
				m().field("sample", 5).build(),
			},
			encoded: []telegraf.Metric{
				m().field("sample", 5).build(),
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := &Encoding{
				Field:       "sample",
				Operation:   "encode",
				Compression: "gzip",
				Encoding:    "base64",
				Log:         testutil.Logger{},
			}

			if testCase.configure != nil {
				testCase.configure(e)
			}

			assert.NoError(t, e.Init())
			results := e.Apply(testCase.metrics...)

			testutil.RequireMetricsEqual(t, testCase.encoded, results)
		})
	}
}

func TestDecodingMetrics(t *testing.T) {
	tests := []struct {
		name      string
		metrics   []telegraf.Metric
		decoded   []telegraf.Metric
		configure func(e *Encoding)
	}{
		{
			name: "decompressed by compression",
			metrics: []telegraf.Metric{
				m().field("sample", "H4sIAAAAAAAA/ypJLS5RKEvMKU0FBAAA//8rQd3sCgAAAA").build(),
			},
			decoded: []telegraf.Metric{
				m().field("sample", "test value").build(),
			},
		},
		{
			name: "skips decompression if missing",
			configure: func(e *Encoding) {
				e.Compression = ""
			},
			metrics: []telegraf.Metric{
				m().field("sample", "dGVzdCB2YWx1ZQ").build(),
			},
			decoded: []telegraf.Metric{
				m().field("sample", "test value").build(),
			},
		},
		{
			name: "decompresses by field",
			configure: func(e *Encoding) {
				e.Compression = ""
				e.CompressionField = "compression"
			},
			metrics: []telegraf.Metric{
				m().
					field("sample", "H4sIAAAAAAAA/ypJLS5RKEvMKU0FBAAA//8rQd3sCgAAAA").
					field("compression", "gzip").build(),
			},
			decoded: []telegraf.Metric{
				m().field("compression", "gzip").field("sample", "test value").build(),
			},
		},
		{
			name: "decompresses by tag",
			configure: func(e *Encoding) {
				e.Compression = ""
				e.CompressionTag = "compression"
			},
			metrics: []telegraf.Metric{
				m().
					tag("compression", "gzip").
					field("sample", "H4sIAAAAAAAA/ypJLS5RKEvMKU0FBAAA//8rQd3sCgAAAA").
					build(),
			},
			decoded: []telegraf.Metric{
				m().tag("compression", "gzip").field("sample", "test value").build(),
			},
		},
		{
			name: "decompresses by tag over field",
			configure: func(e *Encoding) {
				e.Compression = ""
				e.CompressionTag = "compression"
				e.CompressionField = "compression"
			},
			metrics: []telegraf.Metric{
				m().
					tag("compression", "").
					field("compression", "gzip").
					field("sample", "dGVzdCB2YWx1ZQ").build(),
			},
			decoded: []telegraf.Metric{
				m().
					tag("compression", "").
					field("compression", "gzip").
					field("sample", "test value").build(),
			},
		},
		{
			name: "decompresses by field over default",
			configure: func(e *Encoding) {
				e.Compression = "gzip"
				e.CompressionField = "compression"
			},
			metrics: []telegraf.Metric{
				m().
					field("compression", "").
					field("sample", "dGVzdCB2YWx1ZQ").build(),
			},
			decoded: []telegraf.Metric{
				m().
					field("compression", "").
					field("sample", "test value").build(),
			},
		},
		{
			name: "removes original",
			configure: func(e *Encoding) {
				e.DestField = "target"
				e.RemoveOriginal = true
				e.Compression = ""
			},
			metrics: []telegraf.Metric{
				m().field("sample", "dGVzdCB2YWx1ZQ").build(),
			},
			decoded: []telegraf.Metric{
				m().field("target", "test value").build(),
			},
		},
		{
			name: "leaves original",
			configure: func(e *Encoding) {
				e.DestField = "target"
				e.RemoveOriginal = false
				e.Compression = ""
			},
			metrics: []telegraf.Metric{
				m().field("sample", "dGVzdCB2YWx1ZQ").build(),
			},
			decoded: []telegraf.Metric{
				m().field("sample", "dGVzdCB2YWx1ZQ").field("target", "test value").build(),
			},
		},
		{
			name: "leaves failed decode",
			metrics: []telegraf.Metric{
				m().field("sample", "test value").build(),
			},
			decoded: []telegraf.Metric{
				m().field("sample", "test value").build(),
			},
		},
		{
			name: "leaves failed decompression",
			metrics: []telegraf.Metric{
				m().field("sample", "dGVzdCB2YWx1ZQ").build(),
			},
			decoded: []telegraf.Metric{
				m().field("sample", "dGVzdCB2YWx1ZQ").build(),
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := &Encoding{
				Field:       "sample",
				Operation:   "decode",
				Compression: "gzip",
				Encoding:    "base64",
				Log:         testutil.Logger{},
			}

			if testCase.configure != nil {
				testCase.configure(e)
			}

			assert.NoError(t, e.Init())
			results := e.Apply(testCase.metrics...)

			testutil.RequireMetricsEqual(t, testCase.decoded, results)
		})
	}
}

package logfmt

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		measurement string
		bytes       []byte
		want        []telegraf.Metric
		wantErr     bool
	}{
		{
			name: "no bytes returns no metrics",
			want: []telegraf.Metric{},
		},
		{
			name:        "test without trailing end",
			bytes:       []byte("foo=\"bar\""),
			measurement: "testlog",
			want: []telegraf.Metric{
				testutil.MustMetric(
					"testlog",
					map[string]string{},
					map[string]interface{}{
						"foo": "bar",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:        "test with trailing end",
			bytes:       []byte("foo=\"bar\"\n"),
			measurement: "testlog",
			want: []telegraf.Metric{
				testutil.MustMetric(
					"testlog",
					map[string]string{},
					map[string]interface{}{
						"foo": "bar",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:        "logfmt parser returns all the fields",
			bytes:       []byte(`ts=2018-07-24T19:43:40.275Z lvl=info msg="http request" method=POST`),
			measurement: "testlog",
			want: []telegraf.Metric{
				testutil.MustMetric(
					"testlog",
					map[string]string{},
					map[string]interface{}{
						"lvl":    "info",
						"msg":    "http request",
						"method": "POST",
						"ts":     "2018-07-24T19:43:40.275Z",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:        "logfmt parser parses every line",
			bytes:       []byte("ts=2018-07-24T19:43:40.275Z lvl=info msg=\"http request\" method=POST\nparent_id=088876RL000 duration=7.45 log_id=09R4e4Rl000"),
			measurement: "testlog",
			want: []telegraf.Metric{
				testutil.MustMetric(
					"testlog",
					map[string]string{},
					map[string]interface{}{
						"lvl":    "info",
						"msg":    "http request",
						"method": "POST",
						"ts":     "2018-07-24T19:43:40.275Z",
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"testlog",
					map[string]string{},
					map[string]interface{}{
						"parent_id": "088876RL000",
						"duration":  7.45,
						"log_id":    "09R4e4Rl000",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:    "keys without = or values are ignored",
			bytes:   []byte(`i am no data.`),
			want:    []telegraf.Metric{},
			wantErr: false,
		},
		{
			name:    "keys without values are ignored",
			bytes:   []byte(`foo="" bar=`),
			want:    []telegraf.Metric{},
			wantErr: false,
		},
		{
			name:        "unterminated quote produces error",
			measurement: "testlog",
			bytes:       []byte(`bar=baz foo="bar`),
			want:        []telegraf.Metric{},
			wantErr:     true,
		},
		{
			name:        "malformed key",
			measurement: "testlog",
			bytes:       []byte(`"foo=" bar=baz`),
			want:        []telegraf.Metric{},
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Parser{
				MetricName: tt.measurement,
			}
			got, err := l.Parse(tt.bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Logfmt.Parse error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			testutil.RequireMetricsEqual(t, tt.want, got, testutil.IgnoreTime())
		})
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		name        string
		s           string
		measurement string
		want        telegraf.Metric
		wantErr     bool
	}{
		{
			name:    "No Metric In line",
			want:    nil,
			wantErr: true,
		},
		{
			name:        "Log parser fmt returns all fields",
			measurement: "testlog",
			s:           `ts=2018-07-24T19:43:35.207268Z lvl=5 msg="Write failed" log_id=09R4e4Rl000`,
			want: testutil.MustMetric(
				"testlog",
				map[string]string{},
				map[string]interface{}{
					"ts":     "2018-07-24T19:43:35.207268Z",
					"lvl":    int64(5),
					"msg":    "Write failed",
					"log_id": "09R4e4Rl000",
				},
				time.Unix(0, 0),
			),
		},
		{
			name:        "ParseLine only returns metrics from first string",
			measurement: "testlog",
			s:           "ts=2018-07-24T19:43:35.207268Z lvl=5 msg=\"Write failed\" log_id=09R4e4Rl000\nmethod=POST parent_id=088876RL000 duration=7.45 log_id=09R4e4Rl000",
			want: testutil.MustMetric(
				"testlog",
				map[string]string{},
				map[string]interface{}{
					"ts":     "2018-07-24T19:43:35.207268Z",
					"lvl":    int64(5),
					"msg":    "Write failed",
					"log_id": "09R4e4Rl000",
				},
				time.Unix(0, 0),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Parser{
				MetricName: tt.measurement,
			}
			got, err := l.ParseLine(tt.s)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Logfmt.Parse error = %v, wantErr %v", err, tt.wantErr)
			}
			testutil.RequireMetricEqual(t, tt.want, got, testutil.IgnoreTime())
		})
	}
}

func TestTags(t *testing.T) {
	tests := []struct {
		name        string
		measurement string
		tagKeys     []string
		s           string
		want        telegraf.Metric
		wantErr     bool
	}{
		{
			name:        "logfmt parser returns tags and fields",
			measurement: "testlog",
			tagKeys:     []string{"lvl"},
			s:           "ts=2018-07-24T19:43:40.275Z lvl=info msg=\"http request\" method=POST",
			want: testutil.MustMetric(
				"testlog",
				map[string]string{
					"lvl": "info",
				},
				map[string]interface{}{
					"msg":    "http request",
					"method": "POST",
					"ts":     "2018-07-24T19:43:40.275Z",
				},
				time.Unix(0, 0),
			),
		},
		{
			name:        "logfmt parser returns no empty metrics",
			measurement: "testlog",
			tagKeys:     []string{"lvl"},
			s:           "lvl=info",
			want: testutil.MustMetric(
				"testlog",
				map[string]string{
					"lvl": "info",
				},
				map[string]interface{}{},
				time.Unix(0, 0),
			),
		},
		{
			name:        "logfmt parser returns all keys as tag",
			measurement: "testlog",
			tagKeys:     []string{"*"},
			s:           "ts=2018-07-24T19:43:40.275Z lvl=info msg=\"http request\" method=POST",
			want: testutil.MustMetric(
				"testlog",
				map[string]string{
					"lvl":    "info",
					"msg":    "http request",
					"method": "POST",
					"ts":     "2018-07-24T19:43:40.275Z",
				},
				map[string]interface{}{},
				time.Unix(0, 0),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewParser(tt.measurement, map[string]string{}, tt.tagKeys)
			assert.NoError(t, l.Init())

			got, err := l.ParseLine(tt.s)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			testutil.RequireMetricEqual(t, tt.want, got, testutil.IgnoreTime())
		})
	}
}

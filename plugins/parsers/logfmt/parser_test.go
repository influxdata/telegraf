package logfmt

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func MustMetric(t *testing.T, m *testutil.Metric) telegraf.Metric {
	t.Helper()
	v, err := metric.New(m.Measurement, m.Tags, m.Fields, m.Time)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		measurement string
		now         func() time.Time
		bytes       []byte
		want        []testutil.Metric
		wantErr     bool
	}{
		{
			name: "no bytes returns no metrics",
			now:  func() time.Time { return time.Unix(0, 0) },
			want: []testutil.Metric{},
		},
		{
			name:        "test without trailing end",
			bytes:       []byte("foo=\"bar\""),
			now:         func() time.Time { return time.Unix(0, 0) },
			measurement: "testlog",
			want: []testutil.Metric{
				testutil.Metric{
					Measurement: "testlog",
					Tags:        map[string]string{},
					Fields: map[string]interface{}{
						"foo": "bar",
					},
					Time: time.Unix(0, 0),
				},
			},
		},
		{
			name:        "test with trailing end",
			bytes:       []byte("foo=\"bar\"\n"),
			now:         func() time.Time { return time.Unix(0, 0) },
			measurement: "testlog",
			want: []testutil.Metric{
				testutil.Metric{
					Measurement: "testlog",
					Tags:        map[string]string{},
					Fields: map[string]interface{}{
						"foo": "bar",
					},
					Time: time.Unix(0, 0),
				},
			},
		},
		{
			name:        "logfmt parser returns all the fields",
			bytes:       []byte(`ts=2018-07-24T19:43:40.275Z lvl=info msg="http request" method=POST`),
			now:         func() time.Time { return time.Unix(0, 0) },
			measurement: "testlog",
			want: []testutil.Metric{
				testutil.Metric{
					Measurement: "testlog",
					Tags:        map[string]string{},
					Fields: map[string]interface{}{
						"lvl":    "info",
						"msg":    "http request",
						"method": "POST",
						"ts":     "2018-07-24T19:43:40.275Z",
					},
					Time: time.Unix(0, 0),
				},
			},
		},
		{
			name:        "logfmt parser parses every line",
			bytes:       []byte("ts=2018-07-24T19:43:40.275Z lvl=info msg=\"http request\" method=POST\nparent_id=088876RL000 duration=7.45 log_id=09R4e4Rl000"),
			now:         func() time.Time { return time.Unix(0, 0) },
			measurement: "testlog",
			want: []testutil.Metric{
				testutil.Metric{
					Measurement: "testlog",
					Tags:        map[string]string{},
					Fields: map[string]interface{}{
						"lvl":    "info",
						"msg":    "http request",
						"method": "POST",
						"ts":     "2018-07-24T19:43:40.275Z",
					},
					Time: time.Unix(0, 0),
				},
				testutil.Metric{
					Measurement: "testlog",
					Tags:        map[string]string{},
					Fields: map[string]interface{}{
						"parent_id": "088876RL000",
						"duration":  7.45,
						"log_id":    "09R4e4Rl000",
					},
					Time: time.Unix(0, 0),
				},
			},
		},
		{
			name:    "keys without = or values are ignored",
			now:     func() time.Time { return time.Unix(0, 0) },
			bytes:   []byte(`i am no data.`),
			want:    []testutil.Metric{},
			wantErr: false,
		},
		{
			name:    "keys without values are ignored",
			now:     func() time.Time { return time.Unix(0, 0) },
			bytes:   []byte(`foo="" bar=`),
			want:    []testutil.Metric{},
			wantErr: false,
		},
		{
			name:        "unterminated quote produces error",
			now:         func() time.Time { return time.Unix(0, 0) },
			measurement: "testlog",
			bytes:       []byte(`bar=baz foo="bar`),
			want:        []testutil.Metric{},
			wantErr:     true,
		},
		{
			name:        "malformed key",
			now:         func() time.Time { return time.Unix(0, 0) },
			measurement: "testlog",
			bytes:       []byte(`"foo=" bar=baz`),
			want:        []testutil.Metric{},
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Parser{
				MetricName: tt.measurement,
				Now:        tt.now,
			}
			got, err := l.Parse(tt.bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Logfmt.Parse error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, len(tt.want), len(got))
			for i, m := range got {
				testutil.MustEqual(t, m, tt.want[i])
			}
		})
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		name        string
		s           string
		measurement string
		now         func() time.Time
		want        testutil.Metric
		wantErr     bool
	}{
		{
			name:    "No Metric In line",
			now:     func() time.Time { return time.Unix(0, 0) },
			want:    testutil.Metric{},
			wantErr: true,
		},
		{
			name:        "Log parser fmt returns all fields",
			now:         func() time.Time { return time.Unix(0, 0) },
			measurement: "testlog",
			s:           `ts=2018-07-24T19:43:35.207268Z lvl=5 msg="Write failed" log_id=09R4e4Rl000`,
			want: testutil.Metric{
				Measurement: "testlog",
				Fields: map[string]interface{}{
					"ts":     "2018-07-24T19:43:35.207268Z",
					"lvl":    int64(5),
					"msg":    "Write failed",
					"log_id": "09R4e4Rl000",
				},
				Tags: map[string]string{},
				Time: time.Unix(0, 0),
			},
		},
		{
			name:        "ParseLine only returns metrics from first string",
			now:         func() time.Time { return time.Unix(0, 0) },
			measurement: "testlog",
			s:           "ts=2018-07-24T19:43:35.207268Z lvl=5 msg=\"Write failed\" log_id=09R4e4Rl000\nmethod=POST parent_id=088876RL000 duration=7.45 log_id=09R4e4Rl000",
			want: testutil.Metric{
				Measurement: "testlog",
				Fields: map[string]interface{}{
					"ts":     "2018-07-24T19:43:35.207268Z",
					"lvl":    int64(5),
					"msg":    "Write failed",
					"log_id": "09R4e4Rl000",
				},
				Tags: map[string]string{},
				Time: time.Unix(0, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Parser{
				MetricName: tt.measurement,
				Now:        tt.now,
			}
			got, err := l.ParseLine(tt.s)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Logfmt.Parse error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != nil {
				testutil.MustEqual(t, got, tt.want)
			}
		})
	}
}

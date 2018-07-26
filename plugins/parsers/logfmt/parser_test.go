package logfmt

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
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
			name:    "poorly formatted logfmt returns error",
			now:     func() time.Time { return time.Unix(0, 0) },
			bytes:   []byte(`i am garbage data.`),
			want:    []testutil.Metric{},
			wantErr: true,
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
			name: "test something",
			now:  func() time.Time { return time.Unix(0, 0) },
			want: testutil.Metric{
				Tags:   map[string]string{},
				Fields: map[string]interface{}{},
				Time:   time.Unix(0, 0),
			},
		},
		{
			name:        "log parser fmt returns all fields",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Parser{
				MetricName: tt.measurement,
				Now:        tt.now,
			}
			got, err := l.ParseLine(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Logfmt.Parse error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			testutil.MustEqual(t, got, tt.want)
		})
	}
}

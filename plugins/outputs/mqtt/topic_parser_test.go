package mqtt

import (
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_parse(t *testing.T) {
	s := serializers.NewInfluxSerializer()
	m := &MQTT{
		Servers:    []string{"tcp://localhost:502"},
		serializer: s,
		KeepAlive:  30,
		Log:        testutil.Logger{},
	}
	tests := []struct {
		name    string
		pattern string
		want    string
	}{
		{
			name:    "matches default legacy format",
			pattern: "<topic_prefix>/<hostname>/<pluginname>",
			want:    "prefix/hostname/metric-name",
		},
		{
			name:    "respect hardcoded strings",
			pattern: "this/is/a/topic",
			want:    "this/is/a/topic",
		},
		{
			name:    "allows the use of tags",
			pattern: "<topic_prefix>/<tag::tag1>",
			want:    "prefix/value1",
		},
		{
			name:    "uses the plugin name when no pattern is provided",
			pattern: "",
			want:    "metric-name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.Topic = tt.pattern
			m.TopicPrefix = "prefix"
			met := metric.New(
				"metric-name",
				map[string]string{"tag1": "value1"},
				map[string]interface{}{"value": 123},
				time.Date(2022, time.November, 10, 23, 0, 0, 0, time.UTC),
			)
			err := m.Init()
			require.NoError(t, err)
			if got := parse(m, met, "hostname"); got != tt.want {
				t.Errorf("parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

package mqtt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
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
			pattern: "{{ .TopicPrefix }}/{{ .Hostname }}/{{ .PluginName }}",
			want:    "prefix/hostname/metric-name",
		},
		{
			name:    "respect hardcoded strings",
			pattern: "this/is/a/topic",
			want:    "this/is/a/topic",
		},
		{
			name:    "allows the use of tags",
			pattern: "{{ .TopicPrefix }}/{{ .Tag \"tag1\" }}",
			want:    "prefix/value1",
		},
		{
			name:    "uses the plugin name when no pattern is provided",
			pattern: "",
			want:    "metric-name",
		},
		{
			name:    "ignores tag when tag does not exists",
			pattern: "{{ .TopicPrefix }}/{{ .Tag \"not-a-tag\" }}",
			want:    "prefix",
		},
		{
			name:    "ignores empty forward slashes",
			pattern: "double//slashes//are//ignored",
			want:    "double/slashes/are/ignored",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.Topic = tt.pattern
			tp := "prefix"
			m.TopicPrefix = tp
			met := metric.New(
				"metric-name",
				map[string]string{"tag1": "value1"},
				map[string]interface{}{"value": 123},
				time.Date(2022, time.November, 10, 23, 0, 0, 0, time.UTC),
			)
			err := m.Init()
			require.NoError(t, err)
			topic := &TemplateTopic{Hostname: "hostname", metric: met}
			if got := topic.Parse(m); got != tt.want {
				t.Errorf("parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

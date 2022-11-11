package mqtt

import (
	"github.com/influxdata/telegraf"
	"strings"
)

func parse(m *MQTT, metric telegraf.Metric, hostname string) string {
	var t []string
	if m.Topic != "" {
		for _, p := range strings.Split(m.Topic, "/") {
			switch {
			case p == "*topic_prefix*":
				t = append(t, m.TopicPrefix)
			case p == "*hostname*":
				if hostname != "" {
					t = append(t, hostname)
				}
			case p == "*pluginname*":
				t = append(t, metric.Name())
			case strings.Contains(p, "*tag::"):
				k := strings.TrimSuffix(strings.TrimPrefix(p, "*tag::"), "*")
				var tag string
				tag, ok := metric.GetTag(k)
				if ok {
					t = append(t, tag)
				}
			default:
				t = append(t, p)
			}
		}
	}
	return strings.Join(t, "/")
}

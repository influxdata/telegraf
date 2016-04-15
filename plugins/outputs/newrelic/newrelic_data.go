package newrelic

import(
  "time"
  "github.com/influxdata/telegraf"
)

type NewRelicData struct {
  LastWrite time.Time
  Hosts map[string][]NewRelicComponent
  GuidBase string
}

func (nrd *NewRelicData) AddMetric(metric telegraf.Metric) {
	component := NewRelicComponent{
    Duration: int(time.Since(nrd.LastWrite).Seconds()),
    TMetric: metric,
    GuidBase: nrd.GuidBase}
  host      := component.Hostname()
	nrd.Hosts[host] = append(nrd.Hosts[host],component)
}

func (nrd *NewRelicData) AddMetrics(metrics []telegraf.Metric) {
	for _, metric := range(metrics) {
		nrd.AddMetric(metric)
	}
}

func (nrd *NewRelicData) DataSets() []interface{} {
  result := make([]interface{}, 0)
  for host, components := range(nrd.Hosts) {
		result = append(result, map[string]interface{} { "agent": map[string]string { "host": host, "version": "0.0.1" }, "components": components })
	}
	return result
}

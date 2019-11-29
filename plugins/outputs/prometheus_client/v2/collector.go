package v2

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	serializer "github.com/influxdata/telegraf/plugins/serializers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type Metric struct {
	family *dto.MetricFamily
	metric *dto.Metric
}

func (m *Metric) Desc() *prometheus.Desc {
	labelNames := make([]string, 0, len(m.metric.Label))
	for _, label := range m.metric.Label {
		labelNames = append(labelNames, *label.Name)
	}

	desc := prometheus.NewDesc(*m.family.Name, *m.family.Help, labelNames, nil)

	return desc
}

func (m *Metric) Write(out *dto.Metric) error {
	out.Label = m.metric.Label
	out.Counter = m.metric.Counter
	out.Untyped = m.metric.Untyped
	out.Gauge = m.metric.Gauge
	out.Histogram = m.metric.Histogram
	out.Summary = m.metric.Summary
	out.TimestampMs = m.metric.TimestampMs
	return nil
}

type Collector struct {
	sync.Mutex
	expireDuration time.Duration
	coll           *serializer.Collection
}

func NewCollector(expire time.Duration, stringsAsLabel bool) *Collector {
	config := serializer.FormatConfig{}
	if stringsAsLabel {
		config.StringHandling = serializer.StringAsLabel
	}
	return &Collector{
		expireDuration: expire,
		coll:           serializer.NewCollection(config),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	// Sending no descriptor at all marks the Collector as "unchecked",
	// i.e. no checks will be performed at registration time, and the
	// Collector may yield any Metric it sees fit in its Collect method.
	return
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.Lock()
	defer c.Unlock()

	for _, family := range c.coll.GetProto() {
		for _, metric := range family.Metric {
			ch <- &Metric{family: family, metric: metric}
		}
	}
}

func (c *Collector) Add(metrics []telegraf.Metric) error {
	c.Lock()
	defer c.Unlock()

	for _, metric := range metrics {
		c.coll.Add(metric)
	}

	if c.expireDuration != 0 {
		c.coll.Expire(time.Now(), c.expireDuration)
	}
	return nil
}

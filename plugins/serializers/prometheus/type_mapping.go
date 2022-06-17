package prometheus

import (
	"strings"

	"github.com/influxdata/telegraf"
)

type PrometheusMetricType string

const (
	Gauge   PrometheusMetricType = "gauge"
	Counter PrometheusMetricType = "counter"
)

type TypeMapping struct {
	Suffixes []string             `toml:"suffixes"`
	Type     PrometheusMetricType `toml:"type"`
}

func (t *TypeMapping) anySuffixMatches(name string) bool {
	for _, s := range t.Suffixes {
		if strings.HasSuffix(name, s) {
			return true
		}
	}

	return false
}

func (t *TypeMapping) telegrafValueType() telegraf.ValueType {
	switch t.Type {
	case Gauge:
		return telegraf.Gauge
	case Counter:
		return telegraf.Counter
	default:
		return telegraf.Untyped
	}
}

func (t TypeMapping) InferValueType(prometheusMetricName string) (telegraf.ValueType, bool) {
	if t.anySuffixMatches(prometheusMetricName) {
		return t.telegrafValueType(), true
	}

	return telegraf.Untyped, false
}

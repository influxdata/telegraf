package prometheus

import (
	"fmt"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Parser struct {
	IgnoreTimestamp bool              `toml:"prometheus_ignore_timestamp"`
	MetricVersion   int               `toml:"prometheus_metric_version"`
	Header          http.Header       `toml:"-"` // set by the prometheus input
	DefaultTags     map[string]string `toml:"-"`
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	switch p.MetricVersion {
	case 0, 2:
		return p.parse_v2(buf)
	case 1:
		return p.parse_v1(buf)
	}
	return nil, fmt.Errorf("unknown prometheus metric version %d", p.MetricVersion)
}

func init() {
	parsers.Add("prometheus",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{}
		},
	)
}

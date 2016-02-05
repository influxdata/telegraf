package influx

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/encoding"
)

type InfluxParser struct {
}

func (p *InfluxParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	metrics, err := telegraf.ParseMetrics(buf)

	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (p *InfluxParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line + "\n"))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("Can not parse the line: %s, for data format: influx ", line)
	}

	return metrics[0], nil
}

func NewParser() *InfluxParser {
	return &InfluxParser{}
}

func (p *InfluxParser) InitConfig(configs map[string]interface{}) error {
	return nil
}

func init() {
	encoding.Add("influx", func() encoding.Parser {
		return NewParser()
	})
}

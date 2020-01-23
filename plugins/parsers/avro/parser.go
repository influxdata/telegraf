package avro

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type Parser struct {
	DefaultTags       map[string]string
	TimeFunc          func() time.Time
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	return p.createMeasures()
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	return p.createMeasure()
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) createMeasure() (telegraf.Metric, error) {
	recordFields := make(map[string]interface{})
	tags := make(map[string]string)

	for k, v := range p.DefaultTags {
		tags[k] = v
	}

	tags["tagName"] = "tagValue"

	recordFields["fieldName"] = "fieldValue"

	measurementName := "measurementName"

	metricTime := p.TimeFunc()

	m, err := metric.New(measurementName, tags, recordFields, metricTime)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (p *Parser) createMeasures() ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)

	m, err := p.createMeasure()
	if err != nil {
		return metrics, err
	}
	metrics = append(metrics, m)
	
	return metrics, nil
}
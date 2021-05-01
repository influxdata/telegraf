package jsonpath

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

type TimeFunc func() time.Time

type Parser struct {
	Configs     []Config
	DefaultTags map[string]string
	Log         telegraf.Logger
	TimeFunc    func() time.Time
}

type Config struct {
	MetricSelection string `toml:"metric_selection"`
	MetricName      string `toml:"metric_name"`
	Fields          FieldKeys
}

type FieldKeys struct {
	FieldName string `toml:"fieldname"`
	Query     string `toml:"query"`
	FieldType string `toml:"type"`
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	return []telegraf.Metric{}, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {

	err := oj.ValidateString(line)
	if err != nil {
		return nil, fmt.Errorf("The provided JSON is invalid: %v", err)
	}

	obj, err := oj.ParseString(line)
	if err != nil {
		return nil, err
	}

	x, err := jp.ParseString(p.Configs[0].MetricSelection)
	if err != nil {
		return nil, err
	}
	result := x.Get(obj)
	fmt.Println(oj.JSON(result))

	metricname := p.Configs[0].MetricName
	tags := map[string]string{}
	fields := map[string]interface{}{}
	fields[p.Configs[0].Fields.FieldName] = result[0]

	if p.TimeFunc == nil {
		p.TimeFunc = time.Now
	}

	return metric.New(metricname, tags, fields, p.TimeFunc()), nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {

}

func (p *Parser) SetTimeFunc(fn TimeFunc) {
	p.TimeFunc = fn
}

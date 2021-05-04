package jsonpath

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

// NOTE: To test changes quickly you can run the following
// 1. make go-install
// 2. telegraf --config ./plugins/parsers/jsonpath/testdata/simple/simple.conf --debug

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

	err := oj.Validate(buf)
	if err != nil {
		return nil, fmt.Errorf("The provided JSON is invalid: %v", err)
	}

	obj, err := oj.Parse(buf)
	if err != nil {
		return nil, err
	}

	var t []telegraf.Metric
	m, err := p.query(obj)
	if err != nil {
		return nil, err
	}
	t = append(t, m)

	return t, nil
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

	return p.query(obj)
}

func (p *Parser) query(obj interface{}) (telegraf.Metric, error) {

	if len(p.Configs) < 0 {
		return nil, fmt.Errorf("No metric selection!")
	}
	fmt.Println(p.Configs[0].Fields.Query)
	x, err := jp.ParseString(p.Configs[0].Fields.Query)
	if err != nil {
		return nil, err
	}
	result := x.Get(obj)
	fmt.Println(oj.JSON(result))
	metricname := "lol"
	tags := map[string]string{}
	fields := map[string]interface{}{}
	if len(result) > 0 {
		fields[p.Configs[0].Fields.FieldName] = result[0]
	}

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

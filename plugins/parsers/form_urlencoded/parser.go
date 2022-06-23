package form_urlencoded

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
)

var ErrNoMetric = fmt.Errorf("no metric in line")

// Parser decodes "application/x-www-form-urlencoded" data into metrics
type Parser struct {
	MetricName  string            `toml:"metric_name"`
	TagKeys     []string          `toml:"form_urlencoded_tag_keys"`
	DefaultTags map[string]string `toml:"-"`
}

// Parse converts a slice of bytes in "application/x-www-form-urlencoded" format into metrics
func (p Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	buf = bytes.TrimSpace(buf)
	if len(buf) == 0 {
		return make([]telegraf.Metric, 0), nil
	}

	values, err := url.ParseQuery(string(buf))
	if err != nil {
		return nil, err
	}

	tags := p.extractTags(values)
	fields := p.parseFields(values)

	for key, value := range p.DefaultTags {
		tags[key] = value
	}

	m := metric.New(p.MetricName, tags, fields, time.Now().UTC())

	return []telegraf.Metric{m}, nil
}

// ParseLine delegates a single line of text to the Parse function
func (p Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, ErrNoMetric
	}

	return metrics[0], nil
}

// SetDefaultTags sets the default tags for every metric
func (p Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p Parser) extractTags(values url.Values) map[string]string {
	tags := make(map[string]string)
	for _, key := range p.TagKeys {
		value, exists := values[key]

		if !exists || len(key) == 0 {
			continue
		}

		tags[key] = value[0]
		delete(values, key)
	}

	return tags
}

func (p Parser) parseFields(values url.Values) map[string]interface{} {
	fields := make(map[string]interface{})

	for key, value := range values {
		if len(key) == 0 || len(value) == 0 {
			continue
		}

		field, err := strconv.ParseFloat(value[0], 64)
		if err != nil {
			continue
		}

		fields[key] = field
	}

	return fields
}

func (p *Parser) Init() error {
	return nil
}

func (p *Parser) InitFromConfig(config *parsers.Config) error {
	p.MetricName = config.MetricName
	p.TagKeys = config.FormUrlencodedTagKeys
	return p.Init()
}

func init() {
	parsers.Add("form_urlencoded",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{MetricName: defaultMetricName}
		})
}

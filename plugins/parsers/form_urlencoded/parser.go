package form_urlencoded

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var (
	// ErrNoMetric is returned when no metric is found in input line
	ErrNoMetric = fmt.Errorf("no metric in line")
)

// Parser decodes "application/x-www-form-urlencoded" data into metrics
type Parser struct {
	MetricName  string
	DefaultTags map[string]string
	TagKeys     []string
	AllowedKeys []string
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

	if len(p.AllowedKeys) > 0 {
		values = p.filterAllowedKeys(values)
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
func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p Parser) filterAllowedKeys(original url.Values) url.Values {
	result := make(url.Values)

	for _, key := range p.AllowedKeys {
		value, exists := original[key]
		if !exists {
			continue
		}

		result[key] = value
	}

	return result
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

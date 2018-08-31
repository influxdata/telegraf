package logfmt

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/go-logfmt/logfmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var (
	ErrNoMetric = fmt.Errorf("no metric in line")
)

// Parser decodes logfmt formatted messages into metrics.
type Parser struct {
	MetricName  string
	DefaultTags map[string]string
	Now         func() time.Time
}

// NewParser creates a parser.
func NewParser(metricName string, defaultTags map[string]string) *Parser {
	return &Parser{
		MetricName:  metricName,
		DefaultTags: defaultTags,
		Now:         time.Now,
	}
}

// Parse converts a slice of bytes in logfmt format to metrics.
func (p *Parser) Parse(b []byte) ([]telegraf.Metric, error) {
	reader := bytes.NewReader(b)
	decoder := logfmt.NewDecoder(reader)
	metrics := make([]telegraf.Metric, 0)
	for {
		ok := decoder.ScanRecord()
		if !ok {
			err := decoder.Err()
			if err != nil {
				return nil, err
			}
			break
		}
		fields := make(map[string]interface{})
		for decoder.ScanKeyval() {
			if string(decoder.Value()) == "" {
				continue
			}

			//type conversions
			value := string(decoder.Value())
			if iValue, err := strconv.ParseInt(value, 10, 64); err == nil {
				fields[string(decoder.Key())] = iValue
			} else if fValue, err := strconv.ParseFloat(value, 64); err == nil {
				fields[string(decoder.Key())] = fValue
			} else if bValue, err := strconv.ParseBool(value); err == nil {
				fields[string(decoder.Key())] = bValue
			} else {
				fields[string(decoder.Key())] = value
			}
		}
		if len(fields) == 0 {
			continue
		}

		m, err := metric.New(p.MetricName, map[string]string{}, fields, p.Now())
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, m)
	}
	p.applyDefaultTags(metrics)
	return metrics, nil
}

// ParseLine converts a single line of text in logfmt format to metrics.
func (p *Parser) ParseLine(s string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(s))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, ErrNoMetric
	}
	return metrics[0], nil
}

// SetDefaultTags adds tags to the metrics outputs of Parse and ParseLine.
func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) applyDefaultTags(metrics []telegraf.Metric) {
	if len(p.DefaultTags) == 0 {
		return
	}

	for _, m := range metrics {
		for k, v := range p.DefaultTags {
			if !m.HasTag(k) {
				m.AddTag(k, v)
			}
		}
	}
}

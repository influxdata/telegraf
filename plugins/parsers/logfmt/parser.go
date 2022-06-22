package logfmt

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-logfmt/logfmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/metric"
)

var ErrNoMetric = errors.New("no metric in line")

// Parser decodes logfmt formatted messages into metrics.
type Parser struct {
	TagKeys     []string          `toml:"logfmt_tag_keys"`
	DefaultTags map[string]string `toml:"-"`

	metricName string
	tagFilter  filter.Filter
}

// NewParser creates a parser.
func NewParser(metricName string, defaultTags map[string]string, tagKeys []string) *Parser {
	return &Parser{
		metricName:  metricName,
		DefaultTags: defaultTags,
		TagKeys:     tagKeys,
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
		tags := make(map[string]string)
		for decoder.ScanKeyval() {
			if string(decoder.Value()) == "" {
				continue
			}

			//type conversions
			value := string(decoder.Value())
			if p.tagFilter != nil && p.tagFilter.Match(string(decoder.Key())) {
				tags[string(decoder.Key())] = value
			} else if iValue, err := strconv.ParseInt(value, 10, 64); err == nil {
				fields[string(decoder.Key())] = iValue
			} else if fValue, err := strconv.ParseFloat(value, 64); err == nil {
				fields[string(decoder.Key())] = fValue
			} else if bValue, err := strconv.ParseBool(value); err == nil {
				fields[string(decoder.Key())] = bValue
			} else {
				fields[string(decoder.Key())] = value
			}
		}
		if len(fields) == 0 && len(tags) == 0 {
			continue
		}

		m := metric.New(p.metricName, tags, fields, time.Now())

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

func (p *Parser) Init() error {
	var err error

	// Compile tag key patterns
	if p.tagFilter, err = filter.Compile(p.TagKeys); err != nil {
		return fmt.Errorf("error compiling tag pattern: %w", err)
	}

	return nil
}

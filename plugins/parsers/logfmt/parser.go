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
	"github.com/influxdata/telegraf/plugins/parsers"
)

var ErrNoMetric = errors.New("no metric in line")

// Parser decodes logfmt formatted messages into metrics.
type Parser struct {
	TagKeys     []string          `toml:"logfmt_tag_keys"`
	DefaultTags map[string]string `toml:"-"`

	metricName string
	timeFunc   func() time.Time
	tagFilter  filter.Filter
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) SetTimeFunc(f func() time.Time) {
	p.timeFunc = f
}

func (p *Parser) Init() error {
	// Compile tag key patterns
	f, err := filter.Compile(p.TagKeys)
	if err != nil {
		return fmt.Errorf("error compiling tag pattern: %w", err)
	}
	p.tagFilter = f

	if p.timeFunc == nil {
		p.timeFunc = time.Now
	}

	return nil
}

// Parse converts a slice of bytes in logfmt format to metrics.
func (p *Parser) Parse(b []byte) ([]telegraf.Metric, error) {
	reader := bytes.NewReader(b)
	decoder := logfmt.NewDecoder(reader)
	metrics := make([]telegraf.Metric, 0)
	for {
		ok := decoder.ScanRecord()
		if !ok {
			if err := decoder.Err(); err != nil {
				return nil, err
			}
			break
		}
		fields := make(map[string]interface{})
		tags := make(map[string]string)
		for decoder.ScanKeyval() {
			if len(decoder.Value()) == 0 {
				continue
			}

			// type conversions
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

		m := metric.New(p.metricName, tags, fields, p.timeFunc())

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

func init() {
	// Register parser
	parsers.Add("logfmt",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{metricName: defaultMetricName}
		},
	)
}

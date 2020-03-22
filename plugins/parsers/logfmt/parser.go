package logfmt

import (
	"log"
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
	TagKeys     []string
	DefaultTags map[string]string
	Now         func() time.Time
}

// NewParser creates a parser.
func NewParser(metricName string, tagKeys []string, defaultTags map[string]string) *Parser {
	return &Parser{
		MetricName:  metricName,
		TagKeys:     tagKeys,
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
		tags := make(map[string]string)
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

		//m, err := metric.New(p.MetricName, map[string]string{}, fields, p.Now())

		tags, nFields := p.switchFieldToTag(tags, fields)
		m, err := metric.New(p.MetricName, tags, nFields, p.Now())

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

//will take in field map with strings and bools,
//search for TagKeys that match fieldnames and add them to tags
//assumes that any non-numeric values in TagKeys should be displayed as tags
func (p *Parser) switchFieldToTag(tags map[string]string, fields map[string]interface{}) (map[string]string, map[string]interface{}) {
	for _, name := range p.TagKeys {
		//switch any fields in tagkeys into tags
		if fields[name] == nil {
			continue
		}
		switch value := fields[name].(type) {
		case string:
			tags[name] = value
			delete(fields, name)
		case bool:
			tags[name] = strconv.FormatBool(value)
			delete(fields, name)
		case int64:
			tags[name] = strconv.FormatInt(value, 10)
			delete(fields, name)
		case float64:
			tags[name] = strconv.FormatFloat(value, 'f', -1, 64)
			delete(fields, name)
		default:
			log.Printf("E! [parsers.logfmt] Unrecognized value type %T", value)
		}
	}
	return tags, fields
}

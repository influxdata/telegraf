// Package logfmt converts logfmt data into metrics.
package logfmt

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	glogfmt "github.com/go-logfmt/logfmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
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
func (l *Parser) Parse(b []byte) ([]telegraf.Metric, error) {
	reader := bytes.NewReader(b)
	decoder := glogfmt.NewDecoder(reader)
	metrics := make([]telegraf.Metric, 0)
	for decoder.ScanRecord() {
		tags := make(map[string]string)
		fields := make(map[string]interface{})
		//add default tags
		for k, v := range l.DefaultTags {
			tags[k] = v
		}

		for decoder.ScanKeyval() {
			if string(decoder.Value()) == "" {
				return metrics, fmt.Errorf("value could not be found for key: %v", string(decoder.Key()))
			}

			//attempt type conversions
			value := string(decoder.Value())
			if iValue, err := strconv.Atoi(value); err == nil {
				fields[string(decoder.Key())] = iValue
			} else if fValue, err := strconv.ParseFloat(value, 64); err == nil {
				fields[string(decoder.Key())] = fValue
			} else if bValue, err := strconv.ParseBool(value); err == nil {
				fields[string(decoder.Key())] = bValue
			} else {
				fields[string(decoder.Key())] = value
			}
		}
		m, err := metric.New(l.MetricName, tags, fields, l.Now())
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// ParseLine converts a single line of text in logfmt to metrics.
func (l *Parser) ParseLine(s string) (telegraf.Metric, error) {
	reader := strings.NewReader(s)
	decoder := glogfmt.NewDecoder(reader)

	decoder.ScanRecord()
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	//add default tags
	for k, v := range l.DefaultTags {
		tags[k] = v
	}

	for decoder.ScanKeyval() {
		if string(decoder.Value()) == "" {
			return nil, fmt.Errorf("value could not be found for key: %v", string(decoder.Key()))
		}
		//attempt type conversions
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
	m, err := metric.New(l.MetricName, tags, fields, l.Now())
	if err != nil {
		return nil, err
	}
	return m, nil
}

// SetDefaultTags adds tags to the metrics outputs of Parse and ParseLine.
func (l *Parser) SetDefaultTags(tags map[string]string) {
	l.DefaultTags = tags
}

// Package logfmt converts logfmt data into metrics.
package logfmt

import (
	"bytes"
	"fmt"
	"log"
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
			log.Printf("k: %v, v: %v", string(decoder.Key()), string(decoder.Value()))
			if string(decoder.Value()) == "" {
				return metrics, fmt.Errorf("value could not be found for key: %v", string(decoder.Key()))
			}
			fields[string(decoder.Key())] = string(decoder.Value())
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
		fields[string(decoder.Key())] = string(decoder.Value())
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

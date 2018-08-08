// Package logfmt converts logfmt data into metrics.  New comment
package logfmt

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"

	glogfmt "github.com/go-logfmt/logfmt"
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
	decoder := glogfmt.NewDecoder(reader)
	metrics := make([]telegraf.Metric, 0)
	for decoder.ScanRecord() {
		tags := make(map[string]string)
		fields := make(map[string]interface{})
		for decoder.ScanKeyval() {
			if string(decoder.Value()) == "" {
				return metrics, fmt.Errorf("value could not be found for key: %v", string(decoder.Key()))
			}

			//attempt type conversions
			value := string(decoder.Value())
			if iValue, err := strconv.Atoi(value); err == nil {
				//log.Printf("Print Atoi Value Here:", iValue)
				//log.Printf("DECODER =", decoder.Key())
				fields[string(decoder.Key())] = iValue
			} else if fValue, err := strconv.ParseFloat(value, 64); err == nil {
				log.Printf("key:%s, value:%s", decoder.Key(), value)
				//log.Printf("Print ParseFloat Value Here:", fValue)
				fields[string(decoder.Key())] = fValue
			} else if bValue, err := strconv.ParseBool(value); err == nil {
				//log.Printf("Print ParseBool Value Here:", bValue)
				fields[string(decoder.Key())] = bValue
			} else {
				log.Printf("key:%s, value:%s", decoder.Key(), value)
				//				log.Printf("Print Value Here:", value)
				fields[string(decoder.Key())] = value
				//log.Printf("DECODER =", decoder.Key())
			}
		}
		log.Printf("All fields: %s", fields)
		m, err := metric.New(p.MetricName, tags, fields, p.Now())
		//log.Printf("Return all the info in metric", p.MetricName, tags, fields)
		if err != nil {
			log.Println("Error occurred")
			return nil, err
		}

		//add default tags
		metrics = append(metrics, m)
		p.applyDefaultTags(metrics)
	}
	return metrics, nil
}

// ParseLine converts a single line of text in logfmt to metrics.
func (p *Parser) ParseLine(s string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(s))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		//if metrics[1] == nil {
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

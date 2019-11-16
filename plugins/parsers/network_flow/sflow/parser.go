package sflow

import (
	"bytes"
	"fmt"

	"github.com/influxdata/telegraf"

	"github.com/influxdata/telegraf/plugins/parsers/network_flow/decoder"
)

// Parser is Telegraf parser capable of parsing an sFlow v5 network packet
type Parser struct {
	metricName  string
	defaultTags map[string]string
	sflowFormat decoder.Directive
}

// NewParser creates a new SFlow Parser
func NewParser(metricName string, defaultTags map[string]string, sflowConfig V5FormatOptions) (*Parser, error) {
	if metricName == "" {
		return nil, fmt.Errorf("metric name cannot be empty")
	}
	result := &Parser{metricName: metricName, sflowFormat: V5Format(sflowConfig), defaultTags: defaultTags}
	return result, nil
}

// Parse takes a byte buffer separated by newlines
// ie, `cpu.usage.idle 90\ncpu.usage.busy 10`
// and parses it into telegraf metrics
//
// Must be thread-safe.
func (sfp *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	dc := decoder.NewDecodeContext(false)
	if err := dc.Decode(sfp.sflowFormat, bytes.NewBuffer(buf)); err != nil {
		return nil, err
	}
	return dc.GetMetrics(), nil
}

// ParseLine takes a single string metric
// ie, "cpu.usage.idle 90"
// and parses it into a telegraf metric.
//
// Must be thread-safe.
func (sfp *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := sfp.Parse([]byte(line))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("Can not parse the line: %s, for data format: value", line)
	}

	return metrics[0], nil
}

// SetDefaultTags tells the parser to add all of the given tags
// to each parsed metric.
// NOTE: do _not_ modify the map after you've passed it here!!
func (sfp *Parser) SetDefaultTags(tags map[string]string) {
	sfp.defaultTags = tags
}

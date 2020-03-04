package sflow

import (
	"bytes"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/sflow/parser/decoder"
)

// Parser is Telegraf parser capable of parsing an sFlow v5 network packet
type Parser struct {
	metricName  string
	sflowFormat decoder.Directive
}

// NewParser creates a new SFlow Parser
func NewParser(metricName string, sflowConfig V5FormatOptions) (*Parser, error) {
	if metricName == "" {
		return nil, fmt.Errorf("metric name cannot be empty")
	}
	result := &Parser{metricName: metricName, sflowFormat: V5Format(sflowConfig)}
	return result, nil
}

func (sfp *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	dc := decoder.NewDecodeContext(false)
	if err := dc.Decode(sfp.sflowFormat, bytes.NewBuffer(buf)); err != nil {
		return nil, err
	}
	return dc.GetMetrics(), nil
}

package influx

import (
	"errors"
	"fmt"

	"github.com/influxdata/telegraf"
)

const (
	maxErrorBufferSize = 1024
)

var (
	ErrNoMetric = errors.New("no metric in line")
)

type Handler interface {
	SetMeasurement(name []byte)
	AddTag(key []byte, value []byte)
	AddInt(key []byte, value []byte)
	AddFloat(key []byte, value []byte)
	AddString(key []byte, value []byte)
	AddBool(key []byte, value []byte)
	SetTimestamp(tm []byte)
	Reset()
}

type ParseError struct {
	Offset int
	msg    string
	buf    string
}

func (e *ParseError) Error() string {
	buffer := e.buf
	if len(buffer) > maxErrorBufferSize {
		buffer = buffer[:maxErrorBufferSize] + "..."
	}
	return fmt.Sprintf("metric parse error: %s at offset %d: %q", e.msg, e.Offset, buffer)
}

type Parser struct {
	DefaultTags map[string]string

	*machine
	handler *MetricHandler
}

func NewParser(handler *MetricHandler) *Parser {
	return &Parser{
		machine: NewMachine(handler),
		handler: handler,
	}
}

func (p *Parser) Parse(input []byte) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)
	p.machine.SetData(input)

	for p.machine.ParseLine() {
		err := p.machine.Err()
		if err != nil {
			return nil, &ParseError{
				Offset: p.machine.Position(),
				msg:    err.Error(),
				buf:    string(input),
			}
		}

		metric, err := p.handler.Metric()
		if err != nil {
			return nil, err
		}
		p.handler.Reset()
		metrics = append(metrics, metric)
	}

	p.applyDefaultTags(metrics)
	return metrics, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line + "\n"))
	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, ErrNoMetric
	}

	return metrics[0], nil
}

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

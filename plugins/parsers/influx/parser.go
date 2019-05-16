package influx

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
)

const (
	maxErrorBufferSize = 1024
)

var (
	ErrNoMetric = errors.New("no metric in line")
)

type ParseError struct {
	Offset     int
	LineOffset int
	LineNumber int
	Column     int
	msg        string
	buf        string
}

func (e *ParseError) Error() string {
	buffer := e.buf[e.LineOffset:]
	eol := strings.IndexAny(buffer, "\r\n")
	if eol >= 0 {
		buffer = buffer[:eol]
	}
	if len(buffer) > maxErrorBufferSize {
		buffer = buffer[:maxErrorBufferSize] + "..."
	}
	return fmt.Sprintf("metric parse error: %s at %d:%d: %q", e.msg, e.LineNumber, e.Column, buffer)
}

type Parser struct {
	DefaultTags map[string]string

	sync.Mutex
	*machine
	handler *MetricHandler
}

// NewParser returns a Parser than accepts line protocol
func NewParser(handler *MetricHandler) *Parser {
	return &Parser{
		machine: NewMachine(handler),
		handler: handler,
	}
}

// NewSeriesParser returns a Parser than accepts a measurement and tagset
func NewSeriesParser(handler *MetricHandler) *Parser {
	return &Parser{
		machine: NewSeriesMachine(handler),
		handler: handler,
	}
}

func (p *Parser) Parse(input []byte) ([]telegraf.Metric, error) {
	p.Lock()
	defer p.Unlock()
	metrics := make([]telegraf.Metric, 0)
	p.machine.SetData(input)

	for {
		err := p.machine.Next()
		if err == EOF {
			break
		}

		if err != nil {
			p.handler.Reset()
			return nil, &ParseError{
				Offset:     p.machine.Position(),
				LineOffset: p.machine.LineOffset(),
				LineNumber: p.machine.LineNumber(),
				Column:     p.machine.Column(),
				msg:        err.Error(),
				buf:        string(input),
			}
		}

		metric, err := p.handler.Metric()
		if err != nil {
			p.handler.Reset()
			return nil, err
		}

		if metric == nil {
			continue
		}

		metrics = append(metrics, metric)
	}

	p.applyDefaultTags(metrics)
	return metrics, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
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

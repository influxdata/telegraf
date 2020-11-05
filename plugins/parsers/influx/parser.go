package influx

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
)

const (
	maxErrorBufferSize = 1024
)

var (
	ErrNoMetric = errors.New("no metric in line")
)

type TimeFunc func() time.Time

// ParseError indicates a error in the parsing of the text.
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
	eol := strings.IndexAny(buffer, "\n")
	if eol >= 0 {
		buffer = strings.TrimSuffix(buffer[:eol], "\r")
	}
	if len(buffer) > maxErrorBufferSize {
		startEllipsis := true
		offset := e.Offset - e.LineOffset
		start := offset - maxErrorBufferSize
		if start < 0 {
			startEllipsis = false
			start = 0
		}
		// if we trimmed it the column won't line up. it'll always be the last character,
		// because the parser doesn't continue past it, but point it out anyway so
		// it's obvious where the issue is.
		buffer = buffer[start:offset] + "<-- here"
		if startEllipsis {
			buffer = "..." + buffer
		}
	}
	return fmt.Sprintf("metric parse error: %s at %d:%d: %q", e.msg, e.LineNumber, e.Column, buffer)
}

// Parser is an InfluxDB Line Protocol parser that implements the
// parsers.Parser interface.
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

func (h *Parser) SetTimeFunc(f TimeFunc) {
	h.handler.SetTimeFunc(f)
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
		p.applyDefaultTagsSingle(m)
	}
}

func (p *Parser) applyDefaultTagsSingle(metric telegraf.Metric) {
	for k, v := range p.DefaultTags {
		if !metric.HasTag(k) {
			metric.AddTag(k, v)
		}
	}
}

// StreamParser is an InfluxDB Line Protocol parser.  It is not safe for
// concurrent use in multiple goroutines.
type StreamParser struct {
	machine *streamMachine
	handler *MetricHandler
}

func NewStreamParser(r io.Reader) *StreamParser {
	handler := NewMetricHandler()
	return &StreamParser{
		machine: NewStreamMachine(r, handler),
		handler: handler,
	}
}

// SetTimeFunc changes the function used to determine the time of metrics
// without a timestamp.  The default TimeFunc is time.Now.  Useful mostly for
// testing, or perhaps if you want all metrics to have the same timestamp.
func (h *StreamParser) SetTimeFunc(f TimeFunc) {
	h.handler.SetTimeFunc(f)
}

func (h *StreamParser) SetTimePrecision(u time.Duration) {
	h.handler.SetTimePrecision(u)
}

// Next parses the next item from the stream.  You can repeat calls to this
// function if it returns ParseError to get the next metric or error.
func (p *StreamParser) Next() (telegraf.Metric, error) {
	err := p.machine.Next()
	if err == EOF {
		return nil, err
	}

	if e, ok := err.(*readErr); ok {
		return nil, e.Err
	}

	if err != nil {
		return nil, &ParseError{
			Offset:     p.machine.Position(),
			LineOffset: p.machine.LineOffset(),
			LineNumber: p.machine.LineNumber(),
			Column:     p.machine.Column(),
			msg:        err.Error(),
			buf:        p.machine.LineText(),
		}
	}

	metric, err := p.handler.Metric()
	if err != nil {
		return nil, err
	}

	return metric, nil
}

// Position returns the current byte offset into the data.
func (p *StreamParser) Position() int {
	return p.machine.Position()
}

// LineOffset returns the byte offset of the current line.
func (p *StreamParser) LineOffset() int {
	return p.machine.LineOffset()
}

// LineNumber returns the current line number.  Lines are counted based on the
// regular expression `\r?\n`.
func (p *StreamParser) LineNumber() int {
	return p.machine.LineNumber()
}

// Column returns the current column.
func (p *StreamParser) Column() int {
	return p.machine.Column()
}

// LineText returns the text of the current line that has been parsed so far.
func (p *StreamParser) LineText() string {
	return p.machine.LineText()
}

package influx

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/influxdata/line-protocol/v2/lineprotocol"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

const (
	maxErrorBufferSize = 1024
)

var (
	ErrNoMetric = errors.New("no metric in line")
)

type TimeFunc func() time.Time

// nthIndexAny finds the nth index of some unicode code point in a string or returns -1
func nthIndexAny(s, chars string, n int) int {
	offset := 0
	for found := 1; found <= n; found++ {
		i := strings.IndexAny(s[offset:], chars)
		if i < 0 {
			break
		}

		offset += i
		if found == n {
			return offset
		}

		offset += len(chars)
	}

	return -1
}

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
	lineStart := nthIndexAny(e.buf, "\n", e.LineNumber-1) + 1
	buffer := e.buf[lineStart:]
	eol := strings.IndexAny(buffer, "\n")
	if eol >= 0 {
		buffer = strings.TrimSuffix(buffer[:eol], "\r")
	}
	if len(buffer) > maxErrorBufferSize {
		startEllipsis := true
		offset := e.Column - 1 - lineStart
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

// convertToParseError attempts to convert a lineprotocol.DecodeError to a ParseError
func convertToParseError(input []byte, rawErr error) error {
	err, ok := rawErr.(*lineprotocol.DecodeError)
	if !ok {
		return rawErr
	}

	return &ParseError{
		LineNumber: int(err.Line),
		Column:     err.Column,
		msg:        err.Err.Error(),
		buf:        string(input),
	}
}

// Parser is an InfluxDB Line Protocol parser that implements the
// parsers.Parser interface.
type Parser struct {
	DefaultTags map[string]string

	defaultTime  TimeFunc
	precision    lineprotocol.Precision
	allowPartial bool
}

// NewParser returns a Parser than accepts line protocol
func NewParser() *Parser {
	return &Parser{
		defaultTime: time.Now,
		precision:   lineprotocol.Nanosecond,
	}
}

// NewSeriesParser returns a Parser than accepts a measurement and tagset
func NewSeriesParser() *Parser {
	return &Parser{
		defaultTime:  time.Now,
		precision:    lineprotocol.Nanosecond,
		allowPartial: true,
	}
}

func (p *Parser) SetTimeFunc(f TimeFunc) {
	p.defaultTime = f
}

func (p *Parser) Parse(input []byte) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)
	decoder := lineprotocol.NewDecoderWithBytes(input)

	for decoder.Next() {
		measurement, err := decoder.Measurement()
		if err != nil {
			return nil, convertToParseError(input, err)
		}
		m := metric.New(string(measurement), nil, nil, time.Time{})

		for {
			key, value, err := decoder.NextTag()
			if err != nil {
				// Allow empty tags for series parser
				if strings.Contains(err.Error(), "empty tag name") && p.allowPartial {
					break
				}

				return nil, convertToParseError(input, err)
			} else if key == nil {
				break
			}

			m.AddTag(string(key), string(value))
		}

		for {
			key, value, err := decoder.NextField()
			if err != nil {
				// Allow empty fields for series parser
				if strings.Contains(err.Error(), "expected field key") && p.allowPartial {
					break
				}

				return nil, convertToParseError(input, err)
			} else if key == nil {
				break
			}

			m.AddField(string(key), value.Interface())
		}

		t, err := decoder.Time(p.precision, p.defaultTime())
		if err != nil && !p.allowPartial {
			return nil, convertToParseError(input, err)
		}

		m.SetTime(t)
		metrics = append(metrics, m)
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
func (sp *StreamParser) SetTimeFunc(f TimeFunc) {
	sp.handler.SetTimeFunc(f)
}

func (sp *StreamParser) SetTimePrecision(u time.Duration) {
	sp.handler.SetTimePrecision(u)
}

// Next parses the next item from the stream.  You can repeat calls to this
// function if it returns ParseError to get the next metric or error.
func (sp *StreamParser) Next() (telegraf.Metric, error) {
	err := sp.machine.Next()
	if err == EOF {
		return nil, err
	}

	if e, ok := err.(*readErr); ok {
		return nil, e.Err
	}

	if err != nil {
		return nil, &ParseError{
			Offset:     sp.machine.Position(),
			LineOffset: sp.machine.LineOffset(),
			LineNumber: sp.machine.LineNumber(),
			Column:     sp.machine.Column(),
			msg:        err.Error(),
			buf:        sp.machine.LineText(),
		}
	}

	metric, err := sp.handler.Metric()
	if err != nil {
		return nil, err
	}

	return metric, nil
}

// Position returns the current byte offset into the data.
func (sp *StreamParser) Position() int {
	return sp.machine.Position()
}

// LineOffset returns the byte offset of the current line.
func (sp *StreamParser) LineOffset() int {
	return sp.machine.LineOffset()
}

// LineNumber returns the current line number.  Lines are counted based on the
// regular expression `\r?\n`.
func (sp *StreamParser) LineNumber() int {
	return sp.machine.LineNumber()
}

// Column returns the current column.
func (sp *StreamParser) Column() int {
	return sp.machine.Column()
}

// LineText returns the text of the current line that has been parsed so far.
func (sp *StreamParser) LineText() string {
	return sp.machine.LineText()
}

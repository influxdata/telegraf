package influx_upstream

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
	ErrEOF      = errors.New("EOF")
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
	*lineprotocol.DecodeError
	buf string
}

func (e *ParseError) Error() string {
	// When an error occurs within the stream decoder, we do not have access
	// to the internal buffer, so we cannot display the contents of the invalid
	// metric
	if e.buf == "" {
		return fmt.Sprintf("metric parse error: %s at %d:%d", e.Err, e.Line, e.Column)
	}

	lineStart := nthIndexAny(e.buf, "\n", int(e.Line-1)) + 1
	buffer := e.buf[lineStart:]
	eol := strings.IndexAny(buffer, "\n")
	if eol >= 0 {
		buffer = strings.TrimSuffix(buffer[:eol], "\r")
	}
	if len(buffer) > maxErrorBufferSize {
		startEllipsis := true
		offset := e.Column - 1 - lineStart
		if offset > len(buffer) || offset < 0 {
			offset = len(buffer)
		}
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
	return fmt.Sprintf("metric parse error: %s at %d:%d: %q", e.Err, e.Line, e.Column, buffer)
}

// convertToParseError attempts to convert a lineprotocol.DecodeError to a ParseError
func convertToParseError(input []byte, rawErr error) error {
	err, ok := rawErr.(*lineprotocol.DecodeError)
	if !ok {
		return rawErr
	}

	return &ParseError{
		DecodeError: err,
		buf:         string(input),
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
		m, err := nextMetric(decoder, p.precision, p.defaultTime, p.allowPartial)
		if err != nil {
			return nil, convertToParseError(input, err)
		}
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

func (p *Parser) SetTimePrecision(u time.Duration) {
	switch u {
	case time.Nanosecond:
		p.precision = lineprotocol.Nanosecond
	case time.Microsecond:
		p.precision = lineprotocol.Microsecond
	case time.Millisecond:
		p.precision = lineprotocol.Millisecond
	case time.Second:
		p.precision = lineprotocol.Second
	}
}

func (p *Parser) applyDefaultTags(metrics []telegraf.Metric) {
	if len(p.DefaultTags) == 0 {
		return
	}

	for _, m := range metrics {
		p.applyDefaultTagsSingle(m)
	}
}

func (p *Parser) applyDefaultTagsSingle(m telegraf.Metric) {
	for k, v := range p.DefaultTags {
		if !m.HasTag(k) {
			m.AddTag(k, v)
		}
	}
}

// StreamParser is an InfluxDB Line Protocol parser.  It is not safe for
// concurrent use in multiple goroutines.
type StreamParser struct {
	decoder     *lineprotocol.Decoder
	defaultTime TimeFunc
	precision   lineprotocol.Precision
	lastError   error
}

func NewStreamParser(r io.Reader) *StreamParser {
	return &StreamParser{
		decoder:     lineprotocol.NewDecoder(r),
		defaultTime: time.Now,
		precision:   lineprotocol.Nanosecond,
	}
}

// SetTimeFunc changes the function used to determine the time of metrics
// without a timestamp.  The default TimeFunc is time.Now.  Useful mostly for
// testing, or perhaps if you want all metrics to have the same timestamp.
func (sp *StreamParser) SetTimeFunc(f TimeFunc) {
	sp.defaultTime = f
}

func (sp *StreamParser) SetTimePrecision(u time.Duration) {
	switch u {
	case time.Nanosecond:
		sp.precision = lineprotocol.Nanosecond
	case time.Microsecond:
		sp.precision = lineprotocol.Microsecond
	case time.Millisecond:
		sp.precision = lineprotocol.Millisecond
	case time.Second:
		sp.precision = lineprotocol.Second
	}
}

// Next parses the next item from the stream.  You can repeat calls to this
// function if it returns ParseError to get the next metric or error.
func (sp *StreamParser) Next() (telegraf.Metric, error) {
	if !sp.decoder.Next() {
		if err := sp.decoder.Err(); err != nil && err != sp.lastError {
			sp.lastError = err
			return nil, err
		}

		return nil, ErrEOF
	}

	m, err := nextMetric(sp.decoder, sp.precision, sp.defaultTime, false)
	if err != nil {
		return nil, convertToParseError([]byte{}, err)
	}

	return m, nil
}

func nextMetric(decoder *lineprotocol.Decoder, precision lineprotocol.Precision, defaultTime TimeFunc, allowPartial bool) (telegraf.Metric, error) {
	measurement, err := decoder.Measurement()
	if err != nil {
		return nil, err
	}
	m := metric.New(string(measurement), nil, nil, time.Time{})

	for {
		key, value, err := decoder.NextTag()
		if err != nil {
			// Allow empty tags for series parser
			if strings.Contains(err.Error(), "empty tag name") && allowPartial {
				break
			}

			return nil, err
		} else if key == nil {
			break
		}

		m.AddTag(string(key), string(value))
	}

	for {
		key, value, err := decoder.NextField()
		if err != nil {
			// Allow empty fields for series parser
			if strings.Contains(err.Error(), "expected field key") && allowPartial {
				break
			}

			return nil, err
		} else if key == nil {
			break
		}

		m.AddField(string(key), value.Interface())
	}

	t, err := decoder.Time(precision, defaultTime())
	if err != nil && !allowPartial {
		return nil, err
	}
	m.SetTime(t)

	return m, nil
}

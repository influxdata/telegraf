package wavefront

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

const MAX_BUFFER_SIZE = 2

type Point struct {
	Name      string
	Value     string
	Timestamp int64
	Source    string
	Tags      map[string]string
}

// Parser represents a parser.
type PointParser struct {
	s   *PointScanner
	buf struct {
		tok []Token  // last read n tokens
		lit []string // last read n literals
		n   int      // unscanned buffer size (max=2)
	}
	scanBuf     bytes.Buffer // buffer reused for scanning tokens
	writeBuf    bytes.Buffer // buffer reused for parsing elements
	Elements    []ElementParser
	defaultTags map[string]string
}

// Returns a slice of ElementParser's for the Graphite format
func NewWavefrontElements() []ElementParser {
	var elements []ElementParser
	wsParser := WhiteSpaceParser{}
	wsParserNextOpt := WhiteSpaceParser{nextOptional: true}
	repeatParser := LoopedParser{wrappedParser: &TagParser{}, wsPaser: &wsParser}
	elements = append(elements, &NameParser{}, &wsParser, &ValueParser{}, &wsParserNextOpt,
		&TimestampParser{optional: true}, &wsParserNextOpt, &repeatParser)
	return elements
}

func NewWavefrontParser(defaultTags map[string]string) *PointParser {
	elements := NewWavefrontElements()
	return &PointParser{Elements: elements, defaultTags: defaultTags}
}

func (p *PointParser) Parse(buf []byte) ([]telegraf.Metric, error) {

	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	// add newline to end if not exists:
	if len(buf) > 0 && !bytes.HasSuffix(buf, []byte("\n")) {
		buf = append(buf, []byte("\n")...)
	}

	points := make([]Point, 0)

	buffer := bytes.NewBuffer(buf)
	reader := bufio.NewReader(buffer)
	for {
		// Read up to the next newline.
		buf, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}

		p.reset(buf)
		point := Point{}
		for _, element := range p.Elements {
			err := element.parse(p, &point)
			if err != nil {
				return nil, err
			}
		}

		points = append(points, point)
	}

	metrics, err := p.convertPointToTelegrafMetric(points)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (p *PointParser) ParseLine(line string) (telegraf.Metric, error) {
	buf := []byte(line)
	metrics, err := p.Parse(buf)
	if err != nil {
		return nil, err
	}

	if len(metrics) > 0 {
		return metrics[0], nil
	}

	return nil, nil
}

func (p *PointParser) SetDefaultTags(tags map[string]string) {
	p.defaultTags = tags
}

func (p *PointParser) convertPointToTelegrafMetric(points []Point) ([]telegraf.Metric, error) {

	metrics := make([]telegraf.Metric, 0)

	for _, point := range points {
		tags := make(map[string]string)
		for k, v := range point.Tags {
			tags[k] = v
		}
		// apply default tags after parsed tags
		for k, v := range p.defaultTags {
			tags[k] = v
		}

		// single field for value
		fields := make(map[string]interface{})
		v, err := strconv.ParseFloat(point.Value, 64)
		if err != nil {
			return nil, err
		}
		fields["value"] = v

		m, err := metric.New(point.Name, tags, fields, time.Unix(point.Timestamp, 0))
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that from the internal buffer instead.
func (p *PointParser) scan() (Token, string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		idx := p.buf.n % MAX_BUFFER_SIZE
		tok, lit := p.buf.tok[idx], p.buf.lit[idx]
		p.buf.n -= 1
		return tok, lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit := p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buffer(tok, lit)

	return tok, lit
}

func (p *PointParser) buffer(tok Token, lit string) {
	// create the buffer if it is empty
	if len(p.buf.tok) == 0 {
		p.buf.tok = make([]Token, MAX_BUFFER_SIZE)
		p.buf.lit = make([]string, MAX_BUFFER_SIZE)
	}

	// for now assume a simple circular buffer of length two
	p.buf.tok[0], p.buf.lit[0] = p.buf.tok[1], p.buf.lit[1]
	p.buf.tok[1], p.buf.lit[1] = tok, lit
}

// unscan pushes the previously read token back onto the buffer.
func (p *PointParser) unscan() {
	p.unscanTokens(1)
}

func (p *PointParser) unscanTokens(n int) {
	if n > MAX_BUFFER_SIZE {
		// just log for now
		log.Printf("cannot unscan more than %d tokens", MAX_BUFFER_SIZE)
	}
	p.buf.n += n
}

func (p *PointParser) reset(buf []byte) {

	// reset the scan buffer and write new byte
	p.scanBuf.Reset()
	p.scanBuf.Write(buf)

	if p.s == nil {
		p.s = NewScanner(&p.scanBuf)
	} else {
		// reset p.s.r passing in the buffer as the reader
		p.s.r.Reset(&p.scanBuf)
	}
	p.buf.n = 0
}

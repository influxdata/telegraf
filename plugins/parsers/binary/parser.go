package binary

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Parser struct {
	AllowNoMatch bool            `toml:"allow_no_match"`
	Endianess    string          `toml:"endianess" deprecated:"1.27.4;use 'endianness' instead"`
	Endianness   string          `toml:"endianness"`
	Configs      []Config        `toml:"binary"`
	HexEncoding  bool            `toml:"hex_encoding"`
	Log          telegraf.Logger `toml:"-"`

	metricName  string
	defaultTags map[string]string
	converter   binary.ByteOrder
}

func (p *Parser) Init() error {
	if p.Endianess != "" && p.Endianness == "" {
		p.Endianness = p.Endianess
	}

	switch p.Endianness {
	case "le":
		p.converter = binary.LittleEndian
	case "be":
		p.converter = binary.BigEndian
	case "", "host":
		p.converter = internal.HostEndianness
	default:
		return fmt.Errorf("unknown endianness %q", p.Endianness)
	}

	// Pre-process the configurations
	if len(p.Configs) == 0 {
		return errors.New("no configuration given")
	}
	for i, cfg := range p.Configs {
		if err := cfg.preprocess(p.metricName); err != nil {
			return fmt.Errorf("config %d invalid: %w", i, err)
		}
		p.Configs[i] = cfg
	}

	return nil
}

func (p *Parser) Parse(data []byte) ([]telegraf.Metric, error) {
	t := time.Now()

	// If the data is encoded in HEX, we need to decode it first
	buf := data
	if p.HexEncoding {
		s := strings.ReplaceAll(string(data), " ", "")
		s = strings.ReplaceAll(s, "\t", "")
		var err error
		buf, err = hex.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("decoding hex failed: %w", err)
		}
	}

	matches := 0
	metrics := make([]telegraf.Metric, 0)
	for i, cfg := range p.Configs {
		// Apply the filter and see if we should match this
		if !cfg.matches(buf) {
			p.Log.Debugf("ignoring data in config %d", i)
			continue
		}
		matches++

		// Collect the metric
		m, err := cfg.collect(buf, p.converter, t)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	if matches == 0 && !p.AllowNoMatch {
		return nil, errors.New("no matching configuration")
	}

	return metrics, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	switch len(metrics) {
	case 0:
		return nil, nil
	case 1:
		return metrics[0], nil
	default:
		return metrics[0], fmt.Errorf("cannot parse line with multiple (%d) metrics", len(metrics))
	}
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.defaultTags = tags
}

func init() {
	// Register all variants
	parsers.Add("binary",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{metricName: defaultMetricName}
		},
	)
}

func extractPart(in []byte, offset, bits uint64) ([]byte, error) {
	inLen := uint64(len(in))

	start := offset / 8
	bitend := offset%8 + bits
	length := bitend / 8
	if bitend%8 != 0 {
		length++
	}

	if start+length > inLen {
		return nil, fmt.Errorf("out-of-bounds @%d with %d bits", offset, bits)
	}

	var out []byte
	out = append(out, in[start:start+length]...)

	if offset%8 != 0 {
		// Mask the start-byte with the non-aligned bit-mask
		startmask := (byte(1) << (8 - offset%8)) - 1
		out[0] = out[0] & startmask
	}

	if bitend%8 == 0 {
		// The end is aligned to byte-boundaries
		return out, nil
	}

	shift := 8 - bitend%8
	carryshift := bitend % 8

	// We need to shift right in case of not ending at a byte boundary
	// to make the bits right aligned.
	// Carry over the bits from the byte left to fill in...
	var carry byte
	for i, x := range out {
		out[i] = (x >> shift) | carry
		carry = x << carryshift
	}

	if bits%8 == 0 {
		// Avoid an empty leading byte
		return out[1:], nil
	}

	return out, nil
}

func bitsForType(t string) (uint64, error) {
	switch t {
	case "uint8", "int8":
		return 8, nil
	case "uint16", "int16":
		return 16, nil
	case "uint32", "int32", "float32":
		return 32, nil
	case "uint64", "int64", "float64":
		return 64, nil
	}
	return 0, fmt.Errorf("cannot determine length for type %q", t)
}

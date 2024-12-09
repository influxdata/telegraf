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
	"github.com/influxdata/telegraf/metric"
)

type BinaryPart struct {
	Offset uint64 `toml:"offset"`
	Bits   uint64 `toml:"bits"`
	Match  string `toml:"match"`

	val []byte
}

type Filter struct {
	Selection []BinaryPart `toml:"selection"`
	LengthMin uint64       `toml:"length_min"`
	Length    uint64       `toml:"length"`
}

type Config struct {
	MetricName string  `toml:"metric_name"`
	Filter     *Filter `toml:"filter"`
	Entries    []Entry `toml:"entries"`
}

func (c *Config) preprocess(defaultName string) error {
	// Preprocess filter part
	if c.Filter != nil {
		if c.Filter.Length != 0 && c.Filter.LengthMin != 0 {
			return errors.New("length and length_min cannot be used together")
		}

		var length uint64
		for i, s := range c.Filter.Selection {
			end := (s.Offset + s.Bits) / 8
			if (s.Offset+s.Bits)%8 != 0 {
				end++
			}
			if end > length {
				length = end
			}
			var err error
			s.val, err = hex.DecodeString(strings.TrimPrefix(s.Match, "0x"))
			if err != nil {
				return fmt.Errorf("decoding match %d failed: %w", i, err)
			}
			c.Filter.Selection[i] = s
		}

		if c.Filter.Length != 0 && length > c.Filter.Length {
			return fmt.Errorf("filter length (%d) larger than constraint (%d)", length, c.Filter.Length)
		}

		if c.Filter.Length == 0 && length > c.Filter.LengthMin {
			c.Filter.LengthMin = length
		}
	}

	// Preprocess entries part
	var hasField, hasMeasurement bool
	defined := make(map[string]bool)
	for i, e := range c.Entries {
		if err := e.check(); err != nil {
			return fmt.Errorf("entry %q (%d): %w", e.Name, i, err)
		}
		// Store the normalized entry
		c.Entries[i] = e

		if e.Omit {
			continue
		}

		// Check for duplicate entries
		key := e.Assignment + "_" + e.Name
		if defined[key] {
			return fmt.Errorf("multiple definitions of %q", e.Name)
		}
		defined[key] = true
		hasMeasurement = hasMeasurement || e.Assignment == "measurement"
		hasField = hasField || e.Assignment == "field"
	}

	if !hasMeasurement && c.MetricName == "" {
		if defaultName == "" {
			return errors.New("no metric name given")
		}
		c.MetricName = defaultName
	}
	if !hasField {
		return errors.New("no field defined")
	}

	return nil
}

func (c *Config) matches(in []byte) bool {
	// If no filter is given, just match everything
	if c.Filter == nil {
		return true
	}

	// Checking length constraints
	length := uint64(len(in))
	if c.Filter.Length != 0 && length != c.Filter.Length {
		return false
	}
	if c.Filter.LengthMin != 0 && length < c.Filter.LengthMin {
		return false
	}

	// Matching elements
	for _, s := range c.Filter.Selection {
		data, err := extractPart(in, s.Offset, s.Bits)
		if err != nil {
			return false
		}
		if len(data) != len(s.val) {
			return false
		}
		for i, v := range data {
			if v != s.val[i] {
				return false
			}
		}
	}

	return true
}

func (c *Config) collect(in []byte, order binary.ByteOrder, defaultTime time.Time) (telegraf.Metric, error) {
	t := defaultTime
	name := c.MetricName
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	var offset uint64
	for _, e := range c.Entries {
		data, n, err := e.extract(in, offset)
		if err != nil {
			return nil, err
		}
		offset += n

		switch e.Assignment {
		case "measurement":
			name = convertStringType(data)
		case "field":
			v, err := e.convertType(data, order)
			if err != nil {
				return nil, fmt.Errorf("field %q failed: %w", e.Name, err)
			}
			fields[e.Name] = v
		case "tag":
			raw, err := e.convertType(data, order)
			if err != nil {
				return nil, fmt.Errorf("tag %q failed: %w", e.Name, err)
			}
			v, err := internal.ToString(raw)
			if err != nil {
				return nil, fmt.Errorf("tag %q failed: %w", e.Name, err)
			}
			tags[e.Name] = v
		case "time":
			var err error
			t, err = e.convertTimeType(data, order)
			if err != nil {
				return nil, fmt.Errorf("time failed: %w", err)
			}
		}
	}

	return metric.New(name, tags, fields, t), nil
}

package influx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
)

const (
	needMoreSpace = "need more space"
	invalidName   = "invalid name"
	noFields      = "no serializable fields"
)

type Serializer struct {
	MaxLineBytes  int  `toml:"influx_max_line_bytes"`
	SortFields    bool `toml:"influx_sort_fields"`
	UintSupport   bool `toml:"influx_uint_support"`
	OmitTimestamp bool `toml:"influx_omit_timestamp"`

	bytesWritten int

	buf    bytes.Buffer
	header []byte
	footer []byte
	pair   []byte
}

// metricError is an error causing an entire metric to be unserializable.
type metricError struct {
	series string
	reason string
}

func (e metricError) Error() string {
	if e.series != "" {
		return fmt.Sprintf("%q: %s", e.series, e.reason)
	}
	return e.reason
}

// fieldError is an error causing a field to be unserializable.
type fieldError struct {
	reason string
}

func (e fieldError) Error() string {
	return e.reason
}

func (s *Serializer) Init() error {
	s.header = make([]byte, 0, 50)
	s.footer = make([]byte, 0, 21)
	s.pair = make([]byte, 0, 50)

	return nil
}

func (s *Serializer) Serialize(m telegraf.Metric) ([]byte, error) {
	s.buf.Reset()
	err := s.writeMetric(&s.buf, m)
	if err != nil {
		return nil, err
	}

	out := make([]byte, 0, s.buf.Len())
	return append(out, s.buf.Bytes()...), nil
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	s.buf.Reset()
	for _, m := range metrics {
		err := s.write(&s.buf, m)
		if err != nil {
			var mErr *metricError
			if errors.As(err, &mErr) {
				continue
			}
			return nil, err
		}
	}
	out := make([]byte, 0, s.buf.Len())
	return append(out, s.buf.Bytes()...), nil
}

func (s *Serializer) write(w io.Writer, m telegraf.Metric) error {
	return s.writeMetric(w, m)
}

func (s *Serializer) writeString(w io.Writer, str string) error {
	n, err := io.WriteString(w, str)
	s.bytesWritten += n
	return err
}

func (s *Serializer) writeBytes(w io.Writer, b []byte) error {
	n, err := w.Write(b)
	s.bytesWritten += n
	return err
}

func (s *Serializer) buildHeader(m telegraf.Metric) error {
	s.header = s.header[:0]

	name := nameEscape(m.Name())
	if name == "" {
		return s.newMetricError(invalidName)
	}

	s.header = append(s.header, name...)

	for _, tag := range m.TagList() {
		key := escape(tag.Key)
		value := escape(tag.Value)

		// Tag keys and values that end with a backslash cannot be encoded by
		// line protocol.
		if strings.HasSuffix(key, `\`) {
			key = strings.TrimRight(key, `\`)
		}
		if strings.HasSuffix(value, `\`) {
			value = strings.TrimRight(value, `\`)
		}

		// Tag keys and values must not be the empty string.
		if key == "" || value == "" {
			continue
		}

		s.header = append(s.header, ',')
		s.header = append(s.header, key...)
		s.header = append(s.header, '=')
		s.header = append(s.header, value...)
	}

	s.header = append(s.header, ' ')
	return nil
}

func (s *Serializer) buildFooter(m telegraf.Metric) {
	s.footer = s.footer[:0]
	if !s.OmitTimestamp {
		s.footer = append(s.footer, ' ')
		s.footer = strconv.AppendInt(s.footer, m.Time().UnixNano(), 10)
	}
	s.footer = append(s.footer, '\n')
}

func (s *Serializer) buildFieldPair(key string, value interface{}) error {
	s.pair = s.pair[:0]
	key = escape(key)

	// Some keys are not encodeable as line protocol, such as those with a
	// trailing '\' or empty strings.
	if key == "" {
		return &fieldError{"invalid field key"}
	}

	s.pair = append(s.pair, key...)
	s.pair = append(s.pair, '=')
	pair, err := s.appendFieldValue(s.pair, value)
	if err != nil {
		return err
	}
	s.pair = pair
	return nil
}

func (s *Serializer) writeMetric(w io.Writer, m telegraf.Metric) error {
	var err error

	err = s.buildHeader(m)
	if err != nil {
		return err
	}

	s.buildFooter(m)

	if s.SortFields {
		sort.Slice(m.FieldList(), func(i, j int) bool {
			return m.FieldList()[i].Key < m.FieldList()[j].Key
		})
	}

	pairsLen := 0
	firstField := true
	for _, field := range m.FieldList() {
		err = s.buildFieldPair(field.Key, field.Value)
		if err != nil {
			log.Printf(
				"D! [serializers.influx] could not serialize field %q: %v; discarding field",
				field.Key, err)
			continue
		}

		bytesNeeded := len(s.header) + pairsLen + len(s.pair) + len(s.footer)

		// Additional length needed for field separator `,`
		if !firstField {
			bytesNeeded++
		}

		if s.MaxLineBytes > 0 && bytesNeeded > s.MaxLineBytes {
			// Need at least one field per line, this metric cannot be fit
			// into the max line bytes.
			if firstField {
				return s.newMetricError(needMoreSpace)
			}

			err = s.writeBytes(w, s.footer)
			if err != nil {
				return err
			}

			pairsLen = 0
			firstField = true
			bytesNeeded = len(s.header) + len(s.pair) + len(s.footer)

			if bytesNeeded > s.MaxLineBytes {
				return s.newMetricError(needMoreSpace)
			}
		}

		if firstField {
			err = s.writeBytes(w, s.header)
			if err != nil {
				return err
			}
		} else {
			err = s.writeString(w, ",")
			if err != nil {
				return err
			}
		}

		err = s.writeBytes(w, s.pair)
		if err != nil {
			return err
		}

		pairsLen += len(s.pair)
		firstField = false
	}

	if firstField {
		return s.newMetricError(noFields)
	}

	return s.writeBytes(w, s.footer)
}

func (s *Serializer) newMetricError(reason string) *metricError {
	if len(s.header) != 0 {
		series := bytes.TrimRight(s.header, " ")
		return &metricError{series: string(series), reason: reason}
	}
	return &metricError{reason: reason}
}

func (s *Serializer) appendFieldValue(buf []byte, value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case uint64:
		if s.UintSupport {
			return appendUintField(buf, v), nil
		}
		if v <= uint64(math.MaxInt64) {
			return appendIntField(buf, int64(v)), nil
		}
		return appendIntField(buf, math.MaxInt64), nil
	case int64:
		return appendIntField(buf, v), nil
	case float64:
		if math.IsNaN(v) {
			return nil, &fieldError{"is NaN"}
		}

		if math.IsInf(v, 0) {
			return nil, &fieldError{"is Inf"}
		}

		return appendFloatField(buf, v), nil
	case string:
		return appendStringField(buf, v), nil
	case bool:
		return appendBoolField(buf, v), nil
	default:
		return buf, &fieldError{fmt.Sprintf("invalid value type: %T", v)}
	}
}

func appendUintField(buf []byte, value uint64) []byte {
	return append(strconv.AppendUint(buf, value, 10), 'u')
}

func appendIntField(buf []byte, value int64) []byte {
	return append(strconv.AppendInt(buf, value, 10), 'i')
}

func appendFloatField(buf []byte, value float64) []byte {
	return strconv.AppendFloat(buf, value, 'f', -1, 64)
}

func appendBoolField(buf []byte, value bool) []byte {
	return strconv.AppendBool(buf, value)
}

func appendStringField(buf []byte, value string) []byte {
	buf = append(buf, '"')
	buf = append(buf, stringFieldEscape(value)...)
	buf = append(buf, '"')
	return buf
}

func init() {
	serializers.Add("influx",
		func() telegraf.Serializer {
			return &Serializer{}
		},
	)
}

package influx

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"sort"
	"strconv"

	"github.com/influxdata/telegraf"
)

const MaxInt64 = int64(^uint64(0) >> 1)

type FieldSortOrder int

const (
	NoSortFields FieldSortOrder = iota
	SortFields
)

type FieldTypeSupport int

const (
	UintSupport FieldTypeSupport = 1 << iota
)

var (
	NeedMoreSpace = "need more space"
	InvalidName   = "invalid name"
	NoFields      = "no serializable fields"
)

// MetricError is an error causing an entire metric to be unserializable.
type MetricError struct {
	series string
	reason string
}

func (e MetricError) Error() string {
	if e.series != "" {
		return fmt.Sprintf("%q: %s", e.series, e.reason)
	}
	return e.reason
}

// FieldError is an error causing a field to be unserializable.
type FieldError struct {
	reason string
}

func (e FieldError) Error() string {
	return e.reason
}

// Serializer is a serializer for line protocol.
type Serializer struct {
	maxLineBytes     int
	bytesWritten     int
	fieldSortOrder   FieldSortOrder
	fieldTypeSupport FieldTypeSupport

	buf    bytes.Buffer
	header []byte
	footer []byte
	pair   []byte
}

func NewSerializer() *Serializer {
	serializer := &Serializer{
		fieldSortOrder: NoSortFields,

		header: make([]byte, 0, 50),
		footer: make([]byte, 0, 21),
		pair:   make([]byte, 0, 50),
	}
	return serializer
}

func (s *Serializer) SetMaxLineBytes(bytes int) {
	s.maxLineBytes = bytes
}

func (s *Serializer) SetFieldSortOrder(order FieldSortOrder) {
	s.fieldSortOrder = order
}

func (s *Serializer) SetFieldTypeSupport(typeSupport FieldTypeSupport) {
	s.fieldTypeSupport = typeSupport
}

// Serialize writes the telegraf.Metric to a byte slice.  May produce multiple
// lines of output if longer than maximum line length.  Lines are terminated
// with a newline (LF) char.
func (s *Serializer) Serialize(m telegraf.Metric) ([]byte, error) {
	s.buf.Reset()
	err := s.writeMetric(&s.buf, m)
	if err != nil {
		return nil, err
	}

	out := make([]byte, s.buf.Len())
	copy(out, s.buf.Bytes())
	return out, nil
}

// SerializeBatch writes the slice of metrics and returns a byte slice of the
// results.  The returned byte slice may contain multiple lines of data.
func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	s.buf.Reset()
	for _, m := range metrics {
		_, err := s.Write(&s.buf, m)
		if err != nil {
			return nil, err
		}
	}
	out := make([]byte, s.buf.Len())
	copy(out, s.buf.Bytes())
	return out, nil
}
func (s *Serializer) Write(w io.Writer, m telegraf.Metric) (int, error) {
	err := s.writeMetric(w, m)
	return s.bytesWritten, err
}

func (s *Serializer) writeString(w io.Writer, str string) error {
	n, err := io.WriteString(w, str)
	s.bytesWritten += n
	return err
}

func (s *Serializer) write(w io.Writer, b []byte) error {
	n, err := w.Write(b)
	s.bytesWritten += n
	return err
}

func (s *Serializer) buildHeader(m telegraf.Metric) error {
	s.header = s.header[:0]

	name := nameEscape(m.Name())
	if name == "" {
		return s.newMetricError(InvalidName)
	}

	s.header = append(s.header, name...)

	for _, tag := range m.TagList() {
		key := escape(tag.Key)
		value := escape(tag.Value)

		// Some keys and values are not encodeable as line protocol, such as
		// those with a trailing '\' or empty strings.
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
	s.footer = append(s.footer, ' ')
	s.footer = strconv.AppendInt(s.footer, m.Time().UnixNano(), 10)
	s.footer = append(s.footer, '\n')
}

func (s *Serializer) buildFieldPair(key string, value interface{}) error {
	s.pair = s.pair[:0]
	key = escape(key)

	// Some keys are not encodeable as line protocol, such as those with a
	// trailing '\' or empty strings.
	if key == "" {
		return &FieldError{"invalid field key"}
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

	if s.fieldSortOrder == SortFields {
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
			bytesNeeded += 1
		}

		if s.maxLineBytes > 0 && bytesNeeded > s.maxLineBytes {
			// Need at least one field per line, this metric cannot be fit
			// into the max line bytes.
			if firstField {
				return s.newMetricError(NeedMoreSpace)
			}

			err = s.write(w, s.footer)
			if err != nil {
				return err
			}

			pairsLen = 0
			firstField = true
			bytesNeeded = len(s.header) + len(s.pair) + len(s.footer)

			if bytesNeeded > s.maxLineBytes {
				return s.newMetricError(NeedMoreSpace)
			}
		}

		if firstField {
			err = s.write(w, s.header)
			if err != nil {
				return err
			}
		} else {
			err = s.writeString(w, ",")
			if err != nil {
				return err
			}
		}

		err = s.write(w, s.pair)
		if err != nil {
			return err
		}

		pairsLen += len(s.pair)
		firstField = false
	}

	if firstField {
		return s.newMetricError(NoFields)
	}

	return s.write(w, s.footer)
}

func (s *Serializer) newMetricError(reason string) *MetricError {
	if len(s.header) != 0 {
		series := bytes.TrimRight(s.header, " ")
		return &MetricError{series: string(series), reason: reason}
	}
	return &MetricError{reason: reason}
}

func (s *Serializer) appendFieldValue(buf []byte, value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case uint64:
		if s.fieldTypeSupport&UintSupport != 0 {
			return appendUintField(buf, v), nil
		} else {
			if v <= uint64(MaxInt64) {
				return appendIntField(buf, int64(v)), nil
			} else {
				return appendIntField(buf, int64(MaxInt64)), nil
			}
		}
	case int64:
		return appendIntField(buf, v), nil
	case float64:
		if math.IsNaN(v) {
			return nil, &FieldError{"is NaN"}
		}

		if math.IsInf(v, 0) {
			return nil, &FieldError{"is Inf"}
		}

		return appendFloatField(buf, v), nil
	case string:
		return appendStringField(buf, v), nil
	case bool:
		return appendBoolField(buf, v), nil
	default:
		return buf, &FieldError{fmt.Sprintf("invalid value type: %T", v)}
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

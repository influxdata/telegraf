package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

type Serializer struct {
	TimestampFormat string `toml:"timestamp_format"`
	Separator       string `toml:"csv_separator"`
	Header          bool   `toml:"csv_header"`
	Prefix          bool   `toml:"csv_column_prefix"`

	buffer bytes.Buffer
	writer *csv.Writer
}

func NewSerializer(timestampFormat, separator string, header, prefix bool) (*Serializer, error) {
	// Setting defaults
	if separator == "" {
		separator = ","
	}

	// Check inputs
	if len(separator) > 1 {
		return nil, fmt.Errorf("invalid separator %q", separator)
	}
	switch timestampFormat {
	case "":
		timestampFormat = "unix"
	case "unix", "unix_ms", "unix_us", "unix_ns":
	default:
		if time.Now().Format(timestampFormat) == timestampFormat {
			return nil, fmt.Errorf("invalid timestamp format %q", timestampFormat)
		}
	}

	s := &Serializer{
		TimestampFormat: timestampFormat,
		Separator:       separator,
		Header:          header,
		Prefix:          prefix,
	}

	// Initialize the writer
	s.writer = csv.NewWriter(&s.buffer)
	s.writer.Comma = []rune(separator)[0]

	return s, nil
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	// Clear the buffer
	s.buffer.Truncate(0)

	// Write the header if the user wants us to
	if s.Header {
		if err := s.writeHeader(metric); err != nil {
			return nil, fmt.Errorf("writing header failed: %w", err)
		}
		s.Header = false
	}

	// Write the data
	if err := s.writeData(metric); err != nil {
		return nil, fmt.Errorf("writing data failed: %w", err)
	}

	// Finish up
	s.writer.Flush()
	return s.buffer.Bytes(), nil
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	if len(metrics) < 1 {
		return nil, nil
	}

	// Clear the buffer
	s.buffer.Truncate(0)

	// Write the header if the user wants us to
	if s.Header {
		if err := s.writeHeader(metrics[0]); err != nil {
			return nil, fmt.Errorf("writing header failed: %w", err)
		}
		s.Header = false
	}

	for _, m := range metrics {
		if err := s.writeData(m); err != nil {
			return nil, fmt.Errorf("writing data failed: %w", err)
		}
	}

	// Finish up
	s.writer.Flush()
	return s.buffer.Bytes(), nil
}

func (s *Serializer) writeHeader(metric telegraf.Metric) error {
	columns := []string{
		"timestamp",
		"measurement",
	}
	for _, tag := range metric.TagList() {
		if s.Prefix {
			columns = append(columns, "tag_"+tag.Key)
		} else {
			columns = append(columns, tag.Key)
		}
	}

	// Sort the fields by name
	sort.Slice(metric.FieldList(), func(i, j int) bool {
		return metric.FieldList()[i].Key < metric.FieldList()[j].Key
	})
	for _, field := range metric.FieldList() {
		if s.Prefix {
			columns = append(columns, "field_"+field.Key)
		} else {
			columns = append(columns, field.Key)
		}
	}

	return s.writer.Write(columns)
}

func (s *Serializer) writeData(metric telegraf.Metric) error {
	var timestamp string

	// Format the time
	switch s.TimestampFormat {
	case "unix":
		timestamp = strconv.FormatInt(metric.Time().Unix(), 10)
	case "unix_ms":
		timestamp = strconv.FormatInt(metric.Time().UnixNano()/1_000_000, 10)
	case "unix_us":
		timestamp = strconv.FormatInt(metric.Time().UnixNano()/1_000, 10)
	case "unix_ns":
		timestamp = strconv.FormatInt(metric.Time().UnixNano(), 10)
	default:
		timestamp = metric.Time().Format(s.TimestampFormat)
	}

	columns := []string{
		timestamp,
		metric.Name(),
	}
	for _, tag := range metric.TagList() {
		columns = append(columns, tag.Value)
	}

	// Sort the fields by name
	sort.Slice(metric.FieldList(), func(i, j int) bool {
		return metric.FieldList()[i].Key < metric.FieldList()[j].Key
	})
	for _, field := range metric.FieldList() {
		v, err := internal.ToString(field.Value)
		if err != nil {
			return fmt.Errorf("converting field %q to string failed: %w", field.Key, err)
		}
		columns = append(columns, v)
	}

	return s.writer.Write(columns)
}

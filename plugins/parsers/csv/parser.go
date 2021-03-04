package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
)

type TimeFunc func() time.Time

type Config struct {
	ColumnNames       []string `toml:"csv_column_names"`
	ColumnTypes       []string `toml:"csv_column_types"`
	Comment           string   `toml:"csv_comment"`
	Delimiter         string   `toml:"csv_delimiter"`
	HeaderRowCount    int      `toml:"csv_header_row_count"`
	MeasurementColumn string   `toml:"csv_measurement_column"`
	MetricName        string   `toml:"metric_name"`
	SkipColumns       int      `toml:"csv_skip_columns"`
	SkipRows          int      `toml:"csv_skip_rows"`
	TagColumns        []string `toml:"csv_tag_columns"`
	TimestampColumn   string   `toml:"csv_timestamp_column"`
	TimestampFormat   string   `toml:"csv_timestamp_format"`
	Timezone          string   `toml:"csv_timezone"`
	TrimSpace         bool     `toml:"csv_trim_space"`
	SkipValues        []string `toml:"csv_skip_values"`

	gotColumnNames bool

	TimeFunc    func() time.Time
	DefaultTags map[string]string
}

// Parser is a CSV parser, you should use NewParser to create a new instance.
type Parser struct {
	*Config
}

func NewParser(c *Config) (*Parser, error) {
	if c.HeaderRowCount == 0 && len(c.ColumnNames) == 0 {
		return nil, fmt.Errorf("`csv_header_row_count` must be defined if `csv_column_names` is not specified")
	}

	if c.Delimiter != "" {
		runeStr := []rune(c.Delimiter)
		if len(runeStr) > 1 {
			return nil, fmt.Errorf("csv_delimiter must be a single character, got: %s", c.Delimiter)
		}
	}

	if c.Comment != "" {
		runeStr := []rune(c.Comment)
		if len(runeStr) > 1 {
			return nil, fmt.Errorf("csv_delimiter must be a single character, got: %s", c.Comment)
		}
	}

	if len(c.ColumnNames) > 0 && len(c.ColumnTypes) > 0 && len(c.ColumnNames) != len(c.ColumnTypes) {
		return nil, fmt.Errorf("csv_column_names field count doesn't match with csv_column_types")
	}

	c.gotColumnNames = len(c.ColumnNames) > 0

	if c.TimeFunc == nil {
		c.TimeFunc = time.Now
	}

	return &Parser{Config: c}, nil
}

func (p *Parser) SetTimeFunc(fn TimeFunc) {
	p.TimeFunc = fn
}

func (p *Parser) compile(r io.Reader) (*csv.Reader, error) {
	csvReader := csv.NewReader(r)
	// ensures that the reader reads records of different lengths without an error
	csvReader.FieldsPerRecord = -1
	if p.Delimiter != "" {
		csvReader.Comma = []rune(p.Delimiter)[0]
	}
	if p.Comment != "" {
		csvReader.Comment = []rune(p.Comment)[0]
	}
	csvReader.TrimLeadingSpace = p.TrimSpace
	return csvReader, nil
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	r := bytes.NewReader(buf)
	csvReader, err := p.compile(r)
	if err != nil {
		return nil, err
	}
	// skip first rows
	for i := 0; i < p.SkipRows; i++ {
		_, err := csvReader.Read()
		if err != nil {
			return nil, err
		}
	}
	// if there is a header and we did not get DataColumns
	// set DataColumns to names extracted from the header
	// we always reread the header to avoid side effects
	// in cases where multiple files with different
	// headers are read
	if !p.gotColumnNames {
		headerNames := make([]string, 0)
		for i := 0; i < p.HeaderRowCount; i++ {
			header, err := csvReader.Read()
			if err != nil {
				return nil, err
			}
			//concatenate header names
			for i := range header {
				name := header[i]
				if p.TrimSpace {
					name = strings.Trim(name, " ")
				}
				if len(headerNames) <= i {
					headerNames = append(headerNames, name)
				} else {
					headerNames[i] = headerNames[i] + name
				}
			}
		}
		p.ColumnNames = headerNames[p.SkipColumns:]
	} else {
		// if columns are named, just skip header rows
		for i := 0; i < p.HeaderRowCount; i++ {
			_, err := csvReader.Read()
			if err != nil {
				return nil, err
			}
		}
	}

	table, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	metrics := make([]telegraf.Metric, 0)
	for _, record := range table {
		m, err := p.parseRecord(record)
		if err != nil {
			return metrics, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// ParseLine does not use any information in header and assumes DataColumns is set
// it will also not skip any rows
func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	r := bytes.NewReader([]byte(line))
	csvReader, err := p.compile(r)
	if err != nil {
		return nil, err
	}

	// if there is nothing in DataColumns, ParseLine will fail
	if len(p.ColumnNames) == 0 {
		return nil, fmt.Errorf("[parsers.csv] data columns must be specified")
	}

	record, err := csvReader.Read()
	if err != nil {
		return nil, err
	}
	m, err := p.parseRecord(record)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (p *Parser) parseRecord(record []string) (telegraf.Metric, error) {
	recordFields := make(map[string]interface{})
	tags := make(map[string]string)

	// skip columns in record
	record = record[p.SkipColumns:]
outer:
	for i, fieldName := range p.ColumnNames {
		if i < len(record) {
			value := record[i]
			if p.TrimSpace {
				value = strings.Trim(value, " ")
			}

			// don't record fields where the value matches a skip value
			for _, s := range p.SkipValues {
				if value == s {
					continue outer
				}
			}

			for _, tagName := range p.TagColumns {
				if tagName == fieldName {
					tags[tagName] = value
					continue outer
				}
			}

			// If the field name is the timestamp column, then keep field name as is.
			if fieldName == p.TimestampColumn {
				recordFields[fieldName] = value
				continue
			}

			// Try explicit conversion only when column types is defined.
			if len(p.ColumnTypes) > 0 {
				// Throw error if current column count exceeds defined types.
				if i >= len(p.ColumnTypes) {
					return nil, fmt.Errorf("column type: column count exceeded")
				}

				var val interface{}
				var err error

				switch p.ColumnTypes[i] {
				case "int":
					val, err = strconv.ParseInt(value, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("column type: parse int error %s", err)
					}
				case "float":
					val, err = strconv.ParseFloat(value, 64)
					if err != nil {
						return nil, fmt.Errorf("column type: parse float error %s", err)
					}
				case "bool":
					val, err = strconv.ParseBool(value)
					if err != nil {
						return nil, fmt.Errorf("column type: parse bool error %s", err)
					}
				default:
					val = value
				}

				recordFields[fieldName] = val
				continue
			}

			// attempt type conversions
			if iValue, err := strconv.ParseInt(value, 10, 64); err == nil {
				recordFields[fieldName] = iValue
			} else if fValue, err := strconv.ParseFloat(value, 64); err == nil {
				recordFields[fieldName] = fValue
			} else if bValue, err := strconv.ParseBool(value); err == nil {
				recordFields[fieldName] = bValue
			} else {
				recordFields[fieldName] = value
			}
		}
	}

	// add default tags
	for k, v := range p.DefaultTags {
		tags[k] = v
	}

	// will default to plugin name
	measurementName := p.MetricName
	if p.MeasurementColumn != "" {
		if recordFields[p.MeasurementColumn] != nil && recordFields[p.MeasurementColumn] != "" {
			measurementName = fmt.Sprintf("%v", recordFields[p.MeasurementColumn])
		}
	}

	metricTime, err := parseTimestamp(p.TimeFunc, recordFields, p.TimestampColumn, p.TimestampFormat, p.Timezone)
	if err != nil {
		return nil, err
	}

	// Exclude `TimestampColumn` and `MeasurementColumn`
	delete(recordFields, p.TimestampColumn)
	delete(recordFields, p.MeasurementColumn)

	m, err := metric.New(measurementName, tags, recordFields, metricTime)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ParseTimestamp return a timestamp, if there is no timestamp on the csv it
// will be the current timestamp, else it will try to parse the time according
// to the format.
func parseTimestamp(timeFunc func() time.Time, recordFields map[string]interface{},
	timestampColumn, timestampFormat string, Timezone string,
) (time.Time, error) {
	if timestampColumn != "" {
		if recordFields[timestampColumn] == nil {
			return time.Time{}, fmt.Errorf("timestamp column: %v could not be found", timestampColumn)
		}

		switch timestampFormat {
		case "":
			return time.Time{}, fmt.Errorf("timestamp format must be specified")
		default:
			metricTime, err := internal.ParseTimestamp(timestampFormat, recordFields[timestampColumn], Timezone)
			if err != nil {
				return time.Time{}, err
			}
			return metricTime, err
		}
	}

	return timeFunc(), nil
}

// SetDefaultTags set the DefaultTags
func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type Parser struct {
	MetricName        string
	HeaderRowCount    int
	SkipRows          int
	SkipColumns       int
	Delimiter         string
	Comment           string
	TrimSpace         bool
	ColumnNames       []string
	TagColumns        []string
	MeasurementColumn string
	TimestampColumn   string
	TimestampFormat   string
	DefaultTags       map[string]string
}

func (p *Parser) compile(r *bytes.Reader) (*csv.Reader, error) {
	csvReader := csv.NewReader(r)
	// ensures that the reader reads records of different lengths without an error
	csvReader.FieldsPerRecord = -1
	if p.Delimiter != "" {
		runeStr := []rune(p.Delimiter)
		if len(runeStr) > 1 {
			return nil, fmt.Errorf("delimiter must be a single character, got: %s", p.Delimiter)
		}
		csvReader.Comma = runeStr[0]
	}
	if p.Comment != "" {
		runeStr := []rune(p.Comment)
		if len(runeStr) > 1 {
			return nil, fmt.Errorf("comment must be a single character, got: %s", p.Comment)
		}
		csvReader.Comment = runeStr[0]
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
		csvReader.Read()
	}
	// if there is a header and nothing in DataColumns
	// set DataColumns to names extracted from the header
	headerNames := make([]string, 0)
	if len(p.ColumnNames) == 0 {
		for i := 0; i < p.HeaderRowCount; i++ {
			header, err := csvReader.Read()
			if err != nil {
				return nil, err
			}
			//concatenate header names
			for i := range header {
				if len(headerNames) <= i {
					headerNames = append(headerNames, header[i])
				} else {
					headerNames[i] = headerNames[i] + header[i]
				}
			}
		}
		p.ColumnNames = headerNames[p.SkipColumns:]
	} else if len(p.ColumnNames) > 0 {
		// if columns are named, just skip header rows
		for i := 0; i < p.HeaderRowCount; i++ {
			csvReader.Read()
		}
	} else if p.HeaderRowCount == 0 && len(p.ColumnNames) == 0 {
		// if there is no header and no DataColumns, that's an error
		return nil, fmt.Errorf("there must be a header if `csv_data_columns` is not specified")
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
	for i, fieldName := range p.ColumnNames {
		if i < len(record) {
			value := record[i]
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
	if recordFields[p.MeasurementColumn] != nil {
		measurementName = recordFields[p.MeasurementColumn].(string)
	}

	for _, tagName := range p.TagColumns {
		if recordFields[tagName] == nil {
			return nil, fmt.Errorf("could not find field: %v", tagName)
		}
		tags[tagName] = recordFields[tagName].(string)
		delete(recordFields, tagName)
	}

	metricTime := time.Now()
	if p.TimestampColumn != "" {
		if recordFields[p.TimestampColumn] == nil {
			return nil, fmt.Errorf("timestamp column: %v could not be found", p.TimestampColumn)
		}
		tStr := recordFields[p.TimestampColumn].(string)
		if p.TimestampFormat == "" {
			return nil, fmt.Errorf("timestamp format must be specified")
		}

		var err error
		metricTime, err = time.Parse(p.TimestampFormat, tStr)
		if err != nil {
			return nil, err
		}
	}

	m, err := metric.New(measurementName, tags, recordFields, metricTime)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

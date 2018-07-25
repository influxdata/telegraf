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

type CSVParser struct {
	MetricName      string
	Header          bool
	Delimiter       string
	DataColumns     []string
	TagColumns      []string
	FieldColumns    []string
	NameColumn      string
	TimestampColumn string
	TimestampFormat string
	DefaultTags     map[string]string
}

func (p *CSVParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	r := bytes.NewReader(buf)
	csvReader := csv.NewReader(r)
	csvReader.FieldsPerRecord = len(p.DataColumns)
	if p.Delimiter != "" {
		runeStr := []rune(p.Delimiter)
		if len(runeStr) > 1 {
			return nil, fmt.Errorf("delimiter must be a single character, got: %v", p.Delimiter)
		}
		csvReader.Comma = runeStr[0]
	}

	//if there is a header and nothing in DataColumns
	//set DataColumns to names extracted from the header
	if p.Header && len(p.DataColumns) == 0 {
		header, err := csvReader.Read()
		if err != nil {
			return nil, err
		}
		p.DataColumns = header

	} else if p.Header {
		//if there is a header and DataColumns is specified, just skip header
		csvReader.Read()

	} else if !p.Header && len(p.DataColumns) == 0 {
		//if there is no header and no DataColumns, that's an error
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

//does not use any information in header and assumes DataColumns is set
func (p *CSVParser) ParseLine(line string) (telegraf.Metric, error) {
	r := bytes.NewReader([]byte(line))
	csvReader := csv.NewReader(r)
	csvReader.FieldsPerRecord = len(p.DataColumns)
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

func (p *CSVParser) parseRecord(record []string) (telegraf.Metric, error) {
	recordFields := make(map[string]string)
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	for i, fieldName := range p.DataColumns {
		recordFields[fieldName] = record[i]
	}

	//add default tags
	for k, v := range p.DefaultTags {
		tags[k] = v
	}

	for _, tagName := range p.TagColumns {
		if recordFields[tagName] == "" {
			return nil, fmt.Errorf("could not find field: %v", tagName)
		}
		tags[tagName] = recordFields[tagName]
	}

	for _, fieldName := range p.FieldColumns {
		if recordFields[fieldName] == "" {
			return nil, fmt.Errorf("could not find field: %v", fieldName)
		}

		//attempt type conversions
		value := recordFields[fieldName]
		if iValue, err := strconv.Atoi(value); err == nil {
			fields[fieldName] = iValue
		} else if fValue, err := strconv.ParseFloat(value, 64); err == nil {
			fields[fieldName] = fValue
		} else if bValue, err := strconv.ParseBool(value); err == nil {
			fields[fieldName] = bValue
		} else {
			fields[fieldName] = value
		}
	}

	//will default to plugin name
	measurementName := p.MetricName
	if recordFields[p.NameColumn] != "" {
		measurementName = recordFields[p.NameColumn]
	}

	metricTime := time.Now()
	if p.TimestampColumn != "" {
		tStr := recordFields[p.TimestampColumn]
		if p.TimestampFormat == "" {
			return nil, fmt.Errorf("timestamp format must be specified")
		}

		var err error
		metricTime, err = time.Parse(p.TimestampFormat, tStr)
		if err != nil {
			return nil, err
		}
	}

	m, err := metric.New(measurementName, tags, fields, metricTime)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (p *CSVParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

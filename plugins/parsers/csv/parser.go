package csv

import (
	"log"
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
)

type TimeFunc func() time.Time

type Parser struct {
	MetricName        string
	HeaderRowCount    int
	SkipRows          int
	SkipColumns       int
	Delimiter         string
	Comment           string
	TrimSpace         bool
	ColumnNames       []string
	Columns		  map[int][]string
	ColumnTypes       []string
	TagColumns        []string
	MeasurementColumn string
	TimestampColumn   string
	TimestampFormat   string
	DateColumn	  string
	DateFormat	  string
	DefaultTags       map[string]string
	TimeFunc          func() time.Time
}

func (p *Parser) SetTimeFunc(fn TimeFunc) {
	p.TimeFunc = fn
}

func (p *Parser) compile(r *bytes.Reader) (*csv.Reader, error) {
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
		csvReader.Read()
	}
	// if there is a header and nothing in DataColumns
	// set DataColumns to names extracted from the header
	headersMap := make(map[int][]string)
	if len(p.ColumnNames) == 0 {
		log.Printf("header row count: %d", p.HeaderRowCount)
		for i := 0; i < p.HeaderRowCount; i++ {
			headerNames := make([]string, 0)
			header, err := csvReader.Read()
			log.Printf("line %d: [%s] | columns: %d", i, header, len(header))
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
				} //else {
				//	headerNames[i] = headerNames[i] + name
				//}
			}			
			p.ColumnNames = headerNames[p.SkipColumns:]
			headersMap[len(p.ColumnNames)] = p.ColumnNames
		}	
		//p.Columns[len(p.ColumnNames)] = p.ColumnNames
	} else {
		// if columns are named, just skip header rows
		for i := 0; i < p.HeaderRowCount; i++ {
			csvReader.Read()
		}
	}
	p.Columns = headersMap
	log.Printf("done reading in headers: [%s]", p.Columns)
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
	if len(p.ColumnNames) == 0 && len(p.Columns) == 0 {
		return nil, fmt.Errorf("[parsers.csv] data columns must be specified")
	}

	record, err := csvReader.Read()
	if err != nil {
		return nil, err
	}
	log.Printf("parsing record..")
	m, err := p.parseRecord(record)
	log.Printf("parse complete")
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
	if len(p.Columns) >= 2 {
		p.ColumnNames = p.Columns[len(record)]
	}
	if p.ColumnNames != nil {

	outer:
		for i, fieldName := range p.ColumnNames {
			if i < len(record) {
			 	value := record[i]
				if p.TrimSpace {
					value = strings.Trim(value, " ")
				}

				for _, tagName := range p.TagColumns {
					if tagName == fieldName {
						tags[tagName] = value
						continue outer
					}
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
				//if iValue, err := strconv.ParseInt(value, 10, 64); err == nil {
				//	recordFields[fieldName] = iValue
				//} else if fValue, err := strconv.ParseFloat(value, 64); err == nil {
				//	recordFields[fieldName] = fValue
				//} else if bValue, err := strconv.ParseBool(value); err == nil {
				//	recordFields[fieldName] = bValue
				//} else {
					recordFields[fieldName] = value
				//}
			}
		}
	}
	// add default tags
	for k, v := range p.DefaultTags {
		tags[k] = v
	}

	// will default to plugin name
	measurementName := p.MetricName
	if recordFields[p.MeasurementColumn] != nil && recordFields[p.MeasurementColumn] != "" {
		measurementName = fmt.Sprintf("%v", recordFields[p.MeasurementColumn])
	}

	metricTime, err := parseTimestamp(p.TimeFunc, recordFields, p.TimestampColumn, p.TimestampFormat, p.DateColumn, p.DateFormat)
	if err != nil {
		return nil, err
	}

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
	timestampColumn, timestampFormat string, dateColumn string, dateFormat string) (time.Time, error) {
	if dateColumn != "" {
		if dateColumn, ok := recordFields[dateColumn].(string); ok {
			if dateFormat != "" {
				//Convert '/' to '-' so no checks are needed
				dateColumn = strings.ReplaceAll(dateColumn, "/", "-")
				dateFormat = strings.ReplaceAll(dateFormat, "/", "-")
				dateColumnSplit := strings.SplitN(dateFormat, "-", -1)
				dateFormatSplit := strings.SplitN(dateFormat, "-", -1)
				if len(dateColumnSplit) != 3  {
					return time.Time{}, fmt.Errorf("date column: %v could not be parsed", dateColumn)	
				} 
				if len(dateFormatSplit) != 3  {
					return time.Time{}, fmt.Errorf("date format: %v is invalid", dateFormat)	
				}
				var day string
				var month string
				var year string
				for i, val := range dateFormatSplit {
					switch val {
						case "2006":
							year = dateColumnSplit[i]
						case "01": 
							month = dateColumnSplit[i]
						case "02":
							day += dateColumnSplit[i]
					}
				}
				goDate := year + "-" + month + "-" + day
				recordFields[timestampColumn] = goDate + "T" + timeStampColumn + "Z"	
			}	
		}
		
	}
	if timestampColumn != "" {
		if recordFields[timestampColumn] == nil {
			return time.Time{}, fmt.Errorf("timestamp column: %v could not be found", timestampColumn)
		}
		//timestampFormat should remain complete; dateFormat does NOT affect timestamp format
		switch timestampFormat {
		case "":
			return time.Time{}, fmt.Errorf("timestamp format must be specified")
		default:
			metricTime, err := internal.ParseTimestamp(timestampFormat, recordFields[timestampColumn], "UTC")
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

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
	"github.com/araddon/dateparse"
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
	AltTimestamp       []string
	DefaultTags       map[string]string
	UniqueTimestamp   string
	TimeFunc          func() time.Time
	tsModder          *tsModder
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
	if p.UniqueTimestamp == "" {
		p.UniqueTimestamp = "auto"
	}
	csvReader.TrimLeadingSpace = p.TrimSpace
	return csvReader, nil
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	
	p.tsModder = &tsModder{}
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
			log.Printf("%d: ", i)
			headerNames := make([]string, 0)
			header, err := csvReader.Read()

			log.Printf("\t [%s] | columns: %d", header, len(header))
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
	log.Printf("record: %v", record)
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

	metricTime, err := parseTimestamp(p.TimeFunc, recordFields, p.TimestampColumn, p.TimestampFormat, p.AltTimestamp)
	if err != nil {
		return nil, err
	}
	
	if p.UniqueTimestamp == "auto" {
		//increment the metricTime to treat the current metric as a unique entry.
		m, err := metric.New(measurementName, tags, recordFields, p.tsModder.tsMod(metricTime))
		if err != nil {
			return nil, err
		}
		return m, nil
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
	timestampColumn, timestampFormat string, altTimestamp []string) (time.Time, error) {
	log.Printf("recordFields: %v", recordFields)
	log.Printf("recordFields[Date]: %v", recordFields["Date"])
	if len(altTimestamp) != 0 && timestampColumn == ""{
		newRecordFields := make(map[string]interface{})
		var altTimestampValues []string
		for _, columnName := range altTimestamp {
			//if recordFields[columnName] != nil {
			//	return time.Time{}, fmt.Errorf("column: %v could not be found", columnName)
			//}
			columnValue := fmt.Sprint(recordFields[columnName])
			altTimestampValues = append(altTimestampValues, columnValue)
		}
		t := strings.Join(altTimestampValues, " ")
		ts, err := dateparse.ParseLocal(t)
		
		newRecordFields["altTimestamp"] = ts.Format(timestampFormat)
		//Return format will be 2014-04-08 22:05:00 +0000 UTC
		if err != nil {
			return time.Time{}, fmt.Errorf("altTimestamp could not be parsed")
		}
		//Convert to format: 2014-04-08T22:05:00Z
		switch timestampFormat{
			case "":
				return time.Time{}, fmt.Errorf("timestamp format must be specified")
			default:	
				metricTime, err := internal.ParseTimestamp(timestampFormat, newRecordFields["altTimestamp"], "America/Toronto")
				if err != nil{
					return time.Time{}, err
				}
				return metricTime, err		
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


// tsModder is a struct for incrementing identical timestamps of log lines
// so that we don't push identical metrics that will get overwritten.
type tsModder struct {
	dupe     time.Time
	last     time.Time
	incr     time.Duration
	incrn    time.Duration
	rollover time.Duration
}

// tsMod increments the given timestamp one unit more from the previous
// duplicate timestamp.
// the increment unit is determined as the next smallest time unit below the
// most significant time unit of ts.
//   ie, if the input is at ms precision, it will increment it 1µs.
func (t *tsModder) tsMod(ts time.Time) time.Time {
	if ts.IsZero() {
		return ts
	}
	defer func() { t.last = ts }()
	
	// don't mod the time if we don't need to
	if t.last.IsZero() || ts.IsZero() {
		t.incrn = 0
		t.rollover = 0
		return ts
	}
	if !ts.Equal(t.last) && !ts.Equal(t.dupe) {
		t.incr = 0
		t.incrn = 0
		t.rollover = 0
		return ts
	}
	if ts.Equal(t.last) {
		t.dupe = ts
	}

	if ts.Equal(t.dupe) && t.incr == time.Duration(0) {
		tsNano := ts.UnixNano()

		d := int64(10)
		counter := 1
		for {
			a := tsNano % d
			if a > 0 {
				break
			}
			d = d * 10
			counter++
		}

		switch {
		case counter <= 6:
			t.incr = time.Nanosecond
		case counter <= 9:
			t.incr = time.Microsecond
		case counter > 9:
			t.incr = time.Millisecond
		}
	}

	t.incrn++
	if t.incrn == 999 && t.incr > time.Nanosecond {
		t.rollover = t.incr * t.incrn
		t.incrn = 1
		t.incr = t.incr / 1000
		if t.incr < time.Nanosecond {
			t.incr = time.Nanosecond
		}
	}
	return ts.Add(t.incr*t.incrn + t.rollover)
}

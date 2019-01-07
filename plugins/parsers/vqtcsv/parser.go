package vqtcsv

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type VqtCsvMetric struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]interface{}
	Timestamp   time.Time
}

type DLEquipmentModelObject struct {
	Site    string
	Fleet   string
	Metrics map[string]DLTagTypeObject
}

type DLTagTypeObject struct {
	PhdName string
	Name    string
	Systems []string
}

type VqtCsvParser struct {
	DefaultTags  map[string]string
	FieldReplace map[string]string
	Location     *time.Location
	Format       string

	metricName string
}

func NewVqtCsvMetric(measurement string, timestamp time.Time) *VqtCsvMetric {
	m := &VqtCsvMetric{
		Measurement: measurement,
		Fields:      make(map[string]interface{}),
		Tags:        make(map[string]string),
		Timestamp:   timestamp,
	}

	return m
}

func NewVqtCsvParser(format string, timezone string) (*VqtCsvParser, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}
	v := &VqtCsvParser{
		Location: loc,
		Format:   format,
	}

	return v, nil
}

func (p *VqtCsvParser) ProcessFull(csvline []string) (telegraf.Metric, error) {
	if len(csvline) < 4 {
		l := fmt.Sprintf("%s: Does not meet the length requirements", csvline)
		return nil, errors.New(l)
	}

	timestamp, err := time.ParseInLocation("02-01-2006 15:04:05.000", csvline[3], p.Location)
	if err != nil {
		log.Println("Error parsing time", csvline[3], err)
		return nil, err
	}

	currentMetric := NewVqtCsvMetric(p.metricName, timestamp)

	for i := 0; i < len(csvline); i += 4 {
		if (len(csvline)-i)%4 != 0 {
			continue
		}

		splitName := strings.Split(csvline[i], "(")
		idName := strings.Split(splitName[0], ".")
		fieldName := splitName[0]

		switch len(idName) {
		case 2:
			currentMetric.Measurement = idName[0]
			fieldName = idName[1]
			break
		case 1:
			fieldName = idName[0]
			break
		default:
			break
		}

		// Replace the field if configured
		if _, ok := p.FieldReplace[fieldName]; ok {
			fieldName = p.FieldReplace[fieldName]
		}

		// We for sure have a metric now. Get it!
		for i := 1; i < len(splitName); i++ {
			tag := splitName[i]
			tagClean := strings.TrimRight(tag, ")")
			tagPair := strings.Split(tagClean, "=")
			if len(tagPair) != 2 {
				log.Println("ERROR: [parse]: Expected )", csvline)
				continue
			}

			currentMetric.Tags[tagPair[0]] = tagPair[1]
		}

		var value interface{}
		//intval, err := strconv.ParseInt(csvline[i+1], 10, 64)
		//if err != nil {
		// Not an integer
		floatval, err := strconv.ParseFloat(csvline[i+1], 64)
		if err != nil {
			// Not a float. Use the string value then.
			value = string(csvline[i+1])
		} else {
			value = floatval
		}
		//} else {
		//	value = intval
		//}

		qualityInt, err := strconv.ParseInt(csvline[i+2], 10, 16)
		if err != nil {
			log.Println("Error parsing quality", csvline[i+2], err)
		}

		currentMetric.Fields["quality"] = qualityInt
		currentMetric.Fields[fieldName] = value
	}

	newMetric, err := metric.New(currentMetric.Measurement, currentMetric.Tags, currentMetric.Fields, currentMetric.Timestamp)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return newMetric, nil
}

func (p *VqtCsvParser) ProcessSimple(csvline []string) (telegraf.Metric, error) {
	if len(csvline) < 4 {
		l := fmt.Sprintf("%s: Does not meet the length requirements", csvline)
		return nil, errors.New(l)
	}

	timestamp, err := time.ParseInLocation("02-01-2006 15:04:05.000", csvline[3], p.Location)
	if err != nil {
		log.Println("Error parsing time", csvline[3], err)
		return nil, err
	}

	currentMetric := NewVqtCsvMetric(p.metricName, timestamp)

	for i := 0; i < len(csvline); i += 4 {
		if (len(csvline)-i)%4 != 0 {
			continue
		}

		fieldName := csvline[0]

		var value interface{}
		//intval, err := strconv.ParseInt(csvline[i+1], 10, 64)
		//if err != nil {
		// Not an integer
		floatval, err := strconv.ParseFloat(csvline[i+1], 64)
		if err != nil {
			// Not a float. Use the string value then.
			value = string(csvline[i+1])
		} else {
			value = floatval
		}
		//} else {
		//	value = intval
		//}

		qualityInt, err := strconv.ParseInt(csvline[i+2], 10, 16)
		if err != nil {
			log.Println("Error parsing quality", csvline[i+2], err)
		}

		currentMetric.Fields["quality"] = qualityInt
		currentMetric.Fields[fieldName] = value
	}

	newMetric, err := metric.New(currentMetric.Measurement, currentMetric.Tags, currentMetric.Fields, currentMetric.Timestamp)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return newMetric, nil
}

func (p *VqtCsvParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	l := len(buf)
	content := string(buf[:l])
	var metrics []telegraf.Metric

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		metric, err := p.ParseLine(line)

		if err != nil {
			log.Println(err, line)
			continue
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (p *VqtCsvParser) ParseLine(line string) (telegraf.Metric, error) {
	var ret telegraf.Metric
	var err error
	trimline := strings.Trim(line, "\r\n")
	csvline := strings.Split(trimline, ",")
	switch p.Format {
	case "simple":
		ret, err = p.ProcessSimple(csvline)
	case "full":
		ret, err = p.ProcessFull(csvline)
	}

	return ret, err
}

func (p *VqtCsvParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

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

func NewVqtCsvParser(metricName string, location *time.Location, fieldReplace map[string]string) (*VqtCsvParser, error) {
	v := &VqtCsvParser{
		metricName:   metricName,
		Location:     location,
		FieldReplace: fieldReplace,
	}

	return v, nil
}

func parseOpcQuality(qualityIn int64) string {
	qualityStr := "Good [Non-Specific]"
	switch qualityIn {
	case 0:
		qualityStr = "Bad [Non-Specific]"
		break
	case 4:
		qualityStr = "Bad [Configuration Error]"
		break
	case 8:
		qualityStr = "Bad [Not Connected]"
		break
	case 12:
		qualityStr = "Bad [Device Failure]"
		break
	case 16:
		qualityStr = "Bad [Sensor Failure]"
		break
	case 20:
		qualityStr = "Bad [Last Known Value]"
		break
	case 24:
		qualityStr = "Bad [Communication Failure]"
		break
	case 28:
		qualityStr = "Bad [Out of Service]"
		break
	case 64:
		qualityStr = "Uncertain [Non-Specific]"
		break
	case 65:
		qualityStr = "Uncertain [Non-Specific] (Low Limited)"
		break
	case 66:
		qualityStr = "Uncertain [Non-Specific] (High Limited)"
		break
	case 67:
		qualityStr = "Uncertain [Non-Specific] (Constant)"
		break
	case 68:
		qualityStr = "Uncertain [Last Usable]"
		break
	case 69:
		qualityStr = "Uncertain [Last Usable] (Low Limited)"
		break
	case 70:
		qualityStr = "Uncertain [Last Usable] (High Limited)"
		break
	case 71:
		qualityStr = "Uncertain [Last Usable] (Constant)"
		break
	case 80:
		qualityStr = "Uncertain [Sensor Not Accurate]"
		break
	case 81:
		qualityStr = "Uncertain [Sensor Not Accurate] (Low Limited)"
		break
	case 82:
		qualityStr = "Uncertain [Sensor Not Accurate] (High Limited)"
		break
	case 83:
		qualityStr = "Uncertain [Sensor Not Accurate] (Constant)"
		break
	case 84:
		qualityStr = "Uncertain [EU Exceeded]"
		break
	case 85:
		qualityStr = "Uncertain [EU Exceeded] (Low Limited)"
		break
	case 86:
		qualityStr = "Uncertain [EU Exceeded] (High Limited)"
		break
	case 87:
		qualityStr = "Uncertain [EU Exceeded] (Constant)"
		break
	case 88:
		qualityStr = "Uncertain [Sub-Normal]"
		break
	case 89:
		qualityStr = "Uncertain [Sub-Normal] (Low Limited)"
		break
	case 90:
		qualityStr = "Uncertain [Sub-Normal] (High Limited)"
		break
	case 91:
		qualityStr = "Uncertain [Sub-Normal] (Constant)"
		break
	case 192:
		qualityStr = "Good [Non-Specific]"
		break
	case 193:
		qualityStr = "Good [Non-Specific] (Low Limited)"
		break
	case 194:
		qualityStr = "Good [Non-Specific] (High Limited)"
		break
	case 195:
		qualityStr = "Good [Non-Specific] (Constant)"
		break
	case 216:
		qualityStr = "Good [Local Override]"
		break
	case 217:
		qualityStr = "Good [Local Override] (Low Limited)"
		break
	case 218:
		qualityStr = "Good [Local Override] (High Limited)"
		break
	case 219:
		qualityStr = "Good [Local Override] (Constant)"
		break
	}

	return qualityStr
}

func (p *VqtCsvParser) Process(csvline []string) (telegraf.Metric, error) {
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
				log.Println("ERROR: [parse]: Expected )")
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

		currentMetric.Tags["quality"] = parseOpcQuality(qualityInt)
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
	trimline := strings.Trim(line, "\r\n")
	csvline := strings.Split(trimline, ",")
	return p.Process(csvline)
}

func (p *VqtCsvParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

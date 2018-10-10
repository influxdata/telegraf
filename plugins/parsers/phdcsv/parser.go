package phdcsv

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/srclosson/telegraf/metric"
)

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

type PhdCsvData struct {
	Equipment string
	Metric    string
}

type PhdCsvParser struct {
	DefaultTags map[string]string

	phdModel map[string]PhdCsvData
	model    map[string]DLEquipmentModelObject
	acc      telegraf.Accumulator
}

func NewPhdCsvParser(acc telegraf.Accumulator) (*PhdCsvParser, error) {

	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)

	v := &PhdCsvParser{
		acc: acc,
	}

	return v, nil
}

func parseConfidence(qualityIn int) string {
	qualityStr := "Good [Non-Specific]"
	switch qualityIn {
	case 100:
		qualityStr = "Good [Non-Specific]"
		break
	default:
		qualityStr = "Bad [Non-Specific]"
		break
	}

	return qualityStr
}

func (p *PhdCsvParser) ProcessLine(line string) (telegraf.Metric, error) {

	var value interface{}
	entry := strings.Split(line, ",")
	if len(entry) != 4 {
		return nil, errors.New("Expected 4 columns but did get four")
	}

	epoch, err := strconv.Atoi(entry[0])
	if err != nil {
		return nil, err
	}

	confidence, err := strconv.Atoi(entry[3])
	if err != nil {
		return nil, err
	}

	timestamp := time.Unix(int64(epoch), 0)

	phdTagName := entry[1]
	splitTag := strings.Split(phdTagName, ".")
	if len(splitTag) != 2 {
		return nil, err
	}

	measurement := splitTag[0]
	equipment := splitTag[0]
	fieldName := splitTag[1]

	tags := make(map[string]string)
	tags["Equipment"] = equipment
	tags["Quality"] = parseConfidence(confidence)
	tags["Site"] = equipment[0:3]

	parsedVal, err := strconv.ParseFloat(entry[2], 64)
	if err != nil {
		// Not a float. Use the string value then.
		value = entry[2]
	} else {
		value = parsedVal
	}

	fields := make(map[string]interface{})
	fields[fieldName] = value

	newMetric, err := metric.New(measurement, tags, fields, timestamp)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return newMetric, nil
}

func (p *PhdCsvParser) Process(content string) ([]telegraf.Metric, error) {
	lines := strings.Split(content, "\n")
	var metrics []telegraf.Metric

	for _, line := range lines {
		metric, err := p.ProcessLine(line)
		if err != nil {
			log.Println("ERROR: [process.line]: ", err)
			continue
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (p *PhdCsvParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	l := len(buf)
	fileName := string(buf[:l])

	fmt.Printf("Parsing phdcsv: %s\n", fileName)

	p.Process(fileName)

	return nil, nil
}

func (p *PhdCsvParser) ParseLine(line string) (telegraf.Metric, error) {
	return p.ProcessLine(line)
}

func (p *PhdCsvParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

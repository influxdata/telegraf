package phdcsv

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
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

func (p *PhdCsvParser) Process(fileName string) ([]telegraf.Metric, error) {
	f, err := os.Open(fileName)
	if err != nil {
		// If we can't open the file... ignore and move on
		return nil, nil
	}
	defer f.Close()

	cr := csv.NewReader(f)

	err = nil
	var entry []string
	for entry, err = cr.Read(); err != io.EOF; entry, err = cr.Read() {
		if err != nil {
			log.Println("ERROR: [process.line]: ", err)
			continue
		}

		var value interface{}

		epoch, err := strconv.Atoi(entry[0])
		if err != nil {
			continue
		}

		confidence, err := strconv.Atoi(entry[3])
		if err != nil {
			continue
		}

		timestamp := time.Unix(int64(epoch), 0)

		phdTagName := entry[1]
		splitTag := strings.Split(phdTagName, ".")
		if len(splitTag) != 2 {
			continue
		}

		newMetric := p.phdModel[phdTagName].Equipment
		equipment := p.phdModel[phdTagName].Equipment
		fieldName := p.phdModel[phdTagName].Metric

		if len(newMetric) == 0 {
			newMetric = splitTag[0]
			equipment = splitTag[0]
			fieldName = splitTag[1]
		}

		tags := make(map[string]string)
		tags["Equipment"] = equipment
		tags["Quality"] = parseConfidence(confidence)
		tags["Site"] = equipment[0:3]

		// Substitute the model tag config
		if val, ok := p.model[equipment].Metrics[fieldName]; ok {
			tags["description"] = val.Name

			var tagName = "System"
			for _, system := range val.Systems {
				tags[tagName] = system
				tagName = "Sub" + tagName
			}
		}

		parsedVal, err := strconv.ParseFloat(entry[2], 64)
		if err != nil {
			// Not a float. Use the string value then.
			value = entry[2]
		} else {
			value = parsedVal
		}

		fields := make(map[string]interface{})
		fields[fieldName] = value

		p.acc.AddFields(newMetric, fields, tags, timestamp)
	}

	return nil, nil
}

func (p *PhdCsvParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	l := len(buf)
	fileName := string(buf[:l])

	fmt.Printf("Parsing phdcsv: %s\n", fileName)

	p.Process(fileName)

	return nil, nil
}

func (p *PhdCsvParser) ParseLine(line string) (telegraf.Metric, error) {
	return nil, nil
}

func (p *PhdCsvParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

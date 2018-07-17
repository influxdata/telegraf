package csv

import (
	"bytes"
	"encoding/csv"

	"github.com/influxdata/telegraf"
)

type CSVParser struct {
	MetricName      string
	Delimiter       string
	DataColumns     []string
	TagColumns      []string
	FieldColumns    []string
	NameColumn      string
	TimestampColumn string
	DefaultTags     map[string]string
	csvReader       *csv.Reader
}

func (p *CSVParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	r := bytes.NewReader(buf)
	p.csvReader = csv.NewReader(r)
}

func (p *CSVParser) ParseLine(line string) (telegraf.Metric, error) {

}

func (p *CSVParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

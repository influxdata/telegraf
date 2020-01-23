package avro

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"fmt"

    "github.com/linkedin/goavro"
)

type Parser struct {
	DefaultTags       map[string]string
	TimeFunc          func() time.Time
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	codec, err := goavro.NewCodec(`
        {"namespace": "com.example.plant", "type": "record", "name": "Value", "fields": [{"name": "value", "type": "long"}]}`)
    
    if err != nil {
        fmt.Println(err)
    }

	fmt.Println(buf)

    // Convert binary Avro data back to native Go form
    native, _, err := codec.NativeFromBinary(buf[5:])
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println(native)

    // Convert native Go form to textual Avro data
    textual, err := codec.TextualFromNative(nil, native)
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println(textual)

	return p.createMeasures(0)
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	return p.createMeasure(0)
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) createMeasure(value int) (telegraf.Metric, error) {
	recordFields := make(map[string]interface{})
	tags := make(map[string]string)

	for k, v := range p.DefaultTags {
		tags[k] = v
	}

	tags["tagName"] = "tagValue"

	recordFields["value"] = value

	measurementName := "measurementName"

	metricTime := p.TimeFunc()

	m, err := metric.New(measurementName, tags, recordFields, metricTime)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (p *Parser) createMeasures(value int) ([]telegraf.Metric, error) {
	metrics := make([]telegraf.Metric, 0)

	m, err := p.createMeasure(value)
	if err != nil {
		return metrics, err
	}
	metrics = append(metrics, m)
	
	return metrics, nil
}
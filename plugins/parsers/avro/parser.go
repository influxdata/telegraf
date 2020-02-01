package avro

import (
	"time"
	"encoding/binary"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"fmt"

    "github.com/linkedin/goavro"
)


type Parser struct {
	Measurement 	  string
	Tags 			  []string
	Fields 			  []string
	Timestamp 		  string
	DefaultTags       map[string]string
	TimeFunc          func() time.Time
}



func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {

	fmt.Println(p.Measurement)
	fmt.Println(p.Tags)
	fmt.Println(p.Fields)
	fmt.Println(p.Timestamp)

	schemaRegistry := NewSchemaRegistry("http:localhost:8081")
	
	schemaId := int(binary.BigEndian.Uint32(buf[1:5]))
	fmt.Println(schemaId)

	schema, err := schemaRegistry.getSchema(schemaId)
	if err != nil {
        fmt.Println(err)
    }
	fmt.Println(schema)

	codec, err := goavro.NewCodec(schema)
    if err != nil {
        fmt.Println(err)
    }

    // Convert binary Avro data back to native Go form
    native, _, err := codec.NativeFromBinary(buf[5:])
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println(native)

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
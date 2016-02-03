package encoding

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/encoding/graphite"
)

type Parser struct {
	graphiteParser *graphite.Parser
}

func NewParser(parser *graphite.Parser) *Parser {
	return &Parser{graphiteParser: parser}
}

func (p *Parser) Parse(dataFormat string, out []byte, acc telegraf.Accumulator) error {
	var err error
	var metrics []telegraf.Metric
	var metric telegraf.Metric

	switch dataFormat {
	case "","json":
		var jsonOut interface{}
		err = json.Unmarshal(out, &jsonOut)
		if err != nil {
			err = fmt.Errorf("unable to parse out as JSON, %s", err)
			break
		}

		f := internal.JSONFlattener{}
		err = f.FlattenJSON("", jsonOut)
		if err != nil {
			break
		}
		acc.AddFields("exec", f.Fields, nil)
	case "influx":
		now := time.Now()
		metrics, err = telegraf.ParseMetrics(out)
		for _, metric = range metrics {
			acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), now)
		}
	case "graphite":
		metrics, err = p.graphiteParser.ParseMetrics(out)
		for _, metric = range metrics {
			acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
		}
	default:
		err = fmt.Errorf("Unsupported data format: %s. Must be either json, influx or graphite ", dataFormat)
	}

	return err
}

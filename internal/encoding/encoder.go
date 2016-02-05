package encoding

import (
	"fmt"

	"github.com/influxdata/telegraf"
)

type Parser interface {
	InitConfig(configs map[string]interface{}) error
	Parse(buf []byte) ([]telegraf.Metric, error)
	ParseLine(line string) (telegraf.Metric, error)
}

type Creator func() Parser

var Parsers = map[string]Creator{}

func Add(name string, creator Creator) {
	Parsers[name] = creator
}

func NewParser(dataFormat string, configs map[string]interface{}) (parser Parser, err error) {
	creator := Parsers[dataFormat]
	if creator == nil {
		return nil, fmt.Errorf("Unsupported data format: %s. ", dataFormat)
	}
	parser = creator()
	err = parser.InitConfig(configs)
	return parser, err
}

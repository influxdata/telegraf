package concat

import (
	"bytes"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Parser struct {
	Splitter Splitter `toml:"split"`

	parser telegraf.Parser
	buffer bytes.Buffer
}

// SetParser sets the embedded parser applied to each extracted message. It
// implements telegraf.ParserPlugin so the config can build the parser from the
// embedded_parser sub-table.
func (p *Parser) SetParser(parser telegraf.Parser) {
	p.parser = parser
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	p.buffer.Write(buf)
	var metrics []telegraf.Metric
	for {
		advance, msg, err := p.Splitter.Split(p.buffer.Bytes(), false)
		if err != nil {
			return metrics, err
		}
		if advance == 0 {
			// Not enough data for a complete message yet.
			break
		}
		if msg != nil {
			m := make([]byte, len(msg))
			copy(m, msg)
			metricsPart, err := p.parser.Parse(m)
			if err != nil {
				p.buffer.Next(advance)
				return metrics, err
			}
			metrics = append(metrics, metricsPart...)
		}
		p.buffer.Next(advance)
	}
	return metrics, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	return p.parser.ParseLine(line)
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.parser.SetDefaultTags(tags)
}

func (p *Parser) Init() error {
	return p.Splitter.Init()
}

func init() {
	parsers.Add("concat", func(string) telegraf.Parser {
		return &Parser{}
	})
}

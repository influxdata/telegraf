package wavefront

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
)

const maxBufferSize = 2

type Parser struct {
	DefaultTags map[string]string `toml:"-"`
	Log         telegraf.Logger   `toml:"-"`

	parsers  *sync.Pool
	timeFunc func() time.Time
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

func (p *Parser) SetTimeFunc(f func() time.Time) {
	p.timeFunc = f
	p.parsers = &sync.Pool{New: newPointParserFactory(p)}
}

func (p *Parser) Init() error {
	if p.timeFunc == nil {
		p.timeFunc = time.Now
	}

	p.parsers = &sync.Pool{New: newPointParserFactory(p)}
	return nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	buf := []byte(line)

	metrics, err := p.Parse(buf)
	if err != nil {
		return nil, err
	}

	if len(metrics) > 0 {
		return metrics[0], nil
	}

	return nil, nil
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	pp := p.parsers.Get().(*pointParser)
	defer p.parsers.Put(pp)
	return pp.Parse(buf)
}

func init() {
	parsers.Add("wavefront",
		func(string) telegraf.Parser {
			return &Parser{}
		})
}

package benchmark

import (
	"context"
	"io/ioutil"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Benchmark struct {
	Filename string `toml:"filename"`

	parser parsers.Parser
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

func (i *Benchmark) SampleConfig() string {
	return `
  ## File containing input data
  filename = "/tmp/testdata"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`
}

func (i *Benchmark) Description() string {
	return "Generate test data for performance testing"
}

func (i *Benchmark) SetParser(p parsers.Parser) {
	i.parser = p
}

func (i *Benchmark) Start(acc telegraf.Accumulator) error {
	octets, err := ioutil.ReadFile(i.Filename)
	if err != nil {
		return err
	}

	metrics, err := i.parser.Parse(octets)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	i.cancel = cancel
	i.wg.Add(1)
	go func() {
		defer func() {
			i.cancel()
			i.wg.Done()
		}()

		err := i.add(ctx, acc, metrics)
		if err != nil {
			acc.AddError(err)
		}
	}()

	return nil
}

func (i *Benchmark) add(ctx context.Context, acc telegraf.Accumulator, metrics []telegraf.Metric) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			for _, m := range metrics {
				acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
			}
		}
	}
}

func (i *Benchmark) Stop() {
	if i.cancel != nil {
		i.cancel()
	}
	i.wg.Wait()
}

func (i *Benchmark) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("benchmark", func() telegraf.Input {
		return &Benchmark{}
	})
}

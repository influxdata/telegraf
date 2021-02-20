package shim

import (
	"fmt"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/processors"
)

// AddProcessor adds the processor to the shim. Later calls to Run() will run this.
func (s *Shim) AddProcessor(processor telegraf.Processor) error {
	setLoggerOnPlugin(processor, s.Log())
	p := processors.NewStreamingProcessorFromProcessor(processor)
	return s.AddStreamingProcessor(p)
}

// AddStreamingProcessor adds the processor to the shim. Later calls to Run() will run this.
func (s *Shim) AddStreamingProcessor(processor telegraf.StreamingProcessor) error {
	setLoggerOnPlugin(processor, s.Log())
	if p, ok := processor.(telegraf.Initializer); ok {
		err := p.Init()
		if err != nil {
			return fmt.Errorf("failed to init input: %s", err)
		}
	}

	s.Processor = processor
	return nil
}

func (s *Shim) RunProcessor() error {
	acc := agent.NewAccumulator(s, s.metricCh)
	acc.SetPrecision(time.Nanosecond)

	err := s.Processor.Start(acc)
	if err != nil {
		return fmt.Errorf("failed to start processor: %w", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		s.writeProcessedMetrics()
		wg.Done()
	}()

	parser := influx.NewStreamParser(s.stdin)
	for {
		m, err := parser.Next()
		if err != nil {
			if err == influx.EOF {
				break // stream ended
			}
			if parseErr, isParseError := err.(*influx.ParseError); isParseError {
				fmt.Fprintf(s.stderr, "Failed to parse metric: %s\b", parseErr)
				continue
			}
			fmt.Fprintf(s.stderr, "Failure during reading stdin: %s\b", err)
			continue
		}

		s.Processor.Add(m, acc)
	}

	close(s.metricCh)
	s.Processor.Stop()
	wg.Wait()
	return nil
}

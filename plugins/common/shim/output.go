package shim

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
)

// AddOutput adds the input to the shim. Later calls to Run() will run this.
func (s *Shim) AddOutput(output telegraf.Output) error {
	models.SetLoggerOnPlugin(output, s.Log())
	if p, ok := output.(telegraf.Initializer); ok {
		err := p.Init()
		if err != nil {
			return fmt.Errorf("failed to init input: %w", err)
		}
	}

	s.Output = output
	return nil
}

func (s *Shim) RunOutput() error {
	// Create a parser for receiving the metrics in line-protocol format
	parser := influx.Parser{}
	if err := parser.Init(); err != nil {
		return fmt.Errorf("failed to create new parser: %w", err)
	}

	// Connect the output
	if err := s.Output.Connect(); err != nil {
		return fmt.Errorf("failed to start processor: %w", err)
	}
	defer s.Output.Close()

	// Collect the metrics from stdin. Note, we need to flush the metrics
	// when the batch is full or after the configured time, whatever comes
	// first. We need to lock the batch as we run into race conditions
	// otherwise.
	var mu sync.Mutex
	metrics := make([]telegraf.Metric, 0, s.BatchSize)

	// Prepare the flush timer...
	flush := func(whole bool) {
		mu.Lock()
		defer mu.Unlock()

		// Exit early if there is nothing to do
		if len(metrics) == 0 {
			return
		}

		// Determine the threshold on when to stop flushing depending on the
		// given flag.
		var threshold int
		if whole {
			threshold = s.BatchSize
		}

		// Flush out the metrics in batches of the configured size until we
		// got all of them out or if there is less than a whole batch left.
		for len(metrics) > 0 && len(metrics) >= threshold {
			// Write the metrics and remove the batch
			batch := metrics[:min(len(metrics), s.BatchSize)]
			if err := s.Output.Write(batch); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write metrics: %s\n", err)
			}
			metrics = metrics[len(batch):]
		}
	}

	// Setup the time-based flush
	var timer *time.Timer
	if s.BatchTimeout > 0 {
		timer = time.AfterFunc(s.BatchTimeout, func() { flush(false) })
		defer func() {
			if timer != nil {
				timer.Stop()
			}
		}()
	}

	// Start the processing loop
	scanner := bufio.NewScanner(s.stdin)
	for scanner.Scan() {
		// Read metrics from stdin
		m, err := parser.ParseLine(scanner.Text())
		if err != nil {
			fmt.Fprintf(s.stderr, "Failed to parse metric: %s\n", err)
			continue
		}
		mu.Lock()
		metrics = append(metrics, m)
		shouldFlush := len(metrics) >= s.BatchSize
		mu.Unlock()

		// If we got more enough metrics to fill the batch flush it out and
		// reset the time-based guard.
		if shouldFlush {
			if timer != nil {
				timer.Stop()
			}
			flush(true)
			if s.BatchTimeout > 0 {
				timer = time.AfterFunc(s.BatchTimeout, func() { flush(false) })
			}
		}
	}

	// Output all remaining metrics
	if timer != nil {
		timer.Stop()
	}
	flush(false)

	return nil
}

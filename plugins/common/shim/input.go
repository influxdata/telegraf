package shim

import (
	"bufio"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
)

// AddInput adds the input to the shim. Later calls to Run() will run this input.
func (s *Shim) AddInput(input telegraf.Input) error {
	setLoggerOnPlugin(input, s.Log())
	if p, ok := input.(telegraf.Initializer); ok {
		err := p.Init()
		if err != nil {
			return fmt.Errorf("failed to init input: %s", err)
		}
	}

	s.Input = input
	return nil
}

func (s *Shim) RunInput(pollInterval time.Duration) error {
	// context is used only to close the stdin reader. everything else cascades
	// from that point and closes cleanly when it's done.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.watchForShutdown(cancel)

	acc := agent.NewAccumulator(s, s.metricCh)
	acc.SetPrecision(time.Nanosecond)

	if serviceInput, ok := s.Input.(telegraf.ServiceInput); ok {
		if err := serviceInput.Start(acc); err != nil {
			return fmt.Errorf("failed to start input: %s", err)
		}
	}
	s.gatherPromptCh = make(chan empty, 1)
	go func() {
		s.startGathering(ctx, s.Input, acc, pollInterval)
		if serviceInput, ok := s.Input.(telegraf.ServiceInput); ok {
			serviceInput.Stop()
		}
		// closing the metric channel gracefully stops writing to stdout
		close(s.metricCh)
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		s.writeProcessedMetrics()
		wg.Done()
	}()

	go func() {
		scanner := bufio.NewScanner(s.stdin)
		for scanner.Scan() {
			// push a non-blocking message to trigger metric collection.
			s.pushCollectMetricsRequest()
		}

		cancel() // cancel gracefully stops gathering
	}()

	wg.Wait() // wait for writing to stdout to finish
	return nil
}

func (s *Shim) startGathering(ctx context.Context, input telegraf.Input, acc telegraf.Accumulator, pollInterval time.Duration) {
	if pollInterval == PollIntervalDisabled {
		pollInterval = forever
	}
	t := time.NewTicker(pollInterval)
	defer t.Stop()
	for {
		// give priority to stopping.
		if hasQuit(ctx) {
			return
		}
		// see what's up
		select {
		case <-ctx.Done():
			return
		case <-s.gatherPromptCh:
			if err := input.Gather(acc); err != nil {
				fmt.Fprintf(s.stderr, "failed to gather metrics: %s\n", err)
			}
		case <-t.C:
			if err := input.Gather(acc); err != nil {
				fmt.Fprintf(s.stderr, "failed to gather metrics: %s\n", err)
			}
		}
	}
}

// pushCollectMetricsRequest pushes a non-blocking (nil) message to the
// gatherPromptCh channel to trigger metric collection.
// The channel is defined with a buffer of 1, so while it's full, subsequent
// requests are discarded.
func (s *Shim) pushCollectMetricsRequest() {
	// push a message out to each channel to collect metrics. don't block.
	select {
	case s.gatherPromptCh <- empty{}:
	default:
	}
}

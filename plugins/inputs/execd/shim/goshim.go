package shim

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

type empty struct{}

var (
	gatherPromptChans []chan empty
	stdout            io.Writer = os.Stdout
	stdin             io.Reader = os.Stdin
)

const (
	// PollIntervalDisabled is used to indicate that you want to disable polling,
	// as opposed to duration 0 meaning poll constantly.
	PollIntervalDisabled = time.Duration(0)
)

// RunPlugins runs the input plugins..
func RunPlugins(cfg *config.Config, pollInterval time.Duration) error {
	wg := sync.WaitGroup{}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	collectMetricsPrompt := make(chan os.Signal, 1)
	listenForCollectMetricsSignals(collectMetricsPrompt)

	wg.Add(1) // wait for the metric channel to close
	metricCh := make(chan telegraf.Metric, 1)

	s := influx.NewSerializer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, runningInput := range cfg.Inputs {
		if err := runningInput.Init(); err != nil {
			return fmt.Errorf("failed to init input: %w", err)
		}

		acc := agent.NewAccumulator(runningInput, metricCh)
		acc.SetPrecision(time.Nanosecond)

		if serviceInput, ok := runningInput.Input.(telegraf.ServiceInput); ok {
			wg.Add(1)
			if err := serviceInput.Start(acc); err != nil {
				return fmt.Errorf("failed to start input: %w", err)
			}
		}
		gatherPromptCh := make(chan empty, 1)
		gatherPromptChans = append(gatherPromptChans, gatherPromptCh)
		wg.Add(1)
		go func(i int) {
			startGathering(ctx, cfg.Inputs[i], acc, gatherPromptCh, pollInterval)
			wg.Done()
		}(i)
	}

	go stdinCollectMetricsPrompt(ctx, collectMetricsPrompt)

	hasQuit := false
loop:
	for {
		select {
		case <-quit:
			cancel()
			stopServices(&wg, cfg)
			hasQuit = true
			// keep looping until the metric channel closes.
		case <-collectMetricsPrompt:
			if !hasQuit {
				collectMetrics()
			}
		case m, open := <-metricCh:
			if !open {
				wg.Done()
				break loop
			}
			b, err := s.Serialize(m)
			if err != nil {
				return fmt.Errorf("failed to serialize metric: %w", err)
			}
			// Write this to stdout
			fmt.Fprint(stdout, string(b))
		}
	}

	wg.Wait()
	return nil
}

func stopServices(wg *sync.WaitGroup, cfg *config.Config) {
	for _, runningInput := range cfg.Inputs {
		if serviceInput, ok := runningInput.Input.(telegraf.ServiceInput); ok {
			serviceInput.Stop()
			wg.Done()
		}
	}
}

func stdinCollectMetricsPrompt(ctx context.Context, collectMetricsPrompt chan<- os.Signal) {
	s := bufio.NewScanner(stdin)
	// for every line read from stdin, make sure we're not supposed to quit,
	// then push a message on to the collectMetricsPrompt
	for s.Scan() {
		// first check if we should quit
		select {
		case <-ctx.Done():
			return
		default:
		}

		// now push a non-blocking message to trigger metric collection.
		// The channel is defined with a buffer of 1, so if it blocks, that means
		// that there's already a prompt waiting to be processed, and we don't need
		// to push a second one.
		select {
		case collectMetricsPrompt <- nil:
		default:
		}
	}
}

func collectMetrics() {
	for i := 0; i < len(gatherPromptChans); i++ {
		// push a message out to each channel to collect metrics. don't block.
		select {
		case gatherPromptChans[i] <- empty{}:
		default:
		}
	}
}

func startGathering(ctx context.Context, runningInput *models.RunningInput, acc telegraf.Accumulator, gatherPromptCh <-chan empty, pollInterval time.Duration) {
	if pollInterval == PollIntervalDisabled {
		return // don't poll
	}
	t := time.NewTicker(pollInterval)
	defer t.Stop()
	for {
		// give priority to stopping.
		select {
		case <-ctx.Done():
			return
		default:
		}
		// see what's up
		select {
		case <-ctx.Done():
			return
		case <-gatherPromptCh:
			if err := runningInput.Gather(acc); err != nil {
				fmt.Fprintf(os.Stderr, "failed to gather metrics: %s", err)
			}
		case <-t.C:
			if err := runningInput.Gather(acc); err != nil {
				fmt.Fprintf(os.Stderr, "failed to gather metrics: %s", err)
			}
		}
	}
}

package shim

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

type empty struct{}

var gatherPromptChans []chan empty

const (
	PollIntervalDisabled = time.Duration(0)
)

func RunPlugins(cfg *config.Config, pollInterval time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	collectMetricsPrompt := make(chan os.Signal, 1)
	signal.Notify(collectMetricsPrompt, syscall.SIGUSR1, syscall.SIGHUP)

	metricCh := make(chan telegraf.Metric, 1)

	s := influx.NewSerializer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, runningInput := range cfg.Inputs {
		if err := runningInput.Init(); err != nil {
			handleErr(err)
		}
		acc := agent.NewAccumulator(runningInput, metricCh)
		acc.SetPrecision(time.Nanosecond)

		if serviceInput, ok := runningInput.Input.(telegraf.ServiceInput); ok {
			if err := serviceInput.Start(acc); err != nil {
				handleErr(err)
			}
		} else {
			gatherPromptCh := make(chan empty, 1)
			gatherPromptChans = append(gatherPromptChans, gatherPromptCh)
			go startGathering(ctx, cfg.Inputs[i], acc, gatherPromptCh, pollInterval)
		}
	}

	go stdinCollectMetricsPrompt(ctx, collectMetricsPrompt)

	for {
		select {
		case <-quit:
			cancel()
			os.Stdin.Close()
			stopServices(cfg)
			return
		case <-collectMetricsPrompt:
			collectMetrics()
		case m := <-metricCh:
			b, err := s.Serialize(m)
			if err != nil {
				handleErr(err)
			}
			fmt.Print(string(b))
		}
	}

	// TODO: wait for stop?
}

func stopServices(cfg *config.Config) {
	for _, runningInput := range cfg.Inputs {
		if serviceInput, ok := runningInput.Input.(telegraf.ServiceInput); ok {
			serviceInput.Stop()
		}
	}
}

func stdinCollectMetricsPrompt(ctx context.Context, collectMetricsPrompt chan<- os.Signal) {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		select { // don't block.
		case <-ctx.Done():
			return
		case collectMetricsPrompt <- nil:
		default:
		}
	}
}

func collectMetrics() {
	for i := 0; i < len(gatherPromptChans); i++ {
		// don't block
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
		select {
		case <-ctx.Done():
			return
		case <-gatherPromptCh:
			if err := runningInput.Gather(acc); err != nil {
				handleErr(err)
			}
		case <-t.C:
			if err := runningInput.Gather(acc); err != nil {
				handleErr(err)
			}
		}
	}
}

func handleErr(err error) {
	fmt.Fprintf(os.Stderr, "Err: %s\n", err)
	os.Exit(1)
}

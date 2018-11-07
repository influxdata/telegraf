package agent

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/config"
	"github.com/influxdata/telegraf/internal/models"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

// Agent runs a set of plugins.
type Agent struct {
	Config *config.Config
}

// NewAgent returns an Agent for the given Config.
func NewAgent(config *config.Config) (*Agent, error) {
	a := &Agent{
		Config: config,
	}
	return a, nil
}

// Run starts and runs the Agent until the context is done.
func (a *Agent) Run(ctx context.Context) error {
	log.Printf("I! [agent] Config: Interval:%s, Quiet:%#v, Hostname:%#v, "+
		"Flush Interval:%s",
		a.Config.Agent.Interval.Duration, a.Config.Agent.Quiet,
		a.Config.Agent.Hostname, a.Config.Agent.FlushInterval.Duration)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	log.Printf("D! [agent] Connecting outputs")
	err := a.connectOutputs(ctx)
	if err != nil {
		return err
	}

	inputC := make(chan telegraf.Metric, 100)
	procC := make(chan telegraf.Metric, 100)
	outputC := make(chan telegraf.Metric, 100)

	startTime := time.Now()

	log.Printf("D! [agent] Starting service inputs")
	err = a.startServiceInputs(ctx, inputC)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	src := inputC
	dst := inputC

	wg.Add(1)
	go func(dst chan telegraf.Metric) {
		defer wg.Done()

		err := a.runInputs(ctx, startTime, dst)
		if err != nil {
			log.Printf("E! [agent] Error running inputs: %v", err)
		}

		log.Printf("D! [agent] Stopping service inputs")
		a.stopServiceInputs()

		close(dst)
		log.Printf("D! [agent] Input channel closed")
	}(dst)

	src = dst

	if len(a.Config.Processors) > 0 {
		dst = procC

		wg.Add(1)
		go func(src, dst chan telegraf.Metric) {
			defer wg.Done()

			err := a.runProcessors(src, dst)
			if err != nil {
				log.Printf("E! [agent] Error running processors: %v", err)
			}
			close(dst)
			log.Printf("D! [agent] Processor channel closed")
		}(src, dst)

		src = dst
	}

	if len(a.Config.Aggregators) > 0 {
		dst = outputC

		wg.Add(1)
		go func(src, dst chan telegraf.Metric) {
			defer wg.Done()

			err := a.runAggregators(startTime, src, dst)
			if err != nil {
				log.Printf("E! [agent] Error running aggregators: %v", err)
			}
			close(dst)
			log.Printf("D! [agent] Output channel closed")
		}(src, dst)

		src = dst
	}

	wg.Add(1)
	go func(src chan telegraf.Metric) {
		defer wg.Done()

		err := a.runOutputs(startTime, src)
		if err != nil {
			log.Printf("E! [agent] Error running outputs: %v", err)
		}
	}(src)

	wg.Wait()

	log.Printf("D! [agent] Closing outputs")
	err = a.closeOutputs()
	if err != nil {
		return err
	}

	return nil
}

// Test runs the inputs once and prints the output to stdout in line protocol.
func (a *Agent) Test() error {
	var wg sync.WaitGroup
	metricC := make(chan telegraf.Metric)
	defer func() {
		close(metricC)
		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		s := influx.NewSerializer()
		s.SetFieldSortOrder(influx.SortFields)
		for metric := range metricC {
			octets, err := s.Serialize(metric)
			if err == nil {
				fmt.Print("> ", string(octets))
			}
		}
	}()

	for _, input := range a.Config.Inputs {
		if _, ok := input.Input.(telegraf.ServiceInput); ok {
			log.Printf("W!: [agent] skipping plugin [[%s]]: service inputs not supported in --test mode",
				input.Name())
			continue
		}

		acc := NewAccumulator(input, metricC)
		acc.SetPrecision(a.Config.Agent.Precision.Duration,
			a.Config.Agent.Interval.Duration)
		input.SetDefaultTags(a.Config.Tags)

		if err := input.Input.Gather(acc); err != nil {
			return err
		}

		// Special instructions for some inputs. cpu, for example, needs to be
		// run twice in order to return cpu usage percentages.
		switch input.Name() {
		case "inputs.cpu", "inputs.mongodb", "inputs.procstat":
			time.Sleep(500 * time.Millisecond)
			if err := input.Input.Gather(acc); err != nil {
				return err
			}
		}

	}

	return nil
}

// runInputs starts and triggers the periodic gather for Inputs.
//
// When the context is done the timers are stopped and this function returns
// after all ongoing Gather calls complete.
func (a *Agent) runInputs(
	ctx context.Context,
	startTime time.Time,
	dst chan<- telegraf.Metric,
) error {
	var wg sync.WaitGroup
	for _, input := range a.Config.Inputs {
		interval := a.Config.Agent.Interval.Duration
		precision := a.Config.Agent.Precision.Duration
		jitter := a.Config.Agent.CollectionJitter.Duration

		// Overwrite agent interval if this plugin has its own.
		if input.Config.Interval != 0 {
			interval = input.Config.Interval
		}

		acc := NewAccumulator(input, dst)
		acc.SetPrecision(precision, interval)

		wg.Add(1)
		go func(input *models.RunningInput) {
			defer wg.Done()

			if a.Config.Agent.RoundInterval {
				err := internal.SleepContext(
					ctx, internal.AlignDuration(startTime, interval))
				if err != nil {
					return
				}
			}

			a.gatherOnInterval(ctx, acc, input, interval, jitter)
		}(input)
	}
	wg.Wait()

	return nil
}

// gather runs an input's gather function periodically until the context is
// done.
func (a *Agent) gatherOnInterval(
	ctx context.Context,
	acc telegraf.Accumulator,
	input *models.RunningInput,
	interval time.Duration,
	jitter time.Duration,
) {
	defer panicRecover(input)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		err := internal.SleepContext(ctx, internal.RandomDuration(jitter))
		if err != nil {
			return
		}

		err = a.gatherOnce(acc, input, interval)
		if err != nil {
			acc.AddError(err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

// gatherOnce runs the input's Gather function once, logging a warning each
// interval it fails to complete before.
func (a *Agent) gatherOnce(
	acc telegraf.Accumulator,
	input *models.RunningInput,
	timeout time.Duration,
) error {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	done := make(chan error)
	go func() {
		done <- input.Gather(acc)
	}()

	for {
		select {
		case err := <-done:
			return err
		case <-ticker.C:
			log.Printf("W! [agent] input %q did not complete within its interval",
				input.Name())
		}
	}
}

// runProcessors applies processors to metrics.
func (a *Agent) runProcessors(
	src <-chan telegraf.Metric,
	agg chan<- telegraf.Metric,
) error {
	for metric := range src {
		metrics := a.applyProcessors(metric)

		for _, metric := range metrics {
			agg <- metric
		}
	}

	return nil
}

// applyProcessors applies all processors to a metric.
func (a *Agent) applyProcessors(m telegraf.Metric) []telegraf.Metric {
	metrics := []telegraf.Metric{m}
	for _, processor := range a.Config.Processors {
		metrics = processor.Apply(metrics...)
	}

	return metrics
}

// runAggregators triggers the periodic push for Aggregators.
//
// When the context is done a final push will occur and then this function
// will return.
func (a *Agent) runAggregators(
	startTime time.Time,
	src <-chan telegraf.Metric,
	dst chan<- telegraf.Metric,
) error {
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for metric := range src {
			var dropOriginal bool
			for _, agg := range a.Config.Aggregators {
				if ok := agg.Add(metric); ok {
					dropOriginal = true
				}
			}

			if !dropOriginal {
				dst <- metric
			}
		}
		cancel()
	}()

	precision := a.Config.Agent.Precision.Duration
	interval := a.Config.Agent.Interval.Duration
	aggregations := make(chan telegraf.Metric, 100)
	for _, agg := range a.Config.Aggregators {
		wg.Add(1)
		go func(agg *models.RunningAggregator) {
			defer wg.Done()

			if a.Config.Agent.RoundInterval {
				// Aggregators are aligned to the agent interval regardless of
				// their period.
				err := internal.SleepContext(ctx, internal.AlignDuration(startTime, interval))
				if err != nil {
					return
				}
			}

			agg.SetPeriodStart(startTime)

			acc := NewAccumulator(agg, aggregations)
			acc.SetPrecision(precision, interval)
			a.push(ctx, agg, acc)
			close(aggregations)
		}(agg)
	}

	for metric := range aggregations {
		metrics := a.applyProcessors(metric)
		for _, metric := range metrics {
			dst <- metric
		}
	}

	wg.Wait()
	return nil
}

// push runs the push for a single aggregator every period.  More simple than
// the output/input version as timeout should be less likely.... not really
// because the output channel can block for now.
func (a *Agent) push(
	ctx context.Context,
	aggregator *models.RunningAggregator,
	acc telegraf.Accumulator,
) {
	ticker := time.NewTicker(aggregator.Period())
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			break
		case <-ctx.Done():
			aggregator.Push(acc)
			return
		}

		aggregator.Push(acc)
	}
}

// runOutputs triggers the periodic write for Outputs.
//
// When the context is done, outputs continue to run until their buffer is
// closed, afterwich they run flush once more.
func (a *Agent) runOutputs(
	startTime time.Time,
	src <-chan telegraf.Metric,
) error {
	interval := a.Config.Agent.FlushInterval.Duration
	jitter := a.Config.Agent.FlushJitter.Duration

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	for _, output := range a.Config.Outputs {
		interval := interval
		// Overwrite agent flush_interval if this plugin has its own.
		if output.Config.FlushInterval != 0 {
			interval = output.Config.FlushInterval
		}

		wg.Add(1)
		go func(output *models.RunningOutput) {
			defer wg.Done()

			if a.Config.Agent.RoundInterval {
				err := internal.SleepContext(
					ctx, internal.AlignDuration(startTime, interval))
				if err != nil {
					return
				}
			}

			a.flush(ctx, output, interval, jitter)
		}(output)
	}

	for metric := range src {
		for i, output := range a.Config.Outputs {
			if i == len(a.Config.Outputs)-1 {
				output.AddMetric(metric)
			} else {
				output.AddMetric(metric.Copy())
			}
		}
	}

	log.Println("I! [agent] Hang on, flushing any cached metrics before shutdown")
	cancel()
	wg.Wait()

	return nil
}

// flush runs an output's flush function periodically until the context is
// done.
func (a *Agent) flush(
	ctx context.Context,
	output *models.RunningOutput,
	interval time.Duration,
	jitter time.Duration,
) {
	// since we are watching two channels we need a ticker with the jitter
	// integrated.
	ticker := NewTicker(interval, jitter)
	defer ticker.Stop()

	logError := func(err error) {
		if err != nil {
			log.Printf("E! [agent] Error writing to output [%s]: %v", output.Name, err)
		}
	}

	for {
		// Favor shutdown over other methods.
		select {
		case <-ctx.Done():
			logError(a.flushOnce(output, interval, output.Write))
			return
		default:
		}

		select {
		case <-ticker.C:
			logError(a.flushOnce(output, interval, output.Write))
		case <-output.BatchReady:
			// Favor the ticker over batch ready
			select {
			case <-ticker.C:
				logError(a.flushOnce(output, interval, output.Write))
			default:
				logError(a.flushOnce(output, interval, output.WriteBatch))
			}
		case <-ctx.Done():
			logError(a.flushOnce(output, interval, output.Write))
			return
		}
	}
}

// flushOnce runs the output's Write function once, logging a warning each
// interval it fails to complete before.
func (a *Agent) flushOnce(
	output *models.RunningOutput,
	timeout time.Duration,
	writeFunc func() error,
) error {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	done := make(chan error)
	go func() {
		done <- writeFunc()
	}()

	for {
		select {
		case err := <-done:
			output.LogBufferStatus()
			return err
		case <-ticker.C:
			log.Printf("W! [agent] output %q did not complete within its flush interval",
				output.Name)
			output.LogBufferStatus()
		}
	}

}

// connectOutputs connects to all outputs.
func (a *Agent) connectOutputs(ctx context.Context) error {
	for _, output := range a.Config.Outputs {
		log.Printf("D! [agent] Attempting connection to output: %s\n", output.Name)
		err := output.Output.Connect()
		if err != nil {
			log.Printf("E! [agent] Failed to connect to output %s, retrying in 15s, "+
				"error was '%s' \n", output.Name, err)

			err := internal.SleepContext(ctx, 15*time.Second)
			if err != nil {
				return err
			}

			err = output.Output.Connect()
			if err != nil {
				return err
			}
		}
		log.Printf("D! [agent] Successfully connected to output: %s\n", output.Name)
	}
	return nil
}

// closeOutputs closes all outputs.
func (a *Agent) closeOutputs() error {
	var err error
	for _, output := range a.Config.Outputs {
		err = output.Output.Close()
	}
	return err
}

// startServiceInputs starts all service inputs.
func (a *Agent) startServiceInputs(
	ctx context.Context,
	dst chan<- telegraf.Metric,
) error {
	started := []telegraf.ServiceInput{}

	for _, input := range a.Config.Inputs {
		if si, ok := input.Input.(telegraf.ServiceInput); ok {
			// Service input plugins are not subject to timestamp rounding.
			// This only applies to the accumulator passed to Start(), the
			// Gather() accumulator does apply rounding according to the
			// precision agent setting.
			acc := NewAccumulator(input, dst)
			acc.SetPrecision(time.Nanosecond, 0)

			err := si.Start(acc)
			if err != nil {
				log.Printf("E! [agent] Service for input %s failed to start: %v",
					input.Name(), err)

				for _, si := range started {
					si.Stop()
				}

				return err
			}

			started = append(started, si)
		}
	}

	return nil
}

// stopServiceInputs stops all service inputs.
func (a *Agent) stopServiceInputs() {
	for _, input := range a.Config.Inputs {
		if si, ok := input.Input.(telegraf.ServiceInput); ok {
			si.Stop()
		}
	}
}

// panicRecover displays an error if an input panics.
func panicRecover(input *models.RunningInput) {
	if err := recover(); err != nil {
		trace := make([]byte, 2048)
		runtime.Stack(trace, true)
		log.Printf("E! FATAL: Input [%s] panicked: %s, Stack:\n%s\n",
			input.Name(), err, trace)
		log.Println("E! PLEASE REPORT THIS PANIC ON GITHUB with " +
			"stack trace, configuration, and OS information: " +
			"https://github.com/influxdata/telegraf/issues/new/choose")
	}
}

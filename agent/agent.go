package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
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

	log.Printf("D! [agent] Initializing plugins")
	err := a.initPlugins()
	if err != nil {
		return err
	}

	log.Printf("D! [agent] Connecting outputs")
	err = a.connectOutputs(ctx)
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
	a.closeOutputs()

	log.Printf("D! [agent] Stopped Successfully")
	return nil
}

// Test runs the inputs once and prints the output to stdout in line protocol.
func (a *Agent) Test(ctx context.Context, waitDuration time.Duration) error {
	var wg sync.WaitGroup
	metricC := make(chan telegraf.Metric)
	nulC := make(chan telegraf.Metric)
	defer func() {
		close(metricC)
		close(nulC)
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
			metric.Reject()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range nulC {
		}
	}()

	hasServiceInputs := false
	for _, input := range a.Config.Inputs {
		if _, ok := input.Input.(telegraf.ServiceInput); ok {
			hasServiceInputs = true
			break
		}
	}

	log.Printf("D! [agent] Initializing plugins")
	err := a.initPlugins()
	if err != nil {
		return err
	}

	if hasServiceInputs {
		log.Printf("D! [agent] Starting service inputs")
		err := a.startServiceInputs(ctx, metricC)
		if err != nil {
			return err
		}
	}

	hasErrors := false
	for _, input := range a.Config.Inputs {
		select {
		case <-ctx.Done():
			return nil
		default:
			break
		}

		acc := NewAccumulator(input, metricC)
		acc.SetPrecision(a.Precision())

		// Special instructions for some inputs. cpu, for example, needs to be
		// run twice in order to return cpu usage percentages.
		switch input.Config.Name {
		case "cpu", "mongodb", "procstat":
			nulAcc := NewAccumulator(input, nulC)
			nulAcc.SetPrecision(a.Precision())
			if err := input.Input.Gather(nulAcc); err != nil {
				acc.AddError(err)
				hasErrors = true
			}

			time.Sleep(500 * time.Millisecond)
			if err := input.Input.Gather(acc); err != nil {
				acc.AddError(err)
				hasErrors = true
			}
		default:
			if err := input.Input.Gather(acc); err != nil {
				acc.AddError(err)
				hasErrors = true
			}
		}
	}

	if hasServiceInputs {
		log.Printf("D! [agent] Waiting for service inputs")
		internal.SleepContext(ctx, waitDuration)
		log.Printf("D! [agent] Stopping service inputs")
		a.stopServiceInputs()
	}

	if hasErrors {
		return fmt.Errorf("One or more input plugins had an error")
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
		jitter := a.Config.Agent.CollectionJitter.Duration

		// Overwrite agent interval if this plugin has its own.
		if input.Config.Interval != 0 {
			interval = input.Config.Interval
		}

		var ticker Ticker
		if a.Config.Agent.RoundInterval {
			ticker = NewAlignedTicker(startTime, interval, jitter)
		} else {
			ticker = NewUnalignedTicker(interval, jitter)
		}
		defer ticker.Stop()

		acc := NewAccumulator(input, dst)
		acc.SetPrecision(a.Precision())

		wg.Add(1)
		go func(input *models.RunningInput) {
			defer wg.Done()
			a.gatherLoop(ctx, acc, input, ticker)
		}(input)
	}

	wg.Wait()
	return nil
}

// gather runs an input's gather function periodically until the context is
// done.
func (a *Agent) gatherLoop(
	ctx context.Context,
	acc telegraf.Accumulator,
	input *models.RunningInput,
	ticker Ticker,
) {
	defer panicRecover(input)

	for {
		select {
		case <-ticker.Elapsed():
			err := a.gatherOnce(acc, input, ticker)
			if err != nil {
				acc.AddError(err)
			}
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
	ticker Ticker,
) error {
	done := make(chan error)
	go func() {
		done <- input.Gather(acc)
	}()

	for {
		select {
		case err := <-done:
			return err
		case <-ticker.Elapsed():
			log.Printf("W! [agent] [%s] did not complete within its interval",
				input.LogName())
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

func updateWindow(start time.Time, roundInterval bool, period time.Duration) (time.Time, time.Time) {
	var until time.Time
	if roundInterval {
		until = internal.AlignTime(start, period)
		if until == start {
			until = internal.AlignTime(start.Add(time.Nanosecond), period)
		}
	} else {
		until = start.Add(period)
	}

	since := until.Add(-period)

	return since, until
}

// runAggregators adds metrics to the aggregators and triggers their periodic
// push call.
//
// Runs until src is closed and all metrics have been processed.  Will call
// push one final time before returning.
func (a *Agent) runAggregators(
	startTime time.Time,
	src <-chan telegraf.Metric,
	dst chan<- telegraf.Metric,
) error {
	ctx, cancel := context.WithCancel(context.Background())

	// Before calling Add, initialize the aggregation window.  This ensures
	// that any metric created after start time will be aggregated.
	for _, agg := range a.Config.Aggregators {
		since, until := updateWindow(startTime, a.Config.Agent.RoundInterval, agg.Period())
		agg.UpdateWindow(since, until)
	}

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
			} else {
				metric.Drop()
			}
		}
		cancel()
	}()

	aggregations := make(chan telegraf.Metric, 100)
	wg.Add(1)
	go func() {
		defer wg.Done()

		var aggWg sync.WaitGroup
		for _, agg := range a.Config.Aggregators {
			aggWg.Add(1)
			go func(agg *models.RunningAggregator) {
				defer aggWg.Done()

				acc := NewAccumulator(agg, aggregations)
				acc.SetPrecision(a.Precision())
				a.push(ctx, agg, acc)
			}(agg)
		}

		aggWg.Wait()
		close(aggregations)
	}()

	for metric := range aggregations {
		metrics := a.applyProcessors(metric)
		for _, metric := range metrics {
			dst <- metric
		}
	}

	wg.Wait()
	return nil
}

// push runs the push for a single aggregator every period.
func (a *Agent) push(
	ctx context.Context,
	aggregator *models.RunningAggregator,
	acc telegraf.Accumulator,
) {
	for {
		// Ensures that Push will be called for each period, even if it has
		// already elapsed before this function is called.  This is guaranteed
		// because so long as only Push updates the EndPeriod.  This method
		// also avoids drift by not using a ticker.
		until := time.Until(aggregator.EndPeriod())

		select {
		case <-time.After(until):
			aggregator.Push(acc)
			break
		case <-ctx.Done():
			aggregator.Push(acc)
			return
		}
	}
}

// runOutputs triggers the periodic write for Outputs.
//

// Runs until src is closed and all metrics have been processed.  Will call
// Write one final time before returning.
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

		jitter := jitter
		// Overwrite agent flush_jitter if this plugin has its own.
		if output.Config.FlushJitter != nil {
			jitter = *output.Config.FlushJitter
		}

		ticker := NewRollingTicker(interval, jitter)
		defer ticker.Stop()

		wg.Add(1)
		go func(output *models.RunningOutput) {
			defer wg.Done()
			a.flushLoop(ctx, output, ticker)
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

// flushLoop runs an output's flush function periodically until the context is
// done.
func (a *Agent) flushLoop(
	ctx context.Context,
	output *models.RunningOutput,
	ticker Ticker,
) {
	logError := func(err error) {
		if err != nil {
			log.Printf("E! [agent] Error writing to %s: %v", output.LogName(), err)
		}
	}

	// watch for flush requests
	flushRequested := make(chan os.Signal, 1)
	watchForFlushSignal(flushRequested)
	defer stopListeningForFlushSignal(flushRequested)

	for {
		// Favor shutdown over other methods.
		select {
		case <-ctx.Done():
			logError(a.flushOnce(output, ticker, output.Write))
			return
		default:
		}

		select {
		case <-ctx.Done():
			logError(a.flushOnce(output, ticker, output.Write))
			return
		case <-ticker.Elapsed():
			logError(a.flushOnce(output, ticker, output.Write))
		case <-flushRequested:
			logError(a.flushOnce(output, ticker, output.Write))
		case <-output.BatchReady:
			// Favor the ticker over batch ready
			select {
			case <-ticker.Elapsed():
				logError(a.flushOnce(output, ticker, output.Write))
			default:
				logError(a.flushOnce(output, ticker, output.WriteBatch))
			}
		}
	}
}

// flushOnce runs the output's Write function once, logging a warning each
// interval it fails to complete before.
func (a *Agent) flushOnce(
	output *models.RunningOutput,
	ticker Ticker,
	writeFunc func() error,
) error {
	done := make(chan error)
	go func() {
		done <- writeFunc()
	}()

	for {
		select {
		case err := <-done:
			output.LogBufferStatus()
			return err
		case <-ticker.Elapsed():
			log.Printf("W! [agent] [%q] did not complete within its flush interval",
				output.LogName())
			output.LogBufferStatus()
		}
	}

}

// initPlugins runs the Init function on plugins.
func (a *Agent) initPlugins() error {
	for _, input := range a.Config.Inputs {
		err := input.Init()
		if err != nil {
			return fmt.Errorf("could not initialize input %s: %v",
				input.LogName(), err)
		}
	}
	for _, processor := range a.Config.Processors {
		err := processor.Init()
		if err != nil {
			return fmt.Errorf("could not initialize processor %s: %v",
				processor.Config.Name, err)
		}
	}
	for _, aggregator := range a.Config.Aggregators {
		err := aggregator.Init()
		if err != nil {
			return fmt.Errorf("could not initialize aggregator %s: %v",
				aggregator.Config.Name, err)
		}
	}
	for _, output := range a.Config.Outputs {
		err := output.Init()
		if err != nil {
			return fmt.Errorf("could not initialize output %s: %v",
				output.Config.Name, err)
		}
	}
	return nil
}

// connectOutputs connects to all outputs.
func (a *Agent) connectOutputs(ctx context.Context) error {
	for _, output := range a.Config.Outputs {
		log.Printf("D! [agent] Attempting connection to [%s]", output.LogName())
		err := output.Output.Connect()
		if err != nil {
			log.Printf("E! [agent] Failed to connect to [%s], retrying in 15s, "+
				"error was '%s'", output.LogName(), err)

			err := internal.SleepContext(ctx, 15*time.Second)
			if err != nil {
				return err
			}

			err = output.Output.Connect()
			if err != nil {
				return err
			}
		}
		log.Printf("D! [agent] Successfully connected to %s", output.LogName())
	}
	return nil
}

// closeOutputs closes all outputs.
func (a *Agent) closeOutputs() {
	for _, output := range a.Config.Outputs {
		output.Close()
	}
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
			acc.SetPrecision(time.Nanosecond)

			err := si.Start(acc)
			if err != nil {
				log.Printf("E! [agent] Service for [%s] failed to start: %v",
					input.LogName(), err)

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

// Returns the rounding precision for metrics.
func (a *Agent) Precision() time.Duration {
	precision := a.Config.Agent.Precision.Duration
	interval := a.Config.Agent.Interval.Duration

	if precision > 0 {
		return precision
	}

	switch {
	case interval >= time.Second:
		return time.Second
	case interval >= time.Millisecond:
		return time.Millisecond
	case interval >= time.Microsecond:
		return time.Microsecond
	default:
		return time.Nanosecond
	}
}

// panicRecover displays an error if an input panics.
func panicRecover(input *models.RunningInput) {
	if err := recover(); err != nil {
		trace := make([]byte, 2048)
		runtime.Stack(trace, true)
		log.Printf("E! FATAL: [%s] panicked: %s, Stack:\n%s",
			input.LogName(), err, trace)
		log.Println("E! PLEASE REPORT THIS PANIC ON GITHUB with " +
			"stack trace, configuration, and OS information: " +
			"https://github.com/influxdata/telegraf/issues/new/choose")
	}
}

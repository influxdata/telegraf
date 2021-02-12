package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
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

	// units hold channels and define connections between plugins
	outputUnit         outputUnit
	processorGroupUnit processorGroupUnit
	// inputUnit          inputUnit
	inputGroupUnit inputGroupUnit

	ctx context.Context
}

// NewAgent returns an Agent for the given Config.
func NewAgent(config *config.Config) (*Agent, error) {
	a := &Agent{
		Config: config,
	}
	return a, nil
}

type inputUnit struct {
	accumulator  *accumulator
	input        *models.RunningInput
	cancelGather context.CancelFunc // used to cancel the gather loop for plugin shutdown
}

//  ______     ┌───────────┐     ______
// ()_____)──▶ │ Processor │──▶ ()_____)
//             └───────────┘
type processorUnit struct {
	order       int
	src         chan telegraf.Metric
	dst         chan<- telegraf.Metric
	processor   models.ProcessorRunner
	accumulator *accumulator
}

// outputUnit is a group of Outputs and their source channel.  Metrics on the
// channel are written to all outputs.
//
//                            ┌────────┐
//                       ┌──▶ │ Output │
//                       │    └────────┘
//  ______     ┌─────┐   │    ┌────────┐
// ()_____)──▶ │ Fan │───┼──▶ │ Output │
//             └─────┘   │    └────────┘
//                       │    ┌────────┐
//                       └──▶ │ Output │
//                            └────────┘
type outputUnit struct {
	sync.Mutex
	src     chan telegraf.Metric
	outputs []*models.RunningOutput
}

type processorGroupUnit struct {
	sync.Mutex
	accumulator    telegraf.Accumulator
	processorUnits []*processorUnit
}

func (pg *processorGroupUnit) Find(pr models.ProcessorRunner) *processorUnit {
	for _, unit := range pg.processorUnits {
		if unit.processor.GetID() == pr.GetID() {
			return unit
		}
	}

	return nil
}

// inputGroupUnit is a group of input plugins and the shared channel they write to.
//
// ┌───────┐
// │ Input │───┐
// └───────┘   │
// ┌───────┐   │     ______
// │ Input │───┼──▶ ()_____)
// └───────┘   │
// ┌───────┐   │
// │ Input │───┘
// └───────┘
type inputGroupUnit struct {
	sync.Mutex
	dst        chan<- telegraf.Metric
	inputUnits []inputUnit
}

// RunWithAPI runs Telegraf in API mode where all the plugins are controlled by
// the user through the config API. When running in this mode plugins are not
// loaded from the toml file.
func (a *Agent) RunWithAPI(ctx context.Context, outputCancel context.CancelFunc) {
	a.ctx = ctx
	log.Printf("I! [agent] Config: Interval:%s, Quiet:%#v, Hostname:%#v, "+
		"Flush Interval:%s",
		a.Config.Agent.Interval.Duration, a.Config.Agent.Quiet,
		a.Config.Agent.Hostname, a.Config.Agent.FlushInterval.Duration)

	// a.loadState()
	a.outputUnit.src = make(chan telegraf.Metric)
	go func() {
		a.runOutputFanout()
		outputCancel()
	}()
	<-ctx.Done()
	a.stopInputs()
	// a.saveState()

	// wait for all plugins to stop
	log.Printf("D! [agent] Stopped Successfully")
}

// Run starts and runs the Agent until the context is done.
func (a *Agent) Run(ctx context.Context) error {
	a.ctx = ctx
	log.Printf("I! [agent] Config: Interval:%s, Quiet:%#v, Hostname:%#v, "+
		"Flush Interval:%s",
		a.Config.Agent.Interval.Duration, a.Config.Agent.Quiet,
		a.Config.Agent.Hostname, a.Config.Agent.FlushInterval.Duration)

	log.Printf("D! [agent] Initializing plugins")
	if err := a.initPlugins(); err != nil {
		return err
	}

	log.Printf("D! [agent] Connecting outputs")
	if err := a.startOutputs(ctx); err != nil {
		return err
	}

	if err := a.startProcessors(a.outputUnit.src); err != nil {
		return err
	}

	if err := a.startInputs(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := a.runOutputs()
		if err != nil {
			log.Printf("E! [agent] Error running outputs: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.runProcessors()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := a.runInputs()
		if err != nil {
			log.Printf("E! [agent] Error running inputs: %v", err)
		}
	}()

	wg.Wait()

	log.Printf("D! [agent] Stopped Successfully")
	return nil
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
				processor.LogName(), err)
		}
	}
	for _, output := range a.Config.Outputs {
		err := output.Init()
		if err != nil {
			return fmt.Errorf("could not initialize output %s: %v",
				output.LogName(), err)
		}
	}
	return nil
}

func (a *Agent) startInputs() error {
	log.Printf("D! [agent] Starting service inputs")

	for _, input := range a.Config.Inputs {
		if err := a.StartInput(input); err != nil {
			a.stopInputs()
			return fmt.Errorf("starting input %s: %w", input.LogName(), err)
		}
	}

	return nil
}

func (a *Agent) StartInput(input *models.RunningInput) error {
	a.inputGroupUnit.Lock()
	defer a.inputGroupUnit.Unlock()

	// Service input plugins are not normally subject to timestamp
	// rounding except for when precision is set on the input plugin.
	//
	// This only applies to the accumulator passed to Start(), the
	// Gather() accumulator does apply rounding according to the
	// precision and interval agent/plugin settings.
	var interval time.Duration
	var precision time.Duration
	if input.Config.Precision != 0 {
		precision = input.Config.Precision
	}

	dst := a.outputUnit.src
	if len(a.processorGroupUnit.processorUnits) > 0 {
		dst = a.processorGroupUnit.processorUnits[0].src
	}

	acc := NewAccumulator(input, dst)
	acc.SetPrecision(getPrecision(precision, interval))

	if err := input.Start(acc); err != nil {
		return err
	}

	a.inputGroupUnit.inputUnits = append(a.inputGroupUnit.inputUnits,
		inputUnit{
			accumulator: acc.(*accumulator),
			input:       input,
		})
	return nil
}

// runInputs starts and triggers the periodic gather for Inputs.
//
// When the context is done the timers are stopped and this function returns
// after all ongoing Gather calls complete.
func (a *Agent) runInputs() error {
	startTime := time.Now()
	var wg sync.WaitGroup
	for _, iu := range a.inputGroupUnit.inputUnits {
		wg.Add(1)
		go func(input *models.RunningInput) {
			defer wg.Done()
			a.RunInput(input, startTime)
		}(iu.input)
	}

	wg.Wait()
	a.stopInputs()
	return nil
}

// RunInput is a blocking call that runs an input forever
func (a *Agent) RunInput(input *models.RunningInput, startTime time.Time) {
	// a.inputGroupUnit.Lock() ?
	// Overwrite agent interval if this plugin has its own.
	interval := a.Config.Agent.Interval.Duration
	if input.Config.Interval != 0 {
		interval = input.Config.Interval
	}

	// Overwrite agent precision if this plugin has its own.
	precision := a.Config.Agent.Precision.Duration
	if input.Config.Precision != 0 {
		precision = input.Config.Precision
	}

	// Overwrite agent collection_jitter if this plugin has its own.
	jitter := a.Config.Agent.CollectionJitter.Duration
	if input.Config.CollectionJitter != 0 {
		jitter = input.Config.CollectionJitter
	}

	var ticker Ticker
	if a.Config.Agent.RoundInterval {
		ticker = NewAlignedTicker(startTime, interval, jitter)
	} else {
		ticker = NewUnalignedTicker(interval, jitter)
	}
	defer ticker.Stop()

	acc := NewAccumulator(input, a.inputGroupUnit.dst)
	acc.SetPrecision(getPrecision(precision, interval))
	ctx, cancelFunc := context.WithCancel(a.ctx)
	a.inputGroupUnit.Lock()
	for i, iu := range a.inputGroupUnit.inputUnits {
		if iu.input == input {
			a.inputGroupUnit.inputUnits[i].cancelGather = cancelFunc
			break
		}
	}
	a.inputGroupUnit.Unlock()

	a.gatherLoop(ctx, acc, input, ticker, interval)
}

// testStartInputs is a variation of startInputs for use in --test and --once
// mode.  It differs by logging Start errors and returning only plugins
// successfully started.
func (a *Agent) testStartInputs() error {
	log.Printf("D! [agent] Starting service inputs")

	dst := a.outputUnit.src
	if len(a.processorGroupUnit.processorUnits) > 0 {
		dst = a.processorGroupUnit.processorUnits[0].src
	}
	unit := &inputGroupUnit{
		dst: dst,
	}

	for _, input := range a.Config.Inputs {
		// Overwrite agent interval if this plugin has its own.
		interval := a.Config.Agent.Interval.Duration
		if input.Config.Interval != 0 {
			interval = input.Config.Interval
		}

		// Overwrite agent precision if this plugin has its own.
		precision := a.Config.Agent.Precision.Duration
		if input.Config.Precision != 0 {
			precision = input.Config.Precision
		}

		acc := NewAccumulator(input, dst)
		acc.SetPrecision(getPrecision(precision, interval))

		if si, ok := input.Input.(telegraf.ServiceInput); ok {
			// Service input plugins are not subject to timestamp rounding.
			// This only applies to the accumulator passed to Start(), the
			// Gather() accumulator does apply rounding according to the
			// precision agent setting.

			err := si.Start(acc)
			if err != nil {
				log.Printf("E! [agent] Starting input %s: %v", input.LogName(), err)
			}

		}

		unit.inputUnits = append(unit.inputUnits, inputUnit{
			accumulator: acc.(*accumulator),
			input:       input,
		})
	}

	return nil
}

// testRunInputs is a variation of runInputs for use in --test and --once mode.
// Instead of using a ticker to run the inputs they are called once immediately.
func (a *Agent) testRunInputs(ctx context.Context, wait time.Duration) error {
	var wg sync.WaitGroup

	nul := make(chan telegraf.Metric)
	go func() {
		for range nul {
		}
	}()

	for _, iu := range a.inputGroupUnit.inputUnits {
		wg.Add(1)
		go func(iu inputUnit) {
			defer wg.Done()

			// Run plugins that require multiple gathers to calculate rate
			// and delta metrics twice.
			switch iu.input.Config.Name {
			case "cpu", "mongodb", "procstat":
				iu.accumulator.setOutput(nul)
				if err := iu.input.Gather(iu.accumulator); err != nil {
					iu.accumulator.AddError(err)
				}

				time.Sleep(500 * time.Millisecond)
			}

			if err := iu.input.Gather(iu.accumulator); err != nil {
				iu.accumulator.AddError(err)
			}
		}(iu)
	}
	wg.Wait()

	_ = internal.SleepContext(ctx, wait)

	a.stopInputs()
	return nil
}

// stopInputs stops all service inputs.
func (a *Agent) stopInputs() {
	log.Printf("D! [agent] Stopping service inputs")
	a.inputGroupUnit.Lock()
	defer a.inputGroupUnit.Unlock()

	for _, iu := range a.inputGroupUnit.inputUnits {
		iu.input.Stop()
	}
	log.Printf("D! [agent] Input channel closed")
}

func (a *Agent) StopInput(i *models.RunningInput) {
retry:
	a.inputGroupUnit.Lock()
	defer a.inputGroupUnit.Unlock()
	if len(a.inputGroupUnit.inputUnits) == 0 {
		return
	}
	for pos, iu := range a.inputGroupUnit.inputUnits {
		if iu.input == i {
			if iu.cancelGather == nil {
				a.inputGroupUnit.Unlock()
				time.Sleep(1 * time.Millisecond)
				goto retry
			}
			iu.cancelGather()
			// TODO(steven): do I need to wait for the gather to stop?
			i.Stop()
			// drop from the list
			a.inputGroupUnit.inputUnits = append(a.inputGroupUnit.inputUnits[0:pos], a.inputGroupUnit.inputUnits[pos+1:len(a.inputGroupUnit.inputUnits)]...)
			break
		}
	}
	// close the channel if we're the last one and the context is closed.
	if len(a.inputGroupUnit.inputUnits) == 0 {
		if a.ctx.Err() != nil {
			close(a.inputGroupUnit.dst)
		}
	}
}

// gather runs an input's gather function periodically until the context is
// done.
func (a *Agent) gatherLoop(
	ctx context.Context,
	acc telegraf.Accumulator,
	input *models.RunningInput,
	ticker Ticker,
	interval time.Duration,
) {
	defer panicRecover(input)

	for {
		select {
		case <-ticker.Elapsed():
			err := a.gatherOnce(acc, input, ticker, interval)
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
	interval time.Duration,
) error {
	done := make(chan error)
	go func() {
		done <- input.Gather(acc)
	}()

	// Only warn after interval seconds, even if the interval is started late.
	// Intervals can start late if the previous interval went over or due to
	// clock changes.
	slowWarning := time.NewTicker(interval)
	defer slowWarning.Stop()

	for {
		select {
		case err := <-done:
			return err
		case <-slowWarning.C:
			log.Printf("W! [%s] Collection took longer than expected; not complete after interval of %s",
				input.LogName(), interval)
		case <-ticker.Elapsed():
			log.Printf("D! [%s] Previous collection has not completed; scheduled collection skipped",
				input.LogName())
		}
	}
}

// startProcessors sets up the processor chain and calls Start on all
// processors.  If an error occurs any started processors are Stopped.
func (a *Agent) startProcessors(dst chan<- telegraf.Metric) error {
	// Sort from last to first
	sort.SliceStable(a.Config.Processors, a.Config.Processors.Less)

	for _, processor := range a.Config.Processors {
		if err := a.StartProcessor(processor); err != nil {
			// TODO(steven): stop all processors
			// for _, u := range  {
			// 	a.StopProcessor(u.processor)
			// }
			return err
		}
	}

	return nil
}

func (a *Agent) StartProcessor(processor models.ProcessorRunner) error {
	a.processorGroupUnit.Lock()
	defer a.processorGroupUnit.Unlock()

	pu := processorUnit{
		src:       make(chan telegraf.Metric, 100),
		processor: processor,
	}

	// insertPos is the position in the list after the slice insert operation
	insertPos := 0

	if len(a.processorGroupUnit.processorUnits) > 0 {
		// figure out where in the list to put us
		for i, unit := range a.processorGroupUnit.processorUnits {
			if unit.order > int(processor.Order()) {
				break
			}
			insertPos = i + 1
		}
	}

	// if we're the last processor in the list
	if insertPos == len(a.processorGroupUnit.processorUnits) {
		pu.dst = a.outputUnit.src
	} else {
		// we're not the last processor
		pu.dst = a.processorGroupUnit.processorUnits[insertPos].src
	}

	acc := NewAccumulator(processor.(MetricMaker), pu.dst)
	pu.accumulator = acc.(*accumulator)

	err := processor.Start(acc)
	if err != nil {
		return fmt.Errorf("starting processor %s: %w", processor.(MetricMaker).LogName(), err)
	}

	// only if we're the first processor:
	if insertPos == 0 {
		for _, iu := range a.inputGroupUnit.inputUnits {
			iu.accumulator.setOutput(pu.src)
		}
	} else {
		// not the first processor to be added
		prev := a.processorGroupUnit.processorUnits[insertPos-1]
		prev.accumulator.setOutput(pu.src)
	}

	list := a.processorGroupUnit.processorUnits
	// list[0:insertPos] + pu + list[insertPos:]
	// list = [0,1,2,3,4]
	// [0,1] + pu + [2,3,4]
	a.processorGroupUnit.processorUnits = append(append(list[0:insertPos], &pu), list[insertPos:]...)

	return nil
}

// runProcessors begins processing metrics and runs until the source channel is
// closed and all metrics have been written.
func (a *Agent) runProcessors() {
	var wg sync.WaitGroup
	for _, unit := range a.processorGroupUnit.processorUnits {
		wg.Add(1)
		go func(unit *processorUnit) {
			defer wg.Done()
			a.RunProcessor(unit.processor)

			// log.Printf("D! [agent] Processor channel closed")
		}(unit)
	}
	wg.Wait()
}

// RunProcessor is a blocking call that runs a processor forever
func (a *Agent) RunProcessor(p models.ProcessorRunner) {
	a.processorGroupUnit.Lock()
	pu := a.processorGroupUnit.Find(p)
	a.processorGroupUnit.Unlock()

	if pu == nil {
		panic("Must call StartProcessor before calling RunProcessor")
	}

	for m := range pu.src {
		if err := pu.processor.Add(m, pu.accumulator); err != nil {
			pu.accumulator.AddError(err)
			m.Drop()
		}
	}
	pu.processor.Stop()

	if a.ctx.Err() != nil {
		a.processorGroupUnit.Lock()
		close(pu.dst)
		a.processorGroupUnit.Unlock()
	}
}

// StopProcessor stops processors or aggregators.
// ProcessorRunner could be a *models.RunningProcessor or a *models.RunningAggregator
func (a *Agent) StopProcessor(p models.ProcessorRunner) {
	a.processorGroupUnit.Lock()
	defer a.processorGroupUnit.Unlock()

	for i, curr := range a.processorGroupUnit.processorUnits {
		if p.GetID() == curr.processor.GetID() {
			if i > 0 {
				prev := a.processorGroupUnit.processorUnits[i-1]
				prev.dst = curr.dst
				prev.accumulator.setOutput(curr.dst)
			} else {
				a.inputGroupUnit.dst = curr.dst
				for _, iu := range a.inputGroupUnit.inputUnits {
					iu.accumulator.setOutput(curr.dst)
				}
			}

			close(curr.src)
			go curr.processor.Stop()
			// remove it from the slice
			// Remove the stopped processor from the units list
			a.processorGroupUnit.processorUnits = append(a.processorGroupUnit.processorUnits[0:i], a.processorGroupUnit.processorUnits[i+1:len(a.processorGroupUnit.processorUnits)]...)
		}
	}
}

func (a *Agent) StopOutput(p *models.RunningOutput) {
	// lock
	a.outputUnit.Lock()
	defer a.outputUnit.Unlock()

	// find plugin
	for i, output := range a.outputUnit.outputs {
		if output.ID == p.ID {
			// disconnect it from the output broadcaster
			// ?
			// remove it from the list
			a.outputUnit.outputs = append(a.outputUnit.outputs[0:i], a.outputUnit.outputs[i+1:len(a.outputUnit.outputs)]...)
			// maybe close the source

			// tell the plugin to stop
			go output.Close()
		}
	}
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

// startOutputs calls Connect on all outputs and returns the source channel.
// If an error occurs calling Connect all stared plugins have Close called.
func (a *Agent) startOutputs(
	ctx context.Context,
) error {
	src := make(chan telegraf.Metric, 100)
	a.outputUnit = outputUnit{src: src}
	for _, output := range a.Config.Outputs {
		if err := a.StartOutput(output); err != nil {
			for _, output := range a.outputUnit.outputs {
				output.Close()
			}
			return err
		}
	}

	return nil
}

func (a *Agent) StartOutput(output *models.RunningOutput) error {
	a.outputUnit.Lock()
	defer a.outputUnit.Unlock()

	if err := a.connectOutput(a.ctx, output); err != nil {
		return fmt.Errorf("connecting output %s: %w", output.LogName(), err)
	}
	a.outputUnit.outputs = append(a.outputUnit.outputs, output)

	return nil
}

// connectOutputs connects to all outputs.
func (a *Agent) connectOutput(ctx context.Context, output *models.RunningOutput) error {
	log.Printf("D! [agent] Attempting connection to [%s]", output.LogName())
	err := output.Connect()
	if err != nil {
		log.Printf("E! [agent] Failed to connect to [%s], retrying in 15s, "+
			"error was '%s'", output.LogName(), err)

		err := internal.SleepContext(ctx, 15*time.Second)
		if err != nil {
			return err
		}

		err = output.Connect()
		if err != nil {
			return fmt.Errorf("Error connecting to output %q: %w", output.LogName(), err)
		}
	}
	log.Printf("D! [agent] Successfully connected to %s", output.LogName())
	return nil
}

// runOutputs begins processing metrics and returns until the source channel is
// closed and all metrics have been written.  On shutdown metrics will be
// written one last time and dropped if unsuccessful.
func (a *Agent) runOutputs() error {
	var wg sync.WaitGroup

	// flush management context
	flushCtx, cancelFlush := context.WithCancel(context.Background())

	for _, output := range a.outputUnit.outputs {
		wg.Add(1)
		go func(output *models.RunningOutput) {
			defer wg.Done()

			a.RunOutput(flushCtx, output)
		}(output)
	}

	a.runOutputFanout()

	log.Println("I! [agent] Hang on, flushing any cached metrics before shutdown")
	cancelFlush()
	wg.Wait()

	return nil
}

func (a *Agent) RunOutput(ctx context.Context, output *models.RunningOutput) {
	interval := a.Config.Agent.FlushInterval.Duration
	jitter := a.Config.Agent.FlushJitter.Duration
	// Overwrite agent flush_interval if this plugin has its own.
	if output.Config.FlushInterval != 0 {
		interval = output.Config.FlushInterval
	}

	// Overwrite agent flush_jitter if this plugin has its own.
	if output.Config.FlushJitter != 0 {
		jitter = output.Config.FlushJitter
	}

	ticker := NewRollingTicker(interval, jitter)
	defer ticker.Stop()

	a.flushLoop(ctx, output, ticker)
}

// runOutputFanout does the outputUnit fanout, copying a metric to all outputs
func (a *Agent) runOutputFanout() {
	for metric := range a.outputUnit.src {
		a.outputUnit.Lock()
		outs := a.outputUnit.outputs
		a.outputUnit.Unlock()
		for i, output := range outs {
			if i == len(outs)-1 {
				output.AddMetric(metric)
			} else {
				output.AddMetric(metric.Copy())
			}
		}
	}
}

// flushLoop runs an output's flush function periodically until the context is done.
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

// Test runs the inputs, processors and aggregators for a single gather and
// writes the metrics to stdout.
func (a *Agent) Test(ctx context.Context, wait time.Duration) error {
	src := make(chan telegraf.Metric, 100)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := influx.NewSerializer()
		s.SetFieldSortOrder(influx.SortFields)

		for metric := range src {
			octets, err := s.Serialize(metric)
			if err == nil {
				fmt.Print("> ", string(octets))
			}
			metric.Reject()
		}
	}()

	err := a.test(ctx, wait, src)
	if err != nil {
		return err
	}

	wg.Wait()

	if models.GlobalGatherErrors.Get() != 0 {
		return fmt.Errorf("input plugins recorded %d errors", models.GlobalGatherErrors.Get())
	}
	return nil
}

// Test runs the agent and performs a single gather sending output to the
// outputF.  After gathering pauses for the wait duration to allow service
// inputs to run.
func (a *Agent) test(ctx context.Context, wait time.Duration, outputC chan<- telegraf.Metric) error {
	log.Printf("D! [agent] Initializing plugins")
	err := a.initPlugins()
	if err != nil {
		return err
	}

	if err = a.startProcessors(outputC); err != nil {
		return err
	}

	err = a.testStartInputs()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.runProcessors()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := a.testRunInputs(ctx, wait)
		if err != nil {
			log.Printf("E! [agent] Error running inputs: %v", err)
		}
	}()

	wg.Wait()

	log.Printf("D! [agent] Stopped Successfully")

	return nil
}

// Once runs the full agent for a single gather.
func (a *Agent) Once(ctx context.Context, wait time.Duration) error {
	err := a.once(ctx, wait)
	if err != nil {
		return err
	}

	if models.GlobalGatherErrors.Get() != 0 {
		return fmt.Errorf("input plugins recorded %d errors", models.GlobalGatherErrors.Get())
	}

	unsent := 0
	for _, output := range a.Config.Outputs {
		unsent += output.BufferLength()
	}
	if unsent != 0 {
		return fmt.Errorf("output plugins unable to send %d metrics", unsent)
	}
	return nil
}

// On runs the agent and performs a single gather sending output to the
// outputF.  After gathering pauses for the wait duration to allow service
// inputs to run.
func (a *Agent) once(ctx context.Context, wait time.Duration) error {
	log.Printf("D! [agent] Initializing plugins")
	if err := a.initPlugins(); err != nil {
		return err
	}

	log.Printf("D! [agent] Connecting outputs")
	if err := a.startOutputs(ctx); err != nil {
		return err
	}

	if err := a.startProcessors(a.outputUnit.src); err != nil {
		return err
	}

	err := a.testStartInputs()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := a.runOutputs()
		if err != nil {
			log.Printf("E! [agent] Error running outputs: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.runProcessors()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := a.testRunInputs(ctx, wait)
		if err != nil {
			log.Printf("E! [agent] Error running inputs: %v", err)
		}
	}()

	wg.Wait()

	log.Printf("D! [agent] Stopped Successfully")

	return nil
}

// Returns the rounding precision for metrics.
func getPrecision(precision, interval time.Duration) time.Duration {
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

func (a *Agent) RunningInputs() []*models.RunningInput {
	a.inputGroupUnit.Lock()
	defer a.inputGroupUnit.Unlock()
	runningInputs := []*models.RunningInput{}

	for _, iu := range a.inputGroupUnit.inputUnits {
		runningInputs = append(runningInputs, iu.input)
	}
	return runningInputs
}

func (a *Agent) RunningProcessors() []models.ProcessorRunner {
	a.processorGroupUnit.Lock()
	defer a.processorGroupUnit.Unlock()
	runningProcessors := []models.ProcessorRunner{}
	for _, pu := range a.processorGroupUnit.processorUnits {
		runningProcessors = append(runningProcessors, pu.processor)
	}
	return runningProcessors
}

func (a *Agent) RunningOutputs() []*models.RunningOutput {
	a.outputUnit.Lock()
	defer a.outputUnit.Unlock()
	runningOutputs := []*models.RunningOutput{}
	// make sure we allocate and use a new slice that doesn't need a lock
	runningOutputs = append(runningOutputs, a.outputUnit.outputs...)
	return runningOutputs
}

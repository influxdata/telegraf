package agent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/channel"
	"github.com/influxdata/telegraf/models"
)

// agentState describes the running state of the agent.
// Plugins can only be ran once the agent is running,
// you can add plugins before the agent has started.
// you cannot remove plugins before the agent has started.
// you cannot add or remove plugins once it has started shutting down.
type agentState int8

const (
	agentStateStarting agentState = iota
	agentStateRunning
	agentStateShuttingDown
)

// Agent runs a set of plugins.
type Agent struct {
	Config *config.Config

	// units hold channels and define connections between plugins
	outputGroupUnit    outputGroupUnit
	processorGroupUnit processorGroupUnit
	// inputUnit          inputUnit
	inputGroupUnit   inputGroupUnit
	configPluginUnit configPluginUnit

	ctx context.Context

	stateChanged *sync.Cond // a condition that lets plugins know when when the agent state changes
	state        agentState
}

// NewAgent returns an Agent for the given Config.
func NewAgent(ctx context.Context, cfg *config.Config) *Agent {
	inputDestCh := make(chan telegraf.Metric)
	outputSrcCh := make(chan telegraf.Metric)

	// by default, connect the dest of the inputs directly to the src for the outputs,
	// as processors are added, they will be inserted between these two.

	return &Agent{
		Config:       cfg,
		ctx:          ctx,
		stateChanged: sync.NewCond(&sync.Mutex{}),
		state:        agentStateStarting,
		inputGroupUnit: inputGroupUnit{
			dst:   inputDestCh,
			relay: channel.NewRelay(inputDestCh, outputSrcCh),
		},
		outputGroupUnit: outputGroupUnit{
			src: outputSrcCh,
		},
	}
}

type inputUnit struct {
	input        *models.RunningInput
	cancelGather context.CancelFunc // used to cancel the gather loop for plugin shutdown
}

type configPluginUnit struct {
	sync.Mutex
	plugins []config.ConfigPlugin
}

//  ______     ┌───────────┐     ______
// ()_____)──▶ │ Processor │──▶ ()_____)
//             └───────────┘
type processorUnit struct {
	order       int
	src         chan telegraf.Metric   // owns this src
	dst         chan<- telegraf.Metric // reference to another chan owned elsewhere
	processor   models.ProcessorRunner
	accumulator *accumulator
}

// outputGroupUnit is a group of Outputs and their source channel.  Metrics on the
// channel are written to all outputs.
//
//                            ┌────────┐
//                       ┌──▶ │ Output │
//                       │    └────────┘
//  ______     ┌─────┐   │    ┌────────┐
// ()_____)──▶ │ Fan │───┼──▶ │ Output │
//    src      └─────┘   │    └────────┘
//                       │    ┌────────┐
//                       └──▶ │ Output │
//                            └────────┘
type outputGroupUnit struct {
	sync.Mutex
	src     chan telegraf.Metric
	outputs []outputUnit
}

type outputUnit struct {
	output      *models.RunningOutput
	cancelFlush context.CancelFunc // used to cancel the flush loop for plugin shutdown
}

type processorGroupUnit struct {
	sync.Mutex
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
	dst        chan<- telegraf.Metric // owns channel; must stay open until app is shutting down
	relay      *channel.Relay
	inputUnits []inputUnit
}

// RunWithAPI runs Telegraf in API mode where all the plugins are controlled by
// the user through the config API. When running in this mode plugins are not
// loaded from the toml file.
// if ctx errors (eg cancels), inputs will be notified to shutdown
// during shutdown, once the output fanout has received the last metric,
// outputCancel is called as a sort of callback to notify the outputs to finish up.
// you don't want outputs subscribing to the same ctx as the inputs,
// otherwise they'd stop working before receiving all the messages (this is because
// they maintain their own internal buffers rather than depending on metrics buffered
// in a channel)
func (a *Agent) RunWithAPI(outputCancel context.CancelFunc) {
	go func() {
		a.runOutputFanout()
		// then the fanout closes, notify the outputs that they won't be receiving
		// more metrics via this cancel function.
		outputCancel()
	}()

	log.Printf("I! [agent] Config: Interval:%s, Quiet:%#v, Hostname:%#v, "+
		"Flush Interval:%s",
		time.Duration(a.Config.Agent.Interval), a.Config.Agent.Quiet,
		a.Config.Agent.Hostname, time.Duration(a.Config.Agent.FlushInterval))

	a.inputGroupUnit.relay.Start()

	a.setState(agentStateRunning)

	<-a.Context().Done()

	a.setState(agentStateShuttingDown)

	// wait for all plugins to stop
	a.waitForPluginsToStop()

	log.Printf("D! [agent] Stopped Successfully")
}

// AddInput adds an input to the agent to be managed
func (a *Agent) AddInput(input *models.RunningInput) {
	if a.isState(agentStateShuttingDown) {
		return
	}
	a.inputGroupUnit.Lock()
	defer a.inputGroupUnit.Unlock()

	a.inputGroupUnit.inputUnits = append(a.inputGroupUnit.inputUnits,
		inputUnit{
			input: input,
		})
}

func (a *Agent) startInput(input *models.RunningInput) error {
	// plugins can start before the agent has started; wait until it's asked to
	// start before collecting metrics in case other plugins are still loading.
	a.waitUntilState(agentStateRunning)

	a.inputGroupUnit.Lock()
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

	// the plugin's Start() gets its own accumulator with no rounding, etc
	acc := NewAccumulator(input, a.inputGroupUnit.dst)
	acc.SetPrecision(getPrecision(precision, interval))

	for _, inp := range a.inputGroupUnit.inputUnits {
		if inp.input.GetID() == input.GetID() {
			a.inputGroupUnit.Unlock()
			return input.Start(acc)
		}
	}
	a.inputGroupUnit.Unlock()

	return errors.New("cannot start input; call AddInput first")
}

// RunInput is a blocking call that runs an input forever
func (a *Agent) RunInput(input *models.RunningInput, startTime time.Time) {
	// default to agent interval but check for override
	interval := time.Duration(a.Config.Agent.Interval)
	if input.Config.Interval != 0 {
		interval = input.Config.Interval
	}
	// default to agent precision but check for override
	precision := time.Duration(a.Config.Agent.Precision)
	if input.Config.Precision != 0 {
		precision = input.Config.Precision
	}
	// default to agent collection_jitter but check for override
	jitter := time.Duration(a.Config.Agent.CollectionJitter)
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
	defer cancelFunc() // just to keep linters happy

	err := errors.New("loop at least once")
	for err != nil && ctx.Err() == nil {
		if err = a.startInput(input); err != nil {
			log.Printf("E! [agent] failed to start plugin %q: %v", input.LogName(), err)
			time.Sleep(10 * time.Second)
		}
	}

	a.inputGroupUnit.Lock()
	for i, iu := range a.inputGroupUnit.inputUnits {
		if iu.input == input {
			a.inputGroupUnit.inputUnits[i].cancelGather = cancelFunc
			break
		}
	}
	a.inputGroupUnit.Unlock()

	a.gatherLoop(ctx, acc, input, ticker, interval)
	input.Stop()

	a.inputGroupUnit.Lock()
	for i, iu := range a.inputGroupUnit.inputUnits {
		if iu.input == input {
			// remove from list
			a.inputGroupUnit.inputUnits = append(a.inputGroupUnit.inputUnits[0:i], a.inputGroupUnit.inputUnits[i+1:]...)
			break
		}
	}
	a.inputGroupUnit.Unlock()
}

func (a *Agent) StopInput(i *models.RunningInput) {
	defer a.inputGroupUnit.Unlock()
retry:
	a.inputGroupUnit.Lock()
	for _, iu := range a.inputGroupUnit.inputUnits {
		if iu.input == i {
			if iu.cancelGather == nil {
				// plugin hasn't finished starting, wait longer.
				a.inputGroupUnit.Unlock()
				time.Sleep(1 * time.Millisecond)
				goto retry
			}
			iu.cancelGather()
			break
		}
	}
}

// stopRunningOutputs stops all running outputs.
func stopRunningOutputs(outputs []*models.RunningOutput) {
	for _, output := range outputs {
		output.Close()
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

func (a *Agent) AddProcessor(processor models.ProcessorRunner) {
	if a.isState(agentStateShuttingDown) {
		return
	}
	a.inputGroupUnit.Lock()
	defer a.inputGroupUnit.Unlock()
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
		pu.dst = a.outputGroupUnit.src
	} else {
		// we're not the last processor
		pu.dst = a.processorGroupUnit.processorUnits[insertPos].src
	}

	acc := NewAccumulator(processor.(MetricMaker), pu.dst)
	pu.accumulator = acc.(*accumulator)

	// only if we're the first processor or being inserted at front-of-line:
	if insertPos == 0 {
		a.inputGroupUnit.relay.SetDest(pu.src)
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
}

func (a *Agent) startProcessor(processor models.ProcessorRunner) error {
	a.processorGroupUnit.Lock()

	for _, pu := range a.processorGroupUnit.processorUnits {
		if pu.processor.GetID() == processor.GetID() {
			a.processorGroupUnit.Unlock()
			err := processor.Start(pu.accumulator)
			if err != nil {
				return fmt.Errorf("starting processor %s: %w", processor.LogName(), err)
			}

			return nil
		}
	}

	a.processorGroupUnit.Unlock()
	return nil
}

// RunProcessor is a blocking call that runs a processor forever
func (a *Agent) RunProcessor(p models.ProcessorRunner) {
	a.processorGroupUnit.Lock()
	pu := a.processorGroupUnit.Find(p)
	a.processorGroupUnit.Unlock()

	if pu == nil {
		log.Print("W! [agent] Must call AddProcessor before calling RunProcessor")
		return
	}

	err := errors.New("loop at least once")
	for err != nil && a.ctx.Err() == nil {
		if err = a.startProcessor(p); err != nil {
			log.Printf("E! [agent] failed to start processor %q: %v", p.LogName(), err)
			time.Sleep(10 * time.Second)
		}
	}

	// processors own their src channel, if it's closed either it's been asked to stop or we're shutting down
	for m := range pu.src {
		if err := pu.processor.Add(m, pu.accumulator); err != nil {
			pu.accumulator.AddError(err)
			m.Reject()
		}
	}

	pu.processor.Stop()
	// only close dst channel if we're shutting down
	if a.ctx.Err() != nil {
		a.processorGroupUnit.Lock()
		if pu.dst != nil {
			close(pu.dst)
		}
		a.processorGroupUnit.Unlock()
	}

	a.processorGroupUnit.Lock()
	defer a.processorGroupUnit.Unlock()
	for i, curr := range a.processorGroupUnit.processorUnits {
		if pu == curr {
			// remove it from the slice
			// Remove the stopped processor from the units list
			a.processorGroupUnit.processorUnits = append(a.processorGroupUnit.processorUnits[0:i], a.processorGroupUnit.processorUnits[i+1:len(a.processorGroupUnit.processorUnits)]...)
		}
	}
}

// StopProcessor stops processors or aggregators.
// ProcessorRunner could be a *models.RunningProcessor or a *models.RunningAggregator
func (a *Agent) StopProcessor(p models.ProcessorRunner) {
	if a.isState(agentStateShuttingDown) {
		return
	}

	a.processorGroupUnit.Lock()
	defer a.processorGroupUnit.Unlock()

	for i, curr := range a.processorGroupUnit.processorUnits {
		if p.GetID() == curr.processor.GetID() {
			if i == 0 {
				a.inputGroupUnit.relay.SetDest(curr.dst)
			} else {
				prev := a.processorGroupUnit.processorUnits[i-1]
				prev.accumulator.setOutput(curr.dst)
			}
			close(curr.src) // closing source will tell the processor to stop.
		}
	}
}

// RunProcessor is a blocking call that runs a processor forever
func (a *Agent) RunConfigPlugin(ctx context.Context, plugin config.ConfigPlugin) {
	a.configPluginUnit.Lock()
	a.configPluginUnit.plugins = append(a.configPluginUnit.plugins, plugin)
	a.configPluginUnit.Unlock()

	<-ctx.Done()
	//TODO: we might want to wait for all other plugins to close?
	if err := plugin.Close(); err != nil {
		log.Printf("E! [agent] Configuration plugin failed to close: %v", err)
	}

	// because we don't have wrappers for config plugins, check to see if there's a storage plugin attached and close it.
	p := reflect.ValueOf(plugin)
	if p.Kind() == reflect.Ptr {
		p = p.Elem()
	}
	if v := p.FieldByName("Storage"); !v.IsZero() && v.IsValid() && !v.IsNil() {
		if sp, ok := v.Interface().(config.StoragePlugin); ok {
			if err := sp.Close(); err != nil {
				log.Printf("E! [agent] Storage plugin failed to close: %v", err)
			}
		}
	}

	a.configPluginUnit.Lock()
	pos := -1
	for i, p := range a.configPluginUnit.plugins {
		if p == plugin {
			pos = i
			break
		}
	}
	if pos > -1 {
		a.configPluginUnit.plugins = append(a.configPluginUnit.plugins[0:pos], a.configPluginUnit.plugins[pos+1:]...)
	}
	a.configPluginUnit.Unlock()
}

func (a *Agent) StopOutput(output *models.RunningOutput) {
	if a.isState(agentStateShuttingDown) {
		return
	}
	a.outputGroupUnit.Lock()
	defer a.outputGroupUnit.Unlock()

	// find plugin
	for _, o := range a.outputGroupUnit.outputs {
		if o.output == output {
			if o.cancelFlush != nil {
				o.cancelFlush()
			}
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

func (a *Agent) AddOutput(output *models.RunningOutput) {
	if a.isState(agentStateShuttingDown) {
		return
	}
	a.outputGroupUnit.Lock()
	a.outputGroupUnit.outputs = append(a.outputGroupUnit.outputs, outputUnit{output: output})
	a.outputGroupUnit.Unlock()
}

func (a *Agent) startOutput(output *models.RunningOutput) error {
	if err := a.connectOutput(a.ctx, output); err != nil {
		return fmt.Errorf("connecting output %s: %w", output.LogName(), err)
	}

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

// RunOutput runs an output; note the context should be a special context that
// only cancels when it's time for the outputs to close: when the main context
// has closed AND all the input and processor plugins are done.
func (a *Agent) RunOutput(ctx context.Context, output *models.RunningOutput) {
	var cancel context.CancelFunc
	// wrap with a cancel context so that the StopOutput can stop this individual output without stopping all the outputs.
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	a.outputGroupUnit.Lock()
	for i, o := range a.outputGroupUnit.outputs {
		if o.output == output {
			a.outputGroupUnit.outputs[i].cancelFlush = cancel
		}
	}
	a.outputGroupUnit.Unlock()

	interval := time.Duration(a.Config.Agent.FlushInterval)
	jitter := time.Duration(a.Config.Agent.FlushJitter)
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

	err := errors.New("loop at least once")
	for err != nil && a.ctx.Err() == nil {
		if err = a.startOutput(output); err != nil {
			log.Printf("E! [agent] failed to start output %q: %v", output.LogName(), err)
			time.Sleep(10 * time.Second)
		}
	}

	a.flushLoop(ctx, output, ticker)

	a.outputGroupUnit.Lock()

	// find plugin
	for i, o := range a.outputGroupUnit.outputs {
		if o.output == output {
			// disconnect it from the output broadcaster and remove it from the list
			a.outputGroupUnit.outputs = append(a.outputGroupUnit.outputs[:i], a.outputGroupUnit.outputs[i+1:]...)
		}
	}
	a.outputGroupUnit.Unlock()

	if err = a.flushOnce(output, ticker, output.Write); err != nil {
		log.Printf("E! [agent] Error writing to %s: %v", output.LogName(), err)
	}

	output.Close()
}

// runOutputFanout does the outputGroupUnit fanout, copying a metric to all outputs
func (a *Agent) runOutputFanout() {
	for metric := range a.outputGroupUnit.src {
		// if there are no outputs I guess we're dropping them.
		a.outputGroupUnit.Lock()
		outs := a.outputGroupUnit.outputs
		a.outputGroupUnit.Unlock()
		for i, output := range outs {
			if i == len(outs)-1 {
				output.output.AddMetric(metric)
			} else {
				output.output.AddMetric(metric.Copy())
			}
		}
	}
}

// flushLoop runs an output's flush function periodically until the context is done.
func (a *Agent) flushLoop(
	ctx context.Context,
	output *models.RunningOutput,
	ticker *RollingTicker,
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
			ticker.Reset()
			logError(a.flushOnce(output, ticker, output.Write))
		case <-output.BatchReady:
			ticker.Reset()
			logError(a.flushOnce(output, ticker, output.WriteBatch))
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
	a.outputGroupUnit.Lock()
	defer a.outputGroupUnit.Unlock()
	runningOutputs := []*models.RunningOutput{}
	// make sure we allocate and use a new slice that doesn't need a lock
	for _, o := range a.outputGroupUnit.outputs {
		runningOutputs = append(runningOutputs, o.output)
	}

	return runningOutputs
}

func (a *Agent) Context() context.Context {
	a.inputGroupUnit.Lock()
	defer a.inputGroupUnit.Unlock()
	return a.ctx
}

func (a *Agent) waitForPluginsToStop() {
	for {
		a.inputGroupUnit.Lock()
		if len(a.inputGroupUnit.inputUnits) > 0 {
			// fmt.Printf("waiting for %d inputs\n", len(a.inputGroupUnit.inputUnits))
			a.inputGroupUnit.Unlock()
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}
	close(a.inputGroupUnit.dst)
	a.inputGroupUnit.Unlock()
	for {
		time.Sleep(100 * time.Millisecond)
		a.processorGroupUnit.Lock()
		if len(a.processorGroupUnit.processorUnits) > 0 {
			// fmt.Printf("waiting for %d processors\n", len(a.processorGroupUnit.processorUnits))
			a.processorGroupUnit.Unlock()
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}
	a.processorGroupUnit.Unlock()
	for {
		a.outputGroupUnit.Lock()
		if len(a.outputGroupUnit.outputs) > 0 {
			// fmt.Printf("waiting for %d outputs\n", len(a.outputGroupUnit.outputs))
			a.outputGroupUnit.Unlock()
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}
	a.outputGroupUnit.Unlock()

	for {
		a.configPluginUnit.Lock()
		if len(a.configPluginUnit.plugins) > 0 {
			// fmt.Printf("waiting for %d config plugins\n", len(a.configPluginUnit.plugins))
			a.configPluginUnit.Unlock()
			continue
		}
		break
	}
	a.configPluginUnit.Unlock()

	// everything closed; shut down
}

// setState sets the agent's internal state.
func (a *Agent) setState(newState agentState) {
	a.stateChanged.L.Lock()
	a.state = newState
	a.stateChanged.Broadcast()
	a.stateChanged.L.Unlock()
}

// isState returns true if the agent state matches the state parameter
func (a *Agent) isState(state agentState) bool {
	a.stateChanged.L.Lock()
	defer a.stateChanged.L.Unlock()
	return a.state == state
}

// waitUntilState waits until the agent state is the requested state.
func (a *Agent) waitUntilState(state agentState) {
	a.stateChanged.L.Lock()
	for a.state != state {
		a.stateChanged.Wait()
	}
	a.stateChanged.L.Unlock()
}

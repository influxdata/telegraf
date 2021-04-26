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
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/channel"
	"github.com/influxdata/telegraf/models"
)

// Agent runs a set of plugins.
type Agent struct {
	Config *config.Config

	// units hold channels and define connections between plugins
	outputUnit         outputUnit
	processorGroupUnit processorGroupUnit
	// inputUnit          inputUnit
	inputGroupUnit   inputGroupUnit
	configPluginUnit configPluginUnit

	ctx context.Context

	started    *sync.Cond // a condition that lets plugins know when hasStarted is true (when the agent starts)
	hasStarted bool
}

// NewAgent returns an Agent for the given Config.
func NewAgent(ctx context.Context, config *config.Config) *Agent {
	inputDestCh := make(chan telegraf.Metric)
	outputSrcCh := make(chan telegraf.Metric)

	// by default, connect the dest of the inputs directly to the src for the outputs,
	// as processors are added, they will be inserted between these two.

	return &Agent{
		Config:  config,
		ctx:     ctx,
		started: sync.NewCond(&sync.Mutex{}),
		inputGroupUnit: inputGroupUnit{
			dst:   inputDestCh,
			relay: channel.NewRelay(inputDestCh, outputSrcCh),
		},
		outputUnit: outputUnit{
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
	runningInputUnitCount int32
	dst                   chan<- telegraf.Metric // must stay open until app is shutting down
	relay                 *channel.Relay
	inputUnits            []inputUnit
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
		a.Config.Agent.Interval.Duration, a.Config.Agent.Quiet,
		a.Config.Agent.Hostname, a.Config.Agent.FlushInterval.Duration)

	a.inputGroupUnit.relay.Start()

	a.started.L.Lock()
	a.hasStarted = true
	a.started.Broadcast()
	a.started.L.Unlock()

	<-a.Context().Done()

	a.inputGroupUnit.Lock()
	for _, iu := range a.inputGroupUnit.inputUnits {
		iu.input.Stop()
	}
	a.inputGroupUnit.Unlock()

	// wait for all plugins to stop
	a.waitForPluginsToStop()

	log.Printf("D! [agent] Stopped Successfully")
}

// // Run starts and runs the Agent until the context is done.
// func (a *Agent) Run(ctx context.Context) error {
// 	a.ctx = ctx
// 	log.Printf("I! [agent] Config: Interval:%s, Quiet:%#v, Hostname:%#v, "+
// 		"Flush Interval:%s",
// 		a.Config.Agent.Interval.Duration, a.Config.Agent.Quiet,
// 		a.Config.Agent.Hostname, a.Config.Agent.FlushInterval.Duration)

// 	log.Printf("D! [agent] Connecting outputs")
// 	a.startOutputs(ctx)

// 	if err := a.startProcessors(a.outputUnit.src); err != nil {
// 		return err
// 	}

// 	if err := a.startInputs(); err != nil {
// 		return err
// 	}

// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		err := a.runOutputs()
// 		if err != nil {
// 			log.Printf("E! [agent] Error running outputs: %v", err)
// 		}
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		a.runProcessors()
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		err := a.runInputs()
// 		if err != nil {
// 			log.Printf("E! [agent] Error running inputs: %v", err)
// 		}
// 	}()

// 	wg.Wait()

// 	log.Printf("D! [agent] Stopped Successfully")
// 	return nil
// }

// func (a *Agent) startInputs() error {
// 	log.Printf("D! [agent] Starting service inputs")

// 	for _, inputUnit := range a.inputGroupUnit.inputUnits {
// 		if err := a.StartInput(inputUnit.input); err != nil {
// 			log.Printf("Error starting input %s: %v", inputUnit.input.LogName(), err)
// 		}
// 	}

// 	return nil
// }

func (a *Agent) AddInput(input *models.RunningInput) {
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
	a.started.L.Lock()
	for !a.hasStarted {
		a.started.Wait()
	}
	a.started.L.Unlock()

	a.inputGroupUnit.Lock()
	a.processorGroupUnit.Lock()
	a.outputUnit.Lock()

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
			a.outputUnit.Unlock()
			a.processorGroupUnit.Unlock()
			a.inputGroupUnit.Unlock()
			return input.Start(acc)
		}
	}
	a.outputUnit.Unlock()
	a.processorGroupUnit.Unlock()
	a.inputGroupUnit.Unlock()

	return errors.New("cannot start input; call AddInput first")
}

// // runInputs starts and triggers the periodic gather for Inputs.
// //
// // When the context is done the timers are stopped and this function returns
// // after all ongoing Gather calls complete.
// func (a *Agent) runInputs() error {
// 	startTime := time.Now()
// 	var wg sync.WaitGroup
// 	for _, iu := range a.inputGroupUnit.inputUnits {
// 		wg.Add(1)
// 		go func(input *models.RunningInput) {
// 			defer wg.Done()
// 			a.RunInput(input, startTime)
// 		}(iu.input)
// 	}

// 	wg.Wait()
// 	a.stopInputs()
// 	return nil
// }

// RunInput is a blocking call that runs an input forever
func (a *Agent) RunInput(input *models.RunningInput, startTime time.Time) {
	// default to agent interval but check for override
	interval := a.Config.Agent.Interval.Duration
	if input.Config.Interval != 0 {
		interval = input.Config.Interval
	}

	// default to agent precision but check for override
	precision := a.Config.Agent.Precision.Duration
	if input.Config.Precision != 0 {
		precision = input.Config.Precision
	}

	// default to agent collection_jitter but check for override
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
	defer cancelFunc() // just to keep linters happy

	atomic.AddInt32(&a.inputGroupUnit.runningInputUnitCount, 1)
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

	inputsLeft := atomic.AddInt32(&a.inputGroupUnit.runningInputUnitCount, -1)
	if ctx.Err() != nil && inputsLeft == 0 { // shutting down
		// a.inputGroupUnit.Lock()
		close(a.inputGroupUnit.dst)
		// a.inputGroupUnit.Unlock()
	}
}

// // testStartInputs is a variation of startInputs for use in --test and --once
// // mode.  It differs by logging Start errors and returning only plugins
// // successfully started.
// func (a *Agent) testStartInputs() error {
// 	log.Printf("D! [agent] Starting service inputs")

// 	dst := a.outputUnit.src
// 	if len(a.processorGroupUnit.processorUnits) > 0 {
// 		dst = a.processorGroupUnit.processorUnits[0].src
// 	}
// 	unit := &inputGroupUnit{
// 		dst: dst,
// 	}

// 	for _, inputUnit := range a.inputGroupUnit.inputUnits {
// 		// Overwrite agent interval if this plugin has its own.
// 		interval := a.Config.Agent.Interval.Duration
// 		if inputUnit.input.Config.Interval != 0 {
// 			interval = inputUnit.input.Config.Interval
// 		}

// 		// Overwrite agent precision if this plugin has its own.
// 		precision := a.Config.Agent.Precision.Duration
// 		if inputUnit.input.Config.Precision != 0 {
// 			precision = inputUnit.input.Config.Precision
// 		}

// 		acc := NewAccumulator(inputUnit.input, dst)
// 		acc.SetPrecision(getPrecision(precision, interval))

// 		if si, ok := inputUnit.input.Input.(telegraf.ServiceInput); ok {
// 			// Service input plugins are not subject to timestamp rounding.
// 			// This only applies to the accumulator passed to Start(), the
// 			// Gather() accumulator does apply rounding according to the
// 			// precision agent setting.

// 			err := si.Start(acc)
// 			if err != nil {
// 				log.Printf("E! [agent] Starting input %s: %v", inputUnit.input.LogName(), err)
// 			}

// 		}

// 	return nil
// }

// // testRunInputs is a variation of runInputs for use in --test and --once mode.
// // Instead of using a ticker to run the inputs they are called once immediately.
// func (a *Agent) testRunInputs(ctx context.Context, wait time.Duration) error {
// 	var wg sync.WaitGroup

// 	nul := make(chan telegraf.Metric)
// 	go func() {
// 		for range nul {
// 		}
// 	}()

// 	for _, iu := range a.inputGroupUnit.inputUnits {
// 		wg.Add(1)
// 		go func(iu inputUnit) {
// 			defer wg.Done()

// 			// Run plugins that require multiple gathers to calculate rate
// 			// and delta metrics twice.
// 			switch iu.input.Config.Name {
// 			case "cpu", "mongodb", "procstat":
// 				iu.accumulator.setOutput(nul)
// 				if err := iu.input.Gather(iu.accumulator); err != nil {
// 					iu.accumulator.AddError(err)
// 				}

// 				time.Sleep(500 * time.Millisecond)
// 			}

// 			if err := iu.input.Gather(iu.accumulator); err != nil {
// 				iu.accumulator.AddError(err)
// 			}
// 		}(iu)
// 	}
// 	wg.Wait()

// 	_ = internal.SleepContext(ctx, wait)

// 	a.stopInputs()
// 	return nil
// }

// stopInputs stops all service inputs.
// func (a *Agent) stopInputs() {
// 	log.Printf("D! [agent] Stopping service inputs")
// 	a.inputGroupUnit.Lock()
// 	defer a.inputGroupUnit.Unlock()

// 	for _, iu := range a.inputGroupUnit.inputUnits {
// 		iu.input.Stop()
// 	}
// 	log.Printf("D! [agent] Input channel closed")
// }

func (a *Agent) StopInput(i *models.RunningInput) {
	defer a.inputGroupUnit.Unlock()
retry:
	a.inputGroupUnit.Lock()
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
// func (a *Agent) startProcessors(dst chan<- telegraf.Metric) error {
// 	for _, processorUnit := range a.processorGroupUnit.processorUnits {
// 		if err := a.StartProcessor(processorUnit.processor); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

func (a *Agent) AddProcessor(processor models.ProcessorRunner) {
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
		pu.dst = a.outputUnit.src
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

// // runProcessors begins processing metrics and runs until the source channel is
// // closed and all metrics have been written.
// func (a *Agent) runProcessors() {
// 	var wg sync.WaitGroup
// 	for _, unit := range a.processorGroupUnit.processorUnits {
// 		wg.Add(1)
// 		go func(unit *processorUnit) {
// 			defer wg.Done()
// 			a.RunProcessor(unit.processor)

// 			// log.Printf("D! [agent] Processor channel closed")
// 		}(unit)
// 	}
// 	wg.Wait()
// }

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
}

// StopProcessor stops processors or aggregators.
// ProcessorRunner could be a *models.RunningProcessor or a *models.RunningAggregator
func (a *Agent) StopProcessor(p models.ProcessorRunner) {
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

			// remove it from the slice
			// Remove the stopped processor from the units list
			a.processorGroupUnit.processorUnits = append(a.processorGroupUnit.processorUnits[0:i], a.processorGroupUnit.processorUnits[i+1:len(a.processorGroupUnit.processorUnits)]...)
		}
	}
}

// RunProcessor is a blocking call that runs a processor forever
func (a *Agent) RunConfigPlugin(ctx context.Context, plugin config.ConfigPlugin) {
	a.configPluginUnit.Lock()
	a.configPluginUnit.plugins = append(a.configPluginUnit.plugins, plugin)
	a.configPluginUnit.Unlock()

	<-ctx.Done()
	fmt.Println("config plugin context closed")
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

func (a *Agent) StopOutput(p *models.RunningOutput) {
	a.outputUnit.Lock()
	defer a.outputUnit.Unlock()

	// find plugin
	for i, output := range a.outputUnit.outputs {
		if output.ID == p.ID {
			// disconnect it from the output broadcaster and remove it from the list
			a.outputUnit.outputs = append(a.outputUnit.outputs[0:i], a.outputUnit.outputs[i+1:len(a.outputUnit.outputs)]...)

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
// func (a *Agent) startOutputs(
// 	ctx context.Context,
// ) {
// 	src := make(chan telegraf.Metric, 100)
// 	a.outputUnit = outputUnit{src: src}
// 	for _, output := range a.Config.Outputs {
// 		if err := a.StartOutput(output); err != nil {
// 			log.Printf("E! [%s] Error starting output: %v", output.Name(), err)
// 		}
// 	}
// }

func (a *Agent) AddOutput(output *models.RunningOutput) {
	a.outputUnit.Lock()
	a.outputUnit.outputs = append(a.outputUnit.outputs, output)
	a.outputUnit.Unlock()
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

// // runOutputs begins processing metrics and returns until the source channel is
// // closed and all metrics have been written.  On shutdown metrics will be
// // written one last time and dropped if unsuccessful.
// func (a *Agent) runOutputs() error {
// 	var wg sync.WaitGroup

// 	// flush management context
// 	flushCtx, cancelFlush := context.WithCancel(context.Background())

// 	for _, output := range a.outputUnit.outputs {
// 		wg.Add(1)
// 		go func(output *models.RunningOutput) {
// 			defer wg.Done()

// 			a.RunOutput(flushCtx, output)
// 		}(output)
// 	}

// 	a.runOutputFanout()

// 	log.Println("I! [agent] Hang on, flushing any cached metrics before shutdown")
// 	cancelFlush()
// 	wg.Wait()

// 	return nil
// }

// RunOutput runs an output; note the context should be a special context that
// only cancels when it's time for the outputs to close: when the main context
// has closed AND all the input and processor plugins are done.
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

	err := errors.New("loop at least once")
	for err != nil && a.ctx.Err() == nil {
		if err = a.startOutput(output); err != nil {
			log.Printf("E! [agent] failed to start output %q: %v", output.LogName(), err)
			time.Sleep(10 * time.Second)
		}
	}

	a.flushLoop(ctx, output, ticker)
	output.Close()
}

// runOutputFanout does the outputUnit fanout, copying a metric to all outputs
func (a *Agent) runOutputFanout() {
	for metric := range a.outputUnit.src {
		// if there are no outputs I guess we're dropping them.
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

// // Test runs the inputs, processors and aggregators for a single gather and
// // writes the metrics to stdout.
// func (a *Agent) Test(ctx context.Context, wait time.Duration) error {
// 	src := make(chan telegraf.Metric, 100)

// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		s := influx.NewSerializer()
// 		s.SetFieldSortOrder(influx.SortFields)

// 		for metric := range src {
// 			octets, err := s.Serialize(metric)
// 			if err == nil {
// 				fmt.Print("> ", string(octets))
// 			}
// 			metric.Reject()
// 		}
// 	}()

// 	err := a.test(ctx, wait, src)
// 	if err != nil {
// 		return err
// 	}

// 	wg.Wait()

// 	if models.GlobalGatherErrors.Get() != 0 {
// 		return fmt.Errorf("input plugins recorded %d errors", models.GlobalGatherErrors.Get())
// 	}
// 	return nil
// }

// Test runs the agent and performs a single gather sending output to the
// outputF.  After gathering pauses for the wait duration to allow service
// inputs to run.
// func (a *Agent) test(ctx context.Context, wait time.Duration, outputC chan<- telegraf.Metric) error {
// 	a.startProcessors(outputC)
// 	a.testStartInputs()

// 	var wg sync.WaitGroup

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		a.runProcessors()
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		err := a.testRunInputs(ctx, wait)
// 		if err != nil {
// 			log.Printf("E! [agent] Error running inputs: %v", err)
// 		}
// 	}()

// 	wg.Wait()

// 	log.Printf("D! [agent] Stopped Successfully")

// 	return nil
// }

// // Once runs the full agent for a single gather.
// func (a *Agent) Once(ctx context.Context, wait time.Duration) error {
// 	err := a.once(ctx, wait)
// 	if err != nil {
// 		return err
// 	}

// 	if models.GlobalGatherErrors.Get() != 0 {
// 		return fmt.Errorf("input plugins recorded %d errors", models.GlobalGatherErrors.Get())
// 	}

// 	unsent := 0
// 	for _, output := range a.Config.Outputs {
// 		unsent += output.BufferLength()
// 	}
// 	if unsent != 0 {
// 		return fmt.Errorf("output plugins unable to send %d metrics", unsent)
// 	}
// 	return nil
// }

// // On runs the agent and performs a single gather sending output to the
// // outputF.  After gathering pauses for the wait duration to allow service
// // inputs to run.
// func (a *Agent) once(ctx context.Context, wait time.Duration) error {
// 	log.Printf("D! [agent] Connecting outputs")
// 	a.startOutputs(ctx)

// 	if err := a.startProcessors(a.outputUnit.src); err != nil {
// 		return err
// 	}

// 	err := a.testStartInputs()
// 	if err != nil {
// 		return err
// 	}

// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		err := a.runOutputs()
// 		if err != nil {
// 			log.Printf("E! [agent] Error running outputs: %v", err)
// 		}
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		a.runProcessors()
// 	}()

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		err := a.testRunInputs(ctx, wait)
// 		if err != nil {
// 			log.Printf("E! [agent] Error running inputs: %v", err)
// 		}
// 	}()

// 	wg.Wait()

// 	log.Printf("D! [agent] Stopped Successfully")

// 	return nil
// }

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

func (a *Agent) Context() context.Context {
	a.inputGroupUnit.Lock()
	defer a.inputGroupUnit.Unlock()
	return a.ctx
}

func (a *Agent) waitForPluginsToStop() {
	for {
		time.Sleep(10 * time.Millisecond)
		a.inputGroupUnit.Lock()
		if len(a.inputGroupUnit.inputUnits) > 0 {
			fmt.Printf("waiting for %d inputs\n", len(a.inputGroupUnit.inputUnits))
			a.inputGroupUnit.Unlock()
			continue
		}
		a.processorGroupUnit.Lock()
		if len(a.processorGroupUnit.processorUnits) > 0 {
			fmt.Printf("waiting for %d processors\n", len(a.processorGroupUnit.processorUnits))
			a.processorGroupUnit.Unlock()
			a.inputGroupUnit.Unlock()
			continue
		}
		a.outputUnit.Lock()
		if len(a.outputUnit.outputs) > 0 {
			fmt.Printf("waiting for %d outputs\n", len(a.outputUnit.outputs))
			a.outputUnit.Unlock()
			a.processorGroupUnit.Unlock()
			a.inputGroupUnit.Unlock()
			continue
		}
		a.outputUnit.Unlock()
		a.processorGroupUnit.Unlock()
		a.inputGroupUnit.Unlock()

		a.configPluginUnit.Lock()
		if len(a.configPluginUnit.plugins) > 0 {
			fmt.Printf("waiting for %d config plugins\n", len(a.configPluginUnit.plugins))
			a.configPluginUnit.Unlock()
			continue
		}
		a.configPluginUnit.Unlock()

		// everything closed; shut down
		return
	}
}

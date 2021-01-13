package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers/influx"

	"github.com/google/uuid"
)

// Agent runs a set of plugins.
type Agent struct {
	Config         *config.Config
	Context        context.Context
	iu             *inputUnit
	ou             *outputUnit
	ic             int
	oc             int
	runningPlugins map[string]interface{}

	pluginLock *sync.Mutex
	icLock     *sync.Mutex
	ocLock     *sync.Mutex
}

// NewAgent returns an Agent for the given Config.
func NewAgent(config *config.Config) (*Agent, error) {
	runningPlugins := make(map[string]interface{}) // map uuid:plugin

	a := &Agent{
		Config:         config,
		runningPlugins: runningPlugins,
		ic:             0,
		oc:             0,
		pluginLock:     new(sync.Mutex),
		icLock:         new(sync.Mutex),
		ocLock:         new(sync.Mutex),
	}
	return a, nil
}

// inputUnit is a group of input plugins and the shared channel they write to.
//
// ┌───────┐
// │ Input │───┐
// └───────┘   │
// ┌───────┐   │     ______
// │ Input │───┼──▶ ()_____)
// └───────┘   │
// ┌───────┐   │
// │ Input │───┘
//└───────┘
type inputUnit struct {
	dst    chan<- telegraf.Metric
	inputs []*models.RunningInput
}

//  ______     ┌───────────┐     ______
// ()_____)──▶ │ Processor │──▶ ()_____)
//             └───────────┘
type processorUnit struct {
	src       <-chan telegraf.Metric
	dst       chan<- telegraf.Metric
	processor *models.RunningProcessor
}

// aggregatorUnit is a group of Aggregators and their source and sink channels.
// Typically the aggregators write to a processor channel and pass the original
// metrics to the output channel.  The sink channels may be the same channel.
//
//                 ┌────────────┐
//            ┌──▶ │ Aggregator │───┐
//            │    └────────────┘   │
//  ______    │    ┌────────────┐   │     ______
// ()_____)───┼──▶ │ Aggregator │───┼──▶ ()_____)
//            │    └────────────┘   │
//            │    ┌────────────┐   │
//            ├──▶ │ Aggregator │───┘
//            │    └────────────┘
//            │                           ______
//            └────────────────────────▶ ()_____)
type aggregatorUnit struct {
	src         <-chan telegraf.Metric
	aggC        chan<- telegraf.Metric
	outputC     chan<- telegraf.Metric
	aggregators []*models.RunningAggregator
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
	src     <-chan telegraf.Metric
	outputs []*models.RunningOutput
}

// RunSingleInput runs a single input and can be called after an agent is created
func (a *Agent) RunSingleInput(input *models.RunningInput, ctx context.Context) error {

	// NOTE: we can't just use `defer a.statusMutex.Unlock()` since except for the input validation,
	// this function only returns once the gatherLoop is done -- i.e. when the plugin is stopped.
	// Be careful to manually unlock the statusMutex in this function.

	// validating if an input plugin is already running, and therefore shouldn't be run again
	a.pluginLock.Lock()
	_, ok := a.runningPlugins[input.UniqueId]
	a.pluginLock.Unlock()

	if ok {
		log.Printf("E! [agent] You are trying to run an input that is already running: %s \n", input.Config.Name)
		return errors.New("you are trying to run an input that is already running")
	}

	startTime := time.Now()

	if input.UniqueId == "" {
		uniqueID, err := uuid.NewUUID()
		if err != nil {
			return errors.New("errored while generating UUID for new INPUT")
		}
		input.UniqueId = uniqueID.String()
	}

	a.pluginLock.Lock()
	a.runningPlugins[input.UniqueId] = input
	a.pluginLock.Unlock()

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

	acc := NewAccumulator(input, a.iu.dst)
	acc.SetPrecision(getPrecision(precision, interval))

	a.incrementInputCount(1)
	go func(input *models.RunningInput) {
		defer ticker.Stop()
		a.gatherLoop(ctx, acc, input, ticker, interval)
		a.incrementInputCount(-1)
	}(input)

	a.Config.InputsLock.Lock()
	for _, i := range a.Config.Inputs {
		if i.UniqueId == input.UniqueId {
			a.Config.InputsLock.Unlock()
			return nil
		}
	}

	a.Config.Inputs = append(a.Config.Inputs, input)
	a.Config.InputsLock.Unlock()
	return nil
}

// RunSingleOutput runs a single output and can be called after an agent is created
func (a *Agent) RunSingleOutput(output *models.RunningOutput, ctx context.Context) error {
	// NOTE: we can't just use `defer a.statusMutex.Unlock()` since except for the output validation,
	// this function only returns once the gatherLoop is done -- i.e. when the plugin is stopped.
	// Be careful to manually unlock the statusMutex in this function.

	// validating if an output plugin is already running, and therefore shouldn't be run again
	a.pluginLock.Lock()
	_, ok := a.runningPlugins[output.UniqueId]
	a.pluginLock.Unlock()

	if ok {
		log.Printf("E! [agent] You are trying to run an output that is already running: %s \n", output.Config.Name)
		return errors.New("you are trying to run an output that is already running")
	}

	// Start flush loop
	interval := a.Config.Agent.FlushInterval.Duration
	jitter := a.Config.Agent.FlushJitter.Duration

	if output.UniqueId == "" {
		uniqueID, err := uuid.NewUUID()
		if err != nil {
			return errors.New("errored while generating UUID for new INPUT")
		}
		output.UniqueId = uniqueID.String()
	}

	a.pluginLock.Lock()
	a.runningPlugins[output.UniqueId] = output
	a.pluginLock.Unlock()

	// Overwrite agent flush_interval if this plugin has its own.
	if output.Config.FlushInterval != 0 {
		interval = output.Config.FlushInterval
	}

	// Overwrite agent flush_jitter if this plugin has its own.
	if output.Config.FlushJitter != 0 {
		jitter = output.Config.FlushJitter
	}

	a.incrementOutputCount(1)
	go func(output *models.RunningOutput) {
		ticker := NewRollingTicker(interval, jitter)
		defer ticker.Stop()

		a.flushLoop(ctx, output, ticker)
		a.incrementOutputCount(-1)
	}(output)

	a.Config.OutputsLock.Lock()
	for _, i := range a.Config.Outputs {
		if i.UniqueId == output.UniqueId {
			a.Config.OutputsLock.Unlock()
			return nil
		}
	}

	a.Config.Outputs = append(a.Config.Outputs, output)
	a.Config.OutputsLock.Unlock()
	return nil
}

func (a *Agent) incrementInputCount(by int) {
	a.icLock.Lock()
	a.ic += by
	a.icLock.Unlock()
}

func (a *Agent) incrementOutputCount(by int) {
	a.ocLock.Lock()
	a.oc += by
	a.ocLock.Unlock()
}


func GetAllInputPlugins() []string {
	var res []string
	for name := range inputs.Inputs {
		res = append(res, name)
	}
	return res
}

func GetAllOutputPlugins() []string {
	var res []string
	for name := range outputs.Outputs {
		res = append(res, name)
	}
	return res
}

func (a *Agent) GetRunningInputPlugins() []map[string]string {
	var res []map[string]string
	a.Config.InputsLock.Lock()
	for _, runningInput := range a.Config.Inputs {
		res = append(res, map[string]string{"name": runningInput.Config.Name, "id": runningInput.UniqueId})
	}
	a.Config.InputsLock.Unlock()
	return res
}

func (a *Agent) GetRunningOutputPlugins() []map[string]string {
	var res []map[string]string
	a.Config.OutputsLock.Lock()
	for _, runningOutput := range a.Config.Outputs {
		res = append(res, map[string]string{"name": runningOutput.Config.Name, "id": runningOutput.UniqueId})
	}
	a.Config.OutputsLock.Unlock()
	return res
}

type MapFieldSchema struct {
	Value interface{}
	Key   string
}

type ArrayFieldSchema struct {
	Value  interface{}
	Length int // Will be 0 if slice and array length if array
}

// GetPluginTypes returns a map of a plugin's field names to field value types
func (a *Agent) GetPluginTypes(p interface{}) (map[string]interface{}, error) {

	data := reflect.ValueOf(p).Elem() // extract Value of type interface{} from Value pointer to interface
	schema := getFieldType(data.Type())

	s, ok := schema.(map[string]interface{})

	if ok {
		return s, nil
	}

	return nil, fmt.Errorf("returned schema is not a map")
}

func getFieldType(data reflect.Type) interface{} {
	switch data.Kind() {
	case reflect.Struct:
		dataFields := make(map[string]interface{})
		for i := 0; i < data.NumField(); i++ {
			field := data.Field(i)
			if field.PkgPath == "" { // only take exported fields
				dataFields[field.Name] = getFieldType(field.Type)
			}
		}
		return dataFields
	case reflect.Array:
		valueType := getFieldType(data.Elem())
		return ArrayFieldSchema{valueType, data.Len()}
	case reflect.Slice:
		valueType := getFieldType(data.Elem())
		return ArrayFieldSchema{valueType, 0}
	case reflect.Map:
		keyType := data.Key().Name()
		valueType := getFieldType(data.Elem())
		return MapFieldSchema{valueType, keyType}
	default:
		return data.Kind().String()
	}
}

func (a *Agent) GetPluginValues(p interface{}) (map[string]interface{}, error) {
	data := reflect.ValueOf(p).Elem() // extract Value of type interface{} from Value pointer to interface
	values := getFieldValue(data)
	if values == nil {
		values = make(map[string]interface{})
	}
	v, ok := values.(map[string]interface{})

	if ok {
		return v, nil
	}

	return nil, fmt.Errorf("returned schema is not a map")
}

func getFieldValue(data reflect.Value) interface{} {
	switch data.Kind() {
	case reflect.Struct:
		dataFields := make(map[string]interface{})
		dataType := data.Type()
		for i := 0; i < data.NumField(); i++ {
			field := data.Field(i)
			tField := dataType.Field(i)
			if tField.PkgPath == "" { // only take exported fields
				v := getFieldValue(field)
				if v != nil {
					dataFields[tField.Name] = v
				}
			}
		}
		if len(dataFields) == 0 {
			return nil
		}
		return dataFields
	default:
		// These data types, false or zero will make isZero true. Don't want to return nil!
		if data.Kind() == reflect.Bool || data.Kind() == reflect.Int || data.Kind() == reflect.Float32 {
			return data.Interface()
		}

		// For more complex data types, we will return nil if empty.
		if data.IsZero() {
			return nil
		}
		return data.Interface()
	}
}

// Run starts and runs the Agent until the context is done.
func (a *Agent) Run(ctx context.Context) error {
	a.Context = ctx
	log.Printf("I! [agent] Config: Interval:%s, Quiet:%#v, Hostname:%#v, "+
		"Flush Interval:%s",
		a.Config.Agent.Interval.Duration, a.Config.Agent.Quiet,
		a.Config.Agent.Hostname, a.Config.Agent.FlushInterval.Duration)

	log.Printf("D! [agent] Initializing plugins")
	err := a.initPlugins()
	if err != nil {
		return err
	}

	startTime := time.Now()
	log.Printf("D! [agent] Connecting outputs")
	next, ou, err := a.startOutputs(ctx, a.Config.Outputs)
	if err != nil {
		return err
	}
	a.ou = ou
	var apu []*processorUnit
	var au *aggregatorUnit
	if len(a.Config.Aggregators) != 0 {
		aggC := next
		if len(a.Config.AggProcessors) != 0 {
			aggC, apu, err = a.startProcessors(next, a.Config.AggProcessors)
			if err != nil {
				return err
			}
		}

		next, au, err = a.startAggregators(aggC, next, a.Config.Aggregators)
		if err != nil {
			return err
		}
	}

	var pu []*processorUnit
	if len(a.Config.Processors) != 0 {
		next, pu, err = a.startProcessors(next, a.Config.Processors)
		if err != nil {
			return err
		}
	}
	iu, err := a.startInputs(next, a.Config.Inputs)
	a.iu = iu
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := a.runOutputs(ou)
		if err != nil {
			log.Printf("E! [agent] Error running outputs: %v", err)
		}
	}()

	if au != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := a.runProcessors(apu)
			if err != nil {
				log.Printf("E! [agent] Error running processors: %v", err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := a.runAggregators(startTime, au)
			if err != nil {
				log.Printf("E! [agent] Error running aggregators: %v", err)
			}
		}()
	}

	if pu != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := a.runProcessors(pu)
			if err != nil {
				log.Printf("E! [agent] Error running processors: %v", err)
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := a.runInputs(ctx, startTime, iu)
		if err != nil {
			log.Printf("E! [agent] Error running inputs: %v", err)
		}
	}()

	wg.Wait()

	log.Printf("D! [agent] Stopped Successfully")
	return err
}

// StartInput adds an input plugin with default config
func (a *Agent) StartInput(ctx context.Context, pluginName string) (string, error) {
	inputConfig := models.InputConfig{
		Name: pluginName,
	}

	input, err := a.CreateInput(pluginName)
	if err != nil {
		return "", err
	}

	uniqueId, err := uuid.NewUUID()
	if err != nil {
		return "", errors.New("errored while generating UUID for new INPUT")
	}
	ri := models.NewRunningInput(input, &inputConfig, uniqueId.String())

	err = ri.Init()
	if err != nil {
		return "", err
	}

	err = a.RunSingleInput(ri, ctx)
	if err != nil {
		return "", err
	}

	// add new input to inputunit
	a.iu.inputs = append(a.iu.inputs, ri)

	err = a.Config.UpdateConfig(
		map[string]interface{}{
			"unique_id": uniqueId.String(),
			"name":      pluginName,
		},
		uniqueId.String(), "inputs", "START_PLUGIN")

	if err != nil {
		log.Printf("W! [agent] Unable to save configuration for input %s", uniqueId.String())
	}

	return uniqueId.String(), nil
}

// StartOutput adds an output plugin with default config
func (a *Agent) StartOutput(ctx context.Context, pluginName string) (string, error) {
	outputConfig := models.OutputConfig{
		Name: pluginName,
	}

	output, err := a.CreateOutput(pluginName)
	if err != nil {
		return "", err
	}

	uniqueId, err := uuid.NewUUID()
	if err != nil {
		return "", errors.New("errored while generating UUID for new INPUT")
	}

	ro := models.NewRunningOutput(pluginName, output, &outputConfig,
		a.Config.Agent.MetricBatchSize, a.Config.Agent.MetricBufferLimit, uniqueId.String())

	err = ro.Init()
	if err != nil {
		return "", err
	}

	err = a.connectOutput(ctx, ro)
	if err != nil {
		return "", err
	}

	err = a.RunSingleOutput(ro, ctx)
	if err != nil {
		return "", err
	}

	// add new output to outputunit
	a.ou.outputs = append(a.ou.outputs, ro)

	err = a.Config.UpdateConfig(map[string]interface{}{"unique_id": uniqueId.String(), "name": pluginName}, uniqueId.String(), "outputs", "START_PLUGIN")
	if err != nil {
		log.Printf("W! [agent] Unable to save configuration for output %s", uniqueId.String())
	}
	return uniqueId.String(), nil
}

// generateTomlKeysMap generates a map with the toml keys for a specific plugin
func generateTomlKeysMap(structPtr reflect.Value, config map[string]interface{}) (map[string]interface{}, error) {
	strct := structPtr.Elem()
	tomlMap := map[string]interface{}{}
	pType := strct.Type()

	for configKey, configValue := range config {
		field, found := pType.FieldByName(configKey)

		if !found {
			return map[string]interface{}{}, fmt.Errorf("field %s did not exist on plugin", configKey)
		}

		tomlTag := field.Tag.Get("toml")
		if tomlTag == "" {
			tomlTag = configKey
		}

		tomlMap[tomlTag] = configValue
	}

	return tomlMap, nil

}

// CreateInput creates a new input from the name of an input.
func (a *Agent) CreateInput(name string) (telegraf.Input, error) {
	p, exists := inputs.Inputs[name]
	if exists {
		return p(), nil
	}
	return nil, fmt.Errorf("could not find input plugin with name: %s", name)
}

// CreateOutput creates a new output from the name of an output.
func (a *Agent) CreateOutput(name string) (telegraf.Output, error) {
	p, exists := outputs.Outputs[name]
	if exists {
		return p(), nil
	}
	return nil, fmt.Errorf("could not find output plugin with name: %s", name)
}

// GetRunningPlugin gets the values of a running plugin's struct.
func (a *Agent) GetRunningPlugin(uid string) (map[string]interface{}, error) {
	a.pluginLock.Lock()
	obj, exists := a.runningPlugins[uid]
	a.pluginLock.Unlock()
	if !exists {
		return nil, fmt.Errorf("specified plugin is not running")
	}
	ri, isRi := obj.(*models.RunningInput)
	if isRi {
		return a.GetPluginValues(ri.Input)
	}
	ro, isRo := obj.(*models.RunningOutput)
	if isRo {
		return a.GetPluginValues(ro.Output)
	}
	return nil, fmt.Errorf("invalid running plugin")
}

// If the config is compatible with the struct, return the JSON vesion of the config.
func validateStructConfig(structPtr reflect.Value, config map[string]interface{}) ([]byte, error) {
	strct := structPtr.Elem() // extract Value of type interface{} from Value pointer to interface

	if strct.Kind() == reflect.Struct {
		// Check for if trying to modify unsettable fields and fields that don't exist.
		for configKey := range config {
			pluginField := strct.FieldByName(configKey) // cast new value as Value
			if !(pluginField.IsValid()) {
				return nil, fmt.Errorf("invalid field name %s", configKey)
			} else if !(pluginField.CanSet()) {
				return nil, fmt.Errorf("unsettable field %s", configKey)
			}
		}

		// Creates a copy of the struct and see if JSON Unmarshal works without errors
		empty := reflect.New(strct.Type())
		cJSON, err := json.Marshal(config)
		err = json.Unmarshal(cJSON, empty.Interface())
		if err != nil {
			return nil, err
		}
		return cJSON, nil
	}

	return nil, fmt.Errorf("pointer does not point to a struct")
}

// UpdateInputPlugin gets the config for an input plugin given its name
func (a *Agent) UpdateInputPlugin(uid string, config map[string]interface{}) (telegraf.Input, error) {
	a.pluginLock.Lock()
	plugin, ok := a.runningPlugins[uid]
	a.pluginLock.Unlock()

	if !ok {
		log.Printf("E! [agent] You are trying to update an input that does not exist: %s \n", uid)
		return nil, errors.New("you are trying to update an input that does not exist")
	}

	input := plugin.(*models.RunningInput)

	// This code creates a copy of the struct and see if JSON Unmarshal works without errors
	configJSON, err := validateStructConfig(reflect.ValueOf(input.Input), config)
	if err != nil {
		return nil, fmt.Errorf("could not update input plugin %s with error: %s", uid, err)
	}

	tomlMap, err := generateTomlKeysMap(reflect.ValueOf(input.Input), config)
	if err != nil {
		return nil, fmt.Errorf("could not update input plugin %s with error: %s", uid, err)
	}

	if len(a.Config.Inputs) == 1 {
		a.incrementInputCount(1)
	}

	err = a.StopInputPlugin(uid, false)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(configJSON, &input.Input)
	if err != nil {
		return nil, err
	}

	ri := models.NewRunningInput(input.Input, input.Config, input.UniqueId)
	err = a.RunSingleInput(ri, a.Context)
	if err != nil {
		return nil, err
	}

	err = a.Config.UpdateConfig(tomlMap, input.UniqueId, "inputs", "UPDATE_PLUGIN")
	if err != nil {
		return nil, fmt.Errorf("could not update input plugin %s with error: %s", uid, err)
	}

	if len(a.Config.Inputs) == 1 {
		a.incrementInputCount(-1)
	}

	return input.Input, nil
}

// UpdateOutputPlugin gets the config for an output plugin given its name
func (a *Agent) UpdateOutputPlugin(uid string, config map[string]interface{}) (telegraf.Output, error) {
	a.pluginLock.Lock()
	plugin, ok := a.runningPlugins[uid]
	a.pluginLock.Unlock()

	if !ok {
		log.Printf("E! [agent] You are trying to update an output that does not exist: %s \n", uid)
		return nil, errors.New("you are trying to update an output that does not exist")
	}

	output := plugin.(*models.RunningOutput)

	// This code creates a copy of the struct and see if JSON Unmarshal works without errors
	configJSON, err := validateStructConfig(reflect.ValueOf(output.Output), config)
	if err != nil {
		return nil, fmt.Errorf("could not update output plugin %s with error: %s", uid, err)
	}

	tomlMap, err := generateTomlKeysMap(reflect.ValueOf(output.Output), config)
	if err != nil {
		return nil, fmt.Errorf("could not update output plugin %s with error: %s", uid, err)
	}

	if len(a.Config.Outputs) == 1 {
		a.incrementOutputCount(1)
	}

	err = a.StopOutputPlugin(uid, false)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(configJSON, &output.Output)
	if err != nil {
		return nil, err
	}

	ro := models.NewRunningOutput(output.Config.Name, output.Output, output.Config,
		a.Config.Agent.MetricBatchSize, a.Config.Agent.MetricBufferLimit, output.UniqueId)

	err = a.RunSingleOutput(ro, a.Context)
	if err != nil {
		return nil, err
	}

	err = a.Config.UpdateConfig(tomlMap, output.UniqueId, "outputs", "UPDATE_PLUGIN")
	if err != nil {
		return nil, fmt.Errorf("could not update output plugin %s with error: %s", uid, err)
	}

	if len(a.Config.Outputs) == 1 {
		a.incrementOutputCount(-1)
	}

	return output.Output, nil
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
	for _, processor := range a.Config.AggProcessors {
		err := processor.Init()
		if err != nil {
			return fmt.Errorf("could not initialize processor %s: %v",
				processor.Config.Name, err)
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

func (a *Agent) startInputs(
	dst chan<- telegraf.Metric,
	inputs []*models.RunningInput,
) (*inputUnit, error) {
	log.Printf("D! [agent] Starting service inputs")

	unit := &inputUnit{
		dst: dst,
	}

	for _, input := range inputs {
		if si, ok := input.Input.(telegraf.ServiceInput); ok {
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

			acc := NewAccumulator(input, dst)
			acc.SetPrecision(getPrecision(precision, interval))

			err := si.Start(acc)
			if err != nil {
				stopServiceInputs(unit.inputs)
				return nil, fmt.Errorf("starting input %s: %w", input.LogName(), err)
			}
		}
		unit.inputs = append(unit.inputs, input)
	}

	return unit, nil
}

// runInputs starts and triggers the periodic gather for Inputs.
//
// When the context is done the timers are stopped and this function returns
// after all ongoing Gather calls complete.
func (a *Agent) runInputs(
	ctx context.Context,
	startTime time.Time,
	unit *inputUnit,
) error {

	for _, input := range unit.inputs {
		err := a.RunSingleInput(input, ctx)
		if err != nil {
			return err
		}
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("D! [agent] Stopping service inputs")
			stopServiceInputs(a.Config.Inputs)

			close(unit.dst)
			log.Printf("D! [agent] Input channel closed")
			return nil
		}
	}
}

// testStartInputs is a variation of startInputs for use in --test and --once
// mode.  It differs by logging Start errors and returning only plugins
// successfully started.
func (a *Agent) testStartInputs(
	dst chan<- telegraf.Metric,
	inputs []*models.RunningInput,
) (*inputUnit, error) {
	log.Printf("D! [agent] Starting service inputs")

	unit := &inputUnit{
		dst: dst,
	}

	for _, input := range inputs {
		if si, ok := input.Input.(telegraf.ServiceInput); ok {
			// Service input plugins are not subject to timestamp rounding.
			// This only applies to the accumulator passed to Start(), the
			// Gather() accumulator does apply rounding according to the
			// precision agent setting.
			acc := NewAccumulator(input, dst)
			acc.SetPrecision(time.Nanosecond)

			err := si.Start(acc)
			if err != nil {
				log.Printf("E! [agent] Starting input %s: %v", input.LogName(), err)
			}

		}

		unit.inputs = append(unit.inputs, input)
	}

	return unit, nil
}

// testRunInputs is a variation of runInputs for use in --test and --once mode.
// Instead of using a ticker to run the inputs they are called once immediately.
func (a *Agent) testRunInputs(
	ctx context.Context,
	wait time.Duration,
	unit *inputUnit,
) error {
	var wg sync.WaitGroup

	nul := make(chan telegraf.Metric)
	go func() {
		for range nul {
		}
	}()

	for _, input := range unit.inputs {
		wg.Add(1)
		go func(input *models.RunningInput) {
			defer wg.Done()

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

			// Run plugins that require multiple gathers to calculate rate
			// and delta metrics twice.
			switch input.Config.Name {
			case "cpu", "mongodb", "procstat":
				nulAcc := NewAccumulator(input, nul)
				nulAcc.SetPrecision(getPrecision(precision, interval))
				if err := input.Input.Gather(nulAcc); err != nil {
					nulAcc.AddError(err)
				}
				time.Sleep(500 * time.Millisecond)
			}

			acc := NewAccumulator(input, unit.dst)
			acc.SetPrecision(getPrecision(precision, interval))

			if err := input.Input.Gather(acc); err != nil {
				acc.AddError(err)
			}
		}(input)
	}
	wg.Wait()

	internal.SleepContext(ctx, wait)

	log.Printf("D! [agent] Stopping service inputs")
	stopServiceInputs(unit.inputs)

	close(unit.dst)
	log.Printf("D! [agent] Input channel closed")
	return nil
}

// stopServiceInputs stops all service inputs.
func stopServiceInputs(inputs []*models.RunningInput) {
	for _, input := range inputs {
		if si, ok := input.Input.(telegraf.ServiceInput); ok {
			si.Stop()
		}
	}
}

// StopInputPlugin stops an input plugin
func (a *Agent) StopInputPlugin(uuid string, shouldUpdateConfig bool) error {
	a.pluginLock.Lock()
	plugin, ok := a.runningPlugins[uuid].(*models.RunningInput)
	a.pluginLock.Unlock()

	if !ok {
		log.Printf("E! [agent] Input %s is not runnning.\n", uuid)
		return fmt.Errorf("input %s is not runnning", uuid)
	}

	plugin.Stop()

	if shouldUpdateConfig {
		err := a.Config.UpdateConfig(map[string]interface{}{}, uuid, "inputs", "STOP_PLUGIN")
		if err != nil {
			log.Printf("W! [agent] Unable to update configuration for input plugin %s\n", uuid)
		}
	}

	a.Config.InputsLock.Lock()
	for i, other := range a.Config.Inputs {
		if other.UniqueId == uuid {
			a.Config.Inputs = append(a.Config.Inputs[:i], a.Config.Inputs[i+1:]...)
			a.Config.InputsLock.Unlock()
			return nil
		}
	}
	a.Config.InputsLock.Unlock()

	return fmt.Errorf("input was not found in running inputs list")
}

// StopOutputPlugin stops an output plugin
func (a *Agent) StopOutputPlugin(uuid string, shouldUpdateConfig bool) error {
	a.pluginLock.Lock()
	plugin, ok := a.runningPlugins[uuid].(*models.RunningOutput)
	a.pluginLock.Unlock()

	if !ok {
		log.Printf("E! [agent] Output %s is not runnning.\n", uuid)
		return fmt.Errorf("output %s is not runnning", uuid)
	}

	plugin.Stop()

	if shouldUpdateConfig {
		err := a.Config.UpdateConfig(map[string]interface{}{}, uuid, "outputs", "STOP_PLUGIN")
		if err != nil {
			log.Printf("W! [agent] Unable to update configuration for output plugin %s\n", uuid)
		}
	}

	a.Config.OutputsLock.Lock()
	for i, other := range a.Config.Outputs {
		if other.UniqueId == uuid {
			a.Config.Outputs = append(a.Config.Outputs[:i], a.Config.Outputs[i+1:]...)
			break
		}
	}
	a.Config.OutputsLock.Unlock()

	return nil
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
		case <-input.ShutdownChan:
			log.Println("I! [agent] stopping input plugin", input.Config.Name)
			a.pluginLock.Lock()
			delete(a.runningPlugins, input.UniqueId)
			a.pluginLock.Unlock()
			// delete input from input unit slice
			for i, io := range a.iu.inputs {
				if input == io {
					// swap with last input and truncate slice
					if len(a.iu.inputs) > 1 {
						a.iu.inputs[i] = a.iu.inputs[len(a.iu.inputs)-1]
					}
					a.iu.inputs = a.iu.inputs[:len(a.iu.inputs)-1]
					break
				}
			}
			input.Wg.Done()
			return
		case <-ticker.Elapsed():
			err := a.gatherOnce(acc, input, ticker, interval)
			if err != nil {
				acc.AddError(err)
			}
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
func (a *Agent) startProcessors(
	dst chan<- telegraf.Metric,
	processors models.RunningProcessors,
) (chan<- telegraf.Metric, []*processorUnit, error) {
	var units []*processorUnit

	// Sort from last to first
	sort.SliceStable(processors, func(i, j int) bool {
		return processors[i].Config.Order > processors[j].Config.Order
	})

	var src chan telegraf.Metric
	for _, processor := range processors {
		src = make(chan telegraf.Metric, 100)
		acc := NewAccumulator(processor, dst)

		err := processor.Start(acc)
		if err != nil {
			for _, u := range units {
				u.processor.Stop()
				close(u.dst)
			}
			return nil, nil, fmt.Errorf("starting processor %s: %w", processor.LogName(), err)
		}

		units = append(units, &processorUnit{
			src:       src,
			dst:       dst,
			processor: processor,
		})

		dst = src
	}

	return src, units, nil
}

// runProcessors begins processing metrics and runs until the source channel is
// closed and all metrics have been written.
func (a *Agent) runProcessors(
	units []*processorUnit,
) error {
	var wg sync.WaitGroup
	for _, unit := range units {
		wg.Add(1)
		go func(unit *processorUnit) {
			defer wg.Done()

			acc := NewAccumulator(unit.processor, unit.dst)
			for m := range unit.src {
				if err := unit.processor.Add(m, acc); err != nil {
					acc.AddError(err)
					m.Drop()
				}
			}
			unit.processor.Stop()
			close(unit.dst)
			log.Printf("D! [agent] Processor channel closed")
		}(unit)
	}
	wg.Wait()

	return nil
}

// startAggregators sets up the aggregator unit and returns the source channel.
func (a *Agent) startAggregators(
	aggC chan<- telegraf.Metric,
	outputC chan<- telegraf.Metric,
	aggregators []*models.RunningAggregator,
) (chan<- telegraf.Metric, *aggregatorUnit, error) {
	src := make(chan telegraf.Metric, 100)
	unit := &aggregatorUnit{
		src:         src,
		aggC:        aggC,
		outputC:     outputC,
		aggregators: aggregators,
	}
	return src, unit, nil
}

// runAggregators beings aggregating metrics and runs until the source channel
// is closed and all metrics have been written.
func (a *Agent) runAggregators(
	startTime time.Time,
	unit *aggregatorUnit,
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
		for metric := range unit.src {
			var dropOriginal bool
			for _, agg := range a.Config.Aggregators {
				if ok := agg.Add(metric); ok {
					dropOriginal = true
				}
			}

			if !dropOriginal {
				unit.outputC <- metric // keep original.
			} else {
				metric.Drop()
			}
		}
		cancel()
	}()

	for _, agg := range a.Config.Aggregators {
		wg.Add(1)
		go func(agg *models.RunningAggregator) {
			defer wg.Done()

			interval := a.Config.Agent.Interval.Duration
			precision := a.Config.Agent.Precision.Duration

			acc := NewAccumulator(agg, unit.aggC)
			acc.SetPrecision(getPrecision(precision, interval))
			a.push(ctx, agg, acc)
		}(agg)
	}

	wg.Wait()

	// In the case that there are no processors, both aggC and outputC are the
	// same channel.  If there are processors, we close the aggC and the
	// processor chain will close the outputC when it finishes processing.
	close(unit.aggC)
	log.Printf("D! [agent] Aggregator channel closed")

	return nil
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

// startOutputs calls Connect on all outputs and returns the source channel.
// If an error occurs calling Connect all stared plugins have Close called.
func (a *Agent) startOutputs(
	ctx context.Context,
	outputs []*models.RunningOutput,
) (chan<- telegraf.Metric, *outputUnit, error) {
	src := make(chan telegraf.Metric, 100)
	unit := &outputUnit{src: src}
	for _, output := range outputs {
		err := a.connectOutput(ctx, output)
		if err != nil {
			for _, output := range unit.outputs {
				output.Close()
			}
			return nil, nil, fmt.Errorf("connecting output %s: %w", output.LogName(), err)
		}
		unit.outputs = append(unit.outputs, output)
	}

	return src, unit, nil
}

// connectOutputs connects to all outputs.
func (a *Agent) connectOutput(ctx context.Context, output *models.RunningOutput) error {
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
			return fmt.Errorf("Error connecting to output %q: %w", output.LogName(), err)
		}
	}
	log.Printf("D! [agent] Successfully connected to %s", output.LogName())
	return nil
}

// runOutputs begins processing metrics and returns until the source channel is
// closed and all metrics have been written.  On shutdown metrics will be
// written one last time and dropped if unsuccessful.
func (a *Agent) runOutputs(
	unit *outputUnit,
) error {

	ctx, cancel := context.WithCancel(context.Background())

	for _, output := range unit.outputs {
		err := a.RunSingleOutput(output, ctx)
		if err != nil {
			return err
		}
	}

	for metric := range unit.src {
		for i, output := range unit.outputs {
			a.Config.OutputsLock.Lock()
			if i == len(a.Config.Outputs)-1 {
				output.AddMetric(metric)
			} else {
				output.AddMetric(metric.Copy())
			}
			a.Config.OutputsLock.Unlock()
		}
	}

	log.Println("I! [agent] Hang on, flushing any cached metrics before shutdown")
	cancel()

	for {
		select {
		case <-ctx.Done():
			return nil
		}
	}
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
		case <-output.ShutdownChan:
			log.Println("I! [agent] stopping output plugin", output.Config.Name)
			a.pluginLock.Lock()
			delete(a.runningPlugins, output.UniqueId)
			a.pluginLock.Unlock()
			// delete output from output unit slice
			for i, ro := range a.ou.outputs {
				if output == ro {
					// swap with last output and truncate slice
					if len(a.ou.outputs) > 1 {
						a.ou.outputs[i] = a.ou.outputs[len(a.ou.outputs)-1]
					}
					a.ou.outputs = a.ou.outputs[:len(a.ou.outputs)-1]
					break
				}
			}
			output.Close()
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

	startTime := time.Now()

	next := outputC

	var apu []*processorUnit
	var au *aggregatorUnit
	if len(a.Config.Aggregators) != 0 {
		procC := next
		if len(a.Config.AggProcessors) != 0 {
			procC, apu, err = a.startProcessors(next, a.Config.AggProcessors)
			if err != nil {
				return err
			}
		}

		next, au, err = a.startAggregators(procC, next, a.Config.Aggregators)
		if err != nil {
			return err
		}
	}

	var pu []*processorUnit
	if len(a.Config.Processors) != 0 {
		next, pu, err = a.startProcessors(next, a.Config.Processors)
		if err != nil {
			return err
		}
	}

	iu, err := a.testStartInputs(next, a.Config.Inputs)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	if au != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := a.runProcessors(apu)
			if err != nil {
				log.Printf("E! [agent] Error running processors: %v", err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := a.runAggregators(startTime, au)
			if err != nil {
				log.Printf("E! [agent] Error running aggregators: %v", err)
			}
		}()
	}

	if pu != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := a.runProcessors(pu)
			if err != nil {
				log.Printf("E! [agent] Error running processors: %v", err)
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := a.testRunInputs(ctx, wait, iu)
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
	err := a.initPlugins()
	if err != nil {
		return err
	}

	startTime := time.Now()

	log.Printf("D! [agent] Connecting outputs")
	next, ou, err := a.startOutputs(ctx, a.Config.Outputs)
	if err != nil {
		return err
	}

	var apu []*processorUnit
	var au *aggregatorUnit
	if len(a.Config.Aggregators) != 0 {
		procC := next
		if len(a.Config.AggProcessors) != 0 {
			procC, apu, err = a.startProcessors(next, a.Config.AggProcessors)
			if err != nil {
				return err
			}
		}

		next, au, err = a.startAggregators(procC, next, a.Config.Aggregators)
		if err != nil {
			return err
		}
	}

	var pu []*processorUnit
	if len(a.Config.Processors) != 0 {
		next, pu, err = a.startProcessors(next, a.Config.Processors)
		if err != nil {
			return err
		}
	}

	iu, err := a.testStartInputs(next, a.Config.Inputs)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := a.runOutputs(ou)
		if err != nil {
			log.Printf("E! [agent] Error running outputs: %v", err)
		}
	}()

	if au != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := a.runProcessors(apu)
			if err != nil {
				log.Printf("E! [agent] Error running processors: %v", err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := a.runAggregators(startTime, au)
			if err != nil {
				log.Printf("E! [agent] Error running aggregators: %v", err)
			}
		}()
	}

	if pu != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := a.runProcessors(pu)
			if err != nil {
				log.Printf("E! [agent] Error running processors: %v", err)
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := a.testRunInputs(ctx, wait, iu)
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

package scripting

import (
	"go/build"
	"log"

	"github.com/containous/yaegi/interp"
	"github.com/containous/yaegi/stdlib"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
	"github.com/influxdata/telegraf/plugins/processors/scripting/telegrafSymbols"
)

var sampleConfig = `
  ## General Aggregator Arguments:
  ## The period on which to flush & clear the aggregator.
  period = "30s"
  ## If true, the original metric will be dropped by the
  ## aggregator and will not get sent to the output plugins.
  drop_original = false

  ## Go code to aggregate metrics
  script = '''
package scripting
import (
	"github.com/influxdata/telegraf"
)

var data []telegraf.Metric

func Push(acc telegraf.Accumulator) {
	for _,m := range data {
		acc.AddMetric(m)
	}
}

func Add(in telegraf.Metric) {
	data = append(data, in)
}

func Reset() {
}
'''
`

// Scripting is the main structure of the aggregator, storing the script to be executed
type Scripting struct {
	Script           string
	interpreter      *interp.Interpreter
	interpretedAdd   func(telegraf.Metric)
	interpretedPush  func(telegraf.Accumulator)
	interpretedReset func()
}

// Initialize create the vw with the configuration script.
// Gets the Add and Push functions from the script.
// It is only executed the first time "Add" is called
func (s *Scripting) Initialize() {
	// Initialize interpreter
	s.interpreter = interp.New(interp.Options{GoPath: build.Default.GOPATH})
	s.interpreter.Use(stdlib.Symbols)
	s.interpreter.Use(interp.Symbols)
	s.interpreter.Use(telegrafSymbols.Symbols)

	// Parse the script
	_, err := s.interpreter.Eval(s.Script)
	if err != nil {
		log.Printf("E! [aggregators.scripting] parsing script: %v", err)
	}

	// Get the "Add" function from the interpreted script
	addIface, err := s.interpreter.Eval("scripting.Add")
	if err != nil {
		log.Printf("E! [aggregators.scripting] get Add function from script: %v", err)
	}
	s.interpretedAdd = addIface.Interface().(func(telegraf.Metric))

	// Get the "Push" function from the interpreted script
	pushIface, err := s.interpreter.Eval("scripting.Push")
	if err != nil {
		log.Printf("E! [aggregators.scripting] get Push function from script: %v", err)
	}
	s.interpretedPush = pushIface.Interface().(func(telegraf.Accumulator))

	// Get the "Reset" function from the interpreted script
	resetIface, err := s.interpreter.Eval("scripting.Reset")
	if err != nil {
		log.Printf("E! [aggregators.scripting] get Reset function from script: %v", err)
	}
	s.interpretedReset = resetIface.Interface().(func())
}

// SampleConfig return an example configuration
func (s *Scripting) SampleConfig() string {
	return sampleConfig
}

// Description one line explanation of the aggregator
func (s *Scripting) Description() string {
	return "Define a aggregator dinamically"
}

// Add pass metrics to the aggregator
func (s *Scripting) Add(in telegraf.Metric) {
	if s.interpreter == nil {
		s.Initialize()
	}

	s.interpretedAdd(in)
}

// Push ask the aggregator to generate metrics and put them into de accumulator
func (s *Scripting) Push(acc telegraf.Accumulator) {
	s.interpretedPush(acc)
}

// Reset put the aggregator in an initial state
func (s *Scripting) Reset() {
	s.interpretedReset()
}

func init() {
	aggregators.Add("scripting", func() telegraf.Aggregator {
		return &Scripting{}
	})
}

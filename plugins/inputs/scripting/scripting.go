package scripting

import (
	"go/build"
	"log"

	"github.com/containous/yaegi/interp"
	"github.com/containous/yaegi/stdlib"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/processors/scripting/telegrafSymbols"
)

var sampleConfig = `
  ## Go code to gather metrics
  script = '''
package scripting
import (
	"github.com/influxdata/telegraf"
)

func Gather(acc telegraf.Accumulator) error {
	acc.AddMetric(testutil.MustMetric(
		"name",
		map[string]string{"host": "hostA", "foo": "bar"},
		map[string]interface{}{"value": 1},
		time.Now(),
	))
	return nil
}
'''
`

// Scripting is the main structure of the input, storing the script to be executed
type Scripting struct {
	Script            string
	interpreter       *interp.Interpreter
	interpretedGather func(telegraf.Accumulator) error
}

// Initialize create the vw with the configuration script.
// Gets the Gather function from the script.
// It is only executed the first time "Gather" is called
func (s *Scripting) Initialize() {
	// Initialize interpreter
	s.interpreter = interp.New(interp.Options{GoPath: build.Default.GOPATH})
	s.interpreter.Use(stdlib.Symbols)
	s.interpreter.Use(interp.Symbols)
	s.interpreter.Use(telegrafSymbols.Symbols)

	// Parse the script
	_, err := s.interpreter.Eval(s.Script)
	if err != nil {
		log.Printf("E! [input.scripting] parsing script: %v", err)
	}

	// Get the "Push" function from the interpreted script
	gatherIface, err := s.interpreter.Eval("scripting.Gather")
	if err != nil {
		log.Printf("E! [inputs.scripting] get Gather function from script: %v", err)
	}
	s.interpretedGather = gatherIface.Interface().(func(telegraf.Accumulator) error)
}

// SampleConfig return an example configuration
func (s *Scripting) SampleConfig() string {
	return sampleConfig
}

// Description one line explanation of the input
func (s *Scripting) Description() string {
	return "Define a input dinamically"
}

// Gather collects metrics from the input
func (s *Scripting) Gather(acc telegraf.Accumulator) error {
	if s.interpreter == nil {
		s.Initialize()
	}
	return s.interpretedGather(acc)
}

func init() {
	inputs.Add("scripting", func() telegraf.Input {
		return &Scripting{}
	})
}

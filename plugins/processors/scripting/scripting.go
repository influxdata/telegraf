package scripting

import (
	"go/build"
	"log"

	"github.com/containous/yaegi/interp"
	"github.com/containous/yaegi/stdlib"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/processors/scripting/telegrafSymbols"
)

var sampleConfig = `
  ## Example immitating the printer processor
  script = '''
package scripting
import (
  "fmt"
  "time"
  "github.com/influxdata/telegraf"
)

func Apply(in []telegraf.Metric) ([]telegraf.Metric) {
  fmt.Printf("%+v\n", in)
  return in
}
'''
`

// Scripting is the main structure of the processor, storing the script to be executed
type Scripting struct {
	Script           string
	interpreter      *interp.Interpreter
	interpretedApply func([]telegraf.Metric) []telegraf.Metric
}

// SampleConfig return an example configuration
func (s *Scripting) SampleConfig() string {
	return sampleConfig
}

// Description one line explanation of the processor
func (s *Scripting) Description() string {
	return "Define a processor dinamically"
}

// Initialize create the vw with the configuration script.
// Gets the Apply function from the script.
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
		log.Printf("E! [processors.scripting] parsing script: %v", err)
	}

	// Get the "Add" function from the interpreted script
	applyIface, err := s.interpreter.Eval("scripting.Apply")
	if err != nil {
		log.Printf("E! [processors.scripting] get Apply function from script: %v", err)
	}
	s.interpretedApply = applyIface.Interface().(func([]telegraf.Metric) []telegraf.Metric)
}

// Apply pass metrics through the processor
func (s *Scripting) Apply(in ...telegraf.Metric) []telegraf.Metric {
	if s.interpreter == nil {
		s.Initialize()
	}

	return s.interpretedApply(in)
}

func init() {
	processors.Add("scripting", func() telegraf.Processor {
		return &Scripting{}
	})
}

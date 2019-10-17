package scripting

import (
	"go/build"
	"log"

	"github.com/containous/yaegi/interp"
	"github.com/containous/yaegi/stdlib"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/processors/scripting/telegrafSymbols"
)

var sampleConfig = `
  ## Go code to aggregate metrics
  script = '''
package scripting
import (
  "fmt"
  "time"
	"github.com/influxdata/telegraf"
)

func Connect() error {
}

func Close() error {
}

func Write(metrics []telegraf.Metric) error {
	for _,m := range metrics {
		fmt.Printf("%+v\n", m)
	}
}
'''
`

// Scripting is the main structure of the aggregator, storing the script to be executed
type Scripting struct {
	Script             string
	interpreter        *interp.Interpreter
	interpretedConnect func() error
	interpretedClose   func() error
	interpretedWrite   func([]telegraf.Metric) error
}

// Initialize create the vw with the configuration script.
// Gets the Write, Close and Connect functions from the script.
// It is only executed the first time "Connect" is called
func (s *Scripting) Initialize() {
	// Initialize interpreter
	s.interpreter = interp.New(interp.Options{GoPath: build.Default.GOPATH})
	s.interpreter.Use(stdlib.Symbols)
	s.interpreter.Use(interp.Symbols)
	s.interpreter.Use(telegrafSymbols.Symbols)

	// Parse the script
	_, err := s.interpreter.Eval(s.Script)
	if err != nil {
		log.Printf("E! [outputs.scripting] parsing script: %v", err)
	}

	// Get the "Connect" function from the interpreted script
	connectIface, err := s.interpreter.Eval("scripting.Connect")
	if err != nil {
		log.Printf("E! [outputs.scripting] get Connect function from script: %v", err)
	}
	s.interpretedConnect = connectIface.Interface().(func() error)

	// Get the "Write" function from the interpreted script
	writeIface, err := s.interpreter.Eval("scripting.Write")
	if err != nil {
		log.Printf("E! [outputs.scripting] get Write function from script: %v", err)
	}
	s.interpretedWrite = writeIface.Interface().(func([]telegraf.Metric) error)

	// Get the "Close" function from the interpreted script
	closeIface, err := s.interpreter.Eval("scripting.Close")
	if err != nil {
		log.Printf("E! [outputs.scripting] get Close function from script: %v", err)
	}
	s.interpretedClose = closeIface.Interface().(func() error)
}

// SampleConfig return an example configuration
func (s *Scripting) SampleConfig() string {
	return sampleConfig
}

// Description one line explanation of the output
func (s *Scripting) Description() string {
	return "Define a aggregator dinamically"
}

// Write send metrics to the output
func (s *Scripting) Write(metrics []telegraf.Metric) error {
	return s.interpretedWrite(metrics)
}

// Connect prepare the output
func (s *Scripting) Connect() error {
	if s.interpreter == nil {
		s.Initialize()
	}

	return s.interpretedConnect()
}

// Close terminate output
func (s *Scripting) Close() error {
	return s.interpretedClose()
}

func init() {
	outputs.Add("scripting", func() telegraf.Output {
		return &Scripting{}
	})
}

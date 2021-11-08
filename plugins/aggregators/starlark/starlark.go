package starlark //nolint - Needed to avoid getting import-shadowing: The name 'starlark' shadows an import name (revive)

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
	common "github.com/influxdata/telegraf/plugins/common/starlark"
	"go.starlark.net/starlark"
)

const (
	description  = "Aggregate metrics using a Starlark script"
	sampleConfig = `
  ## The Starlark source can be set as a string in this configuration file, or
  ## by referencing a file containing the script.  Only one source or script
  ## should be set at once.
  ##
  ## Source of the Starlark script.
  source = '''
state = {}

def add(metric):
  state["last"] = metric

def push():
  return state.get("last")

def reset():
  state.clear()
'''

  ## File containing a Starlark script.
  # script = "/usr/local/bin/myscript.star"

  ## The constants of the Starlark script.
  # [aggregators.starlark.constants]
  #   max_size = 10
  #   threshold = 0.75
  #   default_name = "Julia"
  #   debug_mode = true
`
)

type Starlark struct {
	common.StarlarkCommon

	addArgs   starlark.Tuple
	addFunc   *starlark.Function
	pushArgs  starlark.Tuple
	pushFunc  *starlark.Function
	resetArgs starlark.Tuple
	resetFunc *starlark.Function
}

func (s *Starlark) Init() error {
	// Execute source
	globals, err := s.StarlarkCommon.Init()
	if err != nil {
		return err
	}

	// The source should define an add function.
	s.addFunc, s.addArgs, err = common.InitFunction(globals, "add", &common.Metric{})
	if err != nil {
		return err
	}

	// The source should define a push function.
	s.pushFunc, s.pushArgs, err = common.InitFunction(globals, "push")
	if err != nil {
		return err
	}

	// The source should define a reset function.
	s.resetFunc, s.resetArgs, err = common.InitFunction(globals, "reset")
	if err != nil {
		return err
	}

	return nil
}

func (s *Starlark) SampleConfig() string {
	return sampleConfig
}

func (s *Starlark) Description() string {
	return description
}

func (s *Starlark) Add(metric telegraf.Metric) {
	s.addArgs[0].(*common.Metric).Wrap(metric)

	s.Call(s.addFunc, s.addArgs) //nolint - error already checked within the Call function
}

func (s *Starlark) Push(acc telegraf.Accumulator) {
	rv, err := s.Call(s.pushFunc, s.pushArgs)
	if err != nil {
		return
	}

	switch rv := rv.(type) {
	case *starlark.List:
		iter := rv.Iterate()
		defer iter.Done()
		var v starlark.Value
		for iter.Next(&v) {
			switch v := v.(type) {
			case *common.Metric:
				m := v.Unwrap()
				acc.AddMetric(m)
			default:
				s.Log.Errorf("Invalid type returned in list: %s", v.Type())
			}
		}
	case *common.Metric:
		m := rv.Unwrap()
		acc.AddMetric(m)
	case starlark.NoneType:
	default:
		s.Log.Errorf("Invalid type returned: %T", rv)
	}
}

func (s *Starlark) Reset() {
	s.Call(s.resetFunc, s.resetArgs) //nolint - error already checked within the Call function
}

// init initializes starlark aggregator plugin
func init() {
	aggregators.Add("starlark", func() telegraf.Aggregator {
		return &Starlark{
			StarlarkCommon: common.StarlarkCommon{
				StarlarkLoadFunc: common.LoadFunc,
			},
		}
	})
}

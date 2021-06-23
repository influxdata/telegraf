package starlark //nolint

import (
	"strings"

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
def add(cache, metric):
  cache["last"] = metric

def push(cache, accumulator):
  last = cache.get("last")
  if last != None:
	  accumulator.add_fields(last.name, last.fields, last.tags)
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

	addArgs  starlark.Tuple
	addFunc  *starlark.Function
	cache    *starlark.Dict
	pushArgs starlark.Tuple
	pushFunc *starlark.Function
}

func (s *Starlark) Init() error {
	// Execute source
	globals, err := s.InitGlobals("aggregators.starlark")
	if err != nil {
		return err
	}

	// Initialize the cache
	s.Reset()

	// Freeze the global state.  This prevents modifications to the aggregator
	// state and prevents scripts from containing errors storing tracking
	// metrics.  Tasks that require global state will not be possible due to
	// this, so maybe we should relax this in the future.
	globals.Freeze()

	// The source should define an add function.
	s.addFunc, err = common.InitFunction(globals, "add", 2)
	if err != nil {
		return err
	}

	// Prepare the arguments of the add method.
	s.addArgs = make(starlark.Tuple, 2)
	s.addArgs[0] = &starlark.Dict{}
	s.addArgs[1] = &common.Metric{}

	// The source should define a push function.
	s.pushFunc, err = common.InitFunction(globals, "push", 2)
	if err != nil {
		return err
	}

	// Prepare the argument of the push method.
	s.pushArgs = make(starlark.Tuple, 2)
	s.pushArgs[0] = &starlark.Dict{}
	s.pushArgs[1] = &common.Accumulator{}

	return nil
}

func (s *Starlark) SampleConfig() string {
	return sampleConfig
}

func (s *Starlark) Description() string {
	return description
}

func (s *Starlark) Add(metric telegraf.Metric) {
	s.addArgs[0] = s.cache
	s.addArgs[1].(*common.Metric).Wrap(metric)

	s.call(s.addFunc, s.addArgs)
}

func (s *Starlark) Push(acc telegraf.Accumulator) {
	s.pushArgs[0] = s.cache
	s.pushArgs[1].(*common.Accumulator).Wrap(acc)

	s.call(s.pushFunc, s.pushArgs)
}

func (s *Starlark) Reset() {
	s.cache = starlark.NewDict(0)
}

func (s *Starlark) call(fn starlark.Value, args starlark.Tuple) {
	_, err := s.Call(fn, args)
	if err != nil {
		if err, ok := err.(*starlark.EvalError); ok {
			for _, line := range strings.Split(err.Backtrace(), "\n") {
				s.Log.Error(line)
			}
		}
	}
}

// init initializes starlark aggregator plugin
func init() {
	aggregators.Add("starlark", func() telegraf.Aggregator {
		return &Starlark{
			StarlarkCommon: common.NewStarlarkCommon(common.LoadFunc),
		}
	})
}

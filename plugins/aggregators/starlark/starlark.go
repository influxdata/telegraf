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

def apply(cache):
  return cache.get("last")
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
	cache     *starlark.Dict
	applyArgs starlark.Tuple
	applyFunc *starlark.Function
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

	// The source should define a apply function.
	s.applyFunc, err = common.InitFunction(globals, "apply", 1)
	if err != nil {
		return err
	}

	// Prepare the argument of the apply method.
	s.applyArgs = make(starlark.Tuple, 1)
	s.applyArgs[0] = &starlark.Dict{}

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
	s.applyArgs[0] = s.cache

	rv, err := s.Call(s.applyFunc, s.applyArgs)
	if err != nil {
		if err, ok := err.(*starlark.EvalError); ok {
			for _, line := range strings.Split(err.Backtrace(), "\n") {
				s.Log.Error(line)
			}
		}
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

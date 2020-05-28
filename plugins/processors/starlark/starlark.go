package starlark

import (
	"errors"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/processors"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
)

const (
	description  = "Process metrics using a Starlark script"
	sampleConfig = `
  ## Starlark source, only set one of source or script.
  source = """
    def apply(metric):
        pass
  """

  ## File containing Starlark script, only set one of source or script.
  # script = ""

  ## Can be set to pass or drop, if 
  # on_error = "drop"
`

	defaultOnError = "drop"
)

type Starlark struct {
	Source  string `toml:"source"`
	Script  string `toml:"script"`
	OnError string `toml:"on_error"`

	Log telegraf.Logger `toml:"-"`

	thread    *starlark.Thread
	applyFunc *starlark.Function
	args      starlark.Tuple
	results   []telegraf.Metric
}

func (s *Starlark) Init() error {
	err := choice.Check(s.OnError, []string{"pass", "drop"})
	if err != nil {
		return err
	}

	if s.Source == "" && s.Script == "" {
		return errors.New("one of source or script must be set")
	}
	if s.Source != "" && s.Script != "" {
		return errors.New("both source or script cannot be set")
	}

	s.thread = &starlark.Thread{
		Print: func(_ *starlark.Thread, msg string) { s.Log.Debug(msg) },
	}

	predeclared := starlark.StringDict{}
	predeclared["Metric"] = starlark.NewBuiltin("Metric", newMetric)
	predeclared["deepcopy"] = starlark.NewBuiltin("deepcopy", deepcopy)

	var program *starlark.Program
	if s.Source != "" {
		_, program, err = starlark.SourceProgram("processor.starlark", s.Source, predeclared.Has)
		if err != nil {
			return err
		}
	} else if s.Script != "" {
		_, program, err = starlark.SourceProgram(s.Script, nil, predeclared.Has)
		if err != nil {
			return err
		}
	}

	// Execute source
	globals, err := program.Init(s.thread, predeclared)
	if err != nil {
		return err
	}

	// Freeze the global state.  This prevents modifications to the processor
	// state and prevents scripts from containing errors storing tracking
	// metrics.  Tasks that require global state will not be possible due to
	// this, so maybe we should relax this in the future.
	globals.Freeze()

	// The source should define an apply function.
	apply := globals["apply"]

	if apply == nil {
		return errors.New("apply is not defined")
	}

	var ok bool
	if s.applyFunc, ok = apply.(*starlark.Function); !ok {
		return errors.New("apply is not a function")
	}

	if s.applyFunc.NumParams() != 1 {
		return errors.New("apply function must take one parameter")
	}

	// Reusing the same metric wrapper to skip an allocation.  This will cause
	// any saved references point to the new metric, but due to freezing the
	// globals none should exist.
	s.args = make(starlark.Tuple, 1)
	s.args[0] = &Metric{}

	// Preallocate a slice for return values.
	s.results = make([]telegraf.Metric, 0, 10)

	return nil
}

func (s *Starlark) SampleConfig() string {
	return sampleConfig
}

func (s *Starlark) Description() string {
	return description
}

func (s *Starlark) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	s.results = s.results[:0]
	for _, m := range metrics {
		s.args[0].(*Metric).Wrap(m)

		rv, err := starlark.Call(s.thread, s.applyFunc, s.args, nil)
		if err != nil {
			if err, ok := err.(*starlark.EvalError); ok {
				for _, line := range strings.Split(err.Backtrace(), "\n") {
					s.Log.Error(line)
				}
			} else {
				s.Log.Error(err)
			}
			continue
		}

		switch rv := rv.(type) {
		case *starlark.List:
			iter := rv.Iterate()
			defer iter.Done()
			var v starlark.Value
			for iter.Next(&v) {
				switch v := v.(type) {
				case *Metric:
					m := v.Unwrap()
					if containsMetric(s.results, m) {
						s.Log.Errorf("Duplicate metric reference detected")
						continue
					}
					s.results = append(s.results, m)
				default:
					s.Log.Errorf("Invalid type returned in list: %T", v)
				}
			}
		case *Metric:
			s.results = append(s.results, rv.Unwrap())
		case starlark.NoneType:
			return nil
		default:
			s.Log.Errorf("Invalid type returned: %T", rv)
		}
	}
	return s.results
}

func containsMetric(metrics []telegraf.Metric, metric telegraf.Metric) bool {
	for _, m := range metrics {
		if m == metric {
			return true
		}
	}
	return false
}

func init() {
	// https://github.com/bazelbuild/starlark/issues/20
	resolve.AllowNestedDef = true
	resolve.AllowLambda = true
	resolve.AllowFloat = true
	resolve.AllowSet = true
	resolve.AllowGlobalReassign = true
	resolve.AllowRecursion = true
}

func init() {
	processors.Add("starlark", func() telegraf.Processor {
		return &Starlark{
			OnError: defaultOnError,
		}
	})
}

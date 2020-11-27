package starlark

import (
	"errors"
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
)

const (
	description  = "Process metrics using a Starlark script"
	sampleConfig = `
  ## The Starlark source can be set as a string in this configuration file, or
  ## by referencing a file containing the script.  Only one source or script
  ## should be set at once.
  ##
  ## Source of the Starlark script.
  source = '''
def apply(metric):
	return metric
'''

  ## File containing a Starlark script.
  # script = "/usr/local/bin/myscript.star"
`
)

type Starlark struct {
	Source string `toml:"source"`
	Script string `toml:"script"`

	Log telegraf.Logger `toml:"-"`

	thread    *starlark.Thread
	applyFunc *starlark.Function
	args      starlark.Tuple
	results   []telegraf.Metric
	state     State
}

// The struct that defines the shared state used
type State struct {
	content map[string]interface{}
}

// The struct that defines a Tuple
type tuple struct {
	values []interface{}
}

// Load from the shared state, the value corresponding to the given key
func (s *State) Load(key string) starlark.Value {
	var value = s.content[key]
	if value == nil {
		return starlark.None
	}
	return toValue(value)
}

// Convert the given value extrated from the shared state into a starlark value
func toValue(original interface{}) starlark.Value {
	switch v := original.(type) {
	case telegraf.Metric:
		return WrapMetric(v)
	case []interface{}:
		length := len(v)
		array := make([]starlark.Value, length)
		for i := 0; i < length; i++ {
			array[i] = toValue(v[i])
		}
		return starlark.NewList(array)
	case map[interface{}]interface{}:
		dict := starlark.NewDict(len(v))
		for key, val := range v {
			dict.SetKey(toValue(key), toValue(val))
		}
		return dict
	case map[interface{}]bool:
		set := starlark.NewSet(len(v))
		for key := range v {
			set.Insert(toValue(key))
		}
		return set
	case tuple:
		length := len(v.values)
		array := make([]starlark.Value, length)
		for i := 0; i < length; i++ {
			array[i] = toValue(v.values[i])
		}
		return starlark.Tuple(array)
	default:
		return v.(starlark.Value)
	}
}

// Store the pair (key, value) into the shared state. If the value is None, the pair (key, value) will be
// removed from the shared state if it exists
func (s *State) Store(key string, value starlark.Value) {
	if value == starlark.None {
		delete(s.content, key)
	} else {
		s.content[key] = fromValue(value)
	}
}

// Convert the given starlark value into a type that can be stored into the shared cache
func fromValue(original starlark.Value) interface{} {
	switch v := original.(type) {
	case *Metric:
		return v.Unwrap()
	case *starlark.List:
		length := v.Len()
		array := make([]interface{}, length)
		for i := 0; i < length; i++ {
			array[i] = fromValue(v.Index(i))
		}
		return array
	case *starlark.Dict:
		dict := make(map[interface{}]interface{})
		for _, xitem := range v.Items() {
			key, val := xitem[0], xitem[1]
			dict[fromValue(key)] = fromValue(val)
		}
		return dict
	case *starlark.Set:
		set := make(map[interface{}]bool)
		iter := v.Iterate()
		defer iter.Done()
		var key starlark.Value
		for iter.Next(&key) {
			set[fromValue(key)] = true
		}
		return set
	case starlark.Tuple:
		length := v.Len()
		array := make([]interface{}, length)
		for i := 0; i < length; i++ {
			array[i] = fromValue(v.Index(i))
		}
		return tuple{
			values: array,
		}
	default:
		return v
	}
}

func (s *Starlark) Init() error {
	if s.Source == "" && s.Script == "" {
		return errors.New("one of source or script must be set")
	}
	if s.Source != "" && s.Script != "" {
		return errors.New("both source or script cannot be set")
	}

	s.thread = &starlark.Thread{
		Print: func(_ *starlark.Thread, msg string) { s.Log.Debug(msg) },
		Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
			return loadFunc(thread, module, s.Log)
		},
	}

	builtins := starlark.StringDict{}
	builtins["Metric"] = starlark.NewBuiltin("Metric", newMetric)
	builtins["deepcopy"] = starlark.NewBuiltin("deepcopy", deepcopy)
	builtins["catch"] = starlark.NewBuiltin("catch", catch)
	builtins["Load"] = starlark.NewBuiltin("Load",
		func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			return load(&s.state, thread, b, args, kwargs)
		},
	)
	builtins["Store"] = starlark.NewBuiltin("Store",
		func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			return store(&s.state, thread, b, args, kwargs)
		},
	)

	program, err := s.sourceProgram(builtins)
	if err != nil {
		return err
	}

	// Execute source
	globals, err := program.Init(s.thread, builtins)
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
	// any saved references to point to the new metric, but due to freezing the
	// globals none should exist.
	s.args = make(starlark.Tuple, 1)
	s.args[0] = &Metric{}

	// Preallocate a slice for return values.
	s.results = make([]telegraf.Metric, 0, 10)

	// Initialize the shared state
	s.state = State{
		content: make(map[string]interface{}),
	}

	return nil
}

func (s *Starlark) sourceProgram(builtins starlark.StringDict) (*starlark.Program, error) {
	if s.Source != "" {
		_, program, err := starlark.SourceProgram("processor.starlark", s.Source, builtins.Has)
		return program, err
	}
	_, program, err := starlark.SourceProgram(s.Script, nil, builtins.Has)
	return program, err
}

func (s *Starlark) SampleConfig() string {
	return sampleConfig
}

func (s *Starlark) Description() string {
	return description
}

func (s *Starlark) Start(acc telegraf.Accumulator) error {
	return nil
}

func (s *Starlark) Add(metric telegraf.Metric, acc telegraf.Accumulator) error {
	s.args[0].(*Metric).Wrap(metric)

	rv, err := starlark.Call(s.thread, s.applyFunc, s.args, nil)
	if err != nil {
		if err, ok := err.(*starlark.EvalError); ok {
			for _, line := range strings.Split(err.Backtrace(), "\n") {
				s.Log.Error(line)
			}
		}
		metric.Reject()
		return err
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
				acc.AddMetric(m)
			default:
				s.Log.Errorf("Invalid type returned in list: %s", v.Type())
			}
		}

		// If the script didn't return the original metrics, mark it as
		// successfully handled.
		if !containsMetric(s.results, metric) {
			metric.Accept()
		}

		// clear results
		for i := range s.results {
			s.results[i] = nil
		}
		s.results = s.results[:0]
	case *Metric:
		m := rv.Unwrap()

		// If the script returned a different metric, mark this metric as
		// successfully handled.
		if m != metric {
			metric.Accept()
		}
		acc.AddMetric(m)
	case starlark.NoneType:
		metric.Drop()
	default:
		return fmt.Errorf("Invalid type returned: %T", rv)
	}
	return nil
}

func (s *Starlark) Stop() error {
	return nil
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
	processors.AddStreaming("starlark", func() telegraf.StreamingProcessor {
		return &Starlark{}
	})
}

func loadFunc(thread *starlark.Thread, module string, logger telegraf.Logger) (starlark.StringDict, error) {
	switch module {
	case "json.star":
		return starlark.StringDict{
			"json": starlarkjson.Module,
		}, nil
	case "logging.star":
		return starlark.StringDict{
			"log": LogModule(logger),
		}, nil
	default:
		return nil, errors.New("module " + module + " is not available")
	}
}

package starlark //nolint - Needed to avoid getting import-shadowing: The name 'starlark' shadows an import name (revive)

import (
	"errors"
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
	"go.starlark.net/lib/math"
	"go.starlark.net/lib/time"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
)

type StarlarkCommon struct {
	Source    string                 `toml:"source"`
	Script    string                 `toml:"script"`
	Constants map[string]interface{} `toml:"constants"`

	Log              telegraf.Logger `toml:"-"`
	StarlarkLoadFunc func(module string, logger telegraf.Logger) (starlark.StringDict, error)

	thread *starlark.Thread
}

func (s *StarlarkCommon) Init() (starlark.StringDict, error) {
	if s.Source == "" && s.Script == "" {
		return nil, errors.New("one of source or script must be set")
	}
	if s.Source != "" && s.Script != "" {
		return nil, errors.New("both source or script cannot be set")
	}

	s.thread = &starlark.Thread{
		Print: func(_ *starlark.Thread, msg string) { s.Log.Debug(msg) },
		Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
			return s.StarlarkLoadFunc(module, s.Log)
		},
	}

	builtins := starlark.StringDict{}
	builtins["Metric"] = starlark.NewBuiltin("Metric", newMetric)
	builtins["deepcopy"] = starlark.NewBuiltin("deepcopy", deepcopy)
	builtins["catch"] = starlark.NewBuiltin("catch", catch)
	s.addConstants(&builtins)

	program, err := s.sourceProgram(builtins, "")

	if err != nil {
		return nil, err
	}

	// Execute source
	globals, err := program.Init(s.thread, builtins)
	if err != nil {
		return nil, err
	}
	// Make available a shared state to the apply function
	globals["state"] = starlark.NewDict(0)

	// Freeze the global state.  This prevents modifications to the processor
	// state and prevents scripts from containing errors storing tracking
	// metrics.  Tasks that require global state will not be possible due to
	// this, so maybe we should relax this in the future.
	globals.Freeze()

	return globals, nil
}

func InitFunction(globals starlark.StringDict, fnName string, expectedParams int) (*starlark.Function, error) {
	globalFn := globals[fnName]

	if globalFn == nil {
		return nil, fmt.Errorf("%s is not defined", fnName)
	}

	var ok bool
	var fn *starlark.Function
	if fn, ok = globalFn.(*starlark.Function); !ok {
		return nil, fmt.Errorf("%s is not a function", fnName)
	}

	if fn.NumParams() != expectedParams {
		return nil, fmt.Errorf("%s function must take %d parameter(s)", fnName, expectedParams)
	}
	return fn, nil
}

// Add all the constants defined in the plugin as constants of the script
func (s *StarlarkCommon) addConstants(builtins *starlark.StringDict) {
	for key, val := range s.Constants {
		sVal, err := asStarlarkValue(val)
		if err != nil {
			s.Log.Errorf("Unsupported type: %T", val)
		}
		(*builtins)[key] = sVal
	}
}

func (s *StarlarkCommon) sourceProgram(builtins starlark.StringDict, filename string) (*starlark.Program, error) {
	if s.Source != "" {
		_, program, err := starlark.SourceProgram(filename, s.Source, builtins.Has)
		return program, err
	}
	_, program, err := starlark.SourceProgram(s.Script, nil, builtins.Has)
	return program, err
}

// Call calls the function fn with the specified positional and keyword arguments.
func (s *StarlarkCommon) Call(fn starlark.Value, args starlark.Tuple) (starlark.Value, error) {
	rv, err := starlark.Call(s.thread, fn, args, nil)
	if err != nil {
		if err, ok := err.(*starlark.EvalError); ok {
			for _, line := range strings.Split(err.Backtrace(), "\n") {
				s.Log.Error(line)
			}
		}
	}
	return rv, err
}

func LoadFunc(module string, logger telegraf.Logger) (starlark.StringDict, error) {
	switch module {
	case "json.star":
		return starlark.StringDict{
			"json": starlarkjson.Module,
		}, nil
	case "logging.star":
		return starlark.StringDict{
			"log": LogModule(logger),
		}, nil
	case "math.star":
		return starlark.StringDict{
			"math": math.Module,
		}, nil
	case "time.star":
		return starlark.StringDict{
			"time": time.Module,
		}, nil
	default:
		return nil, errors.New("module " + module + " is not available")
	}
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

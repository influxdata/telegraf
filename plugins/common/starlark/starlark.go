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

	thread     *starlark.Thread
	globals    starlark.StringDict
	functions  map[string]*starlark.Function
	parameters map[string]starlark.Tuple
}

func (s *StarlarkCommon) Init() error {
	if s.Source == "" && s.Script == "" {
		return errors.New("one of source or script must be set")
	}
	if s.Source != "" && s.Script != "" {
		return errors.New("both source or script cannot be set")
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
	err := s.addConstants(&builtins)
	if err != nil {
		return err
	}

	program, err := s.sourceProgram(builtins, "")
	if err != nil {
		return err
	}

	// Execute source
	globals, err := program.Init(s.thread, builtins)
	if err != nil {
		return err
	}
	// Make available a shared state to the apply function
	globals["state"] = starlark.NewDict(0)

	// Freeze the global state.  This prevents modifications to the processor
	// state and prevents scripts from containing errors storing tracking
	// metrics.  Tasks that require global state will not be possible due to
	// this, so maybe we should relax this in the future.
	globals.Freeze()

	s.globals = globals
	s.functions = make(map[string]*starlark.Function)
	s.parameters = make(map[string]starlark.Tuple)
	return nil
}

func (s *StarlarkCommon) GetParameters(name string) (starlark.Tuple, bool) {
	parameters, found := s.parameters[name]
	return parameters, found
}

func (s *StarlarkCommon) AddFunction(name string, params ...starlark.Value) error {
	globalFn, found := s.globals[name]
	if !found {
		return fmt.Errorf("%s is not defined", name)
	}

	fn, found := globalFn.(*starlark.Function)
	if !found {
		return fmt.Errorf("%s is not a function", name)
	}

	if fn.NumParams() != len(params) {
		return fmt.Errorf("%s function must take %d parameter(s)", name, len(params))
	}
	p := make(starlark.Tuple, len(params))
	for i, param := range params {
		p[i] = param
	}
	s.functions[name] = fn
	s.parameters[name] = params
	return nil
}

// Add all the constants defined in the plugin as constants of the script
func (s *StarlarkCommon) addConstants(builtins *starlark.StringDict) error {
	for key, val := range s.Constants {
		sVal, err := asStarlarkValue(val)
		if err != nil {
			return fmt.Errorf("converting type %T failed: %v", val, err)
		}
		(*builtins)[key] = sVal
	}
	return nil
}

func (s *StarlarkCommon) sourceProgram(builtins starlark.StringDict, filename string) (*starlark.Program, error) {
	var src interface{}
	if s.Source != "" {
		src = s.Source
	}
	_, program, err := starlark.SourceProgram(s.Script, src, builtins.Has)
	return program, err
}

// Call calls the function corresponding to the given name.
func (s *StarlarkCommon) Call(name string) (starlark.Value, error) {
	fn, ok := s.functions[name]
	if !ok {
		return nil, fmt.Errorf("function %q does not exist", name)
	}
	args, ok := s.parameters[name]
	if !ok {
		return nil, fmt.Errorf("params for function %q do not exist", name)
	}
	return starlark.Call(s.thread, fn, args, nil)
}

func (s *StarlarkCommon) LogError(err error) {
	if err, ok := err.(*starlark.EvalError); ok {
		for _, line := range strings.Split(err.Backtrace(), "\n") {
			s.Log.Error(line)
		}
	} else {
		s.Log.Error(err.Msg)
	}
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

package heartbeat

import (
	"fmt"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/ext"
)

type StatusConfig struct {
	Ok      string   `toml:"ok"`
	Warn    string   `toml:"warn"`
	Fail    string   `toml:"fail"`
	Order   []string `toml:"order"`
	Default string   `toml:"default"`
	Initial string   `toml:"initial"`
}

func environment() (*cel.Env, error) {
	pluginStatisticsType := types.NewMapType(
		types.StringType,
		types.NewListType(types.NewMapType(types.StringType, types.DynType)),
	)

	// Declare the computation environment for the programs
	return cel.NewEnv(
		cel.VariableDecls(
			decls.NewVariable("metrics", types.IntType),
			decls.NewVariable("log_errors", types.IntType),
			decls.NewVariable("log_warnings", types.IntType),
			decls.NewVariable("last_update", types.TimestampType),
			decls.NewVariable("inputs", pluginStatisticsType),
			decls.NewVariable("outputs", types.NewMapType(types.StringType, types.DynType)),
		),
		cel.Function(
			"now",
			cel.Overload("now", nil, cel.TimestampType),
			cel.SingletonFunctionBinding(func(_ ...ref.Val) ref.Val { return types.Timestamp{Time: time.Now()} }),
		),
		ext.Encoders(),
		ext.Math(),
		ext.Strings(),
	)
}

type program struct {
	status string
	prog   cel.Program
}

func (p *program) eval(vars map[string]interface{}) (bool, error) {
	result, _, err := p.prog.Eval(vars)
	if err != nil {
		return false, err
	}
	if r, ok := result.Value().(bool); ok {
		return r, nil
	}
	return false, fmt.Errorf("invalid result type %T", result.Value())
}

package starlark

import (
	"errors"
	"fmt"

	"github.com/influxdata/telegraf"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Builds a module that defines all the supported logging functions which will log using the provided logger
func LogModule(logger telegraf.Logger) *starlarkstruct.Module {
	var logFunc = func(t *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return log(b, args, kwargs, logger)
	}
	return &starlarkstruct.Module{
		Name: "log",
		Members: starlark.StringDict{
			"debug": starlark.NewBuiltin("log.debug", logFunc),
			"info":  starlark.NewBuiltin("log.info", logFunc),
			"warn":  starlark.NewBuiltin("log.warn", logFunc),
			"error": starlark.NewBuiltin("log.error", logFunc),
		},
	}
}

// Logs the provided message according to the level chosen
func log(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple, logger telegraf.Logger) (starlark.Value, error) {
	var msg starlark.String
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &msg); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}
	switch b.Name() {
	case "log.debug":
		logger.Debug(string(msg))
	case "log.info":
		logger.Info(string(msg))
	case "log.warn":
		logger.Warn(string(msg))
	case "log.error":
		logger.Error(string(msg))
	default:
		return nil, errors.New("method " + b.Name() + " is unknown")
	}
	return starlark.None, nil
}

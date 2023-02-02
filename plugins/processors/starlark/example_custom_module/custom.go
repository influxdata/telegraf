package main

import (
	"github.com/influxdata/telegraf"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func InitModule(_ telegraf.Logger) *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "custom",
		Members: starlark.StringDict{
			"test": starlark.NewBuiltin("test", test),
		},
	}
}

func test(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	var message starlark.String = "Hallo from custom module"
	return message, nil
}

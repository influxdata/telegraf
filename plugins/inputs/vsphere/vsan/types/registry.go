package types

import (
	"reflect"
)

var t = map[string]reflect.Type{}

func Add(name string, kind reflect.Type) {
	t[name] = kind
}

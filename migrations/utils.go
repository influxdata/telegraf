package migrations

import (
	"fmt"
	"reflect"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

type pluginTOMLStruct map[string]map[string][]interface{}

func CreateTOMLStruct(category, name string) pluginTOMLStruct {
	return map[string]map[string][]interface{}{
		category: {
			name: make([]interface{}, 0),
		},
	}
}

func (p *pluginTOMLStruct) Add(category, name string, plugin interface{}) {
	cfg := map[string]map[string][]interface{}(*p)
	cfg[category][name] = append(cfg[category][name], plugin)
}

func AsStringSlice(raw interface{}) ([]string, error) {
	rawList, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type : %T", raw)
	}

	converted := make([]string, 0, len(rawList))
	for _, rawElement := range rawList {
		el, ok := rawElement.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected type for list element: %T", rawElement)
		}
		converted = append(converted, el)
	}
	return converted, nil
}

// UnmarshalTableSkipMissing unmarshals a TOML table into a struct, skipping any missing fields.
// This is useful for migration purposes where we want to ignore valid fields in the existing
// configuration that have no effect on migrations.
func UnmarshalTableSkipMissing(tbl *ast.Table, v interface{}) error {
	config := toml.DefaultConfig
	config.MissingField = func(_ reflect.Type, _ string) error {
		return nil
	}

	return config.UnmarshalTable(tbl, v)
}

package migrations

import (
	"fmt"
)

type pluginTOMLStruct map[string]map[string][]any

func CreateTOMLStruct(category, name string) pluginTOMLStruct {
	return map[string]map[string][]any{
		category: {
			name: make([]any, 0),
		},
	}
}

func (p *pluginTOMLStruct) Add(category, name string, plugin any) {
	cfg := map[string]map[string][]any(*p)
	cfg[category][name] = append(cfg[category][name], plugin)
}

func AsStringSlice(raw any) ([]string, error) {
	rawList, ok := raw.([]any)
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

package migrations

import (
	"fmt"
)

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

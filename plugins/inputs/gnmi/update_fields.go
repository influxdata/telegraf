package gnmi

import (
	"encoding/json"
	"fmt"
	"strconv"

	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
	gnmiValue "github.com/openconfig/gnmi/value"
)

type updateField struct {
	path  *gnmiLib.Path
	value interface{}
}

func newFieldsFromUpdate(prefix *gnmiLib.Path, update *gnmiLib.Update) ([]updateField, error) {
	// Prepend the update path with the prefix coming from the header of
	// the GNMI update message
	path := joinPaths(prefix, update.Path)
	if update.Val == nil || update.Val.Value == nil {
		return []updateField{{path, nil}}, nil
	}

	// Apply some special handling for special types
	switch v := update.Val.Value.(type) {
	case *gnmiLib.TypedValue_AsciiVal: // not handled in ToScalar
		return []updateField{{path, v.AsciiVal}}, nil
	case *gnmiLib.TypedValue_JsonVal: // requires special path handling
		return processJson(path, v.JsonVal)
	case *gnmiLib.TypedValue_JsonIetfVal: // requires special path handling
		return processJson(path, v.JsonIetfVal)
	}

	// Convert the protobuf "oneof" data to a Golang type.
	value, err := gnmiValue.ToScalar(update.Val)
	if err != nil {
		return nil, err
	}
	return []updateField{{path, value}}, nil
}

func processJson(prefix *gnmiLib.Path, data []byte) ([]updateField, error) {
	var nested interface{}
	if err := json.Unmarshal(data, &nested); err != nil {
		return nil, fmt.Errorf("failed to parse JSON value: %w", err)
	}

	// Flatten the JSON data to get a key-value map
	entries := flatten(nested)

	// Create an update-field with the complete path for all entries
	fields := make([]updateField, 0, len(entries))
	for key, v := range entries {
		fields = append(fields, updateField{
			path:  joinPaths(prefix, jsonKeyToPath(key)),
			value: v,
		})
	}

	return fields, nil
}

func flatten(nested interface{}) map[string]interface{} {
	fields := make(map[string]interface{})

	switch n := nested.(type) {
	case map[string]interface{}:
		for k, child := range n {
			for ck, cv := range flatten(child) {
				key := k
				if ck != "" {
					key += "/" + ck
				}
				fields[key] = cv
			}
		}
	case []interface{}:
		for i, child := range n {
			k := strconv.Itoa(i)
			for ck, cv := range flatten(child) {
				key := k
				if ck != "" {
					key += "/" + ck
				}
				fields[key] = cv
			}
		}
	case nil:
		return nil
	default:
		return map[string]interface{}{"": nested}
	}
	return fields
}

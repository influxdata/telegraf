package gnmi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"

	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

// Parse path to path-buffer and tag-field
func handlePath(gnmiPath *gnmiLib.Path, tags map[string]string, aliases map[string]string, prefix string) (pathBuffer string, aliasPath string, err error) {
	builder := bytes.NewBufferString(prefix)

	// Some devices do report the origin in the first path element
	// so try to find out if this is the case.
	if gnmiPath.Origin == "" && len(gnmiPath.Elem) > 0 {
		groups := originPattern.FindStringSubmatch(gnmiPath.Elem[0].Name)
		if len(groups) == 2 {
			gnmiPath.Origin = groups[1]
			gnmiPath.Elem[0].Name = gnmiPath.Elem[0].Name[len(groups[1])+1:]
		}
	}

	// Prefix with origin
	if len(gnmiPath.Origin) > 0 {
		if _, err := builder.WriteString(gnmiPath.Origin); err != nil {
			return "", "", err
		}
		if _, err := builder.WriteRune(':'); err != nil {
			return "", "", err
		}
	}

	// Parse generic keys from prefix
	for _, elem := range gnmiPath.Elem {
		if len(elem.Name) > 0 {
			if _, err := builder.WriteRune('/'); err != nil {
				return "", "", err
			}
			if _, err := builder.WriteString(elem.Name); err != nil {
				return "", "", err
			}
		}
		name := builder.String()

		if _, exists := aliases[name]; exists {
			aliasPath = name
		}

		if tags != nil {
			for key, val := range elem.Key {
				key = strings.ReplaceAll(key, "-", "_")

				// Use short-form of key if possible
				if _, exists := tags[key]; exists {
					tags[name+"/"+key] = val
				} else {
					tags[key] = val
				}
			}
		}
	}

	return builder.String(), aliasPath, nil
}

// equalPathNoKeys checks if two gNMI paths are equal, without keys
func equalPathNoKeys(a *gnmiLib.Path, b *gnmiLib.Path) bool {
	if len(a.Elem) != len(b.Elem) {
		return false
	}
	for i := range a.Elem {
		if a.Elem[i].Name != b.Elem[i].Name {
			return false
		}
	}
	return true
}

func pathKeys(gpath *gnmiLib.Path) []*gnmiLib.PathElem {
	var newPath []*gnmiLib.PathElem
	for _, elem := range gpath.Elem {
		if elem.Key != nil {
			newPath = append(newPath, elem)
		}
	}
	return newPath
}

func pathWithPrefix(prefix *gnmiLib.Path, gpath *gnmiLib.Path) *gnmiLib.Path {
	if prefix == nil {
		return gpath
	}
	fullPath := new(gnmiLib.Path)
	fullPath.Origin = prefix.Origin
	fullPath.Target = prefix.Target
	fullPath.Elem = append(prefix.Elem, gpath.Elem...)
	return fullPath
}

func gnmiToFields(name string, updateVal *gnmiLib.TypedValue) (map[string]interface{}, error) {
	var value interface{}
	var jsondata []byte

	// Make sure a value is actually set
	if updateVal == nil || updateVal.Value == nil {
		return nil, nil
	}

	switch val := updateVal.Value.(type) {
	case *gnmiLib.TypedValue_AsciiVal:
		value = val.AsciiVal
	case *gnmiLib.TypedValue_BoolVal:
		value = val.BoolVal
	case *gnmiLib.TypedValue_BytesVal:
		value = val.BytesVal
	case *gnmiLib.TypedValue_DoubleVal:
		value = val.DoubleVal
	case *gnmiLib.TypedValue_DecimalVal:
		//nolint:staticcheck // to maintain backward compatibility with older gnmi specs
		value = float64(val.DecimalVal.Digits) / math.Pow(10, float64(val.DecimalVal.Precision))
	case *gnmiLib.TypedValue_FloatVal:
		//nolint:staticcheck // to maintain backward compatibility with older gnmi specs
		value = val.FloatVal
	case *gnmiLib.TypedValue_IntVal:
		value = val.IntVal
	case *gnmiLib.TypedValue_StringVal:
		value = val.StringVal
	case *gnmiLib.TypedValue_UintVal:
		value = val.UintVal
	case *gnmiLib.TypedValue_JsonIetfVal:
		jsondata = val.JsonIetfVal
	case *gnmiLib.TypedValue_JsonVal:
		jsondata = val.JsonVal
	}

	fields := make(map[string]interface{})
	if value != nil {
		fields[name] = value
	} else if jsondata != nil {
		if err := json.Unmarshal(jsondata, &value); err != nil {
			return nil, fmt.Errorf("failed to parse JSON value: %v", err)
		}
		flattener := jsonparser.JSONFlattener{Fields: fields}
		if err := flattener.FullFlattenJSON(name, value, true, true); err != nil {
			return nil, fmt.Errorf("failed to flatten JSON: %v", err)
		}
	}
	return fields, nil
}

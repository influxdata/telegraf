package json

import (
	"fmt"
	"strconv"
)

type JSONFlattener struct {
	Fields map[string]interface{}
}

// FlattenJSON flattens nested maps/interfaces into a fields map (ignoring bools and string)
func (f *JSONFlattener) FlattenJSON(
	fieldname string,
	v interface{}) error {
	if f.Fields == nil {
		f.Fields = make(map[string]interface{})
	}

	return f.FullFlattenJSON(fieldname, v, false, false)
}

// FullFlattenJSON flattens nested maps/interfaces into a fields map (including bools and string)
func (f *JSONFlattener) FullFlattenJSON(
	fieldName string,
	v interface{},
	convertString bool,
	convertBool bool,
) error {
	if f.Fields == nil {
		f.Fields = make(map[string]interface{})
	}

	switch t := v.(type) {
	case map[string]interface{}:
		for fieldKey, fieldVal := range t {
			if fieldName != "" {
				fieldKey = fieldName + "_" + fieldKey
			}

			err := f.FullFlattenJSON(fieldKey, fieldVal, convertString, convertBool)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		for i, fieldVal := range t {
			fieldKey := strconv.Itoa(i)
			if fieldName != "" {
				fieldKey = fieldName + "_" + fieldKey
			}
			err := f.FullFlattenJSON(fieldKey, fieldVal, convertString, convertBool)
			if err != nil {
				return err
			}
		}
	case float64:
		f.Fields[fieldName] = t
	case string:
		if !convertString {
			return nil
		}
		f.Fields[fieldName] = v.(string)
	case bool:
		if !convertBool {
			return nil
		}
		f.Fields[fieldName] = v.(bool)
	case nil:
		return nil
	default:
		return fmt.Errorf("JSON Flattener: got unexpected type %T with value %v (%s)",
			t, t, fieldName)
	}
	return nil
}

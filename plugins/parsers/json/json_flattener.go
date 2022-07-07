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
	fieldname string,
	v interface{},
	convertString bool,
	convertBool bool,
) error {
	if f.Fields == nil {
		f.Fields = make(map[string]interface{})
	}

	switch t := v.(type) {
	case map[string]interface{}:
		for k, v := range t {
			fieldkey := k
			if fieldname != "" {
				fieldkey = fieldname + "_" + fieldkey
			}

			err := f.FullFlattenJSON(fieldkey, v, convertString, convertBool)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		for i, v := range t {
			fieldkey := strconv.Itoa(i)
			if fieldname != "" {
				fieldkey = fieldname + "_" + fieldkey
			}
			err := f.FullFlattenJSON(fieldkey, v, convertString, convertBool)
			if err != nil {
				return err
			}
		}
	case float64:
		f.Fields[fieldname] = t
	case string:
		if !convertString {
			return nil
		}
		f.Fields[fieldname] = v.(string)
	case bool:
		if !convertBool {
			return nil
		}
		f.Fields[fieldname] = v.(bool)
	case nil:
		return nil
	default:
		return fmt.Errorf("JSON Flattener: got unexpected type %T with value %v (%s)",
			t, t, fieldname)
	}
	return nil
}

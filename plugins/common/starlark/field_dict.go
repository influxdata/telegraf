package starlark

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"go.starlark.net/starlark"

	"github.com/influxdata/telegraf"
)

// FieldDict is a starlark.Value for the metric fields.  It is heavily based on the
// starlark.Dict.
type FieldDict struct {
	*Metric
}

func (d FieldDict) String() string {
	buf := new(strings.Builder)
	buf.WriteString("{") //nolint:revive // from builder.go: "It returns the length of r and a nil error."
	sep := ""
	for _, item := range d.Items() {
		k, v := item[0], item[1]
		buf.WriteString(sep)        //nolint:revive // from builder.go: "It returns the length of r and a nil error."
		buf.WriteString(k.String()) //nolint:revive // from builder.go: "It returns the length of r and a nil error."
		buf.WriteString(": ")       //nolint:revive // from builder.go: "It returns the length of r and a nil error."
		buf.WriteString(v.String()) //nolint:revive // from builder.go: "It returns the length of r and a nil error."
		sep = ", "
	}
	buf.WriteString("}") //nolint:revive // from builder.go: "It returns the length of r and a nil error."
	return buf.String()
}

func (d FieldDict) Type() string {
	return "Fields"
}

func (d FieldDict) Freeze() {
	d.frozen = true
}

func (d FieldDict) Truth() starlark.Bool {
	return len(d.metric.FieldList()) != 0
}

func (d FieldDict) Hash() (uint32, error) {
	return 0, errors.New("not hashable")
}

// AttrNames implements the starlark.HasAttrs interface.
func (d FieldDict) AttrNames() []string {
	return builtinAttrNames(FieldDictMethods)
}

// Attr implements the starlark.HasAttrs interface.
func (d FieldDict) Attr(name string) (starlark.Value, error) {
	return builtinAttr(d, name, FieldDictMethods)
}

var FieldDictMethods = map[string]builtinMethod{
	"clear":      dictClear,
	"get":        dictGet,
	"items":      dictItems,
	"keys":       dictKeys,
	"pop":        dictPop,
	"popitem":    dictPopitem,
	"setdefault": dictSetdefault,
	"update":     dictUpdate,
	"values":     dictValues,
}

// Get implements the starlark.Mapping interface.
func (d FieldDict) Get(key starlark.Value) (v starlark.Value, found bool, err error) {
	if k, ok := key.(starlark.String); ok {
		gv, found := d.metric.GetField(k.GoString())
		if !found {
			return starlark.None, false, nil
		}

		v, err := asStarlarkValue(gv)
		if err != nil {
			return starlark.None, false, err
		}
		return v, true, nil
	}

	return starlark.None, false, errors.New("key must be of type 'str'")
}

// SetKey implements the starlark.HasSetKey interface to support map update
// using x[k]=v syntax, like a dictionary.
func (d FieldDict) SetKey(k, v starlark.Value) error {
	if d.fieldIterCount > 0 {
		return fmt.Errorf("cannot insert during iteration")
	}

	key, ok := k.(starlark.String)
	if !ok {
		return errors.New("field key must be of type 'str'")
	}

	gv, err := asGoValue(v)
	if err != nil {
		return err
	}

	d.metric.AddField(key.GoString(), gv)
	return nil
}

// Items implements the starlark.IterableMapping interface.
func (d FieldDict) Items() []starlark.Tuple {
	items := make([]starlark.Tuple, 0, len(d.metric.FieldList()))
	for _, field := range d.metric.FieldList() {
		key := starlark.String(field.Key)
		sv, err := asStarlarkValue(field.Value)
		if err != nil {
			continue
		}
		pair := starlark.Tuple{key, sv}
		items = append(items, pair)
	}
	return items
}

func (d FieldDict) Clear() error {
	if d.fieldIterCount > 0 {
		return fmt.Errorf("cannot delete during iteration")
	}

	keys := make([]string, 0, len(d.metric.FieldList()))
	for _, field := range d.metric.FieldList() {
		keys = append(keys, field.Key)
	}

	for _, key := range keys {
		d.metric.RemoveField(key)
	}
	return nil
}

func (d FieldDict) PopItem() (v starlark.Value, err error) {
	if d.fieldIterCount > 0 {
		return nil, fmt.Errorf("cannot delete during iteration")
	}

	for _, field := range d.metric.FieldList() {
		k := field.Key
		v := field.Value

		d.metric.RemoveField(k)

		sk := starlark.String(k)
		sv, err := asStarlarkValue(v)
		if err != nil {
			return nil, fmt.Errorf("could not convert to starlark value")
		}

		return starlark.Tuple{sk, sv}, nil
	}

	return nil, errors.New("popitem(): field dictionary is empty")
}

func (d FieldDict) Delete(k starlark.Value) (v starlark.Value, found bool, err error) {
	if d.fieldIterCount > 0 {
		return nil, false, fmt.Errorf("cannot delete during iteration")
	}

	if key, ok := k.(starlark.String); ok {
		value, ok := d.metric.GetField(key.GoString())
		if ok {
			d.metric.RemoveField(key.GoString())
			sv, err := asStarlarkValue(value)
			return sv, ok, err
		}
		return starlark.None, false, nil
	}

	return starlark.None, false, errors.New("key must be of type 'str'")
}

// Iterate implements the starlark.Iterator interface.
func (d FieldDict) Iterate() starlark.Iterator {
	d.fieldIterCount++
	return &FieldIterator{Metric: d.Metric, fields: d.metric.FieldList()}
}

type FieldIterator struct {
	*Metric
	fields []*telegraf.Field
}

// Next implements the starlark.Iterator interface.
func (i *FieldIterator) Next(p *starlark.Value) bool {
	if len(i.fields) == 0 {
		return false
	}

	field := i.fields[0]
	i.fields = i.fields[1:]
	*p = starlark.String(field.Key)

	return true
}

// Done implements the starlark.Iterator interface.
func (i *FieldIterator) Done() {
	i.fieldIterCount--
}

// AsStarlarkValue converts a field value to a starlark.Value.
func asStarlarkValue(value interface{}) (starlark.Value, error) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice:
		length := v.Len()
		array := make([]starlark.Value, length)
		for i := 0; i < length; i++ {
			sVal, err := asStarlarkValue(v.Index(i).Interface())
			if err != nil {
				return starlark.None, err
			}
			array[i] = sVal
		}
		return starlark.NewList(array), nil
	case reflect.Map:
		dict := starlark.NewDict(v.Len())
		iter := v.MapRange()
		for iter.Next() {
			sKey, err := asStarlarkValue(iter.Key().Interface())
			if err != nil {
				return starlark.None, err
			}
			sValue, err := asStarlarkValue(iter.Value().Interface())
			if err != nil {
				return starlark.None, err
			}
			dict.SetKey(sKey, sValue)
		}
		return dict, nil
	case reflect.Float32, reflect.Float64:
		return starlark.Float(v.Float()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return starlark.MakeInt64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return starlark.MakeUint64(v.Uint()), nil
	case reflect.String:
		return starlark.String(v.String()), nil
	case reflect.Bool:
		return starlark.Bool(v.Bool()), nil
	}

	return starlark.None, errors.New("invalid type")
}

// AsGoValue converts a starlark.Value to a field value.
func asGoValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case starlark.Float:
		return float64(v), nil
	case starlark.Int:
		n, ok := v.Int64()
		if !ok {
			return nil, errors.New("cannot represent integer as int64")
		}
		return n, nil
	case starlark.String:
		return string(v), nil
	case starlark.Bool:
		return bool(v), nil
	}

	return nil, errors.New("invalid starlark type")
}

// ToFields converts a starlark.Value to a map of values.
func toFields(value starlark.Value) (map[string]interface{}, error) {
	if value == nil {
		return nil, nil
	}
	items, err := items(value, "The type %T is unsupported as type of collection of fields")
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{}, len(items))
	for _, item := range items {
		key, err := toString(item[0], "The type %T is unsupported as type of key for fields")
		if err != nil {
			return nil, err
		}
		value, err := asGoValue(item[1])
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

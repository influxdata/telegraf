package starlark

import (
	"errors"
	"strings"

	"go.starlark.net/starlark"
)

type MetricDataDict struct {
	data     Accessible
	typename string
}

func (m *MetricDataDict) String() string {
	buf := new(strings.Builder)
	buf.WriteString("{")
	sep := ""
	for _, item := range m.Items() {
		k, v := item[0], item[1]
		buf.WriteString(sep)
		buf.WriteString(k.String())
		buf.WriteString(": ")
		buf.WriteString(v.String())
		sep = ", "
	}
	buf.WriteString("}")
	return buf.String()
}

func (m *MetricDataDict) Type() string {
	return m.typename
}

func (m *MetricDataDict) Freeze() {
	// TODO
}

func (m *MetricDataDict) Truth() starlark.Bool {
	return true
}

func (m *MetricDataDict) Hash() (uint32, error) {
	return 0, errors.New("not hashable")
}

// Get implements the starlark.Mapping interface.
func (m *MetricDataDict) Get(key starlark.Value) (v starlark.Value, found bool, err error) {
	if k, ok := key.(starlark.String); ok {
		gv, found := m.data.Get(k.GoString())
		if !found {
			return starlark.None, false, nil
		}
		v, err = asStarlarkValue(gv)
		return v, true, err
	}

	return starlark.None, false, errors.New("key must be of type 'str'")
}

// SetKey implements the starlark.HasSetKey interface to support map update
// using x[k]=v syntax, like a dictionary.
func (m *MetricDataDict) SetKey(k, v starlark.Value) error {
	var key starlark.String
	var ok bool

	if key, ok = k.(starlark.String); !ok {
		return errors.New("key must be of type 'str'")
	}
	value, err := asGoValue(v)
	if err != nil {
		return err
	}

	return m.data.Add(key.GoString(), value)
}

// Items implements the starlark.IterableMapping interface.
func (m *MetricDataDict) Items() []starlark.Tuple {
	items := make([]starlark.Tuple, 0, m.data.Len())
	for _, fields := range m.data.List() {
		key := starlark.String(fields.Key)
		value, _ := asStarlarkValue(fields.Value)
		pair := starlark.Tuple{key, value}
		items = append(items, pair)
	}
	return items
}

// Items implements the starlark.Mapping interface.
func (m *MetricDataDict) Iterate() starlark.Iterator {
	return &MetricDataIterator{data: m.data}
}

type MetricDataIterator struct {
	data  Accessible
	index int
}

func (i *MetricDataIterator) Next(p *starlark.Value) bool {
	if i.index >= i.data.Len() {
		return false
	}

	fieldkey := i.data.GetIndex(i.index)
	key := starlark.String(fieldkey)

	*p = key
	i.index++
	return true
}

func (i *MetricDataIterator) Done() {
}

// AttrNames implements the starlark.HasAttrs interface.
func (m *MetricDataDict) AttrNames() []string {
	return builtinAttrNames(MetricDataDictMethods)
}

// Attr implements the starlark.HasAttrs interface.
func (m *MetricDataDict) Attr(name string) (starlark.Value, error) {
	return builtinAttr(m, name, MetricDataDictMethods)
}

var MetricDataDictMethods = map[string]builtinMethod{
	"clear":      dict_clear,
	"get":        dict_get,
	"items":      dict_items,
	"keys":       dict_keys,
	"pop":        dict_pop,
	"popitem":    dict_popitem,
	"setdefault": dict_setdefault,
	"update":     dict_update,
	"values":     dict_values,
}

// type Removeable (for builtins)
func (m *MetricDataDict) Clear() error {
	m.data.Clear()
	return nil
}

// FIXME called from builtins
func (m *MetricDataDict) Delete(k starlark.Value) (v starlark.Value, found bool, err error) {
	if key, ok := k.(starlark.String); ok {
		value, ok := m.data.Get(key.GoString())
		if ok {
			m.data.Remove(key.GoString())
			v, err := asStarlarkValue(value)
			return v, ok, err
		}
	}
	return starlark.None, false, errors.New("key must be of type 'str'")
}

// FIXME field conversions
func asStarlarkValue(value interface{}) (starlark.Value, error) {
	switch v := value.(type) {
	case float64:
		return starlark.Float(v), nil
	case int64:
		return starlark.MakeInt64(v), nil
	case uint64:
		return starlark.MakeUint64(v), nil
	case string:
		return starlark.String(v), nil
	case bool:
		return starlark.Bool(v), nil
	}

	return starlark.None, errors.New("invalid type")
}

// FIXME field conversions
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

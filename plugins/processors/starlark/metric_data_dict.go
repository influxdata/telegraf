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

func writeValue(out *strings.Builder, m *MetricDataDict) {
}

func (m *MetricDataDict) Type() string {
	return m.typename
}

func (m *MetricDataDict) Freeze() {
	// To be implemented
}

func (m *MetricDataDict) Truth() starlark.Bool {
	return true
}

func (m *MetricDataDict) Hash() (uint32, error) {
	return 0, errors.New("not hashable")
}

// type Mapping
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

// type HasSetKey
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

	m.data.Add(key.GoString(), value)
	return nil
}

// type IterableMapping
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

func (m *MetricDataDict) Iterate() starlark.Iterator {
	return &MetricDataIterator{entries: m.data.List()}
}

type MetricDataIterator struct {
	entries []AccessibleEntry
	index   int
}

func (i *MetricDataIterator) Next(p *starlark.Value) bool {
	if i.index >= len(i.entries) {
		return false
	}

	field := i.entries[i.index]

	key := starlark.String(field.Key)
	val, _ := asStarlarkValue(field.Value)
	pair := starlark.Tuple{key, val}

	*p = pair
	i.index++
	return true
}

func (i *MetricDataIterator) Done() {
}

// type HasAttrs
func (m *MetricDataDict) Attr(name string) (starlark.Value, error) {
	return builtinAttr(m, name, MetricDataDictMethods)
}

func (m *MetricDataDict) AttrNames() []string {
	return builtinAttrNames(MetricDataDictMethods)
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

// Internal converter functions
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

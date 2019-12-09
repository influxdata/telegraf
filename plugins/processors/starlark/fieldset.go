package starlark

import (
	"errors"
	"fmt"

	"github.com/influxdata/telegraf"
	"go.starlark.net/starlark"
)

// type MutableMapping interface {
// 	starlark.HasSetKey
// }

// type Set interface {
// 	Keys()
// 	Clear()
// 	Delete()
// }

type FieldSet struct {
	metric telegraf.Metric
}

func (m *FieldSet) String() string {
	return "fields"
}

func (m *FieldSet) Type() string {
	return "FieldSet"
}

func (m *FieldSet) Freeze() {
}

func (m *FieldSet) Truth() starlark.Bool {
	return true
}

func (m *FieldSet) Hash() (uint32, error) {
	return 0, errors.New("not hashable")
}

func (m *FieldSet) Get(key starlark.Value) (v starlark.Value, found bool, err error) {
	fmt.Println(1)
	switch key := key.(type) {
	case starlark.String:
		t, ok := m.metric.GetField(key.GoString())
		return AsStarlarkValue(t), ok, nil
	default:
		return starlark.String(""), false, errors.New("type error")
	}
}

func (m *FieldSet) Keys() []starlark.Value {
	items := make([]starlark.Value, 0, len(m.metric.TagList()))
	for _, fields := range m.metric.FieldList() {
		item := starlark.String(fields.Key)
		items = append(items, item)
	}
	return items
}

// type IterableMapping
func (m *FieldSet) Items() []starlark.Tuple {
	items := make([]starlark.Tuple, 0, len(m.metric.TagList()))
	for _, fields := range m.metric.FieldList() {
		pair := starlark.Tuple{
			starlark.String(fields.Key),
			AsStarlarkValue(fields.Value),
		}
		items = append(items, pair)
	}
	return items
}

func AsStarlarkValue(value interface{}) starlark.Value {
	switch v := value.(type) {
	case float64:
		return starlark.Float(v)
	case int64:
		return starlark.MakeInt64(v)
	case uint64:
		return starlark.MakeUint64(v)
	case string:
		return starlark.String(v)
	case bool:
		return starlark.Bool(v)
	default:
		// todo error
	}
	return starlark.None
}

func AsFieldValue(value interface{}) interface{} {
	switch v := value.(type) {
	case starlark.Float:
		return float64(v)
	case starlark.Int:
		n, ok := v.Int64()
		if !ok {
			return nil
		}
		return n
	case starlark.String:
		return string(v)
	case starlark.Bool:
		return bool(v)
	default:
		// todo error
	}
	return starlark.None
}

var fieldsetMethods = map[string]builtinMethod{
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

func (m *FieldSet) AttrNames() []string {
	return builtinAttrNames(tagsetMethods)
}

func (m *FieldSet) Attr(name string) (starlark.Value, error) {
	return builtinAttr(m, name, fieldsetMethods)
}

func (m *FieldSet) Iterate() starlark.Iterator {
	return &FieldIterator{metric: m.metric}
}

func (m *FieldSet) Delete(k starlark.Value) (v starlark.Value, found bool, err error) {
	if key, ok := k.(starlark.String); ok {
		value, ok := m.metric.GetField(key.GoString())
		if ok {
			m.metric.RemoveField(key.GoString())
			return AsStarlarkValue(value), ok, nil
		}
	}
	return nil, false, errors.New("tag key must be of type 'str'")
}

func (m *FieldSet) first() (starlark.Value, bool) {
	for _, field := range m.metric.FieldList() {
		return starlark.String(field.Key), true
	}
	return starlark.None, false
}

func (m *FieldSet) Clear() error {
	for _, field := range m.metric.FieldList() {
		m.metric.RemoveField(field.Key)
	}
	return nil
}

func (m *FieldSet) SetKey(k, v starlark.Value) error {
	var key starlark.String
	var ok bool
	if key, ok = k.(starlark.String); !ok {
		return errors.New("field key must be of type 'str'")
	}
	value := AsFieldValue(v)

	m.metric.AddField(key.GoString(), value)
	return nil
}

func (m *FieldSet) Len() int {
	return len(m.metric.FieldList())
}

var _ starlark.IterableMapping = (*FieldSet)(nil)

type FieldIterator struct {
	metric telegraf.Metric
	index  int
}

var _ starlark.Iterator = (*FieldIterator)(nil)

func (i *FieldIterator) Next(p *starlark.Value) bool {
	if i.index >= len(i.metric.FieldList()) {
		return false
	}

	field := i.metric.FieldList()[i.index]

	key := starlark.String(field.Key)
	val := AsStarlarkValue(field.Value)

	pair := starlark.Tuple{key, val}

	*p = pair
	i.index++
	return true
}

func (i *FieldIterator) Done() {
}

package starlark

import (
	"errors"
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
	"go.starlark.net/starlark"
)

type TagDict struct {
	*Metric
}

func (d TagDict) String() string {
	buf := new(strings.Builder)
	buf.WriteString("{")
	sep := ""
	for _, item := range d.Items() {
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

func (d TagDict) Type() string {
	return "Tags"
}

func (d TagDict) Freeze() {
}

func (d TagDict) Truth() starlark.Bool {
	return len(d.metric.TagList()) != 0
}

func (d TagDict) Hash() (uint32, error) {
	return 0, errors.New("not hashable")
}

// AttrNames implements the starlark.HasAttrs interface.
func (d TagDict) AttrNames() []string {
	return builtinAttrNames(TagDictMethods)
}

// Attr implements the starlark.HasAttrs interface.
func (d TagDict) Attr(name string) (starlark.Value, error) {
	return builtinAttr(d, name, TagDictMethods)
}

var TagDictMethods = map[string]builtinMethod{
	"clear":      tags_clear,
	"get":        dict_get,
	"items":      dict_items,
	"keys":       dict_keys,
	"pop":        tags_pop,
	"popitem":    tags_popitem,
	"setdefault": dict_setdefault,
	"update":     dict_update,
	"values":     dict_values,
}

// Get implements the starlark.Mapping interface.
func (d TagDict) Get(key starlark.Value) (v starlark.Value, found bool, err error) {
	if k, ok := key.(starlark.String); ok {
		gv, found := d.metric.GetTag(k.GoString())
		if !found {
			return starlark.None, false, nil
		}
		return starlark.String(gv), true, err
	}

	return starlark.None, false, errors.New("key must be of type 'str'")
}

// SetKey implements the starlark.HasSetKey interface to support map update
// using x[k]=v syntax, like a dictionary.
func (d TagDict) SetKey(k, v starlark.Value) error {
	key, ok := k.(starlark.String)
	if !ok {
		return errors.New("tag key must be of type 'str'")
	}

	value, ok := v.(starlark.String)
	if !ok {
		return errors.New("tag value must be of type 'str'")
	}

	d.metric.AddTag(key.GoString(), value.GoString())
	return nil
}

// Items implements the starlark.IterableMapping interface.
func (d TagDict) Items() []starlark.Tuple {
	items := make([]starlark.Tuple, 0, len(d.metric.TagList()))
	for _, tag := range d.metric.TagList() {
		key := starlark.String(tag.Key)
		value := starlark.String(tag.Value)
		pair := starlark.Tuple{key, value}
		items = append(items, pair)
	}
	return items
}

func (d TagDict) Clear() error {
	keys := make([]string, 0, len(d.metric.TagList()))
	for _, tag := range d.metric.TagList() {
		keys = append(keys, tag.Key)
	}

	for _, key := range keys {
		d.metric.RemoveTag(key)
	}
	return nil
}

func (d TagDict) Delete(k starlark.Value) (v starlark.Value, found bool, err error) {
	if key, ok := k.(starlark.String); ok {
		value, ok := d.metric.GetTag(key.GoString())
		if ok {
			d.metric.RemoveTag(key.GoString())
			v := starlark.String(value)
			return v, ok, err
		}
	}

	return starlark.None, false, errors.New("key must be of type 'str'")
}

// Items implements the starlark.Mapping interface.
func (d TagDict) Iterate() starlark.Iterator {
	return &TagIterator{tags: d.metric.TagList()}
}

type TagIterator struct {
	tags []*telegraf.Tag
}

func (i *TagIterator) Next(p *starlark.Value) bool {
	if len(i.tags) == 0 {
		return false
	}

	tag := i.tags[0]
	i.tags = i.tags[1:]
	*p = starlark.String(tag.Key)

	return true
}

func (i *TagIterator) Done() {
}

// --- dictionary methods ---

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·clear
func tags_clear(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}
	return starlark.None, b.Receiver().(TagDict).Clear()
}

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·pop
func tags_pop(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var k, d starlark.Value
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &k, &d); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}
	if v, found, err := b.Receiver().(TagDict).Delete(k); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err) // dict is frozen or key is unhashable
	} else if found {
		return v, nil
	} else if d != nil {
		return d, nil
	}
	return starlark.None, fmt.Errorf("%s: missing key", b.Name())
}

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·popitem
func tags_popitem(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}

	recv := b.Receiver().(TagDict)

	for _, tag := range recv.metric.TagList() {
		k := tag.Key
		v := tag.Value

		recv.metric.RemoveTag(k)

		sk := starlark.String(k)
		sv := starlark.String(v)
		return starlark.Tuple{sk, sv}, nil
	}

	return nil, nameErr(b, "empty dict")
}

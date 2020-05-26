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
	"get":        tags_get,
	"items":      tags_items,
	"keys":       tags_keys,
	"pop":        tags_pop,
	"popitem":    tags_popitem,
	"setdefault": tags_setdefault,
	"update":     tags_update,
	"values":     tags_values,
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

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·get
func tags_get(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key, dflt starlark.Value
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &key, &dflt); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}
	if v, ok, err := b.Receiver().(starlark.Mapping).Get(key); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	} else if ok {
		return v, nil
	} else if dflt != nil {
		return dflt, nil
	}
	return starlark.None, nil
}

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·items
func tags_items(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}
	items := b.Receiver().(starlark.IterableMapping).Items()
	res := make([]starlark.Value, len(items))
	for i, item := range items {
		res[i] = item // convert [2]starlark.Value to starlark.Value
	}
	return starlark.NewList(res), nil
}

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·keys
func tags_keys(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}

	items := b.Receiver().(starlark.IterableMapping).Items()
	res := make([]starlark.Value, len(items))
	for i, item := range items {
		res[i] = item[0]
	}
	return starlark.NewList(res), nil
}

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·pop
func tags_pop(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var k, d starlark.Value
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &k, &d); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}
	if v, found, err := b.Receiver().(Removeable).Delete(k); err != nil {
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

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·setdefault
func tags_setdefault(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key, dflt starlark.Value = nil, starlark.None
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &key, &dflt); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}

	recv := b.Receiver().(starlark.HasSetKey)
	v, found, err := recv.Get(key)
	if err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}
	if !found {
		v = dflt
		if err := recv.SetKey(key, dflt); err != nil {
			return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
		}
	}
	return v, nil
}

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·update
func tags_update(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// Unpack the arguments
	if len(args) > 1 {
		return nil, fmt.Errorf("update: got %d arguments, want at most 1", len(args))
	}

	// Get the target
	dict := b.Receiver().(starlark.HasSetKey)

	if len(args) == 1 {
		switch updates := args[0].(type) {
		case starlark.IterableMapping:
			// Iterate over dict's key/value pairs, not just keys.
			for _, item := range updates.Items() {
				if err := dict.SetKey(item[0], item[1]); err != nil {
					return nil, err // dict is frozen
				}
			}
		default:
			// all other sequences
			iter := starlark.Iterate(updates)
			if iter == nil {
				return nil, fmt.Errorf("got %s, want iterable", updates.Type())
			}
			defer iter.Done()
			var pair starlark.Value
			for i := 0; iter.Next(&pair); i++ {
				iter2 := starlark.Iterate(pair)
				if iter2 == nil {
					return nil, fmt.Errorf("dictionary update sequence element #%d is not iterable (%s)", i, pair.Type())

				}
				defer iter2.Done()
				len := starlark.Len(pair)
				if len < 0 {
					return nil, fmt.Errorf("dictionary update sequence element #%d has unknown length (%s)", i, pair.Type())
				} else if len != 2 {
					return nil, fmt.Errorf("dictionary update sequence element #%d has length %d, want 2", i, len)
				}
				var k, v starlark.Value
				iter2.Next(&k)
				iter2.Next(&v)
				if err := dict.SetKey(k, v); err != nil {
					return nil, err
				}
			}
		}
	}

	// Then add the kwargs.
	before := starlark.Len(dict)
	for _, pair := range kwargs {
		if err := dict.SetKey(pair[0], pair[1]); err != nil {
			return nil, err // dict is frozen
		}
	}
	// In the common case, each kwarg will add another dict entry.
	// If that's not so, check whether it is because there was a duplicate kwarg.
	if starlark.Len(dict) < before+len(kwargs) {
		keys := make(map[starlark.String]bool, len(kwargs))
		for _, kv := range kwargs {
			k := kv[0].(starlark.String)
			if keys[k] {
				return nil, fmt.Errorf("duplicate keyword arg: %v", k)
			}
			keys[k] = true
		}
	}

	return starlark.None, nil
}

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·update
func tags_values(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}
	items := b.Receiver().(starlark.IterableMapping).Items()
	res := make([]starlark.Value, len(items))
	for i, item := range items {
		res[i] = item[1]
	}
	return starlark.NewList(res), nil
}

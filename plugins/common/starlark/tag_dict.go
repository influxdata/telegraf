package starlark

import (
	"errors"
	"fmt"
	"strings"

	"go.starlark.net/starlark"

	"github.com/influxdata/telegraf"
)

// TagDict is a starlark.Value for the metric tags.  It is heavily based on the
// starlark.Dict.
type TagDict struct {
	*Metric
}

func (d TagDict) String() string {
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

func (d TagDict) Type() string {
	return "Tags"
}

func (d TagDict) Freeze() {
	d.frozen = true
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
	if d.tagIterCount > 0 {
		return fmt.Errorf("cannot insert during iteration")
	}

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
	if d.tagIterCount > 0 {
		return fmt.Errorf("cannot delete during iteration")
	}

	keys := make([]string, 0, len(d.metric.TagList()))
	for _, tag := range d.metric.TagList() {
		keys = append(keys, tag.Key)
	}

	for _, key := range keys {
		d.metric.RemoveTag(key)
	}
	return nil
}

func (d TagDict) PopItem() (v starlark.Value, err error) {
	if d.tagIterCount > 0 {
		return nil, fmt.Errorf("cannot delete during iteration")
	}

	for _, tag := range d.metric.TagList() {
		k := tag.Key
		v := tag.Value

		d.metric.RemoveTag(k)

		sk := starlark.String(k)
		sv := starlark.String(v)
		return starlark.Tuple{sk, sv}, nil
	}

	return nil, errors.New("popitem(): tag dictionary is empty")
}

func (d TagDict) Delete(k starlark.Value) (v starlark.Value, found bool, err error) {
	if d.tagIterCount > 0 {
		return nil, false, fmt.Errorf("cannot delete during iteration")
	}

	if key, ok := k.(starlark.String); ok {
		value, ok := d.metric.GetTag(key.GoString())
		if ok {
			d.metric.RemoveTag(key.GoString())
			v := starlark.String(value)
			return v, ok, err
		}
		return starlark.None, false, nil
	}

	return starlark.None, false, errors.New("key must be of type 'str'")
}

// Iterate implements the starlark.Iterator interface.
func (d TagDict) Iterate() starlark.Iterator {
	d.tagIterCount++
	return &TagIterator{Metric: d.Metric, tags: d.metric.TagList()}
}

type TagIterator struct {
	*Metric
	tags []*telegraf.Tag
}

// Next implements the starlark.Iterator interface.
func (i *TagIterator) Next(p *starlark.Value) bool {
	if len(i.tags) == 0 {
		return false
	}

	tag := i.tags[0]
	i.tags = i.tags[1:]
	*p = starlark.String(tag.Key)

	return true
}

// Done implements the starlark.Iterator interface.
func (i *TagIterator) Done() {
	i.tagIterCount--
}

// ToTags converts a starlark.Value to a map of string.
func toTags(value starlark.Value) (map[string]string, error) {
	if value == nil {
		return nil, nil
	}
	items, err := items(value, "The type %T is unsupported as type of collection of tags")
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(items))
	for _, item := range items {
		key, err := toString(item[0], "The type %T is unsupported as type of key for tags")
		if err != nil {
			return nil, err
		}
		value, err := toString(item[1], "The type %T is unsupported as type of value for tags")
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

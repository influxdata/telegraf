package starlark

import (
	"fmt"
	"sort"
	"time"

	"github.com/influxdata/telegraf/metric"
	"go.starlark.net/starlark"
)

func newMetric(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name starlark.String
	if err := starlark.UnpackPositionalArgs("Metric", args, kwargs, 1, &name); err != nil {
		return nil, err
	}

	m, err := metric.New(string(name), nil, nil, time.Now())
	if err != nil {
		return nil, err
	}

	return &Metric{metric: m}, nil
}

func deepcopy(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var sm *Metric
	if err := starlark.UnpackPositionalArgs("deepcopy", args, kwargs, 1, &sm); err != nil {
		return nil, err
	}

	dup := sm.metric.Copy()
	dup.Drop()
	return &Metric{metric: dup}, nil
}

type builtinMethod func(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)

func builtinAttr(recv starlark.Value, name string, methods map[string]builtinMethod) (starlark.Value, error) {
	method := methods[name]
	if method == nil {
		return starlark.None, fmt.Errorf("no such method '%s'", name)
	}

	// Allocate a closure over 'method'.
	impl := func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return method(b, args, kwargs)
	}
	return starlark.NewBuiltin(name, impl).BindReceiver(recv), nil
}

func builtinAttrNames(methods map[string]builtinMethod) []string {
	names := make([]string, 0, len(methods))
	for name := range methods {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// nameErr returns an error message of the form "name: msg"
// where name is b.Name() and msg is a string or error.
func nameErr(b *starlark.Builtin, msg interface{}) error {
	return fmt.Errorf("%s: %v", b.Name(), msg)
}

// --- dictionary methods ---

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·clear
func dict_clear(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}

	type HasClear interface {
		Clear() error
	}
	return starlark.None, b.Receiver().(HasClear).Clear()
}

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·pop
func dict_pop(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var k, d starlark.Value
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &k, &d); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}

	type HasDelete interface {
		Delete(k starlark.Value) (starlark.Value, bool, error)
	}
	if v, found, err := b.Receiver().(HasDelete).Delete(k); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err) // dict is frozen or key is unhashable
	} else if found {
		return v, nil
	} else if d != nil {
		return d, nil
	}
	return starlark.None, fmt.Errorf("%s: missing key", b.Name())
}

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·popitem
func dict_popitem(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}

	type HasPopItem interface {
		PopItem() (starlark.Value, error)
	}
	return b.Receiver().(HasPopItem).PopItem()
}

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·get
func dict_get(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·setdefault
func dict_setdefault(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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
func dict_update(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·items
func dict_items(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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
func dict_keys(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·update
func dict_values(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

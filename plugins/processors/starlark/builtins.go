package starlark

import (
	"fmt"
	"sort"

	"go.starlark.net/starlark"
)

type Removeable interface {
	starlark.Value
	Clear() error
  Delete(starlark.Value) (starlark.Value, bool, error)
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

// --- dictionary methods ---

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·clear
func dict_clear(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}
	return starlark.None, b.Receiver().(Removeable).Clear()
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

// https://github.com/google/starlark-go/blob/master/doc/spec.md#dict·pop
func dict_pop(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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
func dict_popitem(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	}

	// Get the item
	iter := b.Receiver().(starlark.Iterable).Iterate()
	if iter == nil {
		return starlark.None, fmt.Errorf("%s: couldn't get iterator", b.Name())
	}
	defer iter.Done()

	var item starlark.Value
	if ok := iter.Next(&item); !ok {
		return starlark.None, fmt.Errorf("%s: empty", b.Name())
	}

	// Remove the item
	tuple,ok := item.(starlark.Tuple)
	if !ok || tuple.Len() != 2 {
		return starlark.None, fmt.Errorf("%s: returned item is not a valid tuple", b.Name())
	}

	if _, found, err := b.Receiver().(Removeable).Delete(tuple[0]); err != nil {
		return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
	} else if !found {
		return starlark.None, fmt.Errorf("%s: key '%s' not found", b.Name(), tuple[0])
	}

	return item, nil
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
	if len(args) != 1 {
		return starlark.None, fmt.Errorf("update: got %d arguments, expected 1", len(args))
	}

	// Get the target
	recv := b.Receiver().(starlark.HasSetKey)

	// Use the specified iterable argument to update if given
	if args[0] != starlark.None {
		// We cannot simply iterate over a dictionary as this will return key and
		// value alternatingly instead of returning pairs.... :-0
		// Anyway we assume that if we get an iterable mapping it is much more
		// efficient to use the Items() method.
		if iter_arg, ok := args[0].(starlark.IterableMapping); ok {
			for _,item := range iter_arg.Items() {
				if err := dict_set_tuple(recv, item); err != nil {
					return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
				}
			}
		} else if iter_arg, ok := args[0].(starlark.Iterable); ok {
			iter := iter_arg.Iterate()
			defer iter.Done()

			var v starlark.Value
			for iter.Next(&v) {
				item, ok := v.(starlark.Tuple)
				if !ok {
					return starlark.None, fmt.Errorf("%s: item is not a tuple", b.Name())
				}
				if err := dict_set_tuple(recv, item); err != nil {
					return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
				}
			}
		} else {
			return starlark.None, fmt.Errorf("%s: argument is not iterable", b.Name())
		}
	}

	// Use the specified keyword-argument(s) to update if any
	for _,item := range kwargs {
		if err := dict_set_tuple(recv, item); err != nil {
			return starlark.None, fmt.Errorf("%s: %v", b.Name(), err)
		}
	}

	return starlark.None, nil
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

// -- internal functions --
func dict_set_tuple(recv starlark.HasSetKey, pair starlark.Tuple) error {
	if pair.Len() != 2 {
		return fmt.Errorf("item is not a key/value pair")
	}
	recv.SetKey(pair[0], pair[1])

	return nil
}

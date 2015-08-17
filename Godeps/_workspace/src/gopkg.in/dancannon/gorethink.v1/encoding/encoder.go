// This code is based on encoding/json and gorilla/schema

package encoding

import (
	"errors"
	"reflect"
	"runtime"
	"sync"
)

type encoderFunc func(v reflect.Value) interface{}

// Encode returns the encoded value of v.
//
// Encode  traverses the value v recursively and looks for structs. If a struct
// is found then it is checked for tagged fields and convert to
// map[string]interface{}
func Encode(v interface{}) (ev interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if v, ok := r.(string); ok {
				err = errors.New(v)
			} else {
				err = r.(error)
			}
		}
	}()

	return encode(reflect.ValueOf(v)), nil
}

func encode(v reflect.Value) interface{} {
	return valueEncoder(v)(v)
}

var encoderCache struct {
	sync.RWMutex
	m map[reflect.Type]encoderFunc
}

func valueEncoder(v reflect.Value) encoderFunc {
	if !v.IsValid() {
		return invalidValueEncoder
	}
	return typeEncoder(v.Type())
}

func typeEncoder(t reflect.Type) encoderFunc {
	encoderCache.RLock()
	f := encoderCache.m[t]
	encoderCache.RUnlock()
	if f != nil {
		return f
	}

	// To deal with recursive types, populate the map with an
	// indirect func before we build it. This type waits on the
	// real func (f) to
	//  be ready and then calls it.  This indirect
	// func is only used for recursive types.
	encoderCache.Lock()
	var wg sync.WaitGroup
	wg.Add(1)
	encoderCache.m[t] = func(v reflect.Value) interface{} {
		wg.Wait()
		return f(v)
	}
	encoderCache.Unlock()

	// Compute fields without lock.
	// Might duplicate effort but won't hold other computations back.
	f = newTypeEncoder(t, true)
	wg.Done()
	encoderCache.Lock()
	encoderCache.m[t] = f
	encoderCache.Unlock()
	return f
}

// IgnoreType causes the encoder to ignore a type when encoding
func IgnoreType(t reflect.Type) {
	encoderCache.Lock()
	encoderCache.m[t] = doNothingEncoder
	encoderCache.Unlock()
}

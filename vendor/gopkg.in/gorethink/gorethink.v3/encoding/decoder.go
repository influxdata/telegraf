package encoding

import (
	"errors"
	"reflect"
	"runtime"
	"sync"
)

var byteSliceType = reflect.TypeOf([]byte(nil))

type decoderFunc func(dv reflect.Value, sv reflect.Value)

// Decode decodes map[string]interface{} into a struct. The first parameter
// must be a pointer.
func Decode(dst interface{}, src interface{}) (err error) {
	return decode(dst, src, true)
}

func Merge(dst interface{}, src interface{}) (err error) {
	return decode(dst, src, false)
}

func decode(dst interface{}, src interface{}, blank bool) (err error) {
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

	dv := reflect.ValueOf(dst)
	sv := reflect.ValueOf(src)
	if dv.Kind() != reflect.Ptr {
		return &DecodeTypeError{
			DestType: dv.Type(),
			SrcType:  sv.Type(),
			Reason:   "must be a pointer",
		}
	}

	dv = dv.Elem()
	if !dv.CanAddr() {
		return &DecodeTypeError{
			DestType: dv.Type(),
			SrcType:  sv.Type(),
			Reason:   "must be addressable",
		}
	}

	decodeValue(dv, sv, blank)
	return nil
}

// decodeValue decodes the source value into the destination value
func decodeValue(dv, sv reflect.Value, blank bool) {
	valueDecoder(dv, sv, blank)(dv, sv)
}

type decoderCacheKey struct {
	dt, st reflect.Type
	blank  bool
}

var decoderCache struct {
	sync.RWMutex
	m map[decoderCacheKey]decoderFunc
}

func valueDecoder(dv, sv reflect.Value, blank bool) decoderFunc {
	if !sv.IsValid() {
		return invalidValueDecoder
	}

	if dv.IsValid() {
		dv = indirect(dv, false)
		if blank {
			dv.Set(reflect.Zero(dv.Type()))
		}
	}

	return typeDecoder(dv.Type(), sv.Type(), blank)
}

func typeDecoder(dt, st reflect.Type, blank bool) decoderFunc {
	decoderCache.RLock()
	f := decoderCache.m[decoderCacheKey{dt, st, blank}]
	decoderCache.RUnlock()
	if f != nil {
		return f
	}

	// To deal with recursive types, populate the map with an
	// indirect func before we build it. This type waits on the
	// real func (f) to be ready and then calls it.  This indirect
	// func is only used for recursive types.
	decoderCache.Lock()
	var wg sync.WaitGroup
	wg.Add(1)
	decoderCache.m[decoderCacheKey{dt, st, blank}] = func(dv, sv reflect.Value) {
		wg.Wait()
		f(dv, sv)
	}
	decoderCache.Unlock()

	// Compute fields without lock.
	// Might duplicate effort but won't hold other computations back.
	f = newTypeDecoder(dt, st, blank)
	wg.Done()
	decoderCache.Lock()
	decoderCache.m[decoderCacheKey{dt, st, blank}] = f
	decoderCache.Unlock()
	return f
}

// indirect walks down v allocating pointers as needed,
// until it gets to a non-pointer.
func indirect(v reflect.Value, decodeNull bool) reflect.Value {
	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && (!decodeNull || e.Elem().Kind() == reflect.Ptr) {
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

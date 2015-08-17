package encoding

import (
	"encoding/base64"
	"math"
	"reflect"
	"time"
)

// newTypeEncoder constructs an encoderFunc for a type.
// The returned encoder only checks CanAddr when allowAddr is true.
func newTypeEncoder(t reflect.Type, allowAddr bool) encoderFunc {
	if t.Implements(marshalerType) {
		return marshalerEncoder
	}
	if t.Kind() != reflect.Ptr && allowAddr {
		if reflect.PtrTo(t).Implements(marshalerType) {
			return newCondAddrEncoder(addrMarshalerEncoder, newTypeEncoder(t, false))
		}
	}

	// Check for psuedo-types first
	switch t {
	case timeType:
		return timePseudoTypeEncoder
	}

	switch t.Kind() {
	case reflect.Bool:
		return boolEncoder
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intEncoder
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintEncoder
	case reflect.Float32, reflect.Float64:
		return floatEncoder
	case reflect.String:
		return stringEncoder
	case reflect.Interface:
		return interfaceEncoder
	case reflect.Struct:
		return newStructEncoder(t)
	case reflect.Map:
		return newMapEncoder(t)
	case reflect.Slice:
		return newSliceEncoder(t)
	case reflect.Array:
		return newArrayEncoder(t)
	case reflect.Ptr:
		return newPtrEncoder(t)
	default:
		return unsupportedTypeEncoder
	}
}

func invalidValueEncoder(v reflect.Value) interface{} {
	return nil
}

func doNothingEncoder(v reflect.Value) interface{} {
	return v.Interface()
}

func marshalerEncoder(v reflect.Value) interface{} {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	m := v.Interface().(Marshaler)
	ev, err := m.MarshalRQL()
	if err != nil {
		panic(&MarshalerError{v.Type(), err})
	}

	return ev
}

func addrMarshalerEncoder(v reflect.Value) interface{} {
	va := v.Addr()
	if va.IsNil() {
		return nil
	}
	m := va.Interface().(Marshaler)
	ev, err := m.MarshalRQL()
	if err != nil {
		panic(&MarshalerError{v.Type(), err})
	}

	return ev
}

func boolEncoder(v reflect.Value) interface{} {
	if v.Bool() {
		return true
	} else {
		return false
	}
}

func intEncoder(v reflect.Value) interface{} {
	return v.Int()
}

func uintEncoder(v reflect.Value) interface{} {
	return v.Uint()
}

func floatEncoder(v reflect.Value) interface{} {
	return v.Float()
}

func stringEncoder(v reflect.Value) interface{} {
	return v.String()
}

func interfaceEncoder(v reflect.Value) interface{} {
	if v.IsNil() {
		return nil
	}
	return encode(v.Elem())
}

func unsupportedTypeEncoder(v reflect.Value) interface{} {
	panic(&UnsupportedTypeError{v.Type()})
}

type structEncoder struct {
	fields    []field
	fieldEncs []encoderFunc
}

func (se *structEncoder) encode(v reflect.Value) interface{} {
	m := make(map[string]interface{})

	for i, f := range se.fields {
		fv := fieldByIndex(v, f.index)
		if !fv.IsValid() || f.omitEmpty && se.isEmptyValue(fv) {
			continue
		}

		m[f.name] = se.fieldEncs[i](fv)
	}

	return m
}

func (se *structEncoder) isEmptyValue(v reflect.Value) bool {
	if v.Type() == timeType {
		return v.Interface().(time.Time) == time.Time{}
	}

	return isEmptyValue(v)
}

func newStructEncoder(t reflect.Type) encoderFunc {
	fields := cachedTypeFields(t)
	se := &structEncoder{
		fields:    fields,
		fieldEncs: make([]encoderFunc, len(fields)),
	}
	for i, f := range fields {
		se.fieldEncs[i] = typeEncoder(typeByIndex(t, f.index))
	}
	return se.encode
}

type mapEncoder struct {
	elemEnc encoderFunc
}

func (me *mapEncoder) encode(v reflect.Value) interface{} {
	if v.IsNil() {
		return nil
	}

	m := make(map[string]interface{})

	for _, k := range v.MapKeys() {
		m[k.String()] = me.elemEnc(v.MapIndex(k))
	}

	return m
}

func newMapEncoder(t reflect.Type) encoderFunc {
	if t.Key().Kind() != reflect.String {
		return unsupportedTypeEncoder
	}
	me := &mapEncoder{typeEncoder(t.Elem())}
	return me.encode
}

// sliceEncoder just wraps an arrayEncoder, checking to make sure the value isn't nil.
type sliceEncoder struct {
	arrayEnc encoderFunc
}

func (se *sliceEncoder) encode(v reflect.Value) interface{} {
	if v.IsNil() {
		return []interface{}{}
	}
	return se.arrayEnc(v)
}

func newSliceEncoder(t reflect.Type) encoderFunc {
	// Byte slices get special treatment; arrays don't.
	if t.Elem().Kind() == reflect.Uint8 {
		return encodeByteSlice
	}
	enc := &sliceEncoder{newArrayEncoder(t)}
	return enc.encode
}

type arrayEncoder struct {
	elemEnc encoderFunc
}

func (ae *arrayEncoder) encode(v reflect.Value) interface{} {
	n := v.Len()

	a := make([]interface{}, n)
	for i := 0; i < n; i++ {
		a[i] = ae.elemEnc(v.Index(i))
	}

	return a
}

func newArrayEncoder(t reflect.Type) encoderFunc {
	enc := &arrayEncoder{typeEncoder(t.Elem())}
	return enc.encode
}

type ptrEncoder struct {
	elemEnc encoderFunc
}

func (pe *ptrEncoder) encode(v reflect.Value) interface{} {
	if v.IsNil() {
		return nil
	}
	return pe.elemEnc(v.Elem())
}

func newPtrEncoder(t reflect.Type) encoderFunc {
	enc := &ptrEncoder{typeEncoder(t.Elem())}
	return enc.encode
}

type condAddrEncoder struct {
	canAddrEnc, elseEnc encoderFunc
}

func (ce *condAddrEncoder) encode(v reflect.Value) interface{} {
	if v.CanAddr() {
		return ce.canAddrEnc(v)
	} else {
		return ce.elseEnc(v)
	}
}

// newCondAddrEncoder returns an encoder that checks whether its value
// CanAddr and delegates to canAddrEnc if so, else to elseEnc.
func newCondAddrEncoder(canAddrEnc, elseEnc encoderFunc) encoderFunc {
	enc := &condAddrEncoder{canAddrEnc: canAddrEnc, elseEnc: elseEnc}
	return enc.encode
}

// Pseudo-type encoders

// Encode a time.Time value to the TIME RQL type
func timePseudoTypeEncoder(v reflect.Value) interface{} {
	t := v.Interface().(time.Time)

	timeVal := float64(t.UnixNano()) / float64(time.Second)

	// use seconds-since-epoch precision if time.Time `t`
	// is before the oldest nanosecond time
	if t.Before(time.Unix(0, math.MinInt64)) {
		timeVal = float64(t.Unix())
	}

	return map[string]interface{}{
		"$reql_type$": "TIME",
		"epoch_time":  timeVal,
		"timezone":    "+00:00",
	}
}

// Encode a byte slice to the BINARY RQL type
func encodeByteSlice(v reflect.Value) interface{} {
	var b []byte
	if !v.IsNil() {
		b = v.Bytes()
	}

	dst := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(dst, b)

	return map[string]interface{}{
		"$reql_type$": "BINARY",
		"data":        string(dst),
	}
}

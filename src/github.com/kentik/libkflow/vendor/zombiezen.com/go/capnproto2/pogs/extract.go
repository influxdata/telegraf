package pogs

import (
	"errors"
	"fmt"
	"math"
	"reflect"

	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/internal/nodemap"
	"zombiezen.com/go/capnproto2/std/capnp/schema"
)

// Extract copies s into val, a pointer to a Go struct.
func Extract(val interface{}, typeID uint64, s capnp.Struct) error {
	e := new(extracter)
	err := e.extractStruct(reflect.ValueOf(val), typeID, s)
	if err != nil {
		return fmt.Errorf("pogs: extract @%#x: %v", typeID, err)
	}
	return nil
}

type extracter struct {
	nodes nodemap.Map
}

func (e *extracter) extractStruct(val reflect.Value, typeID uint64, s capnp.Struct) error {
	if val.Kind() == reflect.Ptr {
		if val.Type().Elem().Kind() != reflect.Struct {
			return fmt.Errorf("can't extract struct into %v", val.Type())
		}
		switch {
		case !val.CanSet() && val.IsNil():
			// Even if the Cap'n Proto pointer isn't valid, this is probably
			// the caller's fault and will be a bug at some point.
			return errors.New("can't extract struct into nil")
		case !s.IsValid() && val.CanSet():
			val.Set(reflect.Zero(val.Type()))
			return nil
		case s.IsValid() && val.CanSet() && val.IsNil():
			val.Set(reflect.New(val.Type().Elem()))
		}
		val = val.Elem()
	} else if val.Kind() != reflect.Struct {
		return fmt.Errorf("can't extract struct into %v", val.Type())
	}
	if !val.CanSet() {
		return errors.New("can't modify struct, did you pass in a pointer to your struct?")
	}
	n, err := e.nodes.Find(typeID)
	if err != nil {
		return err
	}
	if !n.IsValid() || n.Which() != schema.Node_Which_structNode {
		return fmt.Errorf("cannot find struct type %#x", typeID)
	}
	props, err := mapStruct(val.Type(), n)
	if err != nil {
		return fmt.Errorf("can't extract %s: %v", val.Type(), err)
	}
	var discriminant uint16
	hasWhich := false
	if hasDiscriminant(n) {
		discriminant = s.Uint16(capnp.DataOffset(n.StructNode().DiscriminantOffset() * 2))
		if err := props.setWhich(val, discriminant); err == nil {
			hasWhich = true
		} else if !isNoWhichError(err) {
			return err
		}
	}
	fields, err := n.StructNode().Fields()
	if err != nil {
		return err
	}
	for i := 0; i < fields.Len(); i++ {
		f := fields.At(i)
		vf := props.makeFieldByOrdinal(val, i)
		if !vf.IsValid() {
			// Don't have a field for this.
			continue
		}
		if dv := f.DiscriminantValue(); dv != schema.Field_noDiscriminant {
			if !hasWhich {
				return fmt.Errorf("can't extract %s into %v: has union field but no Which field", shortDisplayName(n), val.Type())
			}
			if dv != discriminant {
				continue
			}
		}
		switch f.Which() {
		case schema.Field_Which_slot:
			if err := e.extractField(vf, s, f); err != nil {
				return err
			}
		case schema.Field_Which_group:
			if err := e.extractStruct(vf, f.Group().TypeId(), s); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *extracter) extractField(val reflect.Value, s capnp.Struct, f schema.Field) error {
	typ, err := f.Slot().Type()
	if err != nil {
		return err
	}
	dv, err := f.Slot().DefaultValue()
	if err != nil {
		return err
	}
	if dv.IsValid() && int(typ.Which()) != int(dv.Which()) {
		name, _ := f.NameBytes()
		return fmt.Errorf("extract field %s: default value is a %v, want %v", name, dv.Which(), typ.Which())
	}
	if !isTypeMatch(val.Type(), typ) {
		name, _ := f.NameBytes()
		return fmt.Errorf("can't extract field %s of type %v into a Go %v", name, typ.Which(), val.Type())
	}
	switch typ.Which() {
	case schema.Type_Which_bool:
		v := s.Bit(capnp.BitOffset(f.Slot().Offset()))
		d := dv.Bool()
		val.SetBool(v != d) // != acts as XOR
	case schema.Type_Which_int8:
		v := int8(s.Uint8(capnp.DataOffset(f.Slot().Offset())))
		d := dv.Int8()
		val.SetInt(int64(v ^ d))
	case schema.Type_Which_int16:
		v := int16(s.Uint16(capnp.DataOffset(f.Slot().Offset() * 2)))
		d := dv.Int16()
		val.SetInt(int64(v ^ d))
	case schema.Type_Which_int32:
		v := int32(s.Uint32(capnp.DataOffset(f.Slot().Offset() * 4)))
		d := dv.Int32()
		val.SetInt(int64(v ^ d))
	case schema.Type_Which_int64:
		v := int64(s.Uint64(capnp.DataOffset(f.Slot().Offset() * 8)))
		d := dv.Int64()
		val.SetInt(v ^ d)
	case schema.Type_Which_uint8:
		v := s.Uint8(capnp.DataOffset(f.Slot().Offset()))
		d := dv.Uint8()
		val.SetUint(uint64(v ^ d))
	case schema.Type_Which_uint16:
		v := s.Uint16(capnp.DataOffset(f.Slot().Offset() * 2))
		d := dv.Uint16()
		val.SetUint(uint64(v ^ d))
	case schema.Type_Which_enum:
		v := s.Uint16(capnp.DataOffset(f.Slot().Offset() * 2))
		d := dv.Enum()
		val.SetUint(uint64(v ^ d))
	case schema.Type_Which_uint32:
		v := s.Uint32(capnp.DataOffset(f.Slot().Offset() * 4))
		d := dv.Uint32()
		val.SetUint(uint64(v ^ d))
	case schema.Type_Which_uint64:
		v := s.Uint64(capnp.DataOffset(f.Slot().Offset() * 8))
		d := dv.Uint64()
		val.SetUint(v ^ d)
	case schema.Type_Which_float32:
		v := s.Uint32(capnp.DataOffset(f.Slot().Offset() * 4))
		d := math.Float32bits(dv.Float32())
		val.SetFloat(float64(math.Float32frombits(v ^ d)))
	case schema.Type_Which_float64:
		v := s.Uint64(capnp.DataOffset(f.Slot().Offset() * 8))
		d := math.Float64bits(dv.Float64())
		val.SetFloat(math.Float64frombits(v ^ d))
	case schema.Type_Which_text:
		p, err := s.Ptr(uint16(f.Slot().Offset()))
		if err != nil {
			return err
		}
		var b []byte
		if p.IsValid() {
			b = p.TextBytes()
		} else {
			b, _ = dv.TextBytes()
		}
		if val.Kind() == reflect.String {
			val.SetString(string(b))
		} else {
			// byte slice, as guaranteed by isTypeMatch
			val.SetBytes(b)
		}
	case schema.Type_Which_data:
		p, err := s.Ptr(uint16(f.Slot().Offset()))
		if err != nil {
			return err
		}
		var b []byte
		if p.IsValid() {
			b = p.Data()
		} else {
			b, _ = dv.Data()
		}
		val.SetBytes(b)
	case schema.Type_Which_structType:
		p, err := s.Ptr(uint16(f.Slot().Offset()))
		if err != nil {
			return err
		}
		ss := p.Struct()
		if !ss.IsValid() {
			p, _ = dv.StructValuePtr()
			ss = p.Struct()
		}
		return e.extractStruct(val, typ.StructType().TypeId(), ss)
	case schema.Type_Which_list:
		p, err := s.Ptr(uint16(f.Slot().Offset()))
		if err != nil {
			return err
		}
		l := p.List()
		if !l.IsValid() {
			p, _ = dv.ListPtr()
			l = p.List()
		}
		return e.extractList(val, typ, l)
	default:
		return fmt.Errorf("unknown field type %v", typ.Which())
	}
	return nil
}

func (e *extracter) extractList(val reflect.Value, typ schema.Type, l capnp.List) error {
	vt := val.Type()
	elem, err := typ.List().ElementType()
	if err != nil {
		return err
	}
	if !isTypeMatch(vt, typ) {
		// TODO(light): the error won't be that useful for nested lists.
		return fmt.Errorf("can't extract %v list into a Go %v", elem.Which(), vt)
	}
	if !l.IsValid() {
		val.Set(reflect.Zero(vt))
		return nil
	}
	n := l.Len()
	val.Set(reflect.MakeSlice(vt, n, n))
	switch elem.Which() {
	case schema.Type_Which_bool:
		for i := 0; i < n; i++ {
			val.Index(i).SetBool(capnp.BitList{List: l}.At(i))
		}
	case schema.Type_Which_int8:
		for i := 0; i < n; i++ {
			val.Index(i).SetInt(int64(capnp.Int8List{List: l}.At(i)))
		}
	case schema.Type_Which_int16:
		for i := 0; i < n; i++ {
			val.Index(i).SetInt(int64(capnp.Int16List{List: l}.At(i)))
		}
	case schema.Type_Which_int32:
		for i := 0; i < n; i++ {
			val.Index(i).SetInt(int64(capnp.Int32List{List: l}.At(i)))
		}
	case schema.Type_Which_int64:
		for i := 0; i < n; i++ {
			val.Index(i).SetInt(capnp.Int64List{List: l}.At(i))
		}
	case schema.Type_Which_uint8:
		for i := 0; i < n; i++ {
			val.Index(i).SetUint(uint64(capnp.UInt8List{List: l}.At(i)))
		}
	case schema.Type_Which_uint16, schema.Type_Which_enum:
		for i := 0; i < n; i++ {
			val.Index(i).SetUint(uint64(capnp.UInt16List{List: l}.At(i)))
		}
	case schema.Type_Which_uint32:
		for i := 0; i < n; i++ {
			val.Index(i).SetUint(uint64(capnp.UInt32List{List: l}.At(i)))
		}
	case schema.Type_Which_uint64:
		for i := 0; i < n; i++ {
			val.Index(i).SetUint(capnp.UInt64List{List: l}.At(i))
		}
	case schema.Type_Which_float32:
		for i := 0; i < n; i++ {
			val.Index(i).SetFloat(float64(capnp.Float32List{List: l}.At(i)))
		}
	case schema.Type_Which_float64:
		for i := 0; i < n; i++ {
			val.Index(i).SetFloat(capnp.Float64List{List: l}.At(i))
		}
	case schema.Type_Which_text:
		if val.Type().Elem().Kind() == reflect.String {
			for i := 0; i < n; i++ {
				s, err := capnp.TextList{List: l}.At(i)
				if err != nil {
					// TODO(light): collect errors and finish
					return err
				}
				val.Index(i).SetString(s)
			}
		} else {
			for i := 0; i < n; i++ {
				b, err := capnp.TextList{List: l}.BytesAt(i)
				if err != nil {
					// TODO(light): collect errors and finish
					return err
				}
				val.Index(i).SetBytes(b)
			}
		}
	case schema.Type_Which_data:
		for i := 0; i < n; i++ {
			b, err := capnp.DataList{List: l}.At(i)
			if err != nil {
				// TODO(light): collect errors and finish
				return err
			}
			val.Index(i).SetBytes(b)
		}
	case schema.Type_Which_list:
		for i := 0; i < n; i++ {
			p, err := capnp.PointerList{List: l}.PtrAt(i)
			// TODO(light): collect errors and finish
			if err != nil {
				return err
			}
			if err := e.extractList(val.Index(i), elem, p.List()); err != nil {
				return err
			}
		}
	case schema.Type_Which_structType:
		if val.Type().Elem().Kind() == reflect.Struct {
			for i := 0; i < n; i++ {
				err := e.extractStruct(val.Index(i), elem.StructType().TypeId(), l.Struct(i))
				if err != nil {
					return err
				}
			}
		} else {
			for i := 0; i < n; i++ {
				newval := reflect.New(val.Type().Elem().Elem())
				val.Index(i).Set(newval)
				err := e.extractStruct(newval, elem.StructType().TypeId(), l.Struct(i))
				if err != nil {
					return err
				}
			}
		}
	default:
		return fmt.Errorf("unknown list type %v", elem.Which())
	}
	return nil
}

var typeMap = map[schema.Type_Which]reflect.Kind{
	schema.Type_Which_bool:    reflect.Bool,
	schema.Type_Which_int8:    reflect.Int8,
	schema.Type_Which_int16:   reflect.Int16,
	schema.Type_Which_int32:   reflect.Int32,
	schema.Type_Which_int64:   reflect.Int64,
	schema.Type_Which_uint8:   reflect.Uint8,
	schema.Type_Which_uint16:  reflect.Uint16,
	schema.Type_Which_uint32:  reflect.Uint32,
	schema.Type_Which_uint64:  reflect.Uint64,
	schema.Type_Which_float32: reflect.Float32,
	schema.Type_Which_float64: reflect.Float64,
	schema.Type_Which_enum:    reflect.Uint16,
}

func isTypeMatch(r reflect.Type, s schema.Type) bool {
	switch s.Which() {
	case schema.Type_Which_text:
		return r.Kind() == reflect.String || r.Kind() == reflect.Slice && r.Elem().Kind() == reflect.Uint8
	case schema.Type_Which_data:
		return r.Kind() == reflect.Slice && r.Elem().Kind() == reflect.Uint8
	case schema.Type_Which_structType:
		return isStructOrStructPtr(r)
	case schema.Type_Which_list:
		e, _ := s.List().ElementType()
		return r.Kind() == reflect.Slice && isTypeMatch(r.Elem(), e)
	}
	k, ok := typeMap[s.Which()]
	return ok && k == r.Kind()
}

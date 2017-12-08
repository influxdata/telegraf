package pogs

import (
	"fmt"
	"math"
	"reflect"

	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/internal/nodemap"
	"zombiezen.com/go/capnproto2/std/capnp/schema"
)

// Insert copies val, a pointer to a Go struct, into s.
func Insert(typeID uint64, s capnp.Struct, val interface{}) error {
	ins := new(inserter)
	err := ins.insertStruct(typeID, s, reflect.ValueOf(val))
	if err != nil {
		return fmt.Errorf("pogs: insert @%#x: %v", typeID, err)
	}
	return nil
}

type inserter struct {
	nodes nodemap.Map
}

func (ins *inserter) insertStruct(typeID uint64, s capnp.Struct, val reflect.Value) error {
	if val.Kind() == reflect.Ptr {
		// TODO(light): ignore if nil?
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("can't insert %v into a struct", val.Kind())
	}
	n, err := ins.nodes.Find(typeID)
	if err != nil {
		return err
	}
	if !n.IsValid() || n.Which() != schema.Node_Which_structNode {
		return fmt.Errorf("cannot find struct type %#x", typeID)
	}
	props, err := mapStruct(val.Type(), n)
	if err != nil {
		return fmt.Errorf("can't insert into %v: %v", val.Type(), err)
	}
	var discriminant uint16
	hasWhich := false
	if hasDiscriminant(n) {
		discriminant, hasWhich = props.which(val)
		if hasWhich {
			off := capnp.DataOffset(n.StructNode().DiscriminantOffset() * 2)
			if s.Size().DataSize < capnp.Size(off+2) {
				return fmt.Errorf("can't set discriminant for %s: allocated struct is too small", shortDisplayName(n))
			}
			s.SetUint16(off, discriminant)
		}
	}
	fields, err := n.StructNode().Fields()
	if err != nil {
		return err
	}
	for i := 0; i < fields.Len(); i++ {
		f := fields.At(i)
		vf := props.fieldByOrdinal(val, i)
		if !vf.IsValid() {
			// Don't have a field for this.
			continue
		}
		if dv := f.DiscriminantValue(); dv != schema.Field_noDiscriminant {
			if !hasWhich {
				sname, _ := f.NameBytes()
				return fmt.Errorf("can't insert %s from %v: has union field %s but no Which field", shortDisplayName(n), val.Type(), sname)
			}
			if dv != discriminant {
				continue
			}
		}
		switch f.Which() {
		case schema.Field_Which_slot:
			if err := ins.insertField(s, f, vf); err != nil {
				return err
			}
		case schema.Field_Which_group:
			if err := ins.insertStruct(f.Group().TypeId(), s, vf); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ins *inserter) insertField(s capnp.Struct, f schema.Field, val reflect.Value) error {
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
		return fmt.Errorf("insert field %s: default value is a %v, want %v", name, dv.Which(), typ.Which())
	}
	if !isTypeMatch(val.Type(), typ) {
		name, _ := f.NameBytes()
		return fmt.Errorf("can't insert field %s of type Go %v into a %v", name, val.Type(), typ.Which())
	}
	if !isFieldInBounds(s.Size(), f.Slot().Offset(), typ) {
		name, _ := f.NameBytes()
		return fmt.Errorf("can't insert field %s: allocated struct is too small", name)
	}
	switch typ.Which() {
	case schema.Type_Which_bool:
		v := val.Bool()
		d := dv.Bool()
		s.SetBit(capnp.BitOffset(f.Slot().Offset()), v != d) // != acts as XOR
	case schema.Type_Which_int8:
		v := int8(val.Int())
		d := dv.Int8()
		s.SetUint8(capnp.DataOffset(f.Slot().Offset()), uint8(v^d))
	case schema.Type_Which_int16:
		v := int16(val.Int())
		d := dv.Int16()
		s.SetUint16(capnp.DataOffset(f.Slot().Offset()*2), uint16(v^d))
	case schema.Type_Which_int32:
		v := int32(val.Int())
		d := dv.Int32()
		s.SetUint32(capnp.DataOffset(f.Slot().Offset()*4), uint32(v^d))
	case schema.Type_Which_int64:
		v := val.Int()
		d := dv.Int64()
		s.SetUint64(capnp.DataOffset(f.Slot().Offset()*8), uint64(v^d))
	case schema.Type_Which_uint8:
		v := uint8(val.Uint())
		d := dv.Uint8()
		s.SetUint8(capnp.DataOffset(f.Slot().Offset()), v^d)
	case schema.Type_Which_uint16:
		v := uint16(val.Uint())
		d := dv.Uint16()
		s.SetUint16(capnp.DataOffset(f.Slot().Offset()*2), v^d)
	case schema.Type_Which_enum:
		v := uint16(val.Uint())
		d := dv.Enum()
		s.SetUint16(capnp.DataOffset(f.Slot().Offset()*2), v^d)
	case schema.Type_Which_uint32:
		v := uint32(val.Uint())
		d := dv.Uint32()
		s.SetUint32(capnp.DataOffset(f.Slot().Offset()*4), v^d)
	case schema.Type_Which_uint64:
		v := val.Uint()
		d := dv.Uint64()
		s.SetUint64(capnp.DataOffset(f.Slot().Offset()*8), v^d)
	case schema.Type_Which_float32:
		v := math.Float32bits(float32(val.Float()))
		d := math.Float32bits(dv.Float32())
		s.SetUint32(capnp.DataOffset(f.Slot().Offset()*4), v^d)
	case schema.Type_Which_float64:
		v := math.Float64bits(val.Float())
		d := uint64(math.Float64bits(dv.Float64()))
		s.SetUint64(capnp.DataOffset(f.Slot().Offset()*8), v^d)
	case schema.Type_Which_text:
		// TODO(light): don't set if nil or empty. Need to consult default value.
		off := uint16(f.Slot().Offset())
		data, err := capnp.NewUInt8List(s.Segment(), int32(val.Len())+1)
		if err != nil {
			return err
		}
		b := data.ToPtr().TextBytes()
		if val.Kind() == reflect.String {
			copy(b, val.String())
		} else {
			copy(b, val.Bytes())
		}
		return s.SetPtr(off, data.ToPtr())
	case schema.Type_Which_data:
		// TODO(light): don't set if nil or empty. Need to consult default value.
		b := val.Bytes()
		off := uint16(f.Slot().Offset())
		data, err := capnp.NewData(s.Segment(), b)
		if err != nil {
			return err
		}
		return s.SetPtr(off, data.ToPtr())
	case schema.Type_Which_structType:
		off := uint16(f.Slot().Offset())
		sval := val
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return s.SetPtr(off, capnp.Ptr{})
			}
			sval = val.Elem()
		}
		id := typ.StructType().TypeId()
		sz, err := ins.structSize(id)
		if err != nil {
			return err
		}
		ss, err := capnp.NewStruct(s.Segment(), sz)
		if err != nil {
			return err
		}
		if err := s.SetPtr(off, ss.ToPtr()); err != nil {
			return err
		}
		return ins.insertStruct(id, ss, sval)
	case schema.Type_Which_list:
		off := uint16(f.Slot().Offset())
		if val.IsNil() {
			return s.SetPtr(off, capnp.Ptr{})
		}
		elem, err := typ.List().ElementType()
		if err != nil {
			return err
		}
		l, err := ins.newList(s.Segment(), elem, int32(val.Len()))
		if err != nil {
			return err
		}
		if err := s.SetPtr(off, l.ToPtr()); err != nil {
			return err
		}
		return ins.insertList(l, typ, val)
	default:
		return fmt.Errorf("unknown field type %v", typ.Which())
	}
	return nil
}

func (ins *inserter) insertList(l capnp.List, typ schema.Type, val reflect.Value) error {
	elem, err := typ.List().ElementType()
	if err != nil {
		return err
	}
	if !isTypeMatch(val.Type(), typ) {
		// TODO(light): the error won't be that useful for nested lists.
		return fmt.Errorf("can't insert Go %v into a %v list", val.Type(), elem.Which())
	}
	n := val.Len()
	switch elem.Which() {
	case schema.Type_Which_void:
	case schema.Type_Which_bool:
		for i := 0; i < n; i++ {
			capnp.BitList{List: l}.Set(i, val.Index(i).Bool())
		}
	case schema.Type_Which_int8:
		for i := 0; i < n; i++ {
			capnp.Int8List{List: l}.Set(i, int8(val.Index(i).Int()))
		}
	case schema.Type_Which_int16:
		for i := 0; i < n; i++ {
			capnp.Int16List{List: l}.Set(i, int16(val.Index(i).Int()))
		}
	case schema.Type_Which_int32:
		for i := 0; i < n; i++ {
			capnp.Int32List{List: l}.Set(i, int32(val.Index(i).Int()))
		}
	case schema.Type_Which_int64:
		for i := 0; i < n; i++ {
			capnp.Int64List{List: l}.Set(i, val.Index(i).Int())
		}
	case schema.Type_Which_uint8:
		for i := 0; i < n; i++ {
			capnp.UInt8List{List: l}.Set(i, uint8(val.Index(i).Uint()))
		}
	case schema.Type_Which_uint16, schema.Type_Which_enum:
		for i := 0; i < n; i++ {
			capnp.UInt16List{List: l}.Set(i, uint16(val.Index(i).Uint()))
		}
	case schema.Type_Which_uint32:
		for i := 0; i < n; i++ {
			capnp.UInt32List{List: l}.Set(i, uint32(val.Index(i).Uint()))
		}
	case schema.Type_Which_uint64:
		for i := 0; i < n; i++ {
			capnp.UInt64List{List: l}.Set(i, val.Index(i).Uint())
		}
	case schema.Type_Which_float32:
		for i := 0; i < n; i++ {
			capnp.Float32List{List: l}.Set(i, float32(val.Index(i).Float()))
		}
	case schema.Type_Which_float64:
		for i := 0; i < n; i++ {
			capnp.Float64List{List: l}.Set(i, val.Index(i).Float())
		}
	case schema.Type_Which_text:
		if val.Type().Elem().Kind() == reflect.String {
			for i := 0; i < n; i++ {
				err := capnp.TextList{List: l}.Set(i, val.Index(i).String())
				if err != nil {
					// TODO(light): collect errors and finish
					return err
				}
			}
		} else {
			for i := 0; i < n; i++ {
				t, err := capnp.NewTextFromBytes(l.Segment(), val.Index(i).Bytes())
				if err != nil {
					// TODO(light): collect errors and finish
					return err
				}
				err = capnp.PointerList{List: l}.SetPtr(i, t.ToPtr())
				if err != nil {
					// TODO(light): collect errors and finish
					return err
				}
			}
		}
	case schema.Type_Which_data:
		for i := 0; i < n; i++ {
			err := capnp.DataList{List: l}.Set(i, val.Index(i).Bytes())
			if err != nil {
				// TODO(light): collect errors and finish
				return err
			}
		}
	case schema.Type_Which_list:
		pl := capnp.PointerList{List: l}
		for i := 0; i < n; i++ {
			vi := val.Index(i)
			if vi.IsNil() {
				if err := pl.SetPtr(i, capnp.Ptr{}); err != nil {
					return err
				}
				continue
			}
			ee, err := elem.List().ElementType()
			if err != nil {
				return err
			}
			li, err := ins.newList(l.Segment(), ee, int32(vi.Len()))
			if err != nil {
				return err
			}
			if err := pl.SetPtr(i, li.ToPtr()); err != nil {
				return err
			}
			if err := ins.insertList(li, elem, vi); err != nil {
				return err
			}
		}
	case schema.Type_Which_structType:
		id := elem.StructType().TypeId()
		for i := 0; i < n; i++ {
			err := ins.insertStruct(id, l.Struct(i), val.Index(i))
			if err != nil {
				// TODO(light): collect errors and finish
				return err
			}
		}
	default:
		return fmt.Errorf("unknown list type %v", elem.Which())
	}
	return nil
}

func (ins *inserter) newList(s *capnp.Segment, t schema.Type, len int32) (capnp.List, error) {
	switch t.Which() {
	case schema.Type_Which_void:
		l := capnp.NewVoidList(s, len)
		return l.List, nil
	case schema.Type_Which_bool:
		l, err := capnp.NewBitList(s, len)
		return l.List, err
	case schema.Type_Which_int8, schema.Type_Which_uint8:
		l, err := capnp.NewUInt8List(s, len)
		return l.List, err
	case schema.Type_Which_int16, schema.Type_Which_uint16, schema.Type_Which_enum:
		l, err := capnp.NewUInt16List(s, len)
		return l.List, err
	case schema.Type_Which_int32, schema.Type_Which_uint32, schema.Type_Which_float32:
		l, err := capnp.NewUInt32List(s, len)
		return l.List, err
	case schema.Type_Which_int64, schema.Type_Which_uint64, schema.Type_Which_float64:
		l, err := capnp.NewUInt64List(s, len)
		return l.List, err
	case schema.Type_Which_text, schema.Type_Which_data, schema.Type_Which_list, schema.Type_Which_interface, schema.Type_Which_anyPointer:
		l, err := capnp.NewPointerList(s, len)
		return l.List, err
	case schema.Type_Which_structType:
		sz, err := ins.structSize(t.StructType().TypeId())
		if err != nil {
			return capnp.List{}, err
		}
		return capnp.NewCompositeList(s, sz, len)
	default:
		return capnp.List{}, fmt.Errorf("new list: unknown element type: %v", t.Which())
	}
}

func (ins *inserter) structSize(id uint64) (capnp.ObjectSize, error) {
	n, err := ins.nodes.Find(id)
	if err != nil {
		return capnp.ObjectSize{}, err
	}
	if n.Which() != schema.Node_Which_structNode {
		return capnp.ObjectSize{}, fmt.Errorf("insert struct: sizing: node @%#x is not a struct", id)
	}
	return capnp.ObjectSize{
		DataSize:     capnp.Size(n.StructNode().DataWordCount()) * 8,
		PointerCount: n.StructNode().PointerCount(),
	}, nil
}

func isFieldInBounds(sz capnp.ObjectSize, off uint32, t schema.Type) bool {
	switch t.Which() {
	case schema.Type_Which_void:
		return true
	case schema.Type_Which_bool:
		return sz.DataSize >= capnp.Size(off/8+1)
	case schema.Type_Which_int8, schema.Type_Which_uint8:
		return sz.DataSize >= capnp.Size(off+1)
	case schema.Type_Which_int16, schema.Type_Which_uint16, schema.Type_Which_enum:
		return sz.DataSize >= capnp.Size(off+1)*2
	case schema.Type_Which_int32, schema.Type_Which_uint32, schema.Type_Which_float32:
		return sz.DataSize >= capnp.Size(off+1)*4
	case schema.Type_Which_int64, schema.Type_Which_uint64, schema.Type_Which_float64:
		return sz.DataSize >= capnp.Size(off+1)*8
	case schema.Type_Which_text, schema.Type_Which_data, schema.Type_Which_list, schema.Type_Which_structType, schema.Type_Which_interface, schema.Type_Which_anyPointer:
		return sz.PointerCount >= uint16(off+1)
	default:
		return false
	}
}

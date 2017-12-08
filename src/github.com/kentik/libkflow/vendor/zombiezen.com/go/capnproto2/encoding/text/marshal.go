// Package text supports marshaling Cap'n Proto messages as text based on a schema.
package text

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"

	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/internal/nodemap"
	"zombiezen.com/go/capnproto2/schemas"
	"zombiezen.com/go/capnproto2/std/capnp/schema"
)

// Marker strings.
const (
	voidMarker       = "void"
	interfaceMarker  = "<external capability>"
	anyPointerMarker = "<opaque pointer>"
)

// Marshal returns the text representation of a struct.
func Marshal(typeID uint64, s capnp.Struct) (string, error) {
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode(typeID, s); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// An Encoder writes the text format of Cap'n Proto messages to an output stream.
type Encoder struct {
	w     errWriter
	tmp   []byte
	nodes nodemap.Map
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: errWriter{w: w}}
}

// UseRegistry changes the registry that the encoder consults for
// schemas from the default registry.
func (enc *Encoder) UseRegistry(reg *schemas.Registry) {
	enc.nodes.UseRegistry(reg)
}

// Encode writes the text representation of s to the stream.
func (enc *Encoder) Encode(typeID uint64, s capnp.Struct) error {
	if enc.w.err != nil {
		return enc.w.err
	}
	err := enc.marshalStruct(typeID, s)
	if err != nil {
		return err
	}
	return enc.w.err
}

func (enc *Encoder) marshalBool(v bool) {
	if v {
		enc.w.WriteString("true")
	} else {
		enc.w.WriteString("false")
	}
}

func (enc *Encoder) marshalInt(i int64) {
	enc.tmp = strconv.AppendInt(enc.tmp[:0], i, 10)
	enc.w.Write(enc.tmp)
}

func (enc *Encoder) marshalUint(i uint64) {
	enc.tmp = strconv.AppendUint(enc.tmp[:0], i, 10)
	enc.w.Write(enc.tmp)
}

func (enc *Encoder) marshalFloat32(f float32) {
	enc.tmp = strconv.AppendFloat(enc.tmp[:0], float64(f), 'g', -1, 32)
	enc.w.Write(enc.tmp)
}

func (enc *Encoder) marshalFloat64(f float64) {
	enc.tmp = strconv.AppendFloat(enc.tmp[:0], f, 'g', -1, 64)
	enc.w.Write(enc.tmp)
}

func (enc *Encoder) marshalText(t []byte) {
	enc.w.WriteByte('"')
	last := 0
	for i, b := range t {
		if !needsEscape(b) {
			continue
		}
		enc.w.Write(t[last:i])
		switch b {
		case '\a':
			enc.w.WriteString("\\a")
		case '\b':
			enc.w.WriteString("\\b")
		case '\f':
			enc.w.WriteString("\\f")
		case '\n':
			enc.w.WriteString("\\n")
		case '\r':
			enc.w.WriteString("\\r")
		case '\t':
			enc.w.WriteString("\\t")
		case '\v':
			enc.w.WriteString("\\v")
		case '\'':
			enc.w.WriteString("\\'")
		case '"':
			enc.w.WriteString("\\\"")
		case '\\':
			enc.w.WriteString("\\\\")
		default:
			enc.w.WriteString("\\x")
			enc.w.WriteByte(hexDigit(b / 16))
			enc.w.WriteByte(hexDigit(b % 16))
		}
		last = i + 1
	}
	enc.w.Write(t[last:])
	enc.w.WriteByte('"')
}

func needsEscape(b byte) bool {
	return b < 0x20 || b >= 0x7f
}

func hexDigit(b byte) byte {
	const digits = "0123456789abcdef"
	return digits[b]
}

func (enc *Encoder) marshalStruct(typeID uint64, s capnp.Struct) error {
	n, err := enc.nodes.Find(typeID)
	if err != nil {
		return err
	}
	if !n.IsValid() || n.Which() != schema.Node_Which_structNode {
		return fmt.Errorf("cannot find struct type %#x", typeID)
	}
	var discriminant uint16
	if n.StructNode().DiscriminantCount() > 0 {
		discriminant = s.Uint16(capnp.DataOffset(n.StructNode().DiscriminantOffset() * 2))
	}
	enc.w.WriteByte('(')
	fields := codeOrderFields(n.StructNode())
	first := true
	for _, f := range fields {
		if !(f.Which() == schema.Field_Which_slot || f.Which() == schema.Field_Which_group) {
			continue
		}
		if dv := f.DiscriminantValue(); !(dv == schema.Field_noDiscriminant || dv == discriminant) {
			continue
		}
		if !first {
			enc.w.WriteString(", ")
		}
		first = false
		name, err := f.NameBytes()
		if err != nil {
			return err
		}
		enc.w.Write(name)
		enc.w.WriteString(" = ")
		switch f.Which() {
		case schema.Field_Which_slot:
			if err := enc.marshalFieldValue(s, f); err != nil {
				return err
			}
		case schema.Field_Which_group:
			if err := enc.marshalStruct(f.Group().TypeId(), s); err != nil {
				return err
			}
		}
	}
	enc.w.WriteByte(')')
	return nil
}

func (enc *Encoder) marshalFieldValue(s capnp.Struct, f schema.Field) error {
	typ, err := f.Slot().Type()
	if err != nil {
		return err
	}
	dv, err := f.Slot().DefaultValue()
	if err != nil {
		return err
	}
	if dv.IsValid() && int(typ.Which()) != int(dv.Which()) {
		name, _ := f.Name()
		return fmt.Errorf("marshal field %s: default value is a %v, want %v", name, dv.Which(), typ.Which())
	}
	switch typ.Which() {
	case schema.Type_Which_void:
		enc.w.WriteString(voidMarker)
	case schema.Type_Which_bool:
		v := s.Bit(capnp.BitOffset(f.Slot().Offset()))
		d := dv.Bool()
		enc.marshalBool(!d && v || d && !v)
	case schema.Type_Which_int8:
		v := s.Uint8(capnp.DataOffset(f.Slot().Offset()))
		d := uint8(dv.Int8())
		enc.marshalInt(int64(int8(v ^ d)))
	case schema.Type_Which_int16:
		v := s.Uint16(capnp.DataOffset(f.Slot().Offset() * 2))
		d := uint16(dv.Int16())
		enc.marshalInt(int64(int16(v ^ d)))
	case schema.Type_Which_int32:
		v := s.Uint32(capnp.DataOffset(f.Slot().Offset() * 4))
		d := uint32(dv.Int32())
		enc.marshalInt(int64(int32(v ^ d)))
	case schema.Type_Which_int64:
		v := s.Uint64(capnp.DataOffset(f.Slot().Offset() * 8))
		d := uint64(dv.Int64())
		enc.marshalInt(int64(v ^ d))
	case schema.Type_Which_uint8:
		v := s.Uint8(capnp.DataOffset(f.Slot().Offset()))
		d := dv.Uint8()
		enc.marshalUint(uint64(v ^ d))
	case schema.Type_Which_uint16:
		v := s.Uint16(capnp.DataOffset(f.Slot().Offset() * 2))
		d := dv.Uint16()
		enc.marshalUint(uint64(v ^ d))
	case schema.Type_Which_uint32:
		v := s.Uint32(capnp.DataOffset(f.Slot().Offset() * 4))
		d := dv.Uint32()
		enc.marshalUint(uint64(v ^ d))
	case schema.Type_Which_uint64:
		v := s.Uint64(capnp.DataOffset(f.Slot().Offset() * 8))
		d := dv.Uint64()
		enc.marshalUint(v ^ d)
	case schema.Type_Which_float32:
		v := s.Uint32(capnp.DataOffset(f.Slot().Offset() * 4))
		d := math.Float32bits(dv.Float32())
		enc.marshalFloat32(math.Float32frombits(v ^ d))
	case schema.Type_Which_float64:
		v := s.Uint64(capnp.DataOffset(f.Slot().Offset() * 8))
		d := math.Float64bits(dv.Float64())
		enc.marshalFloat64(math.Float64frombits(v ^ d))
	case schema.Type_Which_structType:
		p, err := s.Ptr(uint16(f.Slot().Offset()))
		if err != nil {
			return err
		}
		if !p.IsValid() {
			p, _ = dv.StructValuePtr()
		}
		return enc.marshalStruct(typ.StructType().TypeId(), p.Struct())
	case schema.Type_Which_data:
		p, err := s.Ptr(uint16(f.Slot().Offset()))
		if err != nil {
			return err
		}
		if !p.IsValid() {
			b, _ := dv.Data()
			enc.marshalText(b)
			return nil
		}
		enc.marshalText(p.Data())
	case schema.Type_Which_text:
		p, err := s.Ptr(uint16(f.Slot().Offset()))
		if err != nil {
			return err
		}
		if !p.IsValid() {
			b, _ := dv.TextBytes()
			enc.marshalText(b)
			return nil
		}
		enc.marshalText(p.TextBytes())
	case schema.Type_Which_list:
		elem, err := typ.List().ElementType()
		if err != nil {
			return err
		}
		p, err := s.Ptr(uint16(f.Slot().Offset()))
		if err != nil {
			return err
		}
		if !p.IsValid() {
			p, _ = dv.ListPtr()
		}
		return enc.marshalList(elem, p.List())
	case schema.Type_Which_enum:
		v := s.Uint16(capnp.DataOffset(f.Slot().Offset() * 2))
		d := dv.Uint16()
		return enc.marshalEnum(typ.Enum().TypeId(), v^d)
	case schema.Type_Which_interface:
		enc.w.WriteString(interfaceMarker)
	case schema.Type_Which_anyPointer:
		enc.w.WriteString(anyPointerMarker)
	default:
		return fmt.Errorf("unknown field type %v", typ.Which())
	}
	return nil
}

func codeOrderFields(s schema.Node_structNode) []schema.Field {
	list, _ := s.Fields()
	n := list.Len()
	fields := make([]schema.Field, n)
	for i := 0; i < n; i++ {
		f := list.At(i)
		fields[f.CodeOrder()] = f
	}
	return fields
}

func (enc *Encoder) marshalList(elem schema.Type, l capnp.List) error {
	enc.w.WriteByte('[')
	switch elem.Which() {
	case schema.Type_Which_void:
		for i := 0; i < l.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.w.WriteString(voidMarker)
		}
	case schema.Type_Which_bool:
		bl := capnp.BitList{List: l}
		for i := 0; i < bl.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.marshalBool(bl.At(i))
		}
	case schema.Type_Which_int8:
		il := capnp.Int8List{List: l}
		for i := 0; i < il.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.marshalInt(int64(il.At(i)))
		}
	case schema.Type_Which_int16:
		il := capnp.Int16List{List: l}
		for i := 0; i < il.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.marshalInt(int64(il.At(i)))
		}
	case schema.Type_Which_int32:
		il := capnp.Int32List{List: l}
		for i := 0; i < il.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.marshalInt(int64(il.At(i)))
		}
	case schema.Type_Which_int64:
		il := capnp.Int64List{List: l}
		for i := 0; i < il.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.marshalInt(il.At(i))
		}
	case schema.Type_Which_uint8:
		il := capnp.UInt8List{List: l}
		for i := 0; i < il.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.marshalUint(uint64(il.At(i)))
		}
	case schema.Type_Which_uint16:
		il := capnp.UInt16List{List: l}
		for i := 0; i < il.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.marshalUint(uint64(il.At(i)))
		}
	case schema.Type_Which_uint32:
		il := capnp.UInt32List{List: l}
		for i := 0; i < il.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.marshalUint(uint64(il.At(i)))
		}
	case schema.Type_Which_uint64:
		il := capnp.UInt64List{List: l}
		for i := 0; i < il.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.marshalUint(il.At(i))
		}
	case schema.Type_Which_float32:
		fl := capnp.Float32List{List: l}
		for i := 0; i < fl.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.marshalFloat32(fl.At(i))
		}
	case schema.Type_Which_float64:
		fl := capnp.Float64List{List: l}
		for i := 0; i < fl.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.marshalFloat64(fl.At(i))
		}
	case schema.Type_Which_data:
		dl := capnp.DataList{List: l}
		for i := 0; i < dl.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			d, err := dl.At(i)
			if err != nil {
				return err
			}
			enc.marshalText(d)
		}
	case schema.Type_Which_text:
		tl := capnp.TextList{List: l}
		for i := 0; i < tl.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			t, err := tl.BytesAt(i)
			if err != nil {
				return err
			}
			enc.marshalText(t)
		}
	case schema.Type_Which_structType:
		for i := 0; i < l.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			err := enc.marshalStruct(elem.StructType().TypeId(), l.Struct(i))
			if err != nil {
				return err
			}
		}
	case schema.Type_Which_list:
		ee, err := elem.List().ElementType()
		if err != nil {
			return err
		}
		for i := 0; i < l.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			p, err := capnp.PointerList{List: l}.PtrAt(i)
			if err != nil {
				return err
			}
			err = enc.marshalList(ee, p.List())
			if err != nil {
				return err
			}
		}
	case schema.Type_Which_enum:
		il := capnp.UInt16List{List: l}
		typ := elem.Enum().TypeId()
		// TODO(light): only search for node once
		for i := 0; i < il.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.marshalEnum(typ, il.At(i))
		}
	case schema.Type_Which_interface:
		for i := 0; i < l.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.w.WriteString(interfaceMarker)
		}
	case schema.Type_Which_anyPointer:
		for i := 0; i < l.Len(); i++ {
			if i > 0 {
				enc.w.WriteString(", ")
			}
			enc.w.WriteString(anyPointerMarker)
		}
	default:
		return fmt.Errorf("unknown list type %v", elem.Which())
	}
	enc.w.WriteByte(']')
	return nil
}

func (enc *Encoder) marshalEnum(typ uint64, val uint16) error {
	n, err := enc.nodes.Find(typ)
	if err != nil {
		return err
	}
	if n.Which() != schema.Node_Which_enum {
		return fmt.Errorf("marshaling enum of type @%#x: type is not an enum", typ)
	}
	enums, err := n.Enum().Enumerants()
	if err != nil {
		return err
	}
	if int(val) >= enums.Len() {
		enc.marshalUint(uint64(val))
		return nil
	}
	name, err := enums.At(int(val)).NameBytes()
	if err != nil {
		return err
	}
	enc.w.Write(name)
	return nil
}

type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) Write(p []byte) (int, error) {
	if ew.err != nil {
		return 0, ew.err
	}
	var n int
	n, ew.err = ew.w.Write(p)
	return n, ew.err
}

func (ew *errWriter) WriteString(s string) (int, error) {
	if ew.err != nil {
		return 0, ew.err
	}
	var n int
	n, ew.err = io.WriteString(ew.w, s)
	return n, ew.err
}

func (ew *errWriter) WriteByte(b byte) error {
	if ew.err != nil {
		return ew.err
	}
	if bw, ok := ew.w.(io.ByteWriter); ok {
		ew.err = bw.WriteByte(b)
	} else {
		_, ew.err = ew.w.Write([]byte{b})
	}
	return ew.err
}

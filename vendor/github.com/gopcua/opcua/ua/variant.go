// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package ua

import (
	"reflect"
	"time"

	"github.com/gopcua/opcua/errors"
)

var (
	// MaxVariantArrayLength sets a limit on the number of elements in array
	MaxVariantArrayLength = 0xffff
)

const (
	// VariantArrayDimensions flags whether the array has more than one dimension
	VariantArrayDimensions = 0x40

	// VariantArrayValues flags whether the value is an array.
	VariantArrayValues = 0x80
)

// Variant is a union of the built-in types.
//
// Specification: Part 6, 5.2.2.16
type Variant struct {
	// mask contains the type and the array flags
	// bits 0:5: built-in type id 1-25
	// bit 6: array dimensions
	// bit 7: array values
	mask byte

	// arrayLength is the number of elements in the array.
	// This field is only present if the 'array values'
	// flag is set.
	//
	// Multi-dimensional arrays are encoded as a one-dimensional array and this
	// field specifies the total number of elements. The original array can be
	// reconstructed from the dimensions that are encoded after the value
	// field.
	arrayLength int32

	// arrayDimensionsLength is the number of dimensions.
	// This field is only present if the 'array dimensions' flag
	// is set.
	arrayDimensionsLength int32

	// arrayDimensions is the size for each dimension.
	// This field is only present if the 'array dimensions' flag
	// is set.
	arrayDimensions []int32

	value interface{}
}

func NewVariant(v interface{}) (*Variant, error) {
	va := &Variant{}
	if err := va.set(v); err != nil {
		return nil, err
	}
	return va, nil
}

func MustVariant(v interface{}) *Variant {
	va, err := NewVariant(v)
	if err != nil {
		panic(err)
	}
	return va
}

func (m *Variant) EncodingMask() byte {
	return m.mask
}

// Type returns the type id of the value.
func (m *Variant) Type() TypeID {
	return TypeID(m.mask & 0x3f)
}

func (m *Variant) setType(t TypeID) {
	m.mask |= byte(t & 0x3f)
}

// Has returns whether given encoding mask bits are set.
func (m *Variant) Has(mask byte) bool {
	return m.mask&mask == mask
}

// ArrayLength returns the total number of elements for one and multi-dimensional
// array values.
func (m *Variant) ArrayLength() int32 {
	return m.arrayLength
}

// ArrayDimensions returns the dimensions of multi-dimensional arrays.
func (m *Variant) ArrayDimensions() []int32 {
	return m.arrayDimensions
}

// Value returns the value.
func (m *Variant) Value() interface{} {
	return m.value
}

// Decode implements the codec interface.
func (m *Variant) Decode(b []byte) (int, error) {
	buf := NewBuffer(b)
	m.mask = buf.ReadByte()

	// a null value specifies that no other fields are encoded
	if m.Type() == TypeIDNull {
		return buf.Pos(), buf.Error()
	}

	// check the type
	typ, ok := variantTypeIDToType[m.Type()]
	if !ok {
		return buf.Pos(), errors.Errorf("invalid type id: %d", m.Type())
	}

	// read single value and return
	if !m.Has(VariantArrayValues) {
		m.value = m.decodeValue(buf)
		return buf.Pos(), buf.Error()
	}

	// get total array length (flattened for multi-dimensional arrays)
	m.arrayLength = buf.ReadInt32()

	// read flattened array elements
	n := int(m.arrayLength)
	if n < 0 || n > MaxVariantArrayLength {
		return buf.Pos(), StatusBadEncodingLimitsExceeded
	}

	vals := reflect.MakeSlice(reflect.SliceOf(typ), n, n)
	for i := 0; i < n; i++ {
		vals.Index(i).Set(reflect.ValueOf(m.decodeValue(buf)))
	}

	// check for dimensions of multi-dimensional array
	if m.Has(VariantArrayDimensions) {
		m.arrayDimensionsLength = buf.ReadInt32()
		if m.arrayDimensionsLength < 0 {
			return buf.Pos(), StatusBadEncodingLimitsExceeded
		}
		m.arrayDimensions = make([]int32, m.arrayDimensionsLength)
		for i := 0; i < int(m.arrayDimensionsLength); i++ {
			m.arrayDimensions[i] = buf.ReadInt32()
			if m.arrayDimensions[i] < 1 {
				return buf.Pos(), StatusBadEncodingLimitsExceeded
			}
		}
	}

	// return early if there is an error since the rest of the code
	// depends on the assumption that the array dimensions were read
	// correctly.
	if buf.Error() != nil {
		return buf.Pos(), buf.Error()
	}

	// validate that the total number of elements
	// matches the product of the array dimensions
	if m.arrayDimensionsLength > 0 {
		count := int32(1)
		for i := range m.arrayDimensions {
			count *= m.arrayDimensions[i]
		}
		if count != m.arrayLength {
			return buf.Pos(), errUnbalancedSlice
		}
	}

	// handle one-dimensional arrays
	if m.arrayDimensionsLength < 2 {
		m.value = vals.Interface()
		return buf.Pos(), buf.Error()
	}

	// handle multi-dimensional arrays
	// convert dimensions to []int to avoid lots of type casts
	dims := make([]int, len(m.arrayDimensions))
	for i := range m.arrayDimensions {
		dims[i] = int(m.arrayDimensions[i])
	}
	m.value = split(0, 0, vals.Len(), dims, vals).Interface()

	return buf.Pos(), buf.Error()
}

// split recursively creates a multi-dimensional array from a set of values
// and some given dimensions.
func split(level, i, j int, dims []int, vals reflect.Value) reflect.Value {
	if level == len(dims)-1 {
		a := vals.Slice(i, j)
		// fmt.Printf("split: level:%d i:%d j:%d dims:%v a:%#v\n", level, i, j, dims, a.Interface())
		return a
	}

	// split next level
	var elems []reflect.Value
	if vals.Len() > 0 {
		step := (j - i) / dims[level]
		for ; i < j; i += step {
			elems = append(elems, split(level+1, i, i+step, dims, vals))
		}
	} else {
		for k := 0; k < dims[level]; k++ {
			elems = append(elems, split(level+1, 0, 0, dims, vals))
		}
	}

	// now construct the typed slice, i.e. [](type of inner slice)
	innerT := elems[0].Type()
	a := reflect.MakeSlice(reflect.SliceOf(innerT), len(elems), len(elems))
	for k := range elems {
		a.Index(k).Set(elems[k])
	}
	// fmt.Printf("split: level:%d i:%d j:%d dims:%v a:%#v\n", level, i, j, dims, a.Interface())
	return a
}

// decodeValue reads a single value of the base type from the buffer.
func (m *Variant) decodeValue(buf *Buffer) interface{} {
	switch m.Type() {
	case TypeIDBoolean:
		return buf.ReadBool()
	case TypeIDSByte:
		return buf.ReadInt8()
	case TypeIDByte:
		return buf.ReadByte()
	case TypeIDInt16:
		return buf.ReadInt16()
	case TypeIDUint16:
		return buf.ReadUint16()
	case TypeIDInt32:
		return buf.ReadInt32()
	case TypeIDUint32:
		return buf.ReadUint32()
	case TypeIDInt64:
		return buf.ReadInt64()
	case TypeIDUint64:
		return buf.ReadUint64()
	case TypeIDFloat:
		return buf.ReadFloat32()
	case TypeIDDouble:
		return buf.ReadFloat64()
	case TypeIDString:
		return buf.ReadString()
	case TypeIDDateTime:
		return buf.ReadTime()
	case TypeIDGUID:
		v := new(GUID)
		buf.ReadStruct(v)
		return v
	case TypeIDByteString:
		return buf.ReadBytes()
	case TypeIDXMLElement:
		return XMLElement(buf.ReadString())
	case TypeIDNodeID:
		v := new(NodeID)
		buf.ReadStruct(v)
		return v
	case TypeIDExpandedNodeID:
		v := new(ExpandedNodeID)
		buf.ReadStruct(v)
		return v
	case TypeIDStatusCode:
		return StatusCode(buf.ReadUint32())
	case TypeIDQualifiedName:
		v := new(QualifiedName)
		buf.ReadStruct(v)
		return v
	case TypeIDLocalizedText:
		v := new(LocalizedText)
		buf.ReadStruct(v)
		return v
	case TypeIDExtensionObject:
		v := new(ExtensionObject)
		buf.ReadStruct(v)
		return v
	case TypeIDDataValue:
		v := new(DataValue)
		buf.ReadStruct(v)
		return v
	case TypeIDVariant:
		// todo(fs): limit recursion depth to 100
		v := new(Variant)
		buf.ReadStruct(v)
		return v
	case TypeIDDiagnosticInfo:
		// todo(fs): limit recursion depth to 100
		v := new(DiagnosticInfo)
		buf.ReadStruct(v)
		return v
	default:
		return nil
	}
}

// Encode implements the codec interface.
func (m *Variant) Encode() ([]byte, error) {
	buf := NewBuffer(nil)
	buf.WriteByte(m.mask)

	// a null value specifies that no other fields are encoded
	if m.Type() == TypeIDNull {
		return buf.Bytes(), buf.Error()
	}

	if m.Has(VariantArrayValues) {
		buf.WriteInt32(m.arrayLength)
	}

	m.encode(buf, reflect.ValueOf(m.value))

	if m.Has(VariantArrayDimensions) {
		buf.WriteInt32(m.arrayDimensionsLength)
		for i := 0; i < int(m.arrayDimensionsLength); i++ {
			buf.WriteInt32(m.arrayDimensions[i])
		}
	}

	return buf.Bytes(), buf.Error()
}

// encode recursively writes the values to the buffer.
func (m *Variant) encode(buf *Buffer, val reflect.Value) {
	if val.Kind() != reflect.Slice || m.Type() == TypeIDByteString {
		m.encodeValue(buf, val.Interface())
		return
	}
	for i := 0; i < val.Len(); i++ {
		m.encode(buf, val.Index(i))
	}
}

// encodeValue writes a single value of the base type to the buffer.
func (m *Variant) encodeValue(buf *Buffer, v interface{}) {
	switch x := v.(type) {
	case bool:
		buf.WriteBool(x)
	case int8:
		buf.WriteInt8(x)
	case byte:
		buf.WriteByte(x)
	case int16:
		buf.WriteInt16(x)
	case uint16:
		buf.WriteUint16(x)
	case int32:
		buf.WriteInt32(x)
	case uint32:
		buf.WriteUint32(x)
	case int64:
		buf.WriteInt64(x)
	case uint64:
		buf.WriteUint64(x)
	case float32:
		buf.WriteFloat32(x)
	case float64:
		buf.WriteFloat64(x)
	case string:
		buf.WriteString(x)
	case time.Time:
		buf.WriteTime(x)
	case *GUID:
		buf.WriteStruct(x)
	case []byte:
		buf.WriteByteString(x)
	case XMLElement:
		buf.WriteString(string(x))
	case *NodeID:
		buf.WriteStruct(x)
	case *ExpandedNodeID:
		buf.WriteStruct(x)
	case StatusCode:
		buf.WriteUint32(uint32(x))
	case *QualifiedName:
		buf.WriteStruct(x)
	case *LocalizedText:
		buf.WriteStruct(x)
	case *ExtensionObject:
		buf.WriteStruct(x)
	case *DataValue:
		buf.WriteStruct(x)
	case *Variant:
		buf.WriteStruct(x)
	case *DiagnosticInfo:
		buf.WriteStruct(x)
	}
}

// errUnbalancedSlice indicates a multi-dimensional array has different
// number of elements on the same level.
var errUnbalancedSlice = errors.New("unbalanced multi-dimensional array")

// sliceDim determines the element type, dimensions and the total length
// of a one or multi-dimensional slice.
func sliceDim(v reflect.Value) (typ reflect.Type, dim []int32, count int32, err error) {
	// null type
	if v.Kind() == reflect.Invalid {
		return nil, nil, 0, nil
	}

	// ByteString is its own type
	if v.Type() == reflect.TypeOf([]byte{}) {
		return v.Type(), nil, 1, nil
	}

	// element type
	if v.Kind() != reflect.Slice {
		return v.Type(), nil, 1, nil
	}

	// empty array
	if v.Len() == 0 {
		return v.Type().Elem(), append([]int32{0}, dim...), 0, nil
	}

	// check that inner slices all have the same length
	if v.Index(0).Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			if v.Index(i).Len() != v.Index(0).Len() {
				return nil, nil, 0, errUnbalancedSlice
			}
		}
	}

	// recurse to inner slice or element type
	typ, dim, count, err = sliceDim(v.Index(0))
	if err != nil {
		return nil, nil, 0, err
	}
	return typ, append([]int32{int32(v.Len())}, dim...), count * int32(v.Len()), nil
}

// set sets the value and updates the flags according to the type.
func (m *Variant) set(v interface{}) error {
	// set array length and dimensions if value is a slice
	et, dim, count, err := sliceDim(reflect.ValueOf(v))
	if err != nil {
		return err
	}

	if len(dim) > 0 {
		m.mask |= VariantArrayValues
		m.arrayLength = count
	}

	if len(dim) > 1 {
		m.mask |= VariantArrayDimensions
		m.arrayDimensionsLength = int32(len(dim))
		m.arrayDimensions = dim
	}

	typeid, ok := variantTypeToTypeID[et]
	if !ok {
		return errors.Errorf("cannot set variant to %T", v)
	}
	m.setType(typeid)
	m.value = v
	return nil
}

// todo(fs): this should probably be StringValue or we need to handle all types
// todo(fs): and recursion
func (m *Variant) String() string {
	if m.ArrayLength() > 0 {
		return ""
	}

	switch m.Type() {
	case TypeIDString:
		return m.value.(string)
	case TypeIDXMLElement:
		return string(m.XMLElement())
	case TypeIDLocalizedText:
		return m.value.(*LocalizedText).Text
	case TypeIDQualifiedName:
		return m.value.(*QualifiedName).Name
	default:
		return ""
	}
}

// Bool returns the boolean value if the type is Boolean.
func (m *Variant) Bool() bool {
	if m.ArrayLength() > 0 {
		return false
	}

	switch m.Type() {
	case TypeIDBoolean:
		return m.value.(bool)
	default:
		return false
	}
}

// Float returns the float value if the type is one of the float types.
func (m *Variant) Float() float64 {
	if m.ArrayLength() > 0 {
		return 0
	}

	switch m.Type() {
	case TypeIDFloat:
		return float64(m.value.(float32))
	case TypeIDDouble:
		return m.value.(float64)
	default:
		return 0
	}
}

// Int returns the int value if the type is one of the int types.
func (m *Variant) Int() int64 {
	if m.ArrayLength() > 0 {
		return 0
	}

	switch m.Type() {
	case TypeIDSByte:
		return int64(m.value.(int8))
	case TypeIDInt16:
		return int64(m.value.(int16))
	case TypeIDInt32:
		return int64(m.value.(int32))
	case TypeIDInt64:
		return m.value.(int64)
	default:
		return 0
	}
}

// Uint returns the uint value if the type is one of the uint types.
func (m *Variant) Uint() uint64 {
	if m.ArrayLength() > 0 {
		return 0
	}

	switch m.Type() {
	case TypeIDByte:
		return uint64(m.value.(byte))
	case TypeIDUint16:
		return uint64(m.value.(uint16))
	case TypeIDUint32:
		return uint64(m.value.(uint32))
	case TypeIDUint64:
		return m.value.(uint64)
	default:
		return 0
	}
}

func (m *Variant) ByteString() []byte {
	if m.ArrayLength() > 0 {
		return nil
	}

	switch m.Type() {
	case TypeIDByteString:
		return m.value.([]byte)
	default:
		return nil
	}
}

func (m *Variant) DataValue() *DataValue {
	if m.ArrayLength() > 0 {
		return nil
	}

	switch m.Type() {
	case TypeIDDataValue:
		return m.value.(*DataValue)
	default:
		return nil
	}
}

func (m *Variant) DiagnosticInfo() *DiagnosticInfo {
	if m.ArrayLength() > 0 {
		return nil
	}

	switch m.Type() {
	case TypeIDDiagnosticInfo:
		return m.value.(*DiagnosticInfo)
	default:
		return nil
	}
}

func (m *Variant) ExpandedNodeID() *ExpandedNodeID {
	if m.ArrayLength() > 0 {
		return nil
	}

	switch m.Type() {
	case TypeIDExpandedNodeID:
		return m.value.(*ExpandedNodeID)
	default:
		return nil
	}
}

func (m *Variant) ExtensionObject() *ExtensionObject {
	if m.ArrayLength() > 0 {
		return nil
	}

	switch m.Type() {
	case TypeIDExtensionObject:
		return m.value.(*ExtensionObject)
	default:
		return nil
	}
}

func (m *Variant) GUID() *GUID {
	if m.ArrayLength() > 0 {
		return nil
	}

	switch m.Type() {
	case TypeIDGUID:
		return m.value.(*GUID)
	default:
		return nil
	}
}

func (m *Variant) LocalizedText() *LocalizedText {
	if m.ArrayLength() > 0 {
		return nil
	}

	switch m.Type() {
	case TypeIDLocalizedText:
		return m.value.(*LocalizedText)
	default:
		return nil
	}
}

func (m *Variant) NodeID() *NodeID {
	if m.ArrayLength() > 0 {
		return nil
	}

	switch m.Type() {
	case TypeIDNodeID:
		return m.value.(*NodeID)
	default:
		return nil
	}
}

func (m *Variant) QualifiedName() *QualifiedName {
	if m.ArrayLength() > 0 {
		return nil
	}

	switch m.Type() {
	case TypeIDQualifiedName:
		return m.value.(*QualifiedName)
	default:
		return nil
	}
}

func (m *Variant) StatusCode() StatusCode {
	if m.ArrayLength() > 0 {
		return StatusBadTypeMismatch
	}

	switch m.Type() {
	case TypeIDStatusCode:
		return m.value.(StatusCode)
	default:
		return StatusBadTypeMismatch
	}
}

// Time returns the time value if the type is DateTime.
func (m *Variant) Time() time.Time {
	if m.ArrayLength() > 0 {
		return time.Time{}
	}

	switch m.Type() {
	case TypeIDDateTime:
		return m.value.(time.Time)
	default:
		return time.Time{}
	}
}

func (m *Variant) Variant() *Variant {
	if m.ArrayLength() > 0 {
		return nil
	}
	switch m.Type() {
	case TypeIDVariant:
		return m.value.(*Variant)
	default:
		return nil
	}
}

func (m *Variant) XMLElement() XMLElement {
	if m.ArrayLength() > 0 {
		return ""
	}

	switch m.Type() {
	case TypeIDXMLElement:
		return m.value.(XMLElement)
	default:
		return ""
	}
}

var variantTypeToTypeID = map[reflect.Type]TypeID{}
var variantTypeIDToType = map[TypeID]reflect.Type{
	TypeIDNull:            reflect.TypeOf(nil),
	TypeIDBoolean:         reflect.TypeOf(false),
	TypeIDSByte:           reflect.TypeOf(int8(0)),
	TypeIDByte:            reflect.TypeOf(uint8(0)),
	TypeIDInt16:           reflect.TypeOf(int16(0)),
	TypeIDUint16:          reflect.TypeOf(uint16(0)),
	TypeIDInt32:           reflect.TypeOf(int32(0)),
	TypeIDUint32:          reflect.TypeOf(uint32(0)),
	TypeIDInt64:           reflect.TypeOf(int64(0)),
	TypeIDUint64:          reflect.TypeOf(uint64(0)),
	TypeIDFloat:           reflect.TypeOf(float32(0)),
	TypeIDDouble:          reflect.TypeOf(float64(0)),
	TypeIDString:          reflect.TypeOf(string("")),
	TypeIDDateTime:        reflect.TypeOf(time.Time{}),
	TypeIDGUID:            reflect.TypeOf(new(GUID)),
	TypeIDByteString:      reflect.TypeOf([]byte{}),
	TypeIDXMLElement:      reflect.TypeOf(XMLElement("")),
	TypeIDNodeID:          reflect.TypeOf(new(NodeID)),
	TypeIDExpandedNodeID:  reflect.TypeOf(new(ExpandedNodeID)),
	TypeIDStatusCode:      reflect.TypeOf(StatusCode(0)),
	TypeIDQualifiedName:   reflect.TypeOf(new(QualifiedName)),
	TypeIDLocalizedText:   reflect.TypeOf(new(LocalizedText)),
	TypeIDExtensionObject: reflect.TypeOf(new(ExtensionObject)),
	TypeIDDataValue:       reflect.TypeOf(new(DataValue)),
	TypeIDVariant:         reflect.TypeOf(new(Variant)),
	TypeIDDiagnosticInfo:  reflect.TypeOf(new(DiagnosticInfo)),
}

func init() {
	for id, t := range variantTypeIDToType {
		variantTypeToTypeID[t] = id
	}
}

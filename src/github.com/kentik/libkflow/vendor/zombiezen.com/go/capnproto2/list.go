package capnp

import (
	"errors"
	"math"
)

// A List is a reference to an array of values.
type List struct {
	seg        *Segment
	off        Address
	length     int32
	size       ObjectSize
	depthLimit uint
	flags      listFlags
}

// newPrimitiveList allocates a new list of primitive values, preferring placement in s.
func newPrimitiveList(s *Segment, sz Size, n int32) (List, error) {
	total, ok := sz.times(n)
	if !ok {
		return List{}, errOverflow
	}
	s, addr, err := alloc(s, total)
	if err != nil {
		return List{}, err
	}
	return List{
		seg:        s,
		off:        addr,
		length:     n,
		size:       ObjectSize{DataSize: sz},
		depthLimit: maxDepth,
	}, nil
}

// NewCompositeList creates a new composite list, preferring placement
// in s.
func NewCompositeList(s *Segment, sz ObjectSize, n int32) (List, error) {
	if !sz.isValid() {
		return List{}, errObjectSize
	}
	sz.DataSize = sz.DataSize.padToWord()
	total, ok := sz.totalSize().times(n)
	if !ok || total > maxSize-wordSize {
		return List{}, errOverflow
	}
	s, addr, err := alloc(s, wordSize+total)
	if err != nil {
		return List{}, err
	}
	// Add tag word
	s.writeRawPointer(addr, rawStructPointer(pointerOffset(n), sz))
	return List{
		seg:        s,
		off:        addr + Address(wordSize),
		length:     n,
		size:       sz,
		flags:      isCompositeList,
		depthLimit: maxDepth,
	}, nil
}

// ToList is deprecated in favor of Ptr.List.
func ToList(p Pointer) List {
	return toPtr(p).List()
}

// ToListDefault is deprecated in favor of Ptr.ListDefault.
func ToListDefault(p Pointer, def []byte) (List, error) {
	return toPtr(p).ListDefault(def)
}

// ToPtr converts the list to a generic pointer.
func (p List) ToPtr() Ptr {
	return Ptr{
		seg:        p.seg,
		off:        p.off,
		lenOrCap:   uint32(p.length),
		size:       p.size,
		depthLimit: p.depthLimit,
		flags:      listPtrFlag(p.flags),
	}
}

// Segment returns the segment this pointer references.
func (p List) Segment() *Segment {
	return p.seg
}

// IsValid returns whether the list is valid.
func (p List) IsValid() bool {
	return p.seg != nil
}

// HasData reports whether the list's total size is non-zero.
func (p List) HasData() bool {
	sz, ok := p.size.totalSize().times(p.length)
	if !ok {
		return false
	}
	return sz > 0
}

// readSize returns the list's size for the purposes of read limit
// accounting.
func (p List) readSize() Size {
	if p.seg == nil {
		return 0
	}
	e := p.size.totalSize()
	if e == 0 {
		e = wordSize
	}
	sz, ok := e.times(p.length)
	if !ok {
		return maxSize
	}
	return sz
}

// value returns the equivalent raw list pointer.
func (p List) value(paddr Address) rawPointer {
	if p.seg == nil {
		return 0
	}
	off := makePointerOffset(paddr, p.off)
	if p.flags&isCompositeList != 0 {
		// p.off points to the data not the header
		return rawListPointer(off-1, compositeList, p.length*p.size.totalWordCount())
	}
	if p.flags&isBitList != 0 {
		return rawListPointer(off, bit1List, p.length)
	}
	if p.size.PointerCount == 1 && p.size.DataSize == 0 {
		return rawListPointer(off, pointerList, p.length)
	}
	if p.size.PointerCount != 0 {
		panic(errListSize)
	}
	switch p.size.DataSize {
	case 0:
		return rawListPointer(off, voidList, p.length)
	case 1:
		return rawListPointer(off, byte1List, p.length)
	case 2:
		return rawListPointer(off, byte2List, p.length)
	case 4:
		return rawListPointer(off, byte4List, p.length)
	case 8:
		return rawListPointer(off, byte8List, p.length)
	default:
		panic(errListSize)
	}
}

func (p List) underlying() Pointer {
	return p
}

// Address returns the address the pointer references.
func (p List) Address() Address {
	return p.off
}

// Len returns the length of the list.
func (p List) Len() int {
	if p.seg == nil {
		return 0
	}
	return int(p.length)
}

// elem returns the slice of segment data for a list element.
func (p List) elem(i int) (addr Address, sz Size) {
	if p.seg == nil || i < 0 || i >= int(p.length) {
		panic(errOutOfBounds)
	}
	if p.flags&isBitList != 0 {
		addr = p.off.addOffset(BitOffset(i).offset())
		return addr, 1
	}
	sz = p.size.totalSize()
	addr, _ = p.off.element(int32(i), sz)
	return addr, sz
}

func (p List) slice(i int) []byte {
	addr, sz := p.elem(i)
	return p.seg.slice(addr, sz)
}

// Struct returns the i'th element as a struct.
func (p List) Struct(i int) Struct {
	if p.flags&isBitList != 0 {
		return Struct{}
	}
	addr, _ := p.elem(i)
	return Struct{
		seg:        p.seg,
		off:        addr,
		size:       p.size,
		flags:      isListMember,
		depthLimit: p.depthLimit - 1,
	}
}

// SetStruct set the i'th element to the value in s.
func (p List) SetStruct(i int, s Struct) error {
	if p.flags&isBitList != 0 {
		return errBitListStruct
	}
	return copyStruct(copyContext{}, p.Struct(i), s)
}

// A BitList is a reference to a list of booleans.
type BitList struct{ List }

// NewBitList creates a new bit list, preferring placement in s.
func NewBitList(s *Segment, n int32) (BitList, error) {
	s, addr, err := alloc(s, Size(int64(n+7)/8))
	if err != nil {
		return BitList{}, err
	}
	return BitList{List{
		seg:        s,
		off:        addr,
		length:     n,
		flags:      isBitList,
		depthLimit: maxDepth,
	}}, nil
}

// At returns the i'th bit.
func (p BitList) At(i int) bool {
	b := p.slice(i)
	if b == nil {
		return false
	}
	bit := BitOffset(i)
	return b[0]&bit.mask() != 0
}

// Set sets the i'th bit to v.
func (p BitList) Set(i int, v bool) {
	b := p.slice(i)
	if b == nil {
		panic(errOutOfBounds)
	}
	bit := BitOffset(i)
	if v {
		b[0] |= bit.mask()
	} else {
		b[0] &^= bit.mask()
	}
}

// A PointerList is a reference to an array of pointers.
type PointerList struct{ List }

// NewPointerList allocates a new list of pointers, preferring placement in s.
func NewPointerList(s *Segment, n int32) (PointerList, error) {
	total, ok := wordSize.times(n)
	if !ok {
		return PointerList{}, errOverflow
	}
	s, addr, err := alloc(s, total)
	if err != nil {
		return PointerList{}, err
	}
	return PointerList{List{
		seg:        s,
		off:        addr,
		length:     n,
		size:       ObjectSize{PointerCount: 1},
		depthLimit: maxDepth,
	}}, nil
}

// At is deprecated in favor of PtrAt.
func (p PointerList) At(i int) (Pointer, error) {
	pi, err := p.PtrAt(i)
	return pi.toPointer(), err
}

// PtrAt returns the i'th pointer in the list.
func (p PointerList) PtrAt(i int) (Ptr, error) {
	addr, _ := p.elem(i)
	return p.seg.readPtr(addr, p.depthLimit)
}

// Set is deprecated in favor of SetPtr.
func (p PointerList) Set(i int, v Pointer) error {
	return p.SetPtr(i, toPtr(v))
}

// SetPtr sets the i'th pointer in the list to v.
func (p PointerList) SetPtr(i int, v Ptr) error {
	addr, _ := p.elem(i)
	return p.seg.writePtr(copyContext{}, addr, v)
}

// TextList is an array of pointers to strings.
type TextList struct{ List }

// NewTextList allocates a new list of text pointers, preferring placement in s.
func NewTextList(s *Segment, n int32) (TextList, error) {
	pl, err := NewPointerList(s, n)
	if err != nil {
		return TextList{}, err
	}
	return TextList{pl.List}, nil
}

// At returns the i'th string in the list.
func (l TextList) At(i int) (string, error) {
	addr, _ := l.elem(i)
	p, err := l.seg.readPtr(addr, l.depthLimit)
	if err != nil {
		return "", err
	}
	return p.Text(), nil
}

// BytesAt returns the i'th element in the list as a byte slice.
// The underlying array of the slice is the segment data.
func (l TextList) BytesAt(i int) ([]byte, error) {
	addr, _ := l.elem(i)
	p, err := l.seg.readPtr(addr, l.depthLimit)
	if err != nil {
		return nil, err
	}
	return p.TextBytes(), nil
}

// Set sets the i'th string in the list to v.
func (l TextList) Set(i int, v string) error {
	addr, _ := l.elem(i)
	p, err := NewText(l.seg, v)
	if err != nil {
		return err
	}
	return p.seg.writePtr(copyContext{}, addr, p.List.ToPtr())
}

// DataList is an array of pointers to data.
type DataList struct{ List }

// NewDataList allocates a new list of data pointers, preferring placement in s.
func NewDataList(s *Segment, n int32) (DataList, error) {
	pl, err := NewPointerList(s, n)
	if err != nil {
		return DataList{}, err
	}
	return DataList{pl.List}, nil
}

// At returns the i'th data in the list.
func (l DataList) At(i int) ([]byte, error) {
	addr, _ := l.elem(i)
	p, err := l.seg.readPtr(addr, l.depthLimit)
	if err != nil {
		return nil, err
	}
	return p.Data(), nil
}

// Set sets the i'th data in the list to v.
func (l DataList) Set(i int, v []byte) error {
	addr, _ := l.elem(i)
	p, err := NewData(l.seg, v)
	if err != nil {
		return err
	}
	return p.seg.writePtr(copyContext{}, addr, p.List.ToPtr())
}

// A VoidList is a list of zero-sized elements.
type VoidList struct{ List }

// NewVoidList creates a list of voids.  No allocation is performed;
// s is only used for Segment()'s return value.
func NewVoidList(s *Segment, n int32) VoidList {
	return VoidList{List{
		seg:        s,
		length:     n,
		depthLimit: maxDepth,
	}}
}

// A UInt8List is an array of UInt8 values.
type UInt8List struct{ List }

// NewUInt8List creates a new list of UInt8, preferring placement in s.
func NewUInt8List(s *Segment, n int32) (UInt8List, error) {
	l, err := newPrimitiveList(s, 1, n)
	if err != nil {
		return UInt8List{}, err
	}
	return UInt8List{l}, nil
}

// NewText creates a new list of UInt8 from a string.
func NewText(s *Segment, v string) (UInt8List, error) {
	// TODO(light): error if v is too long
	l, err := NewUInt8List(s, int32(len(v)+1))
	if err != nil {
		return UInt8List{}, err
	}
	copy(l.seg.slice(l.off, Size(len(v))), v)
	return l, nil
}

// NewTextFromBytes creates a NUL-terminated list of UInt8 from a byte slice.
func NewTextFromBytes(s *Segment, v []byte) (UInt8List, error) {
	// TODO(light): error if v is too long
	l, err := NewUInt8List(s, int32(len(v)+1))
	if err != nil {
		return UInt8List{}, err
	}
	copy(l.seg.slice(l.off, Size(len(v))), v)
	return l, nil
}

// NewData creates a new list of UInt8 from a byte slice.
func NewData(s *Segment, v []byte) (UInt8List, error) {
	// TODO(light): error if v is too long
	l, err := NewUInt8List(s, int32(len(v)))
	if err != nil {
		return UInt8List{}, err
	}
	copy(l.seg.slice(l.off, Size(len(v))), v)
	return l, nil
}

// ToText is deprecated in favor of Ptr.Text.
func ToText(p Pointer) string {
	return toPtr(p).TextDefault("")
}

// ToTextDefault is deprecated in favor of Ptr.TextDefault.
func ToTextDefault(p Pointer, def string) string {
	return toPtr(p).TextDefault(def)
}

// ToData is deprecated in favor of Ptr.Data.
func ToData(p Pointer) []byte {
	return toPtr(p).DataDefault(nil)
}

// ToDataDefault is deprecated in favor of Ptr.DataDefault.
func ToDataDefault(p Pointer, def []byte) []byte {
	return toPtr(p).DataDefault(def)
}

func isOneByteList(p Ptr) bool {
	return p.seg != nil && p.flags.ptrType() == listPtrType && p.size.isOneByte() && p.flags.listFlags()&isCompositeList == 0
}

// At returns the i'th element.
func (l UInt8List) At(i int) uint8 {
	b := l.slice(i)
	if b == nil {
		panic(errOutOfBounds)
	}
	return b[0]
}

// Set sets the i'th element to v.
func (l UInt8List) Set(i int, v uint8) {
	b := l.slice(i)
	if b == nil {
		panic(errOutOfBounds)
	}
	b[0] = v
}

// Int8List is an array of Int8 values.
type Int8List struct{ List }

// NewInt8List creates a new list of Int8, preferring placement in s.
func NewInt8List(s *Segment, n int32) (Int8List, error) {
	l, err := newPrimitiveList(s, 1, n)
	if err != nil {
		return Int8List{}, err
	}
	return Int8List{l}, nil
}

// At returns the i'th element.
func (l Int8List) At(i int) int8 {
	b := l.slice(i)
	if b == nil {
		panic(errOutOfBounds)
	}
	return int8(b[0])
}

// Set sets the i'th element to v.
func (l Int8List) Set(i int, v int8) {
	b := l.slice(i)
	if b == nil {
		panic(errOutOfBounds)
	}
	b[0] = uint8(v)
}

// A UInt16List is an array of UInt16 values.
type UInt16List struct{ List }

// NewUInt16List creates a new list of UInt16, preferring placement in s.
func NewUInt16List(s *Segment, n int32) (UInt16List, error) {
	l, err := newPrimitiveList(s, 2, n)
	if err != nil {
		return UInt16List{}, err
	}
	return UInt16List{l}, nil
}

// At returns the i'th element.
func (l UInt16List) At(i int) uint16 {
	addr, _ := l.elem(i)
	return l.seg.readUint16(addr)
}

// Set sets the i'th element to v.
func (l UInt16List) Set(i int, v uint16) {
	addr, _ := l.elem(i)
	l.seg.writeUint16(addr, v)
}

// Int16List is an array of Int16 values.
type Int16List struct{ List }

// NewInt16List creates a new list of Int16, preferring placement in s.
func NewInt16List(s *Segment, n int32) (Int16List, error) {
	l, err := newPrimitiveList(s, 2, n)
	if err != nil {
		return Int16List{}, err
	}
	return Int16List{l}, nil
}

// At returns the i'th element.
func (l Int16List) At(i int) int16 {
	addr, _ := l.elem(i)
	return int16(l.seg.readUint16(addr))
}

// Set sets the i'th element to v.
func (l Int16List) Set(i int, v int16) {
	addr, _ := l.elem(i)
	l.seg.writeUint16(addr, uint16(v))
}

// UInt32List is an array of UInt32 values.
type UInt32List struct{ List }

// NewUInt32List creates a new list of UInt32, preferring placement in s.
func NewUInt32List(s *Segment, n int32) (UInt32List, error) {
	l, err := newPrimitiveList(s, 4, n)
	if err != nil {
		return UInt32List{}, err
	}
	return UInt32List{l}, nil
}

// At returns the i'th element.
func (l UInt32List) At(i int) uint32 {
	addr, _ := l.elem(i)
	return l.seg.readUint32(addr)
}

// Set sets the i'th element to v.
func (l UInt32List) Set(i int, v uint32) {
	addr, _ := l.elem(i)
	l.seg.writeUint32(addr, v)
}

// Int32List is an array of Int32 values.
type Int32List struct{ List }

// NewInt32List creates a new list of Int32, preferring placement in s.
func NewInt32List(s *Segment, n int32) (Int32List, error) {
	l, err := newPrimitiveList(s, 4, n)
	if err != nil {
		return Int32List{}, err
	}
	return Int32List{l}, nil
}

// At returns the i'th element.
func (l Int32List) At(i int) int32 {
	addr, _ := l.elem(i)
	return int32(l.seg.readUint32(addr))
}

// Set sets the i'th element to v.
func (l Int32List) Set(i int, v int32) {
	addr, _ := l.elem(i)
	l.seg.writeUint32(addr, uint32(v))
}

// UInt64List is an array of UInt64 values.
type UInt64List struct{ List }

// NewUInt64List creates a new list of UInt64, preferring placement in s.
func NewUInt64List(s *Segment, n int32) (UInt64List, error) {
	l, err := newPrimitiveList(s, 8, n)
	if err != nil {
		return UInt64List{}, err
	}
	return UInt64List{l}, nil
}

// At returns the i'th element.
func (l UInt64List) At(i int) uint64 {
	addr, _ := l.elem(i)
	return l.seg.readUint64(addr)
}

// Set sets the i'th element to v.
func (l UInt64List) Set(i int, v uint64) {
	addr, _ := l.elem(i)
	l.seg.writeUint64(addr, v)
}

// Int64List is an array of Int64 values.
type Int64List struct{ List }

// NewInt64List creates a new list of Int64, preferring placement in s.
func NewInt64List(s *Segment, n int32) (Int64List, error) {
	l, err := newPrimitiveList(s, 8, n)
	if err != nil {
		return Int64List{}, err
	}
	return Int64List{l}, nil
}

// At returns the i'th element.
func (l Int64List) At(i int) int64 {
	addr, _ := l.elem(i)
	return int64(l.seg.readUint64(addr))
}

// Set sets the i'th element to v.
func (l Int64List) Set(i int, v int64) {
	addr, _ := l.elem(i)
	l.seg.writeUint64(addr, uint64(v))
}

// Float32List is an array of Float32 values.
type Float32List struct{ List }

// NewFloat32List creates a new list of Float32, preferring placement in s.
func NewFloat32List(s *Segment, n int32) (Float32List, error) {
	l, err := newPrimitiveList(s, 8, n)
	if err != nil {
		return Float32List{}, err
	}
	return Float32List{l}, nil
}

// At returns the i'th element.
func (l Float32List) At(i int) float32 {
	addr, _ := l.elem(i)
	return math.Float32frombits(l.seg.readUint32(addr))
}

// Set sets the i'th element to v.
func (l Float32List) Set(i int, v float32) {
	addr, _ := l.elem(i)
	l.seg.writeUint32(addr, math.Float32bits(v))
}

// Float64List is an array of Float64 values.
type Float64List struct{ List }

// NewFloat64List creates a new list of Float64, preferring placement in s.
func NewFloat64List(s *Segment, n int32) (Float64List, error) {
	l, err := newPrimitiveList(s, 8, n)
	if err != nil {
		return Float64List{}, err
	}
	return Float64List{l}, nil
}

// At returns the i'th element.
func (l Float64List) At(i int) float64 {
	addr, _ := l.elem(i)
	return math.Float64frombits(l.seg.readUint64(addr))
}

// Set sets the i'th element to v.
func (l Float64List) Set(i int, v float64) {
	addr, _ := l.elem(i)
	l.seg.writeUint64(addr, math.Float64bits(v))
}

type listFlags uint8

const (
	isCompositeList listFlags = 1 << iota
	isBitList
)

var errBitListStruct = errors.New("capnp: SetStruct called on bit list")

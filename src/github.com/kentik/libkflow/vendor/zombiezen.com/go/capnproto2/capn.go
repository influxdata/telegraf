package capnp

import (
	"encoding/binary"
	"errors"

	"github.com/glycerine/rbtree"
)

// A SegmentID is a numeric identifier for a Segment.
type SegmentID uint32

// A Segment is an allocation arena for Cap'n Proto objects.
// It is part of a Message, which can contain other segments that
// reference each other.
type Segment struct {
	msg  *Message
	id   SegmentID
	data []byte
}

// Message returns the message that contains s.
func (s *Segment) Message() *Message {
	return s.msg
}

// ID returns the segment's ID.
func (s *Segment) ID() SegmentID {
	return s.id
}

// Data returns the raw byte slice for the segment.
func (s *Segment) Data() []byte {
	return s.data
}

func (s *Segment) inBounds(addr Address) bool {
	return addr < Address(len(s.data))
}

func (s *Segment) regionInBounds(base Address, sz Size) bool {
	end, ok := base.addSize(sz)
	if !ok {
		return false
	}
	return end <= Address(len(s.data))
}

// slice returns the segment of data from base to base+sz.
func (s *Segment) slice(base Address, sz Size) []byte {
	// Bounds check should have happened before calling slice.
	return s.data[base : base+Address(sz)]
}

func (s *Segment) readUint8(addr Address) uint8 {
	return s.slice(addr, 1)[0]
}

func (s *Segment) readUint16(addr Address) uint16 {
	return binary.LittleEndian.Uint16(s.slice(addr, 2))
}

func (s *Segment) readUint32(addr Address) uint32 {
	return binary.LittleEndian.Uint32(s.slice(addr, 4))
}

func (s *Segment) readUint64(addr Address) uint64 {
	return binary.LittleEndian.Uint64(s.slice(addr, 8))
}

func (s *Segment) readRawPointer(addr Address) rawPointer {
	return rawPointer(s.readUint64(addr))
}

func (s *Segment) writeUint8(addr Address, val uint8) {
	s.slice(addr, 1)[0] = val
}

func (s *Segment) writeUint16(addr Address, val uint16) {
	binary.LittleEndian.PutUint16(s.slice(addr, 2), val)
}

func (s *Segment) writeUint32(addr Address, val uint32) {
	binary.LittleEndian.PutUint32(s.slice(addr, 4), val)
}

func (s *Segment) writeUint64(addr Address, val uint64) {
	binary.LittleEndian.PutUint64(s.slice(addr, 8), val)
}

func (s *Segment) writeRawPointer(addr Address, val rawPointer) {
	s.writeUint64(addr, uint64(val))
}

// root returns a 1-element pointer list that references the first word
// in the segment.  This only makes sense to call on the first segment
// in a message.
func (s *Segment) root() PointerList {
	sz := ObjectSize{PointerCount: 1}
	if !s.regionInBounds(0, sz.totalSize()) {
		return PointerList{}
	}
	return PointerList{List{
		seg:        s,
		length:     1,
		size:       sz,
		depthLimit: s.msg.depthLimit(),
	}}
}

func (s *Segment) lookupSegment(id SegmentID) (*Segment, error) {
	if s.id == id {
		return s, nil
	}
	return s.msg.Segment(id)
}

func (s *Segment) readPtr(off Address, depthLimit uint) (ptr Ptr, err error) {
	val := s.readRawPointer(off)
	s, off, val, err = s.resolveFarPointer(off, val)
	if err != nil {
		return Ptr{}, err
	}
	if val == 0 {
		return Ptr{}, nil
	}
	if depthLimit == 0 {
		return Ptr{}, errDepthLimit
	}
	// Be wary of overflow. Offset is 30 bits signed. List size is 29 bits
	// unsigned. For both of these we need to check in terms of words if
	// using 32 bit maths as bits or bytes will overflow.
	switch val.pointerType() {
	case structPointer:
		sp, err := s.readStructPtr(off, val)
		if err != nil {
			return Ptr{}, err
		}
		if !s.msg.ReadLimiter().canRead(sp.readSize()) {
			return Ptr{}, errReadLimit
		}
		sp.depthLimit = depthLimit - 1
		return sp.ToPtr(), nil
	case listPointer:
		lp, err := s.readListPtr(off, val)
		if err != nil {
			return Ptr{}, err
		}
		if !s.msg.ReadLimiter().canRead(lp.readSize()) {
			return Ptr{}, errReadLimit
		}
		lp.depthLimit = depthLimit - 1
		return lp.ToPtr(), nil
	case otherPointer:
		if val.otherPointerType() != 0 {
			return Ptr{}, errOtherPointer
		}
		return Interface{
			seg: s,
			cap: val.capabilityIndex(),
		}.ToPtr(), nil
	default:
		// Only other types are far pointers.
		return Ptr{}, errBadLandingPad
	}
}

func (s *Segment) readStructPtr(off Address, val rawPointer) (Struct, error) {
	addr, ok := val.offset().resolve(off)
	if !ok {
		return Struct{}, errPointerAddress
	}
	sz := val.structSize()
	if !s.regionInBounds(addr, sz.totalSize()) {
		return Struct{}, errPointerAddress
	}
	return Struct{
		seg:  s,
		off:  addr,
		size: sz,
	}, nil
}

func (s *Segment) readListPtr(off Address, val rawPointer) (List, error) {
	addr, ok := val.offset().resolve(off)
	if !ok {
		return List{}, errPointerAddress
	}
	lt := val.listType()
	lsize, ok := val.totalListSize()
	if !ok {
		return List{}, errOverflow
	}
	if !s.regionInBounds(addr, lsize) {
		return List{}, errPointerAddress
	}
	limitSize := lsize
	if limitSize == 0 {

	}
	if lt == compositeList {
		hdr := s.readRawPointer(addr)
		var ok bool
		addr, ok = addr.addSize(wordSize)
		if !ok {
			return List{}, errOverflow
		}
		if hdr.pointerType() != structPointer {
			return List{}, errBadTag
		}
		sz := hdr.structSize()
		n := int32(hdr.offset())
		// TODO(light): check that this has the same end address
		if tsize, ok := sz.totalSize().times(n); !ok {
			return List{}, errOverflow
		} else if !s.regionInBounds(addr, tsize) {
			return List{}, errPointerAddress
		}
		return List{
			seg:    s,
			size:   sz,
			off:    addr,
			length: n,
			flags:  isCompositeList,
		}, nil
	}
	if lt == bit1List {
		return List{
			seg:    s,
			off:    addr,
			length: val.numListElements(),
			flags:  isBitList,
		}, nil
	}
	return List{
		seg:    s,
		size:   val.elementSize(),
		off:    addr,
		length: val.numListElements(),
	}, nil
}

func (s *Segment) resolveFarPointer(off Address, val rawPointer) (*Segment, Address, rawPointer, error) {
	switch val.pointerType() {
	case doubleFarPointer:
		// A double far pointer points to a double pointer, where the
		// first points to the actual data, and the second is the tag
		// that would normally be placed right before the data (offset
		// == 0).

		faroff, segid := val.farAddress(), val.farSegment()
		s, err := s.lookupSegment(segid)
		if err != nil {
			return nil, 0, 0, err
		}
		if !s.regionInBounds(faroff, wordSize*2) {
			return nil, 0, 0, errPointerAddress
		}
		far := s.readRawPointer(faroff)
		tagStart, ok := faroff.addSize(wordSize)
		if !ok {
			return nil, 0, 0, errOverflow
		}
		tag := s.readRawPointer(tagStart)
		if far.pointerType() != farPointer || tag.offset() != 0 {
			return nil, 0, 0, errPointerAddress
		}
		segid = far.farSegment()
		if s, err = s.lookupSegment(segid); err != nil {
			return nil, 0, 0, errBadLandingPad
		}
		return s, 0, landingPadNearPointer(far, tag), nil
	case farPointer:
		faroff, segid := val.farAddress(), val.farSegment()
		s, err := s.lookupSegment(segid)
		if err != nil {
			return nil, 0, 0, err
		}
		if !s.regionInBounds(faroff, wordSize) {
			return nil, 0, 0, errPointerAddress
		}
		val = s.readRawPointer(faroff)
		return s, faroff, val, nil
	default:
		return s, off, val, nil
	}
}

type offset struct {
	id         SegmentID
	boff, bend int64 // in bits
	newval     Ptr
}

func makeOffsetKey(p Ptr) offset {
	// Since this is used for copying, the address boundaries should already be clamped.
	switch p.flags.ptrType() {
	case structPtrType:
		s := p.Struct()
		return offset{
			id:   s.seg.id,
			boff: int64(s.off) * 8,
			bend: (int64(s.off) + int64(s.size.totalSize())) * 8,
		}
	case listPtrType:
		l := p.List()
		key := offset{
			id:   l.seg.id,
			boff: int64(l.off) * 8,
		}
		if l.flags&isBitList != 0 {
			key.bend = int64(l.off)*8 + int64(l.length)
		} else {
			key.bend = (int64(l.off) + int64(l.size.totalSize())*int64(l.length)) * 8
		}
		if l.flags&isCompositeList != 0 {
			// Composite lists' offsets are after the tag word.
			key.boff -= int64(wordSize) * 8
		}
		return key
	default:
		panic("unreachable")
	}
}

func compare(a, b rbtree.Item) int {
	ao := a.(offset)
	bo := b.(offset)
	switch {
	case ao.id != bo.id:
		return int(ao.id) - int(bo.id)
	case ao.boff > bo.boff:
		return 1
	case ao.boff < bo.boff:
		return -1
	default:
		return 0
	}
}

func needsCopy(dest *Segment, src Ptr) bool {
	if src.seg.msg != dest.msg {
		return true
	}
	s := src.Struct()
	if s.seg == nil {
		return false
	}
	// Structs can only be referenced if they're not list members.
	return s.flags&isListMember != 0
}

func (s *Segment) writePtr(cc copyContext, off Address, src Ptr) error {
	// handle nulls
	if !src.IsValid() {
		s.writeRawPointer(off, 0)
		return nil
	}
	srcSeg := src.Segment()

	if i := src.Interface(); i.Segment() != nil {
		if s.msg != srcSeg.msg {
			c := s.msg.AddCap(i.Client())
			i = NewInterface(s, c)
		}
		s.writeRawPointer(off, i.value(off))
		return nil
	}
	if s != srcSeg {
		// Different segments
		if needsCopy(s, src) {
			return copyPointer(cc, s, off, src)
		}
		if !hasCapacity(srcSeg.data, wordSize) {
			// Double far pointer needed.
			const landingSize = wordSize * 2
			t, dstAddr, err := alloc(s, landingSize)
			if err != nil {
				return err
			}

			srcAddr := src.address()
			t.writeRawPointer(dstAddr, rawFarPointer(srcSeg.id, srcAddr))
			// alloc guarantees that two words are available.
			t.writeRawPointer(dstAddr+Address(wordSize), src.value(srcAddr-Address(wordSize)))
			s.writeRawPointer(off, rawDoubleFarPointer(t.id, dstAddr))
			return nil
		}
		// Have room in the target for a tag
		_, srcAddr, _ := alloc(srcSeg, wordSize)
		srcSeg.writeRawPointer(srcAddr, src.value(srcAddr))
		s.writeRawPointer(off, rawFarPointer(srcSeg.id, srcAddr))
		return nil
	}
	s.writeRawPointer(off, src.value(off))
	return nil
}

func copyPointer(cc copyContext, dstSeg *Segment, dstAddr Address, src Ptr) error {
	if cc.depth >= 32 {
		return errCopyDepth
	}
	cc = cc.init()
	// First, see if the ptr has already been copied.
	key := makeOffsetKey(src)
	iter := cc.copies.FindLE(key)
	if key.bend > key.boff {
		if !iter.NegativeLimit() {
			other := iter.Item().(offset)
			if key.id == other.id {
				if key.boff == other.boff && key.bend == other.bend {
					return dstSeg.writePtr(cc.incDepth(), dstAddr, other.newval)
				} else if other.bend >= key.bend {
					return errOverlap
				}
			}
		}

		iter = iter.Next()

		if !iter.Limit() {
			other := iter.Item().(offset)
			if key.id == other.id && other.boff < key.bend {
				return errOverlap
			}
		}
	}

	// No copy nor overlap found, so we need to clone the target
	newSeg, newAddr, err := alloc(dstSeg, Size((key.bend-key.boff)/8))
	if err != nil {
		return err
	}
	switch src.flags.ptrType() {
	case structPtrType:
		s := src.Struct()
		dst := Struct{
			seg:        newSeg,
			off:        newAddr,
			size:       s.size,
			depthLimit: maxDepth,
			// clear flags
		}
		key.newval = dst.ToPtr()
		cc.copies.Insert(key)
		if err := copyStruct(cc, dst, s); err != nil {
			return err
		}
	case listPtrType:
		l := src.List()
		dst := List{
			seg:        newSeg,
			off:        newAddr,
			length:     l.length,
			size:       l.size,
			flags:      l.flags,
			depthLimit: maxDepth,
		}
		if dst.flags&isCompositeList != 0 {
			// Copy tag word
			newSeg.writeRawPointer(newAddr, l.seg.readRawPointer(l.off-Address(wordSize)))
			var ok bool
			dst.off, ok = dst.off.addSize(wordSize)
			if !ok {
				return errOverflow
			}
		}
		key.newval = dst.ToPtr()
		cc.copies.Insert(key)
		// TODO(light): fast path for copying text/data
		if dst.flags&isBitList != 0 {
			copy(newSeg.data[newAddr:], l.seg.data[l.off:l.length+7/8])
		} else {
			for i := 0; i < l.Len(); i++ {
				err := copyStruct(cc, dst.Struct(i), l.Struct(i))
				if err != nil {
					return err
				}
			}
		}
	default:
		panic("unreachable")
	}
	return dstSeg.writePtr(cc.incDepth(), dstAddr, key.newval)
}

type copyContext struct {
	copies *rbtree.Tree
	depth  int
}

func (cc copyContext) init() copyContext {
	if cc.copies == nil {
		return copyContext{
			copies: rbtree.NewTree(compare),
		}
	}
	return cc
}

func (cc copyContext) incDepth() copyContext {
	return copyContext{
		copies: cc.copies,
		depth:  cc.depth + 1,
	}
}

var (
	errPointerAddress = errors.New("capnp: invalid pointer address")
	errBadLandingPad  = errors.New("capnp: invalid far pointer landing pad")
	errBadTag         = errors.New("capnp: invalid tag word")
	errOtherPointer   = errors.New("capnp: unknown pointer type")
	errObjectSize     = errors.New("capnp: invalid object size")
	errReadLimit      = errors.New("capnp: read traversal limit reached")
	errDepthLimit     = errors.New("capnp: depth limit reached")
)

var (
	errOverflow    = errors.New("capnp: address or size overflow")
	errOutOfBounds = errors.New("capnp: address out of bounds")
	errCopyDepth   = errors.New("capnp: copy depth too large")
	errOverlap     = errors.New("capnp: overlapping data on copy")
	errListSize    = errors.New("capnp: invalid list size")
)

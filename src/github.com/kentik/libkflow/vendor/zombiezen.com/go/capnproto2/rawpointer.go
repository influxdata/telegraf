package capnp

// pointerOffset is an address offset in multiples of word size.
type pointerOffset int32

// resolve returns the absolute address, given that the pointer is located at paddr.
func (off pointerOffset) resolve(paddr Address) (addr Address, ok bool) {
	// TODO(light): verify
	if off < 0 && Address(-off) > paddr {
		return 0, false
	}
	return paddr + Address(off*8) + 8, true
}

// makePointerOffset computes the offset for a pointer at paddr to point to addr.
func makePointerOffset(paddr, addr Address) pointerOffset {
	// TODO(light): verify
	return pointerOffset(addr/Address(wordSize) - paddr/Address(wordSize) - 1)
}

// rawPointer is an encoded pointer.
type rawPointer uint64

// rawStructPointer returns a struct pointer.  The offset is from the
// end of the pointer to the start of the struct.
func rawStructPointer(off pointerOffset, sz ObjectSize) rawPointer {
	return structPointer | orable30BitOffsetPart(off) | rawPointer(sz.dataWordCount())<<32 | rawPointer(sz.PointerCount)<<48
}

// rawListPointer returns a list pointer.  The offset is the number of
// words relative to the end of the pointer that the list starts.  If
// listType is compositeList, then length is the number of words
// that the list occupies, otherwise it is the number of elements in
// the list.
func rawListPointer(off pointerOffset, listType int, length int32) rawPointer {
	return listPointer | orable30BitOffsetPart(off) | rawPointer(listType)<<32 | rawPointer(length)<<35
}

// rawInterfacePointer returns an interface pointer that references
// a capability number.
func rawInterfacePointer(capability CapabilityID) rawPointer {
	return otherPointer | rawPointer(capability)<<32
}

// rawFarPointer returns a pointer to a pointer in another segment.
func rawFarPointer(segID SegmentID, off Address) rawPointer {
	return farPointer | rawPointer(off&^7) | (rawPointer(segID) << 32)
}

// rawDoubleFarPointer returns a pointer to a pointer in another segment.
func rawDoubleFarPointer(segID SegmentID, off Address) rawPointer {
	return doubleFarPointer | rawPointer(off&^7) | (rawPointer(segID) << 32)
}

// landingPadNearPointer converts a far pointer landing pad into
// a near pointer in the destination segment.  Its offset will be
// relative to the beginning of the segment.
func landingPadNearPointer(far, tag rawPointer) rawPointer {
	return tag | rawPointer(far.farAddress()-Address(wordSize))<<2
}

// Raw pointer types.
const (
	structPointer    = 0
	listPointer      = 1
	farPointer       = 2
	doubleFarPointer = 6
	otherPointer     = 3
)

// Raw list pointer types.
const (
	voidList      = 0
	bit1List      = 1
	byte1List     = 2
	byte2List     = 3
	byte4List     = 4
	byte8List     = 5
	pointerList   = 6
	compositeList = 7
)

func (p rawPointer) pointerType() int {
	t := p & 3
	if t == farPointer {
		return int(p & 7)
	}
	return int(t)
}

func (p rawPointer) structSize() ObjectSize {
	c := uint16(p >> 32)
	d := uint16(p >> 48)
	return ObjectSize{
		DataSize:     Size(c) * wordSize,
		PointerCount: d,
	}
}

func (p rawPointer) listType() int {
	return int((p >> 32) & 7)
}

func (p rawPointer) numListElements() int32 {
	return int32(p >> 35)
}

// elementSize returns the size of an individual element in the list referenced by p.
func (p rawPointer) elementSize() ObjectSize {
	switch p.listType() {
	case voidList:
		return ObjectSize{}
	case bit1List:
		// Size is ignored on bit lists.
		return ObjectSize{}
	case byte1List:
		return ObjectSize{DataSize: 1}
	case byte2List:
		return ObjectSize{DataSize: 2}
	case byte4List:
		return ObjectSize{DataSize: 4}
	case byte8List:
		return ObjectSize{DataSize: 8}
	case pointerList:
		return ObjectSize{PointerCount: 1}
	default:
		panic("elementSize not supposed to be called on composite or unknown list type")
	}
}

// totalListSize returns the total size of the list referenced by p.
func (p rawPointer) totalListSize() (sz Size, ok bool) {
	n := p.numListElements()
	switch p.listType() {
	case voidList:
		return 0, true
	case bit1List:
		return Size((n + 7) / 8), true
	case compositeList:
		// For a composite list, n represents the number of words (excluding the tag word).
		return wordSize.times(n + 1)
	default:
		return p.elementSize().totalSize().times(n)
	}
}

// used in orable30BitOffsetPart(), rawPointer.offset(), and rawPointer.otherPointerType()
const zerohi32 rawPointer = ^(^0 << 32)

// orable30BitOffsetPart(): get an or-able value that handles sign
// conversion. Creates part B in a struct (or list) pointer, leaving
// parts A, C, and D completely zeroed in the returned uint64.
//
// From the spec:
//
// lsb                      struct pointer                       msb
// +-+-----------------------------+---------------+---------------+
// |A|             B               |       C       |       D       |
// +-+-----------------------------+---------------+---------------+
//
// A (2 bits) = 0, to indicate that this is a struct pointer.
// B (30 bits) = Offset, in words, from the end of the pointer to the
//     start of the struct's data section.  Signed.
// C (16 bits) = Size of the struct's data section, in words.
// D (16 bits) = Size of the struct's pointer section, in words.
//
// (B is the same for list pointers, but C and D have different size
// and meaning)
//
func orable30BitOffsetPart(signedOff pointerOffset) rawPointer {
	d32 := signedOff << 2
	return rawPointer(d32) & zerohi32
}

// and convert in the other direction, extracting the count from
// the B section into an int
func (p rawPointer) offset() pointerOffset {
	u64 := p & zerohi32
	u32 := uint32(u64)
	s32 := int32(u32) >> 2
	return pointerOffset(s32)
}

// otherPointerType returns the type of "other pointer" from p.
func (p rawPointer) otherPointerType() uint32 {
	return uint32(p & zerohi32 >> 2)
}

// farAddress returns the address of the landing pad pointer.
func (p rawPointer) farAddress() Address {
	// 29-bit*8 < 32-bits, so overflow is impossible.
	return Address(p&(1<<32-1)>>3) * Address(wordSize)
}

// farSegment returns the segment ID that the far pointer references.
func (p rawPointer) farSegment() SegmentID {
	return SegmentID(p >> 32)
}

// capabilityIndex returns the index of the capability in the message's capability table.
func (p rawPointer) capabilityIndex() CapabilityID {
	return CapabilityID(p >> 32)
}

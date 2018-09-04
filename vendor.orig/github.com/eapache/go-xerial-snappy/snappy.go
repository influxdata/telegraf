package snappy

import (
	"bytes"
	"encoding/binary"
	"errors"

	master "github.com/golang/snappy"
)

const (
	sizeOffset = 16
	sizeBytes  = 4
)

var (
	xerialHeader = []byte{130, 83, 78, 65, 80, 80, 89, 0}
	// ErrMalformed is returned by the decoder when the xerial framing
	// is malformed
	ErrMalformed = errors.New("malformed xerial framing")
)

// Encode encodes data as snappy with no framing header.
func Encode(src []byte) []byte {
	return master.Encode(nil, src)
}

// Decode decodes snappy data whether it is traditional unframed
// or includes the xerial framing format.
func Decode(src []byte) ([]byte, error) {
	var max = len(src)
	if max < len(xerialHeader) {
		return nil, ErrMalformed
	}

	if !bytes.Equal(src[:8], xerialHeader) {
		return master.Decode(nil, src)
	}

	if max < sizeOffset+sizeBytes {
		return nil, ErrMalformed
	}

	var (
		pos   = sizeOffset
		dst   = make([]byte, 0, len(src))
		chunk []byte
		err   error
	)

	for pos+sizeBytes <= max {
		size := int(binary.BigEndian.Uint32(src[pos : pos+sizeBytes]))
		pos += sizeBytes

		nextPos := pos + size
		// On architectures where int is 32-bytes wide size + pos could
		// overflow so we need to check the low bound as well as the
		// high
		if nextPos < pos || nextPos > max {
			return nil, ErrMalformed
		}

		chunk, err = master.Decode(chunk, src[pos:nextPos])
		if err != nil {
			return nil, err
		}
		pos = nextPos
		dst = append(dst, chunk...)
	}
	return dst, nil
}

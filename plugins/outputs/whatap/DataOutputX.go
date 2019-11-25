package whatap

import (
	"bytes"
	"encoding/binary"
	"math"
)

const (
	INT3_MIN_VALUE  = -8388608 /*0xff800000*/
	INT3_MAX_VALUE  = 0x007fffff
	LONG5_MIN_VALUE = -549755813888 /*0xffffff8000000000*/
	LONG5_MAX_VALUE = 0x0000007fffffffff
)

type DataOutputX struct {
	buffer  bytes.Buffer
	written int
}

func NewDataOutputX() *DataOutputX {
	in := new(DataOutputX)
	return in
}

func (out *DataOutputX) ToByteArray() []byte {
	return out.buffer.Bytes()
}

func (out *DataOutputX) WriteIntBytes(b []byte) *DataOutputX {
	if b == nil || len(b) == 0 {
		out.WriteInt(0)
	} else {
		out.WriteInt(int32(len(b)))
		out.WriteBytes(b)
	}
	return out
}

func (out *DataOutputX) WriteBlob(value []byte) *DataOutputX {
	if value == nil || len(value) == 0 {
		out.WriteByte1(0)
	} else {
		sz := len(value)
		if sz <= 253 {
			out.WriteByte1(byte(sz))
			out.WriteBytes(value)
		} else if sz <= 65535 {
			buff := []byte{255, 0, 0}
			out.WriteBytes(SetBytesShort(buff, 1, int16(sz)))
			out.WriteBytes(value)
		} else {
			buff := []byte{254, 0, 0, 0, 0}
			out.WriteBytes(SetBytesInt(buff, 1, int32(sz)))
			out.WriteBytes(value)
		}
	}
	return out
}
func (out *DataOutputX) WriteDecimal(v int64) *DataOutputX {

	switch {
	case v == 0:
		out.WriteByte1(0)
	case math.MinInt8 <= v && v <= math.MaxInt8:
		b := []byte{0, 0}
		b[0] = 1
		b[1] = byte(v)
		out.WriteBytes(b)
	case math.MinInt16 <= v && v <= math.MaxInt16:
		b := []byte{0, 0, 0}
		b[0] = 2
		SetBytesShort(b, 1, int16(v))
		out.WriteBytes(b)
	case INT3_MIN_VALUE <= v && v <= INT3_MAX_VALUE:
		b := []byte{0, 0, 0, 0}
		b[0] = 3
		out.WriteBytes(SetBytesInt3(b, 1, int32(v)))
	case math.MinInt32 <= v && v <= math.MaxInt32:
		b := []byte{0, 0, 0, 0, 0}
		b[0] = 4
		out.WriteBytes(SetBytesInt(b, 1, int32(v)))
	case LONG5_MIN_VALUE <= v && v <= LONG5_MAX_VALUE:
		b := []byte{0, 0, 0, 0, 0, 0}
		b[0] = 5
		out.WriteBytes(SetBytesLong5(b, 1, v))
	case math.MinInt64 <= v && v <= math.MaxInt64:
		b := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0}
		b[0] = 8
		out.WriteBytes(SetBytesLong(b, 1, v))
	}
	return out
}
func (out *DataOutputX) WriteText(s string) *DataOutputX {
	if s == "" {
		out.WriteByte1(0)
	} else {
		out.WriteBlob([]byte(s))
	}
	return out
}

func (out *DataOutputX) WriteBytes(b []byte) *DataOutputX {
	out.written += len(b)
	out.buffer.Write(b)
	return out
}
func (out *DataOutputX) WriteBool(b bool) *DataOutputX {
	out.WriteBytes(ToBytesBool(b))
	return out
}

// go vet : method WriteByte(b byte) (*DataOutputX)
// should have signature WriteByte(byte) error
func (out *DataOutputX) WriteByte1(b byte) (*DataOutputX, error) {
	out.written++
	out.buffer.WriteByte(b)
	return out, nil
}
func (out *DataOutputX) WriteShort(b int16) *DataOutputX {
	out.WriteBytes(ToBytesShort(b))
	return out
}
func (out *DataOutputX) WriteInt(b int32) *DataOutputX {
	out.WriteBytes(ToBytesInt(b))
	return out
}
func (out *DataOutputX) WriteLong(b int64) *DataOutputX {
	out.WriteBytes(ToBytesLong(b))
	return out
}
func (out *DataOutputX) WriteFloat(b float32) *DataOutputX {
	out.WriteBytes(ToBytesFloat(b))
	return out
}
func (out *DataOutputX) WriteDouble(b float64) *DataOutputX {
	out.WriteBytes(ToBytesDouble(b))
	return out
}
func (out *DataOutputX) Size() int {
	return out.written
}

func ToBytesBool(b bool) []byte {
	if b {
		return []byte{1}
	} else {
		return []byte{0}
	}
}

func ToBytesShort(v int16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf[0:], uint16(v))
	return buf
}
func SetBytesShort(buf []byte, off int, v int16) []byte {
	binary.BigEndian.PutUint16(buf[off:], uint16(v))
	return buf
}
func ToBytesInt(v int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf[0:], uint32(v))
	return buf
}
func SetBytesInt(buf []byte, off int, v int32) []byte {
	binary.BigEndian.PutUint32(buf[off:], uint32(v))
	return buf
}
func SetBytesInt3(buf []byte, off int, v int32) []byte {
	buf[off] = byte(v >> 16)
	buf[off+1] = byte(v >> 8)
	buf[off+2] = byte(v >> 0)
	return buf
}
func ToBytesLong(v int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf[0:], uint64(v))
	return buf
}

func SetBytesLong(buf []byte, off int, v int64) []byte {
	binary.BigEndian.PutUint64(buf[off:], uint64(v))
	return buf
}

func SetBytesLong5(buf []byte, off int, v int64) []byte {
	buf[off] = byte(v >> 32)
	buf[off+1] = byte(v >> 24)
	buf[off+2] = byte(v >> 16)
	buf[off+3] = byte(v >> 8)
	buf[off+4] = byte(v >> 0)
	return buf
}
func ToBytesFloat(v float32) []byte {
	return ToBytesInt(int32(math.Float32bits(v)))
}

func SetBytesFloat(buf []byte, off int, v float32) []byte {
	return SetBytesInt(buf, off, int32(math.Float32bits(v)))
}
func ToBytesDouble(v float64) []byte {
	return ToBytesLong(int64(math.Float64bits(v)))
}

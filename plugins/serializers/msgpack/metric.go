package msgpack

import (
	"encoding/binary"
	"time"

	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp

// Metric is structure to define MessagePack message format
// will be used by msgp code generator
type Metric struct {
	Name   string                 `msg:"name"`
	Time   MessagePackTime        `msg:"time,extension"`
	Tags   map[string]string      `msg:"tags"`
	Fields map[string]interface{} `msg:"fields"`
}

// MessagePackTime implements the official timestamp extension type
// https://github.com/msgpack/msgpack/blob/master/spec.md#timestamp-extension-type
//
// tinylib/msgp has been using their own custom extension type and the official extension
// is not available. (https://github.com/tinylib/msgp/issues/214)
type MessagePackTime struct {
	time time.Time
}

func init() {
	msgp.RegisterExtension(-1, func() msgp.Extension { return new(MessagePackTime) })
}

// ExtensionType implements the Extension interface
func (*MessagePackTime) ExtensionType() int8 {
	return -1
}

// Len implements the Extension interface
// The timestamp extension uses variable length encoding depending the input
//
// 32bits: [1970-01-01 00:00:00 UTC, 2106-02-07 06:28:16 UTC) range. If the nanoseconds part is 0
// 64bits: [1970-01-01 00:00:00.000000000 UTC, 2514-05-30 01:53:04.000000000 UTC) range.
// 96bits: [-584554047284-02-23 16:59:44 UTC, 584554051223-11-09 07:00:16.000000000 UTC) range.
func (z *MessagePackTime) Len() int {
	sec := z.time.Unix()
	nsec := z.time.Nanosecond()

	if sec < 0 || sec >= (1<<34) { // 96 bits encoding
		return 12
	}
	if sec >= (1<<32) || nsec != 0 {
		return 8
	}
	return 4
}

// MarshalBinaryTo implements the Extension interface
func (z *MessagePackTime) MarshalBinaryTo(buf []byte) error {
	length := z.Len()

	if length == 4 {
		sec := z.time.Unix()
		binary.BigEndian.PutUint32(buf, uint32(sec))
	} else if length == 8 {
		sec := z.time.Unix()
		nsec := z.time.Nanosecond()

		data := uint64(nsec)<<34 | (uint64(sec) & 0x03_ffff_ffff)
		binary.BigEndian.PutUint64(buf, data)
	} else if length == 12 {
		sec := z.time.Unix()
		nsec := z.time.Nanosecond()

		binary.BigEndian.PutUint32(buf, uint32(nsec))
		binary.BigEndian.PutUint64(buf[4:], uint64(sec))
	}

	return nil
}

// UnmarshalBinary implements the Extension interface
func (z *MessagePackTime) UnmarshalBinary(buf []byte) error {
	length := len(buf)

	if length == 4 {
		sec := binary.BigEndian.Uint32(buf)
		z.time = time.Unix(int64(sec), 0)
	} else if length == 8 {
		data := binary.BigEndian.Uint64(buf)

		nsec := (data & 0xfffffffc_00000000) >> 34
		sec := data & 0x00000003_ffffffff

		z.time = time.Unix(int64(sec), int64(nsec))
	} else if length == 12 {
		nsec := binary.BigEndian.Uint32(buf)
		sec := binary.BigEndian.Uint64(buf[4:])

		z.time = time.Unix(int64(sec), int64(nsec))
	}

	return nil
}

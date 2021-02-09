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
func (t *MessagePackTime) Len() int {
	sec := t.time.Unix()
	nsec := t.time.Nanosecond()

	if sec < 0 || sec > 0x400000000 { // 96 bits encoding
		return 12
	} else if sec > 0xFFFFFFFF || nsec != 0 {
		return 8
	} else {
		return 4
	}
}

// MarshalBinaryTo implements the Extension interface
func (t *MessagePackTime) MarshalBinaryTo(buf []byte) error {
	len := t.Len()

	if len == 4 {
		sec := t.time.Unix()
		binary.BigEndian.PutUint32(buf, uint32(sec))

	} else if len == 8 {
		sec := t.time.Unix()
		nsec := t.time.Nanosecond()

		data := uint64(nsec)<<34 | (uint64(sec) & 0x03_FFFF_FFFF)
		binary.BigEndian.PutUint64(buf, data)

	} else if len == 12 {
		sec := t.time.Unix()
		nsec := t.time.Nanosecond()

		binary.BigEndian.PutUint32(buf, uint32(nsec))
		binary.BigEndian.PutUint64(buf[4:], uint64(sec))
	}

	return nil
}

// UnmarshalBinary implements the Extension interface
func (t *MessagePackTime) UnmarshalBinary(buf []byte) error {
	len := len(buf)

	if len == 4 {
		sec := binary.BigEndian.Uint32(buf)
		t.time = time.Unix(int64(sec), 0)
	} else if len == 8 {
		data := binary.BigEndian.Uint64(buf)

		nsec := (data & 0xfffffffc_00000000) >> 34
		sec := (data & 0x00000003_ffffffff)

		t.time = time.Unix(int64(sec), int64(nsec))
	} else if len == 12 {
		nsec := binary.BigEndian.Uint32(buf)
		sec := binary.BigEndian.Uint64(buf[4:])

		t.time = time.Unix(int64(sec), int64(nsec))
	}

	return nil
}

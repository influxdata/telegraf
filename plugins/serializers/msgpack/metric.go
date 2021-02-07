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
// tinylib/msgp has been used their own custom extenstion type and the offcial extension
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
func (*MessagePackTime) Len() int {
	return 8
}

// MarshalBinaryTo implements the Extension interface
func (t *MessagePackTime) MarshalBinaryTo(buf []byte) error {
	sec := uint64(t.time.Unix())
	nsec := uint64(t.time.Nanosecond())

	data := nsec<<34 | sec

	binary.BigEndian.PutUint64(buf, data)

	return nil
}

// UnmarshalBinary implements the Extension interface
func (t *MessagePackTime) UnmarshalBinary(buf []byte) error {
	data := binary.BigEndian.Uint64(buf)

	nsec := (data & 0xfffffffc_00000000) >> 34
	sec := (data & 0x00000003_ffffffff)

	t.time = time.Unix(int64(sec), int64(nsec))

	return nil
}

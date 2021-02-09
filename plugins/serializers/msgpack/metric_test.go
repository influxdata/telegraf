package msgpack

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMsgPackTime32(t *testing.T) {
	// Maximum of 4 bytes encodable time
	var sec int64 = 0xFFFFFFFF
	var nsec int64 = 0
	t1 := MessagePackTime{time: time.Unix(sec, nsec)}

	assert.Equal(t, t1.Len(), 4)

	buf := make([]byte, t1.Len())
	assert.NoError(t, t1.MarshalBinaryTo(buf))

	t2 := new(MessagePackTime)
	t2.UnmarshalBinary(buf)

	assert.Equal(t, t1.time, t2.time)
}

func TestMsgPackTime64(t *testing.T) {
	// Maximum of 8 bytes encodable time
	var sec int64 = 0x3FFFFFFFF
	var nsec int64 = 999999999
	t1 := MessagePackTime{time: time.Unix(sec, nsec)}

	assert.Equal(t, t1.Len(), 8)

	buf := make([]byte, t1.Len())
	assert.NoError(t, t1.MarshalBinaryTo(buf))

	t2 := new(MessagePackTime)
	t2.UnmarshalBinary(buf)

	assert.Equal(t, t1.time, t2.time)
}

func TestMsgPackTime96(t *testing.T) {
	// Testing 12 bytes timestamp
	var sec int64 = 0x400000001
	var nsec int64 = 111111111
	t1 := MessagePackTime{time: time.Unix(sec, nsec)}

	assert.Equal(t, t1.Len(), 12)

	buf := make([]byte, t1.Len())
	assert.NoError(t, t1.MarshalBinaryTo(buf))

	t2 := new(MessagePackTime)
	t2.UnmarshalBinary(buf)

	assert.True(t, t1.time.Equal(t2.time))

	// Testing the default value: 0001-01-01T00:00:00Z
	t1 = MessagePackTime{}

	assert.Equal(t, t1.Len(), 12)
	assert.NoError(t, t1.MarshalBinaryTo(buf))

	t2 = new(MessagePackTime)
	t2.UnmarshalBinary(buf)

	assert.True(t, t1.time.Equal(t2.time))
}

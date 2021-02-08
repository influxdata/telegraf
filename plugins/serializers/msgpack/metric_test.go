package msgpack

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMsgPackTime(t *testing.T) {
	var sec int64 = 1612703425
	var nsec int64 = 111111111
	t1 := MessagePackTime{time: time.Unix(sec, nsec)}

	buf := make([]byte, t1.Len())
	assert.NoError(t, t1.MarshalBinaryTo(buf))

	t2 := new(MessagePackTime)
	t2.UnmarshalBinary(buf)

	assert.Equal(t, t1.time, t2.time)
}

func TestMsgPackTimeOut(t *testing.T) {
	var sec int64 = 161270342500
	var nsec int64 = 111111111
	t1 := MessagePackTime{time: time.Unix(sec, nsec)}

	buf := make([]byte, t1.Len())
	err := t1.MarshalBinaryTo(buf)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "Time is out of supported range: 7080-06-15T06:21:40+09:00")
}

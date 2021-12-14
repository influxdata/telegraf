package msgpack

import (
	"encoding/hex"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMsgPackTime32(t *testing.T) {
	// Maximum of 4 bytes encodable time
	var sec int64 = 0xFFFFFFFF
	var nsec int64
	t1 := MessagePackTime{time: time.Unix(sec, nsec)}

	require.Equal(t, t1.Len(), 4)

	buf := make([]byte, t1.Len())
	require.NoError(t, t1.MarshalBinaryTo(buf))

	t2 := new(MessagePackTime)
	err := t2.UnmarshalBinary(buf)
	require.NoError(t, err)

	require.Equal(t, t1.time, t2.time)
}

func TestMsgPackTime64(t *testing.T) {
	// Maximum of 8 bytes encodable time
	var sec int64 = 0x3FFFFFFFF
	var nsec int64 = 999999999
	t1 := MessagePackTime{time: time.Unix(sec, nsec)}

	require.Equal(t, t1.Len(), 8)

	buf := make([]byte, t1.Len())
	require.NoError(t, t1.MarshalBinaryTo(buf))

	t2 := new(MessagePackTime)
	err := t2.UnmarshalBinary(buf)
	require.NoError(t, err)

	require.Equal(t, t1.time, t2.time)
}

func TestMsgPackTime96(t *testing.T) {
	// Testing 12 bytes timestamp
	var sec int64 = 0x400000001
	var nsec int64 = 111111111
	t1 := MessagePackTime{time: time.Unix(sec, nsec)}

	require.Equal(t, t1.Len(), 12)

	buf := make([]byte, t1.Len())
	require.NoError(t, t1.MarshalBinaryTo(buf))

	t2 := new(MessagePackTime)
	err := t2.UnmarshalBinary(buf)
	require.NoError(t, err)

	require.True(t, t1.time.Equal(t2.time))

	// Testing the default value: 0001-01-01T00:00:00Z
	t1 = MessagePackTime{}

	require.Equal(t, t1.Len(), 12)
	require.NoError(t, t1.MarshalBinaryTo(buf))

	t2 = new(MessagePackTime)
	err = t2.UnmarshalBinary(buf)
	require.NoError(t, err)

	require.True(t, t1.time.Equal(t2.time))
}

func TestMsgPackTimeEdgeCases(t *testing.T) {
	times := make([]time.Time, 0)
	expected := make([][]byte, 0)

	// Unix epoch. Begin of 4bytes dates
	// Nanoseconds: 0x00000000, Seconds: 0x0000000000000000
	ts, _ := time.Parse(time.RFC3339, "1970-01-01T00:00:00Z")
	bs, _ := hex.DecodeString("d6ff00000000")
	times = append(times, ts)
	expected = append(expected, bs)

	// End of 4bytes dates
	// Nanoseconds: 0x00000000, Seconds: 0x00000000ffffffff
	ts, _ = time.Parse(time.RFC3339, "2106-02-07T06:28:15Z")
	bs, _ = hex.DecodeString("d6ffffffffff")
	times = append(times, ts)
	expected = append(expected, bs)

	// Begin of 8bytes dates
	// Nanoseconds: 0x00000000, Seconds: 0x0000000100000000
	ts, _ = time.Parse(time.RFC3339, "2106-02-07T06:28:16Z")
	bs, _ = hex.DecodeString("d7ff0000000100000000")
	times = append(times, ts)
	expected = append(expected, bs)

	// Just after Unix epoch. Non zero nanoseconds
	// Nanoseconds: 0x00000001, Seconds: 0x0000000000000000
	ts, _ = time.Parse(time.RFC3339Nano, "1970-01-01T00:00:00.000000001Z")
	bs, _ = hex.DecodeString("d7ff0000000400000000")
	times = append(times, ts)
	expected = append(expected, bs)

	// End of 8bytes dates
	// Nanoseconds: 0x00000000, Seconds: 0x00000003ffffffff
	ts, _ = time.Parse(time.RFC3339Nano, "2514-05-30T01:53:03.000000000Z")
	bs, _ = hex.DecodeString("d7ff00000003ffffffff")
	times = append(times, ts)
	expected = append(expected, bs)

	// Begin of 12bytes date
	// Nanoseconds: 0x00000000, Seconds: 0x0000000400000000
	ts, _ = time.Parse(time.RFC3339Nano, "2514-05-30T01:53:04.000000000Z")
	bs, _ = hex.DecodeString("c70cff000000000000000400000000")
	times = append(times, ts)
	expected = append(expected, bs)

	// Zero value, 0001-01-01T00:00:00Z
	// Nanoseconds: 0x00000000, Seconds: 0xfffffff1886e0900
	ts = time.Time{}
	bs, _ = hex.DecodeString("c70cff00000000fffffff1886e0900")
	times = append(times, ts)
	expected = append(expected, bs)

	// Max value
	// Nanoseconds: 0x3b9ac9ff, Seconds: 0x7fffffffffffffff
	ts = time.Unix(math.MaxInt64, 999_999_999).UTC()
	bs, _ = hex.DecodeString("c70cff3b9ac9ff7fffffffffffffff")
	times = append(times, ts)
	expected = append(expected, bs)

	buf := make([]byte, 0)
	for i, ts := range times {
		t1 := MessagePackTime{time: ts}
		m := Metric{Time: t1}

		buf = buf[:0]
		buf, _ = m.MarshalMsg(buf)
		require.Equal(t, expected[i], buf[12:len(buf)-14])
	}
}

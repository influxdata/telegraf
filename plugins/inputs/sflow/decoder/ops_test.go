package decoder

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_U64AsF(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := AsF("out")
	in := uint64(5)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, in, getField(m, "out"))
}

func Test_U32AsF(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := AsF("out")
	in := uint32(5)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, uint64(in), getField(m, "out"))
}

func Test_U16PtrAsF(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := AsF("out")
	in := uint16(5)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, uint64(in), getField(m, "out"))
}

func Test_U16AsF(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := AsF("out")
	in := uint16(5)
	require.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	require.Equal(t, uint64(in), getField(m, "out"))
}

func Test_U8AsF(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := AsF("out")
	in := uint8(5)
	require.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	require.Equal(t, uint64(in), getField(m, "out"))
}

func Test_U8PtrAsF(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := AsF("out")
	in := uint8(5)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, uint64(in), getField(m, "out"))
}

func Test_U32AsT(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := AsT("out")
	in := uint32(5)
	require.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	require.Equal(t, fmt.Sprintf("%d", in), getTag(m, "out"))
}

func Test_U32PtrAsT(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := AsT("out")
	in := uint32(5)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, fmt.Sprintf("%d", in), getTag(m, "out"))
}

func Test_U16AsT(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := AsT("out")
	in := uint16(5)
	require.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	require.Equal(t, fmt.Sprintf("%d", in), getTag(m, "out"))
}

func Test_U16PtrAsT(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := AsT("out")
	in := uint16(5)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, fmt.Sprintf("%d", in), getTag(m, "out"))
}

func Test_U8AsT(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := AsT("out")
	in := uint8(5)
	require.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	require.Equal(t, fmt.Sprintf("%d", in), getTag(m, "out"))
}

func Test_U8PtrAsT(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := AsT("out")
	in := uint8(5)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, fmt.Sprintf("%d", in), getTag(m, "out"))
}

func Test_U32ToU32AsF(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := U32ToU32(func(i uint32) uint32 { return i * 2 })
	ddo2 := ddo.AsF("out")
	require.Equal(t, ddo, ddo2.prev())
	in := uint32(5)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, uint64(in*2), getField(m, "out"))
}

func Test_U16ToU16AsF(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := U16ToU16(func(i uint16) uint16 { return i * 2 })
	ddo2 := ddo.AsF("out")
	require.Equal(t, ddo, ddo2.prev())
	in := uint16(5)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, uint64(in*2), getField(m, "out"))
}

func Test_U32ToStrAsT(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := U32ToStr(func(i uint32) string { return fmt.Sprintf("%d", i*2) })
	ddo2 := ddo.AsT("out")
	require.Equal(t, ddo, ddo2.prev())
	in := uint32(5)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, fmt.Sprintf("%d", (in*2)), getTag(m, "out"))
}

func Test_U16ToStrAsT(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := U16ToStr(func(i uint16) string { return fmt.Sprintf("%d", i*2) })
	ddo2 := ddo.AsT("out")
	require.Equal(t, ddo, ddo2.prev())
	in := uint16(5)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, fmt.Sprintf("%d", (in*2)), getTag(m, "out"))
}

func Test_MapU32ToStrAsT(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	myMap := map[uint32]string{5: "five"}
	ddo := MapU32ToStr(myMap)
	ddo2 := ddo.AsT("out")
	require.Equal(t, ddo, ddo2.prev())
	in := uint32(5)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, "five", getTag(m, "out"))
}

func Test_MapU16ToStrAsT(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	myMap := map[uint16]string{5: "five"}
	ddo := MapU16ToStr(myMap)
	ddo2 := ddo.AsT("out")
	require.Equal(t, ddo, ddo2.prev())
	in := uint16(5)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, "five", getTag(m, "out"))
}

func Test_DecDir_ToU32(t *testing.T) {
	u := U32().
		Do(U32ToU32(func(in uint32) uint32 { return in >> 2 }).AsF("out1")).
		Do(U32ToU32(func(in uint32) uint32 { return in * 2 }).AsF("out2"))
	dd := Seq(OpenMetric(""), u, CloseMetric())

	value := uint32(1001)
	var buffer bytes.Buffer
	require.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))

	dc := NewDecodeContext()
	require.NoError(t, dc.Decode(dd, &buffer))

	x, _ := u.(*valueDirective)
	require.Equal(t, &value, x.value)

	// require field ejected
	require.Equal(t, 1, len(dc.GetMetrics()))
	m := dc.GetMetrics()
	require.Equal(t, uint64(value>>2), getField(m[0], "out1"))
	require.Equal(t, uint64(value*2), getField(m[0], "out2"))
}

func Test_BytesToStrAsT(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	f := func(b []byte) string { return fmt.Sprintf("%d:%d", b[0], b[1]) }
	ddo := BytesToStr(2, f)
	ddo2 := ddo.AsT("out")
	require.Equal(t, ddo, ddo2.prev())
	in := []byte{0x01, 0x02}
	require.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	require.Equal(t, fmt.Sprintf("%d:%d", in[0], in[1]), getTag(m, "out"))
}

func Test_BytesToAsT(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	f := func(b []byte) interface{} { return fmt.Sprintf("%d:%d", b[0], b[1]) }
	ddo := BytesTo(2, f)
	ddo2 := ddo.AsT("out")
	require.Equal(t, ddo, ddo2.prev())
	in := []byte{0x01, 0x02}
	require.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	require.Equal(t, fmt.Sprintf("%d:%d", in[0], in[1]), getTag(m, "out"))
}

func Test_BytesToU32AsF(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	f := func(b []byte) uint32 { return uint32(b[0] * b[1]) }
	ddo := BytesToU32(2, f)
	ddo2 := ddo.AsF("out")
	require.Equal(t, ddo, ddo2.prev())
	in := []byte{0x01, 0x02}
	require.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	require.Equal(t, uint64(in[0]*in[1]), getField(m, "out"))
}

func Test_U32require(t *testing.T) {
	dc := NewDecodeContext()
	ddo := U32Assert(func(in uint32) bool { return false }, "bad")
	in := uint32(5)
	require.Error(t, ddo.process(dc, &in))
}

func Test_U16require(t *testing.T) {
	dc := NewDecodeContext()
	ddo := U16Assert(func(in uint16) bool { return false }, "bad")
	in := uint16(5)
	require.Error(t, ddo.process(dc, &in))
}

func Test_Set(t *testing.T) {
	dc := NewDecodeContext()
	ptr := new(uint32)
	ddo := Set(ptr)
	in := uint32(5)
	require.NoError(t, ddo.process(dc, &in))
	require.Equal(t, *ptr, in)
}

func Test_U16Set(t *testing.T) {
	dc := NewDecodeContext()
	ptr := new(uint16)
	ddo := Set(ptr)
	in := uint16(5)
	require.NoError(t, ddo.process(dc, in))
	require.Equal(t, *ptr, in)
}

func Test_U16PtrSet(t *testing.T) {
	dc := NewDecodeContext()
	ptr := new(uint16)
	ddo := Set(ptr)
	in := uint16(5)
	require.NoError(t, ddo.process(dc, &in))
	require.Equal(t, *ptr, in)
}

func Test_U32toU32Set(t *testing.T) {
	dc := NewDecodeContext()
	ptr := new(uint32)
	ddo := U32ToU32(func(in uint32) uint32 { return in * 2 }).Set(ptr).prev()
	in := uint32(5)
	require.NoError(t, ddo.process(dc, &in))
	require.Equal(t, *ptr, in*2)
}

func Test_U32toU32toString(t *testing.T) {
	dc := NewDecodeContext()
	ptr := new(string)
	ddo := U32ToU32(func(in uint32) uint32 { return in * 2 }).ToString(func(in uint32) string { return fmt.Sprintf("%d", in*2) }).Set(ptr).prev().prev()
	in := uint32(2)
	require.NoError(t, ddo.process(dc, &in))
	require.Equal(t, "8", *ptr)
}

func Test_U32toU32toStringBreakIf(t *testing.T) {
	dc := NewDecodeContext()
	ptr := new(string)
	ddo := U32ToU32(func(in uint32) uint32 { return in * 2 }).ToString(func(in uint32) string { return fmt.Sprintf("%d", in*2) }).BreakIf("8").Set(ptr).prev().prev().prev()
	in := uint32(2)
	require.NoError(t, ddo.process(dc, &in))
	require.Equal(t, "", *ptr)

	in = uint32(1)
	require.NoError(t, ddo.process(dc, &in))
	require.Equal(t, "4", *ptr)
}

func Test_notify(t *testing.T) {
	value := uint32(1001)
	var buffer bytes.Buffer
	require.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))

	ptr := new(uint32)
	*ptr = uint32(2002)
	var notificationOne uint32
	var notificationTwo uint32
	dd := Seq(
		Notify(func() { notificationOne = *ptr }),
		U32().Do(Set(ptr)),
		Notify(func() { notificationTwo = *ptr }),
	)

	require.NoError(t, Execute(dd, &buffer))
	require.Equal(t, uint32(2002), notificationOne)
	require.Equal(t, uint32(1001), notificationTwo)
}

func Test_nop(t *testing.T) {
	value := uint32(1001)
	var buffer bytes.Buffer
	require.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	originalLen := buffer.Len()
	dd := Seq(
		Nop(),
	)

	require.NoError(t, Execute(dd, &buffer))
	require.Equal(t, originalLen, buffer.Len())
}

func Test_AsTimestamp(t *testing.T) {
	dc := NewDecodeContext()
	dc.openMetric("")
	ddo := AsTimestamp()
	now := time.Now()
	in := uint32(now.Unix()) // only handles as uin32 (not uint64)
	require.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	require.Equal(t, now.Unix(), m.Time().Unix())
}

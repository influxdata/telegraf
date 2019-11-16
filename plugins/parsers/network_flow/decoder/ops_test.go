package decoder

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_U32AsF(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	ddo := AsF("out")
	in := uint32(5)
	assert.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	assert.Equal(t, uint64(in), getField(m, "out"))
}

func Test_U16PtrAsF(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	ddo := AsF("out")
	in := uint16(5)
	assert.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	assert.Equal(t, uint64(in), getField(m, "out"))
}

func Test_U16AsF(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	ddo := AsF("out")
	in := uint16(5)
	assert.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	assert.Equal(t, uint64(in), getField(m, "out"))
}

func Test_U8AsF(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	ddo := AsF("out")
	in := uint8(5)
	assert.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	assert.Equal(t, uint64(in), getField(m, "out"))
}

func Test_BadAsF(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	ddo := AsF("out")
	in := "hello"
	assert.Error(t, ddo.process(dc, in))
}

func Test_U32AsT(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	ddo := AsT("out")
	in := uint32(5)
	assert.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	assert.Equal(t, fmt.Sprintf("%d", in), getTag(m, "out"))
}

func Test_U32PtrAsT(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	ddo := AsT("out")
	in := uint32(5)
	assert.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	assert.Equal(t, fmt.Sprintf("%d", in), getTag(m, "out"))
}

func Test_U16AsT(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	ddo := AsT("out")
	in := uint16(5)
	assert.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	assert.Equal(t, fmt.Sprintf("%d", in), getTag(m, "out"))
}

func Test_U16PtrAsT(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	ddo := AsT("out")
	in := uint16(5)
	assert.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	assert.Equal(t, fmt.Sprintf("%d", in), getTag(m, "out"))
}

func Test_U32ToU32AsF(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	ddo := U32ToU32(func(i uint32) uint32 { return i * 2 })
	ddo2 := ddo.AsF("out")
	assert.Equal(t, ddo, ddo2.prev())
	in := uint32(5)
	assert.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	assert.Equal(t, uint64(in*2), getField(m, "out"))
}

func Test_U16ToU16AsF(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	ddo := U16ToU16(func(i uint16) uint16 { return i * 2 })
	ddo2 := ddo.AsF("out")
	assert.Equal(t, ddo, ddo2.prev())
	in := uint16(5)
	assert.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	assert.Equal(t, uint64(in*2), getField(m, "out"))
}

func Test_U32ToStrAsT(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	ddo := U32ToStr(func(i uint32) string { return fmt.Sprintf("%d", i*2) })
	ddo2 := ddo.AsT("out")
	assert.Equal(t, ddo, ddo2.prev())
	in := uint32(5)
	assert.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	assert.Equal(t, fmt.Sprintf("%d", (in*2)), getTag(m, "out"))
}

func Test_MapU32ToStrAsT(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	myMap := map[uint32]string{5: "five"}
	ddo := MapU32ToStr(myMap)
	ddo2 := ddo.AsT("out")
	assert.Equal(t, ddo, ddo2.prev())
	in := uint32(5)
	assert.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	assert.Equal(t, "five", getTag(m, "out"))
}

func Test_MapU16ToStrAsT(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	myMap := map[uint16]string{5: "five"}
	ddo := MapU16ToStr(myMap)
	ddo2 := ddo.AsT("out")
	assert.Equal(t, ddo, ddo2.prev())
	in := uint16(5)
	assert.NoError(t, ddo.process(dc, &in))
	m := dc.currentMetric()
	assert.Equal(t, "five", getTag(m, "out"))
}

func Test_DecDir_ToU32(t *testing.T) {
	u := U32().
		Do(U32ToU32(func(in uint32) uint32 { return in >> 2 }).AsF("out1")).
		Do(U32ToU32(func(in uint32) uint32 { return in * 2 }).AsF("out2"))
	dd := Seq(OpenMetric(), u, CloseMetric())

	value := uint32(1001)
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))

	dc := NewDecodeContext(true)
	assert.NoError(t, dc.Decode(dd, &buffer))

	x, _ := u.(*valueDirective)
	assert.Equal(t, &value, x.value)

	// assert field ejected
	assert.Equal(t, 1, len(dc.GetMetrics()))
	m := dc.GetMetrics()
	assert.Equal(t, uint64(value>>2), getField(m[0], "out1"))
	assert.Equal(t, uint64(value*2), getField(m[0], "out2"))
}

func Test_BytesToStrAsT(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	f := func(b []byte) string { return fmt.Sprintf("%d:%d", b[0], b[1]) }
	ddo := BytesToStr(2, f)
	ddo2 := ddo.AsT("out")
	assert.Equal(t, ddo, ddo2.prev())
	in := []byte{0x01, 0x02}
	assert.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	assert.Equal(t, fmt.Sprintf("%d:%d", in[0], in[1]), getTag(m, "out"))
}

func Test_BytesToAsT(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	f := func(b []byte) interface{} { return fmt.Sprintf("%d:%d", b[0], b[1]) }
	ddo := BytesTo(2, f)
	ddo2 := ddo.AsT("out")
	assert.Equal(t, ddo, ddo2.prev())
	in := []byte{0x01, 0x02}
	assert.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	assert.Equal(t, fmt.Sprintf("%d:%d", in[0], in[1]), getTag(m, "out"))
}

func Test_BytesToU32AsF(t *testing.T) {
	dc := NewDecodeContext(true)
	dc.openMetric()
	f := func(b []byte) uint32 { return uint32(b[0] * b[1]) }
	ddo := BytesToU32(2, f)
	ddo2 := ddo.AsF("out")
	assert.Equal(t, ddo, ddo2.prev())
	in := []byte{0x01, 0x02}
	assert.NoError(t, ddo.process(dc, in))
	m := dc.currentMetric()
	assert.Equal(t, uint64(in[0]*in[1]), getField(m, "out"))
}

func Test_U32Assert(t *testing.T) {
	dc := NewDecodeContext(true)
	ddo := U32Assert(func(in uint32) bool { return false }, "bad")
	in := uint32(5)
	assert.Error(t, ddo.process(dc, &in))
}

func Test_Set(t *testing.T) {
	dc := NewDecodeContext(true)
	ptr := new(uint32)
	ddo := Set(ptr)
	in := uint32(5)
	assert.NoError(t, ddo.process(dc, &in))
	assert.Equal(t, *ptr, in)
}

func Test_U32toU32Set(t *testing.T) {
	dc := NewDecodeContext(true)
	ptr := new(uint32)
	ddo := U32ToU32(func(in uint32) uint32 { return in * 2 }).Set(ptr).prev()
	in := uint32(5)
	assert.NoError(t, ddo.process(dc, &in))
	assert.Equal(t, *ptr, in*2)
}

func Test_U32toU32toString(t *testing.T) {
	dc := NewDecodeContext(true)
	ptr := new(string)
	ddo := U32ToU32(func(in uint32) uint32 { return in * 2 }).ToString(func(in uint32) string { return fmt.Sprintf("%d", in*2) }).Set(ptr).prev().prev()
	in := uint32(2)
	assert.NoError(t, ddo.process(dc, &in))
	assert.Equal(t, "8", *ptr)
}

func Test_U32toU32toStringBreakIf(t *testing.T) {
	dc := NewDecodeContext(true)
	ptr := new(string)
	ddo := U32ToU32(func(in uint32) uint32 { return in * 2 }).ToString(func(in uint32) string { return fmt.Sprintf("%d", in*2) }).BreakIf("8").Set(ptr).prev().prev().prev()
	in := uint32(2)
	assert.NoError(t, ddo.process(dc, &in))
	assert.Equal(t, "", *ptr)

	in = uint32(1)
	assert.NoError(t, ddo.process(dc, &in))
	assert.Equal(t, "4", *ptr)
}

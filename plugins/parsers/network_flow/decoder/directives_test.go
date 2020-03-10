package decoder

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/assert"
)

// Execute will ececute the decode directive relative to the supplied buffer
func Execute(dd Directive, buffer *bytes.Buffer) error {
	dc := &DecodeContext{}
	return dd.Execute(buffer, dc)
}

func Test_basicUI32NotEnoughBytes(t *testing.T) {
	dd := U32()
	value := uint16(1001) // not enough bytes to read a U32 out as only a U16 in
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	assert.Error(t, Execute(dd, &buffer))
}

func Test_basicUI32(t *testing.T) {
	dd := U32()
	value := uint32(1001)
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	assert.NoError(t, Execute(dd, &buffer))
	assert.Equal(t, 0, buffer.Len())
	x, _ := dd.(*valueDirective)
	assert.Equal(t, &value, x.value)
}

func Test_basicBytes(t *testing.T) {
	dd := Bytes(4)
	value := []byte{0x01, 0x02, 0x03, 0x04}
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	assert.NoError(t, Execute(dd, &buffer))
	assert.Equal(t, 0, buffer.Len())
	x, _ := dd.(*valueDirective)
	assert.Equal(t, value, x.value)
}

func Test_basicSeq(t *testing.T) {

	// Seq with no members compiles and executed but buffer is left untouched
	dd := Seq()
	value := uint32(1001)
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	originalLen := buffer.Len()
	assert.NoError(t, Execute(dd, &buffer))
	assert.Equal(t, originalLen, buffer.Len())

	u := U32()
	dd = Seq(
		u,
	)
	value = uint32(1001)
	buffer.Reset()
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	assert.NoError(t, Execute(dd, &buffer))
	assert.Equal(t, 0, buffer.Len())
	x, _ := u.(*valueDirective)
	assert.Equal(t, &value, x.value)
}

func Test_basicSeqOf(t *testing.T) {

	// SeqOf with no members compiles and executed but buffer is left untouched
	dd := SeqOf([]Directive{})
	value := uint32(1001)
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	originalLen := buffer.Len()
	assert.NoError(t, Execute(dd, &buffer))
	assert.Equal(t, originalLen, buffer.Len())

	u := U32()
	dd = SeqOf(
		[]Directive{u},
	)
	value = uint32(1001)
	buffer.Reset()
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	assert.NoError(t, Execute(dd, &buffer))
	assert.Equal(t, 0, buffer.Len())
	x, _ := u.(*valueDirective)
	assert.Equal(t, &value, x.value)
}

func Test_errorInSeq(t *testing.T) {

	// Seq with no members compiles and executed but buffer is left untouched
	dd := Seq(U32(), ErrorDirective())
	value := uint32(1001)
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	assert.Error(t, Execute(dd, &buffer))
}

func Test_basicU32Switch(t *testing.T) {

	c1 := U32()
	c2 := U32()
	dd := U32().Switch(
		Case(uint32(1), c1),
		Case(uint32(2), c2),
	)

	value1 := uint32(3)
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value1))
	value2 := uint32(4)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value2))
	assert.Error(t, Execute(dd, &buffer)) // should error as no path

	value1 = uint32(1)
	buffer.Reset()
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value1))
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value2))
	assert.NoError(t, Execute(dd, &buffer))
	x, _ := c1.(*valueDirective)
	y, _ := c2.(*valueDirective)
	value0 := uint32(0)
	assert.Equal(t, &value2, x.value)
	assert.Equal(t, &value0, y.value)

	// bad path shoudl raise error
	// path 1 should be able to fina value in c1 and not in c2
	// then other way around
}

func Test_basicU16Switch(t *testing.T) {

	c1 := U32()
	c2 := U32()
	dd := U16().Switch(
		Case(uint16(1), c1),
		Case(uint16(2), c2),
	)

	swValue := uint16(3)
	cValue := uint32(1)
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &swValue))
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &cValue))
	assert.Error(t, Execute(dd, &buffer)) // should error as no path

	swValue = uint16(1)
	buffer.Reset()
	dd.Reset()

	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &swValue))
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &cValue))
	assert.NoError(t, Execute(dd, &buffer))
	x, _ := c1.(*valueDirective)
	y, _ := c2.(*valueDirective)
	value0 := uint32(0)
	assert.Equal(t, &cValue, x.value)
	assert.Equal(t, &value0, y.value)

	swValue = uint16(2)
	buffer.Reset()
	dd.Reset()

	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &swValue))
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &cValue))
	assert.NoError(t, Execute(dd, &buffer))
	x, _ = c1.(*valueDirective)
	xUI32, _ := x.value.(*uint32)
	y, _ = c2.(*valueDirective)
	yUI32, _ := y.value.(*uint32)

	assert.Equal(t, cValue, *yUI32)
	assert.Equal(t, value0, *xUI32)

}

func Test_basicBinSwitch(t *testing.T) {

	c1 := U32()
	c2 := U32()
	dd := Bytes(1).Switch(
		Case(byte(1), c1),
		Case(byte(2), c2),
	)

	value1 := byte(3)
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value1))
	value2 := uint32(4)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value2))
	assert.Error(t, Execute(dd, &buffer)) // should error as no path

	value1 = byte(1)
	buffer.Reset()
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value1))
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value2))
	assert.NoError(t, Execute(dd, &buffer))
	x, _ := c1.(*valueDirective)
	y, _ := c2.(*valueDirective)
	value0 := uint32(0)
	assert.Equal(t, &value2, x.value)
	assert.Equal(t, &value0, y.value)

	// bad path shoudl raise error
	// path 1 should be able to fina value in c1 and not in c2
	// then other way around
}

func Test_basicIter(t *testing.T) {
	innerDD := U32()
	dd := U32().Iter(math.MaxInt32, innerDD)

	var buffer bytes.Buffer
	iterations := uint32(2)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &iterations))
	it1Val := uint32(3)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &it1Val))
	it2Val := uint32(4)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &it2Val))
	assert.NoError(t, Execute(dd, &buffer))
	x, _ := dd.(*valueDirective)
	assert.Equal(t, &iterations, x.value)
	y, _ := innerDD.(*valueDirective)
	// we can't test it1Val as it gets overwritten!
	assert.Equal(t, &it2Val, y.value)

	// check reset passes down
	dd.Reset()
	zero := uint32(0)
	assert.Equal(t, &zero, x.value)
	assert.Equal(t, &zero, y.value)
}

func Test_IterLimit(t *testing.T) {
	innerDD := U32()
	dd := U32().Iter(1, innerDD) // limit set at 1
	var buffer bytes.Buffer
	iterations := uint32(2)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &iterations))
	it1Val := uint32(3)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &it1Val))
	it2Val := uint32(4)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &it2Val))
	assert.Error(t, Execute(dd, &buffer))
}

func Test_errorWithinIter(t *testing.T) {
	dd := U32().Iter(math.MaxInt32, ErrorDirective())

	var buffer bytes.Buffer
	iterations := uint32(1)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &iterations))

	assert.Error(t, Execute(dd, &buffer))
}

func Test_errorWithinIter2(t *testing.T) {
	dd := U32().Iter(math.MaxInt32, U32().Do(ErrorOp(false)))
	var buffer bytes.Buffer
	iterations := uint32(1)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &iterations))
	innerValue := uint32(1)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &innerValue))
	assert.Error(t, Execute(dd, &buffer))
}

func Test_errorWithinIter3(t *testing.T) {
	defer expectPanic(t, "Test_cantIterBytes")
	U32().Iter(math.MaxInt32, U32().Do(ErrorOp(true)))
}

func Test_alreadyEncapsulated(t *testing.T) {
	defer expectPanic(t, "Test_cantIterBytes")
	u := U32()
	inner := U32()
	u.Encapsulated(math.MaxInt32, inner)
	u.Encapsulated(math.MaxInt32, inner)
}

func Test_alreadyDoAssigned(t *testing.T) {
	defer expectPanic(t, "Test_cantIterBytes")
	u := U32()
	u.Do(AsF("foo"))
	inner := U32()
	u.Encapsulated(math.MaxInt32, inner)
}

func Test_cantIterBytes(t *testing.T) {
	defer expectPanic(t, "Test_cantIterBytes")
	_ = Bytes(1).Iter(math.MaxInt32, U32())
}

// then open metric
func Test_OpenMetric(t *testing.T) {
	innerDD := U32()
	dd := U32().Iter(math.MaxInt32, Seq(
		OpenMetric(""),
		innerDD,
		CloseMetric(),
	))

	var buffer bytes.Buffer
	iterations := uint32(2)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &iterations))
	it1Val := uint32(3)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &it1Val))
	it2Val := uint32(3)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &it2Val))
	dc := NewDecodeContext(true)
	assert.NoError(t, dc.Decode(dd, &buffer))
	assert.Equal(t, 2, len(dc.GetMetrics()))
}

func Test_AsF(t *testing.T) {
	innerDD := U32().Do(AsF("foo"))
	dd := U32().Iter(math.MaxInt32, Seq(
		OpenMetric(""),
		innerDD,
		CloseMetric(),
	))

	var buffer bytes.Buffer
	iterations := uint32(2)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &iterations))
	it1Val := uint32(3)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &it1Val))
	it2Val := uint32(3)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &it2Val))
	dc := NewDecodeContext(true)
	assert.NoError(t, dc.Decode(dd, &buffer))
	assert.Equal(t, 2, len(dc.GetMetrics()))
	m := dc.GetMetrics()
	assert.Equal(t, uint64(it1Val), getField(m[0], "foo"))
	assert.Equal(t, uint64(it2Val), getField(m[1], "foo"))
}

func Test_AsT(t *testing.T) {
	innerDD := U32().Do(AsT("foo"))
	dd := U32().Iter(math.MaxInt32, Seq(
		OpenMetric(""),
		innerDD,
		CloseMetric(),
	))

	var buffer bytes.Buffer
	iterations := uint32(2)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &iterations))
	it1Val := uint32(3)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &it1Val))
	it2Val := uint32(3)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &it2Val))
	dc := NewDecodeContext(true)
	assert.NoError(t, dc.Decode(dd, &buffer))
	assert.Equal(t, 2, len(dc.GetMetrics()))
	m := dc.GetMetrics()
	assert.Equal(t, fmt.Sprintf("%d", it1Val), getTag(m[0], "foo"))
	assert.Equal(t, fmt.Sprintf("%d", it2Val), getTag(m[1], "foo"))
}

func getField(m telegraf.Metric, name string) interface{} {
	v, _ := m.GetField(name)
	return v
}

func getTag(m telegraf.Metric, name string) string {
	v, _ := m.GetTag(name)
	return v
}

func Test_preMetricNesting(t *testing.T) {
	innerDD := U32().Do(AsF("foo"))
	dd := Seq(
		U32().Do(AsF("bar")),
		U32().Do(AsT("baz")),
		U32().Iter(math.MaxInt32,
			Seq(
				OpenMetric(""),
				innerDD,
				CloseMetric(),
			),
		),
	)

	var buffer bytes.Buffer
	barVal := uint32(55)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &barVal))
	bazVal := uint32(56)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &bazVal))
	iterations := uint32(2)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &iterations))
	it1Val := uint32(3)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &it1Val))
	it2Val := uint32(3)
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &it2Val))
	dc := NewDecodeContext(true)
	assert.NoError(t, dc.Decode(dd, &buffer))
	assert.Equal(t, 2, len(dc.GetMetrics()))
	m := dc.GetMetrics()
	assert.Equal(t, uint64(barVal), getField(m[0], "bar"))
	assert.Equal(t, fmt.Sprintf("%d", bazVal), getTag(m[0], "baz"))
	assert.Equal(t, uint64(it1Val), getField(m[0], "foo"))
	assert.Equal(t, uint64(barVal), getField(m[1], "bar"))
	assert.Equal(t, fmt.Sprintf("%d", bazVal), getTag(m[1], "baz"))
	assert.Equal(t, uint64(it2Val), getField(m[1], "foo"))
}

func Test_BasicEncapsulated(t *testing.T) {

	encap1Value := uint32(2)
	encap2Value := uint32(3)
	var encapBuffer bytes.Buffer
	assert.NoError(t, binary.Write(&encapBuffer, binary.BigEndian, &encap1Value))
	assert.NoError(t, binary.Write(&encapBuffer, binary.BigEndian, &encap2Value))

	encapSize := uint32(encapBuffer.Len())
	envelopeValue := uint32(4)
	var envelopeBuffer bytes.Buffer

	assert.NoError(t, binary.Write(&envelopeBuffer, binary.BigEndian, &encapSize))
	l, e := envelopeBuffer.Write(encapBuffer.Bytes())
	assert.NoError(t, e)
	assert.Equal(t, encapSize, uint32(l))
	assert.NoError(t, binary.Write(&envelopeBuffer, binary.BigEndian, &envelopeValue))

	innerDD := U32()
	envelopeDD := U32() // the buffer contains another U32 but the encpaultation will ignore it
	dd := Seq(
		U32().Encapsulated(math.MaxInt32, innerDD),
		envelopeDD,
	)
	assert.NoError(t, Execute(dd, &envelopeBuffer))

	assert.Equal(t, 0, envelopeBuffer.Len())
	x, _ := envelopeDD.(*valueDirective)
	assert.Equal(t, &envelopeValue, x.value)
	y, _ := innerDD.(*valueDirective)
	assert.Equal(t, &encap1Value, y.value)

	// check Reset works
	dd.Reset()
	zero := uint32(0)
	assert.Equal(t, &zero, x.value)
	assert.Equal(t, &zero, y.value)

}

func Test_EncapsulationLimit(t *testing.T) {

	encap1Value := uint32(2)
	encap2Value := uint32(3)
	var encapBuffer bytes.Buffer
	assert.NoError(t, binary.Write(&encapBuffer, binary.BigEndian, &encap1Value))
	assert.NoError(t, binary.Write(&encapBuffer, binary.BigEndian, &encap2Value))

	encapSize := uint32(encapBuffer.Len())
	envelopeValue := uint32(4)
	var envelopeBuffer bytes.Buffer

	assert.NoError(t, binary.Write(&envelopeBuffer, binary.BigEndian, &encapSize))
	l, e := envelopeBuffer.Write(encapBuffer.Bytes())
	assert.NoError(t, e)
	assert.Equal(t, encapSize, uint32(l))
	assert.NoError(t, binary.Write(&envelopeBuffer, binary.BigEndian, &envelopeValue))

	innerDD := U32()
	envelopeDD := U32()
	dd := Seq(
		U32().Encapsulated(4, innerDD), // 4 bytes, not 8 bytes or higher as max
		envelopeDD,
	)
	assert.Error(t, Execute(dd, &envelopeBuffer))
}

func Test_cantEncapulatedBytes(t *testing.T) {
	defer expectPanic(t, "cantEncapulatedBytes")
	_ = Bytes(1).Encapsulated(math.MaxInt32, U32())
}

func Test_BasicRef(t *testing.T) {

	var x interface{}
	dd1 := U32().Ref(&x)
	dd2 := Ref(x)
	dd := Seq(
		dd1,
		dd2,
	)
	y, ok := dd2.(*valueDirective)
	assert.True(t, ok)
	assert.Equal(t, y.reference, x)

	value := uint32(1001)
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	assert.NoError(t, Execute(dd, &buffer))

	y, _ = dd1.(*valueDirective)
	assert.Equal(t, &value, y.value)

	y, _ = dd2.(*valueDirective)
	assert.Equal(t, &value, y.value)
}

func Test_RefReassignError(t *testing.T) {
	defer expectPanic(t, "iter iter")
	var x interface{}
	U32().Ref(&x)
	U32().Ref(&x)
}

func Test_ToU32(t *testing.T) {
	u := U32().Do(U32ToU32(func(in uint32) uint32 { return in >> 2 }).AsF("x"))
	dd := Seq(OpenMetric(""), u, CloseMetric())

	value := uint32(1001)
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))

	dc := NewDecodeContext(true)
	assert.NoError(t, dc.Decode(dd, &buffer))

	// assert original value decoded
	x, _ := u.(*valueDirective)
	assert.Equal(t, &value, x.value)

	// assert field ejected
	assert.Equal(t, 1, len(dc.GetMetrics()))
	m := dc.GetMetrics()
	assert.Equal(t, uint64(value>>2), getField(m[0], "x"))
}

func expectPanic(t *testing.T, msg string) {
	if r := recover(); r == nil {
		t.Errorf(msg)
	}
}

func Test_U32BlankCanvasIter(t *testing.T) {
	u := U32().Iter(math.MaxInt32, U32())
	func() {
		defer expectPanic(t, "iter iter")
		u.Iter(math.MaxInt32, U32())
	}()
	func() {
		defer expectPanic(t, "iter switch")
		u.Switch(Case(uint32(0), U32()))
	}()
	func() {
		defer expectPanic(t, "iter encap")
		u.Encapsulated(math.MaxInt32, U32())
	}()
	func() {
		defer expectPanic(t, "iter do")
		u.Do(AsF("foo"))
	}()
}
func Test_U32BlankCanvasSwitch(t *testing.T) {
	u := U32().Switch(Case(uint32(0), U32()))
	func() {
		defer expectPanic(t, "switch iter")
		u.Iter(math.MaxInt32, U32())
	}()
	func() {
		defer expectPanic(t, "switch switch")
		u.Switch(Case(uint32(0), U32()))
	}()
	func() {
		defer expectPanic(t, "switch encap")
		u.Encapsulated(math.MaxInt32, U32())
	}()
	func() {
		defer expectPanic(t, "switch do")
		u.Do(AsF("foo"))
	}()
}

func Test_U32BasicSwitch(t *testing.T) {
	s := U32().Switch(Case(uint32(0), nil))
	value := uint32(0)
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	dc := NewDecodeContext(true)
	assert.NoError(t, dc.Decode(s, &buffer))
}

func Test_U32BasicSwitchDefault(t *testing.T) {
	s := U32().Switch(Case(uint32(0), nil), DefaultCase(nil))
	value := uint32(2)
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	dc := NewDecodeContext(true)
	assert.NoError(t, dc.Decode(s, &buffer))
}

func Test_U16(t *testing.T) {
	dd := U16()
	value := uint16(1001)
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	assert.NoError(t, Execute(dd, &buffer))
	assert.Equal(t, 0, buffer.Len())
	x, _ := dd.(*valueDirective)
	assert.Equal(t, &value, x.value)
}

func Test_U16Value(t *testing.T) {
	myU16 := uint16(5)
	dd := U16Value(&myU16)
	var buffer bytes.Buffer
	assert.NoError(t, Execute(dd, &buffer))
	x, _ := dd.(*valueDirective)
	assert.Equal(t, &myU16, x.value)
}

func Test_Bytes(t *testing.T) {
	dd := Bytes(4)
	value := []byte{0x01, 0x02, 0x03, 0x04}
	var buffer bytes.Buffer
	assert.NoError(t, binary.Write(&buffer, binary.BigEndian, &value))
	assert.NoError(t, Execute(dd, &buffer))
	assert.Equal(t, 0, buffer.Len())
	x, _ := dd.(*valueDirective)
	assert.Equal(t, value, x.value)

	// test reset
	dd.Reset()
	zeroBytes := []byte{0x00, 0x0, 0x00, 0x00}
	assert.Equal(t, zeroBytes, x.value)
}

func Test_nilRefAnfWongTypeRef(t *testing.T) {
	func() {
		defer expectPanic(t, "Test_nilRef")
		Ref(nil)
	}()

	func() {
		defer expectPanic(t, "Test_nilRef")
		f := new(uint32)
		Ref(f)
	}()
}

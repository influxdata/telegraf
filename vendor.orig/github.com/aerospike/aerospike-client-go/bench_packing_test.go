// Copyright 2013-2017 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"runtime"
	"strings"
	"testing"
	// "time"

	_ "net/http/pprof"
)

var buf *benchBuffer

func init() {
	buf = &benchBuffer{dataBuffer: make([]byte, 1024*1024), dataOffset: 0}
}

func Benchmark_Pack_binary_Write(b *testing.B) {
	buf := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		binary.Write(buf, binary.BigEndian, int64(0))
	}
}

func Benchmark_Pack_binary_PutUint64(b *testing.B) {
	buf := make([]byte, 8)
	for i := 0; i < b.N; i++ {
		binary.BigEndian.PutUint64(buf, 0)
	}
}

func doPack(val interface{}, b *testing.B) {
	var err error
	v := NewValue(val)
	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.dataOffset = 0
		_, err = v.pack(buf)
		if err != nil {
			panic(err)
		}
	}
}

func Benchmark_Pack_________Int64(b *testing.B) {
	val := rand.Int63()
	doPack(val, b)
}

func Benchmark_Pack_________Int32(b *testing.B) {
	val := rand.Int31()
	doPack(val, b)
}

func Benchmark_Pack_String______1(b *testing.B) {
	val := strings.Repeat("s", 1)
	doPack(val, b)
}

func Benchmark_Pack_String_____10(b *testing.B) {
	val := strings.Repeat("s", 10)
	doPack(val, b)
}

func Benchmark_Pack_String____100(b *testing.B) {
	val := strings.Repeat("s", 100)
	doPack(val, b)
}

func Benchmark_Pack_String___1000(b *testing.B) {
	val := strings.Repeat("s", 1000)
	doPack(val, b)
}

func Benchmark_Pack_String__10000(b *testing.B) {
	val := strings.Repeat("s", 10000)
	doPack(val, b)
}

func Benchmark_Pack_String_100000(b *testing.B) {
	val := strings.Repeat("s", 100000)
	doPack(val, b)
}

func Benchmark_Pack_Complex_IfcArray_Direct(b *testing.B) {
	val := []interface{}{1, 1, 1, "a simple string", nil, rand.Int63(), []byte{12, 198, 211}}
	doPack(val, b)
}

var _ ListIter = myList([]string{})

// supports old generic slices
type myList []string

func (cs myList) PackList(buf BufferEx) (int, error) {
	size := 0
	for _, elem := range cs {
		n, err := __PackString(buf, elem)
		size += n
		if err != nil {
			return size, err
		}
	}
	return size, nil
}

func (m myList) Len() int {
	return len(m)
}

func Benchmark_Pack_Complex_Array_ListIter(b *testing.B) {
	val := myList([]string{strings.Repeat("s", 1), strings.Repeat("s", 2), strings.Repeat("s", 3), strings.Repeat("s", 4), strings.Repeat("s", 5), strings.Repeat("s", 6), strings.Repeat("s", 7), strings.Repeat("s", 8), strings.Repeat("s", 9), strings.Repeat("s", 10)})
	doPack(val, b)
}

func Benchmark_Pack_Complex_ValueArray(b *testing.B) {
	val := []Value{NewValue(1), NewValue(strings.Repeat("s", 100000)), NewValue(1.75), NewValue(nil)}
	doPack(val, b)
}

func Benchmark_Pack_Complex_Map(b *testing.B) {
	val := map[interface{}]interface{}{
		rand.Int63(): rand.Int63(),
		nil:          1,
		"s":          491871,
		15892987:     strings.Repeat("s", 100),
		"s2":         []interface{}{"a simple string", nil, rand.Int63(), []byte{12, 198, 211}},
	}
	doPack(val, b)
}

func Benchmark_Pack_Complex_JsonMap(b *testing.B) {
	val := map[string]interface{}{
		"rand.Int63()": rand.Int63(),
		"nil":          1,
		"s":            491871,
		"15892987":     strings.Repeat("s", 100),
		"s2":           []interface{}{"a simple string", nil, rand.Int63(), []byte{12, 198, 211}},
	}
	doPack(val, b)
}

////////////////////////////////////////////////////////////////////////////////////////
type benchBuffer struct {
	dataBuffer []byte
	dataOffset int
}

// Int64ToBytes converts an int64 into slice of Bytes.
func (bb *benchBuffer) WriteInt64(num int64) (int, error) {
	return bb.WriteUint64(uint64(num))
}

// Uint64ToBytes converts an uint64 into slice of Bytes.
func (bb *benchBuffer) WriteUint64(num uint64) (int, error) {
	binary.BigEndian.PutUint64(bb.dataBuffer[bb.dataOffset:bb.dataOffset+8], num)
	bb.dataOffset += 8
	return 8, nil
}

// Int32ToBytes converts an int32 to a byte slice of size 4
func (bb *benchBuffer) WriteInt32(num int32) (int, error) {
	return bb.WriteUint32(uint32(num))
}

// Uint32ToBytes converts an uint32 to a byte slice of size 4
func (bb *benchBuffer) WriteUint32(num uint32) (int, error) {
	binary.BigEndian.PutUint32(bb.dataBuffer[bb.dataOffset:bb.dataOffset+4], num)
	bb.dataOffset += 4
	return 4, nil
}

// Int16ToBytes converts an int16 to slice of bytes
func (bb *benchBuffer) WriteInt16(num int16) (int, error) {
	return bb.WriteUint16(uint16(num))
}

// Int16ToBytes converts an int16 to slice of bytes
func (bb *benchBuffer) WriteUint16(num uint16) (int, error) {
	binary.BigEndian.PutUint16(bb.dataBuffer[bb.dataOffset:bb.dataOffset+2], num)
	bb.dataOffset += 2
	return 2, nil
}

func (bb *benchBuffer) WriteFloat32(float float32) (int, error) {
	bits := math.Float32bits(float)
	binary.BigEndian.PutUint32(bb.dataBuffer[bb.dataOffset:bb.dataOffset+4], bits)
	bb.dataOffset += 4
	return 4, nil
}

func (bb *benchBuffer) WriteFloat64(float float64) (int, error) {
	bits := math.Float64bits(float)
	binary.BigEndian.PutUint64(bb.dataBuffer[bb.dataOffset:bb.dataOffset+8], bits)
	bb.dataOffset += 8
	return 8, nil
}

func (bb *benchBuffer) WriteByte(b byte) error {
	bb.dataBuffer[bb.dataOffset] = b
	bb.dataOffset++
	return nil
}

func (bb *benchBuffer) WriteString(s string) (int, error) {
	copy(bb.dataBuffer[bb.dataOffset:bb.dataOffset+len(s)], s)
	bb.dataOffset += len(s)
	return len(s), nil
}

func (bb *benchBuffer) Write(b []byte) (int, error) {
	copy(bb.dataBuffer[bb.dataOffset:bb.dataOffset+len(b)], b)
	bb.dataOffset += len(b)
	return len(b), nil
}

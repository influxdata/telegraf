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
	"math/rand"
	"strings"
	"testing"
)

var sv StringValue
var iv IntegerValue
var lv LongValue
var bav BytesValue

var __value Value

func Benchmark_StringValue(b *testing.B) {
	b.N = 1e6
	str := strings.Repeat("a", 1000)
	for i := 0; i < b.N; i++ {
		__value = NewStringValue(str)
	}
}

func Benchmark_IntegerValue(b *testing.B) {
	b.N = 1e6
	in := 1091
	for i := 0; i < b.N; i++ {
		__value = NewIntegerValue(in)
	}
}

func Benchmark_LongValue(b *testing.B) {
	b.N = 1e6
	in := int64(10916927583729485)
	for i := 0; i < b.N; i++ {
		__value = NewLongValue(in)
	}
}

func Benchmark_BytesValue(b *testing.B) {
	b.N = 1e6
	barr := bytes.Repeat([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}, 1000)
	for i := 0; i < b.N; i++ {
		__value = NewBytesValue(barr)
	}
}

func Benchmark_ListValue(b *testing.B) {
	b.N = 1e6
	value := []interface{}{
		rand.Int63(),
		strings.Repeat("s", 100),
		[]interface{}{"a simple string", nil, rand.Int63(), []byte{12, 198, 211}},
		map[interface{}]interface{}{
			rand.Int63(): rand.Int63(),
			nil:          1,
			"s":          491871,
			15892987:     strings.Repeat("s", 100),
			"s2":         []interface{}{"a simple string", nil, rand.Int63(), []byte{12, 198, 211}},
		},
	}
	for i := 0; i < b.N; i++ {
		__value = NewListValue(value)
	}
}

func Benchmark_JsonMapValue(b *testing.B) {
	b.N = 1e6
	value := map[string]interface{}{
		strings.Repeat("a", 16): rand.Int63(),
		strings.Repeat("b", 16): strings.Repeat("s", 100),
		strings.Repeat("c", 16): []interface{}{"a simple string", nil, rand.Int63(), []byte{12, 198, 211}},
		strings.Repeat("d", 16): map[interface{}]interface{}{
			rand.Int63(): rand.Int63(),
			nil:          1,
			"s":          491871,
			15892987:     strings.Repeat("s", 100),
			"s2":         []interface{}{"a simple string", nil, rand.Int63(), []byte{12, 198, 211}},
		},
	}
	for i := 0; i < b.N; i++ {
		__value = NewValue(value)
	}
}

func Benchmark_IfcMapValue(b *testing.B) {
	b.N = 1e6
	value := map[interface{}]interface{}{
		strings.Repeat("a", 16): rand.Int63(),
		strings.Repeat("b", 16): strings.Repeat("s", 100),
		strings.Repeat("c", 16): []interface{}{"a simple string", nil, rand.Int63(), []byte{12, 198, 211}},
		strings.Repeat("d", 16): map[interface{}]interface{}{
			rand.Int63(): rand.Int63(),
			nil:          1,
			"s":          491871,
			15892987:     strings.Repeat("s", 100),
			"s2":         []interface{}{"a simple string", nil, rand.Int63(), []byte{12, 198, 211}},
		},
	}
	for i := 0; i < b.N; i++ {
		__value = NewValue(value)
	}
}

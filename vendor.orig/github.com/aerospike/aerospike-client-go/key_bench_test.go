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
	"math/rand"
	"strings"
	"testing"

	"github.com/aerospike/aerospike-client-go/pkg/ripemd160"
)

var res []byte = make([]byte, 20)

func doTheHash(buf []byte, b *testing.B) {
	hash := ripemd160.New()
	for i := 0; i < b.N; i++ {
		hash.Reset()
		hash.Write(buf)
		hash.Sum(res)
	}
}

func Benchmark_Key_Hash_S_______1(b *testing.B) {
	buffer := []byte(strings.Repeat("s", 1))
	doTheHash(buffer, b)
}

func Benchmark_Key_Hash_S______10(b *testing.B) {
	buffer := []byte(strings.Repeat("s", 10))
	doTheHash(buffer, b)
}

func Benchmark_Key_Hash_S_____100(b *testing.B) {
	buffer := []byte(strings.Repeat("s", 100))
	doTheHash(buffer, b)
}

func Benchmark_Key_Hash_S____1000(b *testing.B) {
	buffer := []byte(strings.Repeat("s", 1000))
	doTheHash(buffer, b)
}

func Benchmark_Key_Hash_S__10_000(b *testing.B) {
	buffer := []byte(strings.Repeat("s", 10000))
	doTheHash(buffer, b)
}

func Benchmark_Key_Hash_S_100_000(b *testing.B) {
	buffer := []byte(strings.Repeat("s", 100000))
	doTheHash(buffer, b)
}

var _key *Key

func makeKeys(val interface{}, b *testing.B) {
	var err error
	for i := 0; i < b.N; i++ {
		_key, err = NewKey("ns", "set", val)
		if err != nil {
			panic(err)
		}
	}
}

func Benchmark_NewKey_String______1(b *testing.B) {
	buffer := strings.Repeat("s", 1)
	makeKeys(buffer, b)
}

func Benchmark_NewKey_String_____10(b *testing.B) {
	buffer := strings.Repeat("s", 10)
	makeKeys(buffer, b)
}

func Benchmark_NewKey_String____100(b *testing.B) {
	buffer := strings.Repeat("s", 100)
	makeKeys(buffer, b)
}

func Benchmark_NewKey_String___1000(b *testing.B) {
	buffer := strings.Repeat("s", 1000)
	makeKeys(buffer, b)
}

func Benchmark_NewKey_String__10000(b *testing.B) {
	buffer := strings.Repeat("s", 10000)
	makeKeys(buffer, b)
}

func Benchmark_NewKey_String_100000(b *testing.B) {
	buffer := strings.Repeat("s", 100000)
	makeKeys(buffer, b)
}
func Benchmark_NewKey_Byte______1(b *testing.B) {
	buffer := []byte(strings.Repeat("s", 1))
	makeKeys(buffer, b)
}

func Benchmark_NewKey_Byte_____10(b *testing.B) {
	buffer := []byte(strings.Repeat("s", 10))
	makeKeys(buffer, b)
}

func Benchmark_NewKey_Byte____100(b *testing.B) {
	buffer := []byte(strings.Repeat("s", 100))
	makeKeys(buffer, b)
}

func Benchmark_NewKey_Byte___1000(b *testing.B) {
	buffer := []byte(strings.Repeat("s", 1000))
	makeKeys(buffer, b)
}

func Benchmark_NewKey_Byte__10000(b *testing.B) {
	buffer := []byte(strings.Repeat("s", 10000))
	makeKeys(buffer, b)
}

func Benchmark_NewKey_Byte_100000(b *testing.B) {
	buffer := []byte(strings.Repeat("s", 100000))
	makeKeys(buffer, b)
}

func Benchmark_NewKey_________Int(b *testing.B) {
	makeKeys(rand.Int63(), b)
}

func Benchmark_NewKey_____Float64(b *testing.B) {
	makeKeys(rand.Float64(), b)
}

func Benchmark_NewKey_List_No_Reflect(b *testing.B) {
	list := []interface{}{
		strings.Repeat("s", 1e3),
		strings.Repeat("s", 1e3),
		strings.Repeat("s", 1e3),
		strings.Repeat("s", 1e3),
		strings.Repeat("s", 1e3),
		strings.Repeat("s", 1e3),
	}
	makeKeys(list, b)
}

func Benchmark_NewKey_List_With_Reflect(b *testing.B) {
	list := []string{
		strings.Repeat("s", 1e3),
		strings.Repeat("s", 1e3),
		strings.Repeat("s", 1e3),
		strings.Repeat("s", 1e3),
		strings.Repeat("s", 1e3),
		strings.Repeat("s", 1e3),
	}
	makeKeys(list, b)
}

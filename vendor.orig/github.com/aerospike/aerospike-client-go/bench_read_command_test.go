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
	"runtime"
	"strings"
	"testing"
	// "time"

	_ "net/http/pprof"
	// . "github.com/aerospike/aerospike-client-go"
)

func doGet(set string, value interface{}, b *testing.B) {
	var err error
	policy := NewPolicy()

	dataBuffer := make([]byte, 1024*1024)

	binNames := []string{}
	key, _ := NewKey("test", set, 1000)

	for i := 0; i < b.N; i++ {
		command := newReadCommand(nil, policy, key, binNames)
		command.baseCommand.dataBuffer = dataBuffer
		err = command.writeBuffer(&command)
		if err != nil {
			panic(err)
		}
	}
}

func Benchmark_ReadCommand_________Int64(b *testing.B) {
	set := "put_bench_integer"
	value := rand.Int63()
	b.N = 1000
	runtime.GC()
	b.ResetTimer()
	doGet(set, value, b)
}

func Benchmark_ReadCommand_________Int32(b *testing.B) {
	set := "put_bench_integer"
	value := rand.Int31()
	b.N = 1000
	runtime.GC()
	b.ResetTimer()
	doGet(set, value, b)
}

func Benchmark_ReadCommand_String______1(b *testing.B) {
	set := "put_bench_str_1"
	value := strings.Repeat("s", 1)
	b.N = 1000
	runtime.GC()
	b.ResetTimer()
	doGet(set, value, b)
}

func Benchmark_ReadCommand_String_____10(b *testing.B) {
	set := "put_bench_str_10"
	value := strings.Repeat("s", 10)
	b.N = 1000
	runtime.GC()
	b.ResetTimer()
	doGet(set, value, b)
}

func Benchmark_ReadCommand_String____100(b *testing.B) {
	set := "put_bench_str_100"
	value := strings.Repeat("s", 100)
	b.N = 1000
	runtime.GC()
	b.ResetTimer()
	doGet(set, value, b)
}

func Benchmark_ReadCommand_String___1000(b *testing.B) {
	set := "put_bench_str_1000"
	value := strings.Repeat("s", 1000)
	b.N = 1000
	runtime.GC()
	b.ResetTimer()
	doGet(set, value, b)
}

func Benchmark_ReadCommand_String__10000(b *testing.B) {
	set := "put_bench_str_10000"
	value := strings.Repeat("s", 10000)
	b.N = 1000
	runtime.GC()
	b.ResetTimer()
	doGet(set, value, b)
}

func Benchmark_ReadCommand_String_100000(b *testing.B) {
	set := "put_bench_str_10000"
	value := strings.Repeat("s", 100000)
	b.N = 1000
	runtime.GC()
	b.ResetTimer()
	doGet(set, value, b)
}

func Benchmark_ReadCommand_Complex_Array(b *testing.B) {
	set := "put_bench_str_10000"
	value := []interface{}{1, 1, 1, "a simple string", nil, rand.Int63(), []byte{12, 198, 211}}
	b.N = 1000
	runtime.GC()
	b.ResetTimer()
	doGet(set, value, b)
}

func Benchmark_ReadCommand_Complex_Map(b *testing.B) {
	set := "put_bench_str_10000"
	value := map[interface{}]interface{}{
		rand.Int63(): rand.Int63(),
		nil:          1,
		"s":          491871,
		15892987:     strings.Repeat("s", 100),
		"s2":         []interface{}{"a simple string", nil, rand.Int63(), []byte{12, 198, 211}},
	}
	b.N = 1000
	runtime.GC()
	b.ResetTimer()
	doGet(set, value, b)
}

func Benchmark_ReadCommand_JSON_Map(b *testing.B) {
	set := "put_bench_str_10000"
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
	b.N = 1000
	runtime.GC()
	b.ResetTimer()
	doGet(set, value, b)
}

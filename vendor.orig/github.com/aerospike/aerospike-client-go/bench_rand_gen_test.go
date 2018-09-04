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
	"testing"
	"time"

	xor "github.com/aerospike/aerospike-client-go/types/rand"
)

func Benchmark_math_rand(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < b.N; i++ {
		r.Int63()
	}
}

func Benchmark_xor_rand(b *testing.B) {
	r := xor.NewXorRand()
	for i := 0; i < b.N; i++ {
		r.Int64()
	}
}

func Benchmark_math_rand_with_new(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Int63()
	}
}

func Benchmark_xor_rand_with_new(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := xor.NewXorRand()
		r.Int64()
	}
}

func Benchmark_math_rand_synched(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rand.Int63()
	}
}

func Benchmark_xor_rand_fast_pool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		xor.Int64()
	}
}

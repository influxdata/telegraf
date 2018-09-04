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

package aerospike_test

import (
	"runtime"

	. "github.com/aerospike/aerospike-client-go"

	"testing"
)

var r *Record
var rs []*Record
var err error

type OBJECT struct {
	Price  int
	DBName string
	// Blob   []byte
	Blob []int64
}

func benchGet(times int, client *Client, key *Key) {
	for i := 0; i < times; i++ {
		if r, err = client.Get(nil, key); err != nil {
			panic(err)
		}
	}
}

func benchBatchGet(times int, client *Client, keys []*Key) {
	for i := 0; i < times; i++ {
		if rs, err = client.BatchGet(nil, keys); err != nil {
			panic(err)
		}
	}
}

func benchPut(times int, client *Client, key *Key, wp *WritePolicy) {
	dbName := NewBin("dbname", "CouchDB")
	price := NewBin("price", 0)
	keywords := NewBin("keywords", []string{"concurrent", "fast"})
	for i := 0; i < times; i++ {
		if err = client.PutBins(wp, key, dbName, price, keywords); err != nil {
			panic(err)
		}
	}
}

func Benchmark_Get(b *testing.B) {
	client, err := NewClientWithPolicy(clientPolicy, *host, *port)
	if err != nil {
		b.Fail()
	}

	key, _ := NewKey("test", "test", "Aerospike")
	// obj := &OBJECT{198, "Jack Shaftoe and Company", []byte(bytes.Repeat([]byte{32}, 1000))}
	// obj := &OBJECT{198, "Jack Shaftoe and Company", []int64{1}}
	client.Delete(nil, key)
	client.PutBins(nil, key, NewBin("b", []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, "a", "b"}))
	// client.PutBins(nil, key, NewBin("b", 1))
	// client.PutObject(nil, key, &obj)

	b.N = 100
	runtime.GC()
	b.ResetTimer()
	benchGet(b.N, client, key)
}

func Benchmark_Put(b *testing.B) {
	client, err := NewClient(*host, *port)
	if err != nil {
		b.Fail()
	}

	key, _ := NewKey("test", "test", "Aerospike")
	writepolicy := NewWritePolicy(0, 0)

	b.N = 100
	runtime.GC()
	b.ResetTimer()
	benchPut(b.N, client, key, writepolicy)
}

func Benchmark_BatchGet(b *testing.B) {
	client, err := NewClientWithPolicy(clientPolicy, *host, *port)
	if err != nil {
		b.Fail()
	}

	var keys []*Key
	for i := 0; i < 10; i++ {
		key, _ := NewKey("test", "test", i)
		if err := client.PutBins(nil, key, NewBin("b", 1)); err == nil {
			keys = append(keys, key)
		}
	}

	b.N = 1e4
	runtime.GC()
	b.ResetTimer()
	benchBatchGet(b.N, client, keys)
}

// Copyright 2013-2016 Aerospike, Inc.
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

package main

import (
	// "fmt"

	. "github.com/aerospike/aerospike-client-go"
)

type Person struct {
	name string
}

// Define The AerospikeBlob interface
func (p Person) EncodeBlob() ([]byte, error) {
	return append([]byte(p.name)), nil
}

// Decoder is optional, and should be used manually
func (p *Person) DecodeBlob(buf []byte) error {
	p.name = string(buf)
	return nil
}

func main() {
	// define a client to connect to
	client, err := NewClient("127.0.0.1", 3000)
	panicOnError(err)

	namespace := "test"
	setName := "people"
	key, err := NewKey(namespace, setName, "key") // user key can be of any supported type
	panicOnError(err)

	// define some bins
	bins := BinMap{
		"bin1": Person{name: "Albert Einstein"},
		"bin2": &Person{name: "Richard Feynman"},
	}

	// write the bins
	writePolicy := NewWritePolicy(0, 0)
	err = client.Put(writePolicy, key, bins)
	panicOnError(err)

	// read it back!
	readPolicy := NewPolicy()
	rec, err := client.Get(readPolicy, key)
	panicOnError(err)

	result := &Person{}

	// decode first object
	result.DecodeBlob(rec.Bins["bin1"].([]byte))

	// decode second object
	result.DecodeBlob(rec.Bins["bin2"].([]byte))
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

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
	"fmt"

	. "github.com/aerospike/aerospike-client-go"
)

func main() {
	// define a client to connect to
	client, err := NewClient("127.0.0.1", 3000)
	panicOnError(err)

	namespace := "test"
	setName := "aerospike"
	key, err := NewKey(namespace, setName, "key") // user key can be of any supported type
	panicOnError(err)

	// define some bins
	bins := BinMap{
		"bin1": 42, // you can pass any supported type as bin value
		"bin2": "An elephant is a mouse with an operating system",
		"bin3": []interface{}{"Go", 17981},
	}

	// write the bins
	writePolicy := NewWritePolicy(0, 0)
	err = client.Put(writePolicy, key, bins)
	panicOnError(err)

	// read it back!
	readPolicy := NewPolicy()
	rec, err := client.Get(readPolicy, key)
	panicOnError(err)

	fmt.Printf("%#v\n", *rec)

	// Add to bin1
	err = client.Add(writePolicy, key, BinMap{"bin1": 1})
	panicOnError(err)

	rec2, err := client.Get(readPolicy, key)
	panicOnError(err)

	fmt.Printf("value of %s: %v\n", "bin1", rec2.Bins["bin1"])

	// prepend and append to bin2
	err = client.Prepend(writePolicy, key, BinMap{"bin2": "Frankly:  "})
	panicOnError(err)
	err = client.Append(writePolicy, key, BinMap{"bin2": "."})
	panicOnError(err)

	rec3, err := client.Get(readPolicy, key)
	panicOnError(err)

	fmt.Printf("value of %s: %v\n", "bin2", rec3.Bins["bin2"])

	// delete bin3
	err = client.Put(writePolicy, key, BinMap{"bin3": nil})
	rec4, err := client.Get(readPolicy, key)
	panicOnError(err)

	fmt.Printf("bin3 does not exist anymore: %#v\n", *rec4)

	// check if key exists
	exists, err := client.Exists(readPolicy, key)
	panicOnError(err)
	fmt.Printf("key exists in the database: %#v\n", exists)

	// delete the key, and check if key exists
	existed, err := client.Delete(writePolicy, key)
	panicOnError(err)
	fmt.Printf("did key exist before delete: %#v\n", existed)

	exists, err = client.Exists(readPolicy, key)
	panicOnError(err)
	fmt.Printf("key exists: %#v\n", exists)
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

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
	"fmt"
	"log"

	as "github.com/aerospike/aerospike-client-go"
)

/*
	myListString
*/
var _ as.ListIter = myListString([]string{})

// your custom list
type myListString []string

func (ml myListString) PackList(buf as.BufferEx) (int, error) {
	size := 0
	for _, elem := range ml {
		n, err := as.PackString(buf, elem)
		size += n
		if err != nil {
			return size, err
		}
	}
	return size, nil
}

func (ml myListString) Len() int {
	return len(ml)
}

func ExampleListIter_string() {
	// Setup the client here
	// client, err := as.NewClient("127.0.0.1", 3000)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	var v as.Value = as.NewValue(myListString([]string{"a", "b", "c"}))
	key, err := as.NewKey("test", "test", 1)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Put(nil, key, as.BinMap{"myBin": v})
	if err != nil {
		log.Fatal(err)
	}

	rec, err := client.Get(nil, key)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rec.Bins["myBin"])
	// Output:
	// [a b c]
}

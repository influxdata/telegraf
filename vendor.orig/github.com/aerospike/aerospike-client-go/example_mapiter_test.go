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
	"time"

	as "github.com/aerospike/aerospike-client-go"
)

/*
	myMapStringTime
*/
var _ as.MapIter = myMapStringTime(map[string]time.Time{})

// your custom list
type myMapStringTime map[string]time.Time

func (mm myMapStringTime) PackMap(buf as.BufferEx) (int, error) {
	size := 0
	for key, val := range mm {
		n, err := as.PackString(buf, key)
		size += n
		if err != nil {
			return size, err
		}

		n, err = as.PackInt64(buf, val.UnixNano())
		size += n
		if err != nil {
			return size, err
		}
	}
	return size, nil
}

func (mm myMapStringTime) Len() int {
	return len(mm)
}

func ExampleMapIter() {
	// Setup the client here
	// client, err := as.NewClient("127.0.0.1", 3000)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	now := time.Unix(123123123, 0)
	var v as.Value = as.NewValue(myMapStringTime(map[string]time.Time{"now": now}))
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
	// map[now:123123123000000000]
}

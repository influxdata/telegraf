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
	"log"
	"time"

	. "github.com/aerospike/aerospike-client-go"
)

func main() {
	// remove timestamps from log messages
	log.SetFlags(0)

	// connect to the host
	if conn, err := NewConnection("localhost:3000", 10*time.Second); err != nil {
		log.Fatalln(err.Error())
	} else {
		if infoMap, err := RequestInfo(conn, ""); err != nil {
			log.Fatalln(err.Error())
		} else {
			cnt := 1
			for k, v := range infoMap {
				log.Printf("%d :  %s\n     %s\n\n", cnt, k, v)
				cnt++
			}
		}
	}
}

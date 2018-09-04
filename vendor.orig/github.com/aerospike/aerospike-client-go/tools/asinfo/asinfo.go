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
	"flag"
	"log"
	"os"
	"strings"
	"time"

	as "github.com/aerospike/aerospike-client-go"
)

var (
	host     = flag.String("h", "127.0.0.1", "host (default 127.0.0.1)")
	port     = flag.Int("p", 3000, "port (default 3000)")
	value    = flag.String("v", "", "(fetch single value - default all)")
	user     = flag.String("U", "", "User.")
	password = flag.String("P", "", "Password.")

	clientPolicy *as.ClientPolicy
)

func main() {
	flag.Parse()
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	clientPolicy = as.NewClientPolicy()
	if *user != "" {
		clientPolicy.User = *user
		clientPolicy.Password = *password
	}
	*value = strings.Trim(*value, " ")

	// connect to the host
	client, err := as.NewClientWithPolicy(clientPolicy, *host, *port)
	dieIfError(err)

	node := client.GetNodes()[0]
	conn, err := node.GetConnection(time.Second)
	dieIfError(err)

	infoMap, err := as.RequestInfo(conn, *value)
	dieIfError(err, func() {
		node.InvalidateConnection(conn)
	})

	node.PutConnection(conn)

	if len(infoMap) == 0 {
		log.Printf("Query successful, no information for -v \"%s\"\n\n", *value)
		return
	}

	outfmt := "%d :  %s\n     %s\n"
	cnt := 1
	for k, v := range infoMap {
		log.Printf(outfmt, cnt, k, v)
		cnt++
	}
}

// dieIfError calls each callback in turn before printing the error via log.Fatalln.
func dieIfError(err error, cleanup ...func()) {
	if err != nil {
		log.Println("Error:")
		for _, cb := range cleanup {
			cb()
		}
		log.Fatalln(err.Error())
	}
	return
}

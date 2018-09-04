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
	// "time"
	"runtime"

	. "github.com/aerospike/aerospike-client-go"
	. "github.com/aerospike/aerospike-client-go/logger"
)

// flag information
var host = flag.String("h", "127.0.0.1", "Aerospike server seed hostnames or IP addresses")
var port = flag.Int("p", 3000, "Aerospike server seed hostname or IP address port number.")
var namespace = flag.String("n", "test", "Aerospike namespace.")
var set = flag.String("s", "testset", "Aerospike set name.")
var operand = flag.String("o", "get", "Operand: get, set, delete")
var binName = flag.String("b", "bin", "Bin name")
var key = flag.String("k", "key", "Key information")
var recordTTL = flag.Int("e", 0, "Record TTL in seconds")
var value = flag.String("v", "", "Value information; used only by get operand")
var verbose = flag.Bool("verbose", false, "Verbose mode")
var showUsage = flag.Bool("u", false, "Show usage information.")

func quitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.SetOutput(os.Stdout)
	// use all cpus in the system for concurrency
	runtime.GOMAXPROCS(runtime.NumCPU())
	// remove timestamp from log messages
	log.SetFlags(0)

	readFlags()

	// connect to server
	client, err := NewClient(*host, *port)
	quitOnError(err)

	theKey, err := NewKey(*namespace, *set, *key)
	quitOnError(err)
	switch *operand {
	case "get":
		policy := NewPolicy()
		rec, err := client.Get(policy, theKey, *binName)
		quitOnError(err)

		if rec != nil {
			log.Println(rec.Bins[*binName])
		} else {
			log.Println("Record not found.")
		}
	case "set":
		policy := NewWritePolicy(0, uint32(*recordTTL))
		err = client.Put(policy, theKey, BinMap{*binName: *value})
		quitOnError(err)
	case "delete":
		existed, err := client.Delete(nil, theKey)
		quitOnError(err)
		if existed {
			log.Println("key deleted successfully.")
		} else {
			log.Println("key didn't exist.")
		}
	}
}

func readFlags() {
	flag.Parse()

	if *showUsage {
		flag.Usage()
		os.Exit(0)
	}

	if *verbose {
		Logger.SetLevel(INFO)
	}

	*operand = strings.ToLower(*operand)
	switch *operand {
	case "get":
		if *key == "" {
			log.Fatalln("Key is required for get operation.")
		}
	case "set":
		if *key == "" {
			log.Fatalln("Key is required for get operation.")
		}
	case "delete":
		if *key == "" {
			log.Fatalln("Key is required for get operation.")
		}
	default:
		log.Fatalf("operand %s not recognized. valid values: get, set, delete.\n", *operand)
	}

}

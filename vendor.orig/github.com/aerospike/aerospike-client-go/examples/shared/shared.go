/*
 * Copyright 2012-2016 Aerospike, Inc.
 *
 * Portions may be licensed to Aerospike, Inc. under one or more contributor
 * license agreements.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License. You may obtain a copy of
 * the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations under
 * the License.
 */

package shared

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"

	as "github.com/aerospike/aerospike-client-go"
)

var WritePolicy = as.NewWritePolicy(0, 0)
var Policy = as.NewPolicy()

var Host = flag.String("h", "127.0.0.1", "Aerospike server seed hostnames or IP addresses")
var Port = flag.Int("p", 3000, "Aerospike server seed hostname or IP address port number.")
var Namespace = flag.String("n", "test", "Aerospike namespace.")
var Set = flag.String("s", "testset", "Aerospike set name.")
var showUsage = flag.Bool("u", false, "Show usage information.")
var Client *as.Client

func ValidateBin(key *as.Key, bin *as.Bin, record *as.Record) {
	if !bytes.Equal(record.Key.Digest(), key.Digest()) {
		log.Fatalln(fmt.Sprintf("key `%s` is not the same as key `%s`.", key, record.Key))
	}

	if record.Bins[bin.Name] != bin.Value.GetObject() {
		log.Fatalln(fmt.Sprintf("bin `%s: %v` is not the same as bin `%v`.", bin.Name, bin.Value, record.Bins[bin.Name]))
	}
}

func PanicOnError(err error) {
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func printParams() {
	log.Printf("hosts:\t\t%s", *Host)
	log.Printf("port:\t\t%d", *Port)
	log.Printf("namespace:\t\t%s", *Namespace)
	log.Printf("set:\t\t%s", *Set)
}

func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	_div := math.Copysign(div, val)
	_roundOn := math.Copysign(roundOn, val)
	if _div >= _roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

// reads input flags and interprets the complex ones
func init() {
	// use all cpus in the system for concurrency
	runtime.GOMAXPROCS(runtime.NumCPU())

	log.SetOutput(os.Stdout)

	flag.Parse()

	if *showUsage {
		flag.Usage()
		os.Exit(0)
	}

	printParams()

	var err error
	Client, err = as.NewClient(*Host, *Port)
	if err != nil {
		PanicOnError(err)
	}
}

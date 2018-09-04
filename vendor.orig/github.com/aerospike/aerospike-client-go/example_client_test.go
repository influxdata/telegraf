/*
 * Copyright 2013-2017 Aerospike, Inc.
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

package aerospike_test

import (
	"fmt"
	"log"

	as "github.com/aerospike/aerospike-client-go"
)

func ExampleClient_Add() {
	key, err := as.NewKey("test", "test", "addkey")
	if err != nil {
		log.Fatal(err)
	}

	if _, err = client.Delete(nil, key); err != nil {
		log.Fatal(err)
	}

	// Add to a non-existing record/bin, should create a record
	bin := as.NewBin("bin", 10)
	if err = client.AddBins(nil, key, bin); err != nil {
		log.Fatal(err)
	}

	// Add to 5 to the original 10
	bin = as.NewBin("bin", 5)
	if err = client.AddBins(nil, key, bin); err != nil {
		log.Fatal(err)
	}

	// Check the result
	record, err := client.Get(nil, key, bin.Name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(record.Bins["bin"])

	// Demonstrate add and get combined.
	bin = as.NewBin("bin", 30)
	if record, err = client.Operate(nil, key, as.AddOp(bin), as.GetOp()); err != nil {
		log.Fatal(err)
	}

	fmt.Println(record.Bins["bin"])
	// Output:
	// 15
	// 45
}

func ExampleClient_Append() {
	key, err := as.NewKey("test", "test", "appendkey")
	if err != nil {
		log.Fatal(err)
	}

	if _, err = client.Delete(nil, key); err != nil {
		log.Fatal(err)
	}

	// Create by appending to non-existing value
	bin1 := as.NewBin("myBin", "Hello")
	if err = client.AppendBins(nil, key, bin1); err != nil {
		log.Fatal(err)
	}

	// Append World
	bin2 := as.NewBin("myBin", " World")
	if err = client.AppendBins(nil, key, bin2); err != nil {
		log.Fatal(err)
	}

	record, err := client.Get(nil, key)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(record.Bins["myBin"])
	// Output:
	// Hello World
}

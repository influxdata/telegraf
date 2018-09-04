/*
 * Copyright 2012-2016 Aerospike, Inc.
 *
 * Portions may be licensed to Aerospike, Inc. under one or more contributor
 * license agreements.
 *
 * Licensed under the Apache License, Version 2.0 (the "License") you may not
 * use this file except in compliance with the License. You may obtain a copy of
 * the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations under
 * the License.
 */
package main

import (
	"fmt"
	"log"
	"time"

	as "github.com/aerospike/aerospike-client-go"
	shared "github.com/aerospike/aerospike-client-go/examples/shared"
)

func main() {
	runSingleBinExample(shared.Client)
	runMultiBinExample(shared.Client)
	runGetHeaderExample(shared.Client)

	log.Println("Example finished successfully.")
}

// Execute put and get on a server configured as multi-bin.  This is the server default.
func runMultiBinExample(client *as.Client) {
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "putgetkey")
	bin1 := as.NewBin("bin1", "value1")
	bin2 := as.NewBin("bin2", "value2")

	log.Printf("Put: namespace=%s set=%s key=%s bin1=%s value1=%s bin2=%s value2=%s",
		key.Namespace(), key.SetName(), key.Value(), bin1.Name, bin1.Value, bin2.Name, bin2.Value)

	client.PutBins(shared.WritePolicy, key, bin1, bin2)

	log.Printf("Get: namespace=%s set=%s key=%s", key.Namespace(), key.SetName(), key.Value())

	record, err := client.Get(shared.Policy, key)
	shared.PanicOnError(err)

	if record == nil {
		panic(fmt.Errorf("Failed to get: namespace=%s set=%s key=%s", key.Namespace(), key.SetName(), key.Value()))
	}

	shared.ValidateBin(key, bin1, record)
	shared.ValidateBin(key, bin2, record)
}

// Execute put and get on a server configured as single-bin.
func runSingleBinExample(client *as.Client) {
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "putgetkey")
	bin := as.NewBin("", "value")

	log.Printf("Single Put: namespace=%s set=%s key=%s value=%s",
		key.Namespace(), key.SetName(), key.Value(), bin.Value)

	client.PutBins(shared.WritePolicy, key, bin)

	log.Printf("Single Get: namespace=%s set=%s key=%s", key.Namespace(), key.SetName(), key.Value())

	record, err := client.Get(shared.Policy, key)
	shared.PanicOnError(err)

	if record == nil {
		panic(fmt.Errorf("Failed to get: namespace=%s set=%s key=%s", key.Namespace(), key.SetName(), key.Value()))
	}

	shared.ValidateBin(key, bin, record)
}

// Read record header data.
func runGetHeaderExample(client *as.Client) {
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "putgetkey")

	log.Printf("Get record header: namespace=%s set=%s key=%s", key.Namespace(), key.SetName(), key.Value())
	record, err := client.GetHeader(shared.Policy, key)
	shared.PanicOnError(err)

	if record == nil {
		panic(fmt.Errorf("Failed to get: namespace=%s set=%s key=%s", key.Namespace(), key.SetName(), key.Value()))
	}

	// Generation should be greater than zero.  Make sure it's populated.
	if record.Generation == 0 {
		panic(fmt.Errorf("Invalid record header: generation=%d expiration=%d", record.Generation, record.Expiration))
	}
	log.Printf("Received: generation=%d expiration=%d (%s)\n", record.Generation, record.Expiration, time.Duration(record.Expiration)*time.Second)
}

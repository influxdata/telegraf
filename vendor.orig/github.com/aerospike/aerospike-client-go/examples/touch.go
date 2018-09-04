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
package main

import (
	"log"
	"time"

	as "github.com/aerospike/aerospike-client-go"
	shared "github.com/aerospike/aerospike-client-go/examples/shared"
)

func main() {
	runExample(shared.Client)
	log.Println("Example finished successfully.")
}

func runExample(client *as.Client) {
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "touchkey")
	bin := as.NewBin("touchbin", "touchvalue")

	log.Printf("Create record with 2 second expiration.")
	writePolicy := as.NewWritePolicy(0, 2)
	client.PutBins(writePolicy, key, bin)

	log.Printf("Touch same record with 5 second expiration.")
	writePolicy.Expiration = 5
	record, err := client.Operate(writePolicy, key, as.TouchOp(), as.GetHeaderOp())

	if record == nil {
		log.Fatalf(
			"Failed to get: namespace=%s set=%s key=%s bin=%s value=nil",
			key.Namespace(), key.SetName(), key.Value(), bin.Name)
	}

	if record.Expiration == 0 {
		log.Fatalf(
			"Failed to get record expiration: namespace=%s set=%s key=%s",
			key.Namespace(), key.SetName(), key.Value())
	}

	log.Printf("Sleep 3 seconds.")
	time.Sleep(3 * time.Second)

	record, err = client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)

	if record == nil {
		log.Fatalf(
			"Failed to get: namespace=%s set=%s key=%s",
			key.Namespace(), key.SetName(), key.Value())
	}

	log.Printf("Success. Record still exists.")
	log.Printf("Sleep 4 seconds.")
	time.Sleep(4 * time.Second)

	record, err = client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)

	if record == nil {
		log.Printf("Success. Record expired as expected.")
	} else {
		log.Fatalf("Found record when it should have expired.")
	}
}

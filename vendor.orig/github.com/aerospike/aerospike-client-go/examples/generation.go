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
	"errors"
	"log"

	as "github.com/aerospike/aerospike-client-go"
	shared "github.com/aerospike/aerospike-client-go/examples/shared"
	ast "github.com/aerospike/aerospike-client-go/types"
)

func main() {
	runExample(shared.Client)

	log.Println("Example finished successfully.")
}

func runExample(client *as.Client) {
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "genkey")
	binName := "genbin"

	// Delete record if it already exists.
	client.Delete(shared.WritePolicy, key)

	// Set some values for the same record.
	bin := as.NewBin(binName, "genvalue1")
	log.Printf("Put: namespace=%s set=%s key=%s bin=%s value=%s",
		key.Namespace(), key.SetName(), key.Value(), bin.Name, bin.Value)

	client.PutBins(shared.WritePolicy, key, bin)

	bin = as.NewBin(binName, "genvalue2")
	log.Printf("Put: namespace=%s set=%s key=%s bin=%s value=%s",
		key.Namespace(), key.SetName(), key.Value(), bin.Name, bin.Value)

	client.PutBins(shared.WritePolicy, key, bin)

	// Retrieve record and its generation count.
	record, err := client.Get(shared.Policy, key, bin.Name)

	if record == nil {
		log.Fatalf(
			"Failed to get: namespace=%s set=%s key=%s",
			key.Namespace(), key.SetName(), key.Value())
	}

	received := record.Bins[bin.Name]
	expected := bin.Value.String()

	if received == expected {
		log.Printf("Get successful: namespace=%s set=%s key=%s bin=%s value=%s generation=%d",
			key.Namespace(), key.SetName(), key.Value(), bin.Name, received, record.Generation)
	} else {
		log.Fatalf("Get mismatch: Expected %s. Received %s.",
			expected, received)
	}

	// Set record and fail if it's not the expected generation.
	bin = as.NewBin(binName, "genvalue3")
	log.Printf("Put: namespace=%s set=%s key=%s bin=%s value=%s expected generation=%d",
		key.Namespace(), key.SetName(), key.Value(), bin.Name, bin.Value, record.Generation)

	writePolicy := as.NewWritePolicy(0, 2)
	writePolicy.GenerationPolicy = as.EXPECT_GEN_EQUAL
	writePolicy.Generation = record.Generation
	client.PutBins(writePolicy, key, bin)

	// Set record with invalid generation and check results .
	bin = as.NewBin(binName, "genvalue4")
	writePolicy.Generation = 9999
	log.Printf("Put: namespace=%s set=%s key=%s bin=%s value=%s expected generation=%d",
		key.Namespace(), key.SetName(), key.Value(), bin.Name, bin.Value, writePolicy.Generation)

	err = client.PutBins(writePolicy, key, bin)
	if err != nil {
		if ae, ok := err.(ast.AerospikeError); ok && ae.ResultCode() != ast.GENERATION_ERROR {
			shared.PanicOnError(errors.New("Should have received generation error instead of success."))
		}
		log.Printf("Success: Generation error returned as expected.")
	} else {
		log.Fatalf(
			"Unexpected set return code: namespace=%s set=%s key=%s bin=%s value=%s code=%s",
			key.Namespace(), key.SetName(), key.Value(), bin.Name, bin.Value, err)
	}

	// Verify results.
	record, err = client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)

	if record == nil {
		log.Fatalf(
			"Failed to get: namespace=%s set=%s key=%s",
			key.Namespace(), key.SetName(), key.Value())
	}

	received = record.Bins[bin.Name]
	expected = "genvalue3"

	if received == expected {
		log.Printf("Get successful: namespace=%s set=%s key=%s bin=%s value=%s generation=%d",
			key.Namespace(), key.SetName(), key.Value(), bin.Name, received, record.Generation)
	} else {
		log.Fatalf("Get mismatch: Expected %s. Received %s.",
			expected, received)
	}
}

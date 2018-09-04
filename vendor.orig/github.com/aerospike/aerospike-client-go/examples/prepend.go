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

	as "github.com/aerospike/aerospike-client-go"
	shared "github.com/aerospike/aerospike-client-go/examples/shared"
)

func main() {
	runExample(shared.Client)
	log.Println("Example finished successfully.")
}

func runExample(client *as.Client) {
	key, err := as.NewKey(*shared.Namespace, *shared.Set, "prependkey")
	shared.PanicOnError(err)

	binName := "prependbin"

	// Delete record if it already exists.
	client.Delete(shared.WritePolicy, key)

	bin := as.NewBin(binName, "World")
	log.Println("Initial prepend will create record.  Initial value is ", bin.Value, ".")
	client.PrependBins(shared.WritePolicy, key, bin)

	bin = as.NewBin(binName, "Hello ")
	log.Println("Prepend \"", bin.Value, "\" to existing record.")
	client.PrependBins(shared.WritePolicy, key, bin)

	record, err := client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)

	if record == nil {
		log.Fatalf(
			"Failed to get: namespace=%s set=%s key=%s",
			key.Namespace(), key.SetName(), key.Value())
	}

	// The value received from the server is an unsigned byte stream.
	// Convert to an integer before comparing with expected.
	received := record.Bins[bin.Name]
	expected := "Hello World"

	if received == expected {
		log.Printf("Prepend successful: ns=%s set=%s key=%s bin=%s value=%s",
			key.Namespace(), key.SetName(), key.Value(), bin.Name, received)
	} else {
		log.Fatalf("Prepend mismatch: Expected %s. Received %s.", expected, received)
	}
}

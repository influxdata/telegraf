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
	"strconv"

	as "github.com/aerospike/aerospike-client-go"
	shared "github.com/aerospike/aerospike-client-go/examples/shared"
	asl "github.com/aerospike/aerospike-client-go/logger"
)

func main() {
	keyPrefix := "batchkey"
	valuePrefix := "batchvalue"
	binName := "batchbin"
	size := 8

	writeRecords(shared.Client, keyPrefix, binName, valuePrefix, size)
	// batchExists(shared.Client, keyPrefix, size)
	// batchReads(shared.Client, keyPrefix, binName, size)
	// batchReadHeaders(shared.Client, keyPrefix, size)

	log.Println("Example finished successfully.")
}

/**
 * Write records individually.
 */
func writeRecords(
	client *as.Client,
	keyPrefix string,
	binName string,
	valuePrefix string,
	size int,
) {
	for i := 1; i <= size; i++ {
		key, _ := as.NewKey(*shared.Namespace, *shared.Set, keyPrefix+strconv.Itoa(i))
		bin := as.NewBin(binName, valuePrefix+strconv.Itoa(i))

		log.Printf("Put: ns=%s set=%s key=%s bin=%s value=%s",
			key.Namespace(), key.SetName(), key.Value(), bin.Name, bin.Value)

		client.PutBins(shared.WritePolicy, key, bin)
	}
}

/**
 * Check existence of records in one batch.
 */
func batchExists(
	client *as.Client,

	keyPrefix string,
	size int,
) {
	// Batch into one call.
	keys := make([]*as.Key, size)
	for i := 0; i < size; i++ {
		keys[i], _ = as.NewKey(*shared.Namespace, *shared.Set, keyPrefix+strconv.Itoa(i+1))
	}

	existsArray, err := client.BatchExists(nil, keys)
	shared.PanicOnError(err)

	for i := 0; i < len(existsArray); i++ {
		key := keys[i]
		exists := existsArray[i]
		log.Printf("Record: ns=%s set=%s key=%s exists=%t",
			key.Namespace(), key.SetName(), key.Value(), exists)
	}
}

/**
 * Read records in one batch.
 */
func batchReads(
	client *as.Client,
	keyPrefix string,
	binName string,
	size int,
) {
	// Batch gets into one call.
	keys := make([]*as.Key, size)
	for i := 0; i < size; i++ {
		keys[i], _ = as.NewKey(*shared.Namespace, *shared.Set, keyPrefix+strconv.Itoa(i+1))
	}

	records, err := client.BatchGet(nil, keys, binName)
	shared.PanicOnError(err)

	for i := 0; i < len(records); i++ {
		key := keys[i]
		record := records[i]
		level := asl.ERR
		var value interface{}

		if record != nil {
			level = asl.INFO
			value = record.Bins[binName]
		}
		asl.Logger.LogAtLevel(level, "Record: ns=%s set=%s key=%s bin=%s value=%s",
			key.Namespace(), key.SetName(), key.Value(), binName, value)
	}

	if len(records) != size {
		log.Fatalf("Record size mismatch. Expected %d. Received %d.", size, len(records))
	}
}

/**
 * Read record header data in one batch.
 */
func batchReadHeaders(
	client *as.Client,
	keyPrefix string,
	size int,
) {
	// Batch gets into one call.
	keys := make([]*as.Key, size)
	for i := 0; i < size; i++ {
		keys[i], _ = as.NewKey(*shared.Namespace, *shared.Set, keyPrefix+strconv.Itoa(i+1))
	}

	records, err := client.BatchGetHeader(nil, keys)
	shared.PanicOnError(err)

	for i := 0; i < len(records); i++ {
		key := keys[i]
		record := records[i]
		level := asl.ERR
		generation := uint32(0)
		expiration := uint32(0)

		if record != nil && (record.Generation > 0 || record.Expiration > 0) {
			level = asl.INFO
			generation = record.Generation
			expiration = record.Expiration
		}
		asl.Logger.LogAtLevel(level, "Record: ns=%s set=%s key=%s generation=%d expiration=%d",
			key.Namespace(), key.SetName(), key.Value(), generation, expiration)
	}

	if len(records) != size {
		log.Fatalf("Record size mismatch. Expected %d. Received %d.", size, len(records))
	}
}

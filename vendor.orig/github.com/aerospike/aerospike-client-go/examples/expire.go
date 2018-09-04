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
	"math"
	"time"

	as "github.com/aerospike/aerospike-client-go"
	shared "github.com/aerospike/aerospike-client-go/examples/shared"
)

func main() {
	expireExample(shared.Client)
	noExpireExample(shared.Client)

	log.Println("Example finished successfully.")
}

/**
 * Write and twice read an expiration record.
 */
func expireExample(client *as.Client) {
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "expirekey ")
	bin := as.NewBin("expirebin", "expirevalue")

	log.Printf("Put: namespace=%s set=%s key=%s bin=%s value=%s expiration=2",
		key.Namespace(), key.SetName(), key.Value(), bin.Name, bin.Value)

	// Specify that record expires 2 seconds after it's written.
	writePolicy := as.NewWritePolicy(0, 2)
	client.PutBins(writePolicy, key, bin)

	// Read the record before it expires, showing it is there.
	log.Printf("Get: namespace=%s set=%s key=%s",
		key.Namespace(), key.SetName(), key.Value())

	record, err := client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)
	if record == nil {
		log.Fatalf(
			"Failed to get record: namespace=%s set=%s key=%s",
			key.Namespace(), key.SetName(), key.Value())
	}

	received := record.Bins[bin.Name]
	expected := bin.Value.String()
	if received == expected {
		log.Printf("Get record successful: namespace=%s set=%s key=%s bin=%s value=%s",
			key.Namespace(), key.SetName(), key.Value(), bin.Name, received)
	} else {
		log.Fatalf("Expire record mismatch: Expected %s. Received %s.",
			expected, received)
	}

	// Read the record after it expires, showing it's gone.
	log.Printf("Sleeping for 3 seconds ...")
	time.Sleep(3 * time.Second)
	record, err = client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)
	if record == nil {
		log.Printf("Expiry of record successful. Record not found.")
	} else {
		log.Fatalf("Found record when it should have expired.")
	}
}

/**
 * Write and twice read a non-expiring tuple using the new "NoExpire" value (-1).
 * This example is most effective when the Default Namespace Time To Live (TTL)
 * is set to a small value, such as 5 seconds.  When we sleep beyond that
 * time, we show that the NoExpire TTL flag actually works.
 */
func noExpireExample(client *as.Client) {
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "expirekey")
	bin := as.NewBin("expirebin", "noexpirevalue")

	log.Printf("Put: namespace=%s set=%s key=%s bin=%s value=%s expiration=NoExpire",
		key.Namespace(), key.SetName(), key.Value(), bin.Name, bin.Value)

	// Specify that record NEVER expires.
	// The "Never Expire" value is -1, or 0xFFFFFFFF.
	writePolicy := as.NewWritePolicy(0, 2)
	writePolicy.Expiration = math.MaxUint32
	client.PutBins(writePolicy, key, bin)

	// Read the record, showing it is there.
	log.Printf("Get: namespace=%s set=%s key=%s",
		key.Namespace(), key.SetName(), key.Value())

	record, err := client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)
	if record == nil {
		log.Fatalf(
			"Failed to get record: namespace=%s set=%s key=%s",
			key.Namespace(), key.SetName(), key.Value())
	}

	received := record.Bins[bin.Name]
	expected := bin.Value.String()
	if received == expected {
		log.Printf("Get record successful: namespace=%s set=%s key=%s bin=%s value=%s",
			key.Namespace(), key.SetName(), key.Value(), bin.Name, received)
	} else {
		log.Fatalf("Expire record mismatch: Expected %s. Received %s.",
			expected, received)
	}

	// Read this Record after the Default Expiration, showing it is still there.
	// We should have set the Namespace TTL at 5 sec.
	log.Printf("Sleeping for 10 seconds ...")
	time.Sleep(10 * time.Second)
	record, err = client.Get(shared.Policy, key, bin.Name)
	shared.PanicOnError(err)

	if record == nil {
		log.Fatalf("Record expired and should NOT have.")
	} else {
		log.Printf("Found Record (correctly) after default TTL.")
	}
}

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
	// Write initial record.
	key, _ := as.NewKey(*shared.Namespace, *shared.Set, "opkey")
	bin1 := as.NewBin("optintbin", 7)
	bin2 := as.NewBin("optstringbin", "string value")
	log.Printf("Put: namespace=%s set=%s key=%s bin1=%s value1=%s bin2=%s value2=%s",
		key.Namespace(), key.SetName(), key.Value(), bin1.Name, bin1.Value, bin2.Name, bin2.Value)
	client.PutBins(shared.WritePolicy, key, bin1, bin2)

	// Add integer, write new string and read record.
	bin3 := as.NewBin(bin1.Name, 4)
	bin4 := as.NewBin(bin2.Name, "new string")
	log.Println("Add: ", bin3.Value)
	log.Println("Write: ", bin4.Value)
	log.Println("Read:")

	record, err := client.Operate(shared.WritePolicy, key, as.AddOp(bin3), as.PutOp(bin4), as.GetOp())
	shared.PanicOnError(err)

	if record == nil {
		log.Fatalf(
			"Failed to get: namespace=%s set=%s key=%s",
			key.Namespace(), key.SetName(), key.Value())
	}

	binExpected := as.NewBin(bin3.Name, 11)
	shared.ValidateBin(key, binExpected, record)
	shared.ValidateBin(key, bin4, record)
}

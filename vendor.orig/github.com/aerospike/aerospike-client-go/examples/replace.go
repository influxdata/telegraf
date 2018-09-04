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
	ast "github.com/aerospike/aerospike-client-go/types"
)

func main() {
	runReplaceExample(shared.Client)
	runReplaceOnlyExample(shared.Client)

	log.Println("Example finished successfully.")
}

func runReplaceExample(client *as.Client) {
	key, err := as.NewKey(*shared.Namespace, *shared.Set, "replacekey")
	shared.PanicOnError(err)
	bin1 := as.NewBin("bin1", "value1")
	bin2 := as.NewBin("bin2", "value2")
	bin3 := as.NewBin("bin3", "value3")

	log.Printf("Put: namespace=%s set=%s key=%s bin1=%s value1=%s bin2=%s value2=%s",
		key.Namespace(), key.SetName(), key.Value(), bin1.Name, bin1.Value, bin2.Name, bin2.Value)

	client.PutBins(shared.WritePolicy, key, bin1, bin2)

	log.Printf("Replace with: namespace=%s set=%s key=%s bin=%s value=%s",
		key.Namespace(), key.SetName(), key.Value(), bin3.Name, bin3.Value)

	wpolicy := as.NewWritePolicy(0, 0)
	wpolicy.RecordExistsAction = as.REPLACE
	client.PutBins(wpolicy, key, bin3)

	log.Printf("Get: namespace=%s set=%s key=%s", key.Namespace(), key.SetName(), key.Value())

	record, err := client.Get(shared.Policy, key)
	shared.PanicOnError(err)

	if record == nil {
		log.Fatalf(
			"Failed to get: namespace=%s set=%s key=%s", key.Namespace(), key.SetName(), key.Value())
	}

	if record.Bins[bin1.Name] == nil {
		log.Printf(bin1.Name + " was deleted as expected.")
	} else {
		log.Fatalln(bin1.Name + " found when it should have been deleted.")
	}

	if record.Bins[bin2.Name] == nil {
		log.Printf(bin2.Name + " was deleted as expected.")
	} else {
		log.Fatalln(bin2.Name + " found when it should have been deleted.")
	}
	shared.ValidateBin(key, bin3, record)
}

func runReplaceOnlyExample(client *as.Client) {
	key, err := as.NewKey(*shared.Namespace, *shared.Set, "replaceonlykey")
	shared.PanicOnError(err)
	bin := as.NewBin("bin", "value")

	// Delete record if it already exists.
	client.Delete(shared.WritePolicy, key)

	log.Printf("Replace record requiring that it exists: namespace=%s set=%s key=%s",
		key.Namespace(), key.SetName(), key.Value())

	wpolicy := as.NewWritePolicy(0, 0)
	wpolicy.RecordExistsAction = as.REPLACE_ONLY
	err = client.PutBins(wpolicy, key, bin)
	if ae, ok := err.(ast.AerospikeError); ok && ae.ResultCode() == ast.KEY_NOT_FOUND_ERROR {
		log.Printf("Success. `Not found` error returned as expected.")
	} else {
		log.Fatalln("Failure. This command should have resulted in an error.")
	}
}

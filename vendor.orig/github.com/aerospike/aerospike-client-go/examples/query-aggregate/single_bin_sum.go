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
	"os"
	"time"

	as "github.com/aerospike/aerospike-client-go"
	shared "github.com/aerospike/aerospike-client-go/examples/shared"
)

const keyCount = 1000

func main() {
	runExample(shared.Client)
	log.Println("Example finished successfully.")
}

func runExample(client *as.Client) {
	// Set LuaPath
	luaPath, _ := os.Getwd()
	luaPath += "/udf/"
	as.SetLuaPath(luaPath)

	filename := "sum_single_bin"
	regTask, err := client.RegisterUDFFromFile(nil, luaPath+filename+".lua", filename+".lua", as.LUA)
	shared.PanicOnError(err)

	// wait until UDF is created
	shared.PanicOnError(<-regTask.OnComplete())

	sum := 0
	for i := 1; i <= keyCount; i++ {
		sum += i
		key, err := as.NewKey(*shared.Namespace, *shared.Set, i)
		shared.PanicOnError(err)

		bin1 := as.NewBin("bin1", i)
		client.PutBins(nil, key, bin1)
	}

	t := time.Now()
	stm := as.NewStatement(*shared.Namespace, *shared.Set)
	res, err := client.QueryAggregate(nil, stm, filename, filename, "bin1")
	shared.PanicOnError(err)

	for rec := range res.Results() {
		log.Printf("Result %f should equal %d\n", rec.Record.Bins["SUCCESS"], sum)
	}
	log.Println("Map/Reduce took", time.Now().Sub(t))
}

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
	log.Printf("Scan parallel: namespace=" + *shared.Namespace + " set=" + *shared.Set)
	recordCount := 0
	begin := time.Now()
	policy := as.NewScanPolicy()
	recordset, err := client.ScanAll(policy, *shared.Namespace, *shared.Set)
	shared.PanicOnError(err)

L:
	for {
		select {
		case rec := <-recordset.Records:
			if rec == nil {
				break L
			}
			recordCount++

			if (recordCount % 10000) == 0 {
				log.Println("Records ", recordCount)
			}
		case err := <-recordset.Errors:
			// if there was an error, stop
			shared.PanicOnError(err)
		}
	}

	end := time.Now()
	seconds := float64(end.Sub(begin)) / float64(time.Second)
	log.Println("Total records returned: ", recordCount)
	log.Println("Elapsed time: ", seconds, " seconds")
	performance := shared.Round(float64(recordCount)/float64(seconds), 0.5, 0)
	log.Println("Records/second: ", performance)
}

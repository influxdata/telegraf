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

type Metrics struct {
	count int
	total int
}

var setMap = make(map[string]Metrics)

func runExample(client *as.Client) {
	log.Println("Scan series: namespace=", *shared.Namespace, " set=", *shared.Set)

	// Use low scan priority.  This will take more time, but it will reduce
	// the load on the server.
	policy := as.NewScanPolicy()
	policy.MaxRetries = 1
	policy.Priority = as.LOW

	nodeList := client.GetNodes()
	begin := time.Now()

	for _, node := range nodeList {
		log.Println("Scan node ", node.GetName())
		recordset, err := client.ScanNode(policy, node, *shared.Namespace, *shared.Set)
		shared.PanicOnError(err)

	L:
		for {
			select {
			case rec := <-recordset.Records:
				if rec == nil {
					break L
				}
				metrics, exists := setMap[rec.Key.SetName()]

				if !exists {
					metrics = Metrics{}
				}
				metrics.count++
				metrics.total++
				setMap[rec.Key.SetName()] = metrics

			case err := <-recordset.Errors:
				// if there was an error, stop
				shared.PanicOnError(err)
			}
		}

		for k, v := range setMap {
			log.Println("Node ", node, " set ", k, " count: ", v.count)
			v.count = 0
		}
	}

	end := time.Now()
	seconds := float64(end.Sub(begin)) / float64(time.Second)
	log.Println("Elapsed time: ", seconds, " seconds")

	total := 0

	for k, v := range setMap {
		log.Println("Total set ", k, " count: ", v.total)
		total += v.total
	}
	log.Println("Grand total: ", total)
	performance := shared.Round(float64(total)/seconds, 0.5, 0)
	log.Println("Records/second: ", performance)
}

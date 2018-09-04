// Copyright 2013-2017 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike_test

import (
	"math/rand"
	"strings"
	"time"

	as "github.com/aerospike/aerospike-client-go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const RANDOM_OPS_RUNS = 1000

// ALL tests are isolated by SetName and Key, which are 50 random characters
var _ = Describe("Aerospike", func() {
	initTestVars()

	Describe("Random Data Operations", func() {
		// connection data
		var err error
		var ns = "test"
		var set = randString(50)
		var key *as.Key
		var wpolicy = as.NewWritePolicy(0, 0)
		var rpolicy = as.NewPolicy()
		var rec *as.Record

		if *useReplicas {
			rpolicy.ReplicaPolicy = as.MASTER_PROLES
		}

		Context("Put/Get operations", func() {

			It("must create, update and read keys consistently", func() {
				key, err = as.NewKey(ns, set, randString(50))
				Expect(err).ToNot(HaveOccurred())

				bin1 := as.NewBin("Aerospike1", 0)
				bin2 := as.NewBin("Aerospike2", "a") // to avoid deletion of key

				i := 0
				for i < RANDOM_OPS_RUNS {
					iters := rand.Intn(10) + 1
					for wr := 0; wr < iters; wr++ {
						i++

						//reset
						err = client.PutBins(wpolicy, key, bin1, bin2)
						Expect(err).ToNot(HaveOccurred())

						// update
						err = client.PutBins(wpolicy, key, as.NewBin("Aerospike1", i), as.NewBin("Aerospike2", strings.Repeat("a", i)))
						Expect(err).ToNot(HaveOccurred())
					}

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(rec.Bins[bin1.Name]).To(Equal(i))
					Expect(rec.Bins[bin2.Name]).To(Equal(strings.Repeat("a", i)))
				}
			})

		}) // context put/get operations

		Context("Parallel Put/Get/Delete operations", func() {

			It("must save, read, delete keys consistently", func() {

				errChan := make(chan error, 100)

				func_delete := func(keys ...*as.Key) {
					defer GinkgoRecover()
					for _, key := range keys {
						existed, err := client.Delete(wpolicy, key)
						Expect(existed).To(BeTrue())
						errChan <- err
					}
				}

				i := 0
				for i < RANDOM_OPS_RUNS {
					iters := rand.Intn(1000) + 1
					for wr := 0; wr < iters; wr++ {
						i++

						key, err = as.NewKey(ns, set, randString(50))
						Expect(err).ToNot(HaveOccurred())

						err = client.PutBins(wpolicy, key, as.NewBin("Aerospike1", i), as.NewBin("Aerospike2", strings.Repeat("a", i)))
						Expect(err).ToNot(HaveOccurred())

						go func_delete(key)
					}

					// Timeout
					timeout := time.After(time.Second * 3)

					// Gather errors
					for i := 0; i < iters; i++ {
						select {
						case err := <-errChan:
							Expect(err).ToNot(HaveOccurred())

						case <-timeout:
							Expect(timeout).To(BeNil())
						}
					} // for i < iters

				} // for i < iters
			})

		}) // context parallel put/get/delete operations

	}) // describe

}) // describe

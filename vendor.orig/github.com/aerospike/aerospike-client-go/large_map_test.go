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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	as "github.com/aerospike/aerospike-client-go"
)

var _ = Describe("LargeMap Test", func() {
	initTestVars()

	var err error
	var ns = "test"
	var set = randString(50)
	var key *as.Key
	var wpolicy = as.NewWritePolicy(0, 0)

	if nsInfo(ns, "ldt-enabled") != "true" {
		By("LargeMap Tests are not supported since LDT is disabled.")
		return
	}

	BeforeEach(func() {
		key, err = as.NewKey(ns, set, randString(50))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create a valid LargeMap; Support PutMap() and Size(), Destroy()", func() {
		lmap := client.GetLargeMap(wpolicy, key, randString(10), "")
		res, err := lmap.Size()
		Expect(err).ToNot(HaveOccurred()) // bin not exists
		Expect(res).To(Equal(0))

		testMap := make(map[interface{}]interface{})
		for i := 1; i <= 100; i++ {
			testMap[i*100] = i
		}

		err = lmap.PutMap(testMap)
		Expect(err).ToNot(HaveOccurred())

		// confirm that the LMAP size has been increased to the expected size
		sz, err := lmap.Size()
		Expect(err).ToNot(HaveOccurred())
		Expect(sz).To(Equal(100))

		err = lmap.Destroy()
		Expect(err).ToNot(HaveOccurred())

		resMap, err := lmap.Scan()
		Expect(len(resMap)).To(Equal(0))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create a valid LargeMap; Support Put(), Exists(), Get(), Remove(), Find(), Size(), Scan()", func() {
		lmap := client.GetLargeMap(wpolicy, key, randString(10), "")
		res, err := lmap.Size()
		Expect(err).ToNot(HaveOccurred()) // bin not exists
		Expect(res).To(Equal(0))

		for i := 1; i <= 100; i++ {
			err = lmap.Put(as.NewValue(i*100), as.NewValue(i))
			Expect(err).ToNot(HaveOccurred())

			// check if it can be retrieved
			elem, err := lmap.Get(i * 100)
			Expect(err).ToNot(HaveOccurred())
			Expect(elem).To(Equal(map[interface{}]interface{}{i * 100: i}))

			// check if it exists
			exists, err := lmap.Exists(i * 100)
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeTrue())

			// check for a non-existing element
			// This test only passes in Aerospike 3.4.1 and above
			elem, err = lmap.Get(i * 70000)
			Expect(err).NotTo(HaveOccurred())
			Expect(elem).To(Equal(map[interface{}]interface{}{}))

			// confirm that the LMAP size has been increased to the expected size
			sz, err := lmap.Size()
			Expect(err).ToNot(HaveOccurred())
			Expect(sz).To(Equal(i))
		}

		// Scan() the map
		scanResult, err := lmap.Scan()
		scanExpectation := make(map[interface{}]interface{})
		for i := 1; i <= 100; i++ {
			scanExpectation[interface{}(i*100)] = i
		}
		Expect(err).ToNot(HaveOccurred())
		Expect(len(scanResult)).To(Equal(100))
		Expect(scanResult).To(Equal(scanExpectation))

		// remove all keys
		for i := 1; i <= 100; i++ {
			err = lmap.Remove(i * 100)
			Expect(err).ToNot(HaveOccurred())
		}

		scanExpectation = make(map[interface{}]interface{})
		scanResult, err = lmap.Scan()
		Expect(err).ToNot(HaveOccurred())
		Expect(scanResult).To(Equal(scanExpectation))
	})

	It("should correctly GetConfig()", func() {
		lmap := client.GetLargeMap(wpolicy, key, randString(10), "")
		err = lmap.Put(as.NewValue(0), as.NewValue(0))
		Expect(err).ToNot(HaveOccurred())

		config, err := lmap.GetConfig()
		Expect(err).ToNot(HaveOccurred())
		Expect(config["SUMMARY"]).To(Equal("LMAP Summary"))
	})

}) // describe

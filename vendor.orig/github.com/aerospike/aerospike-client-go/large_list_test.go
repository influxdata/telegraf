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

var _ = Describe("LargeList Test", func() {
	initTestVars()

	var err error
	var ns = "test"
	var set = randString(50)
	var key *as.Key
	var wpolicy = as.NewWritePolicy(0, 0)

	if nsInfo(ns, "ldt-enabled") != "true" {
		By("LargeList Tests are not supported since LDT is disabled.")
		return
	}

	BeforeEach(func() {
		key, err = as.NewKey(ns, set, randString(50))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create a valid LargeList; Support Add(), Remove(), Find(), Size(), Scan(), Range(), Destroy()", func() {
		llist := client.GetLargeList(wpolicy, key, randString(10), "")
		res, err := llist.Size()
		Expect(err).ToNot(HaveOccurred()) // bin not exists
		Expect(res).To(Equal(0))

		for i := 1; i <= 100; i++ {
			err = llist.Add(as.NewValue(i))
			Expect(err).ToNot(HaveOccurred())

			// confirm that the LLIST size has been increased to the expected size
			sz, err := llist.Size()
			Expect(err).ToNot(HaveOccurred())
			Expect(sz).To(Equal(i))
		}

		// Scan() the list
		scanResult, err := llist.Scan()
		scanExpectation := []interface{}{}
		for i := 1; i <= 100; i++ {
			scanExpectation = append(scanExpectation, interface{}(i))
		}
		Expect(err).ToNot(HaveOccurred())
		Expect(len(scanResult)).To(Equal(100))
		Expect(scanResult).To(Equal(scanExpectation))

		// check for range
		rangeResult, err := llist.Range(0, 100)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(rangeResult)).To(Equal(100))

		for i := 1; i <= 100; i++ {
			// confirm that the value already exists in the LLIST
			findResult, err := llist.Find(as.NewValue(i))
			Expect(err).ToNot(HaveOccurred())
			Expect(findResult).To(Equal([]interface{}{i}))

			// check for a non-existing element
			findResult, err = llist.Find(i * 70000)
			Expect(err).To(HaveOccurred())
			Expect(findResult).To(BeNil())

			// remove the value
			err = llist.Remove(as.NewValue(i))
			Expect(err).ToNot(HaveOccurred())

			// make sure the value has been removed
			findResult, err = llist.Find(as.NewValue(i))
			Expect(len(findResult)).To(Equal(0))
			// TODO: Revert in the future
			// Expect(err).To(HaveOccurred())
			// Expect(err.(AerospikeError).ResultCode()).To(Equal(LARGE_ITEM_NOT_FOUND))
		}

		err = llist.Destroy()
		Expect(err).ToNot(HaveOccurred())

		scanResult, err = llist.Scan()
		Expect(err).ToNot(HaveOccurred())
		Expect(len(scanResult)).To(Equal(0))

		err = llist.Add(1, 2, 3, 4, 5)
		existsResult, err := llist.Exist(1)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(existsResult)).To(Equal(1))
		Expect(existsResult[0]).To(Equal(true))

		existsResult, err = llist.Exist(3, 4, 5, 6)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(existsResult)).To(Equal(4))
		Expect(existsResult).To(Equal([]bool{true, true, true, false}))

		ffResult, err := llist.FindFrom(3, 2)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(ffResult)).To(Equal(2))
		Expect(ffResult).To(Equal([]interface{}{3, 4}))

		ffResult2, err := llist.FindFrom(3, 1)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(ffResult2)).To(Equal(1))
		Expect(ffResult2).To(Equal([]interface{}{3}))
	})

	It("should correctly GetConfig()", func() {
		llist := client.GetLargeList(wpolicy, key, randString(10), "")
		err = llist.Add(as.NewValue(0))
		Expect(err).ToNot(HaveOccurred())

		config, err := llist.GetConfig()
		Expect(err).ToNot(HaveOccurred())
		Expect(config["SUMMARY"]).To(Equal("LList Summary"))
	})

}) // describe

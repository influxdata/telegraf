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

/////////////////////////////////////////////////////////////
//
// NOTICE:
// 			THIS FEATURE HAS BEEN DEPRECATED ON SERVER.
//			THE API WILL BE REMOVED FROM THE CLIENT IN THE FUTURE.
//
/////////////////////////////////////////////////////////////

package aerospike_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	as "github.com/aerospike/aerospike-client-go"
)

var _ = Describe("LargeSet Test", func() {
	initTestVars()

	var err error
	var ns = "test"
	var set = randString(50)
	var key *as.Key
	var wpolicy = as.NewWritePolicy(0, 0)

	if nsInfo(ns, "ldt-enabled") != "true" {
		By("LargeSet Tests are not supported since LDT is disabled.")
		return
	}

	BeforeEach(func() {
		key, err = as.NewKey(ns, set, randString(50))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create a valid LargeSet; Support Add(), Get(), Remove(), Exists(), Size(), Scan(), Destroy()", func() {
		const elems = 100

		lset := client.GetLargeSet(wpolicy, key, randString(10), "")
		res, err := lset.Size()
		Expect(err).ToNot(HaveOccurred()) // bin not exists
		Expect(res).To(Equal(0))

		for i := 1; i <= elems; i++ {
			err = lset.Add(i)
			Expect(err).ToNot(HaveOccurred())

			// check if it can be retrieved
			elem, err := lset.Get(i)
			Expect(err).ToNot(HaveOccurred())
			Expect(elem).To(Equal(i))

			// check for a non-existing element
			elem, err = lset.Get(i * 70000)
			Expect(err).To(HaveOccurred())
			Expect(elem).To(BeNil())

			// confirm that the LSET size has been increased to the expected size
			sz, err := lset.Size()
			Expect(err).ToNot(HaveOccurred())
			Expect(sz).To(Equal(i))
		}

		// Scan() the set
		scanResult, err := lset.Scan()
		for i := 1; i <= elems; i++ {
			Expect(scanResult).To(ContainElement(i))
		}
		Expect(err).ToNot(HaveOccurred())
		Expect(len(scanResult)).To(Equal(elems))

		for i := 1; i <= elems; i++ {
			// confirm that the value already exists in the LSET
			exists, err := lset.Exists(as.NewValue(i))
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeTrue())

			// remove the value
			err = lset.Remove(as.NewValue(i))
			Expect(err).ToNot(HaveOccurred())

			// make sure the value has been removed
			exists, err = lset.Exists(as.NewValue(i))
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeFalse())
		}

		err = lset.Destroy()
		Expect(err).ToNot(HaveOccurred())

		scanResult, err = lset.Scan()
		Expect(err).ToNot(HaveOccurred())
		Expect(len(scanResult)).To(Equal(0))
	})

	It("should correctly GetConfig()", func() {
		lset := client.GetLargeSet(wpolicy, key, randString(10), "")
		err = lset.Add(as.NewValue(0))
		Expect(err).ToNot(HaveOccurred())

		config, err := lset.GetConfig()
		Expect(err).ToNot(HaveOccurred())
		Expect(config["SUMMARY"]).To(Equal("LSET Summary"))
	})

}) // describe

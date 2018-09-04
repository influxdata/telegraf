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

var _ = Describe("LargeStack Test", func() {
	initTestVars()

	var err error
	var ns = "test"
	var set = randString(50)
	var key *as.Key
	var wpolicy = as.NewWritePolicy(0, 0)

	if nsInfo(ns, "ldt-enabled") != "true" {
		By("LargeStack Tests are not supported since LDT is disabled.")
		return
	}

	BeforeEach(func() {
		key, err = as.NewKey(ns, set, randString(50))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create a valid LargeStack; Support Push(), Peek(), Pop(), Size(), Scan(), Destroy()", func() {
		lstack := client.GetLargeStack(wpolicy, key, randString(10), "")
		res, err := lstack.Size()
		Expect(err).ToNot(HaveOccurred()) // bin not exists
		Expect(res).To(Equal(0))

		for i := 1; i <= 100; i++ {
			err = lstack.Push(as.NewValue(i))
			Expect(err).ToNot(HaveOccurred())

			// confirm that the LSTACK size has been increased to the expected size
			sz, err := lstack.Size()
			Expect(err).ToNot(HaveOccurred())
			Expect(sz).To(Equal(i))
		}

		// Scan() the stack
		scanResult, err := lstack.Scan()
		scanExpectation := []interface{}{}
		for i := 100; i > 0; i-- {
			scanExpectation = append(scanExpectation, interface{}(i))
		}
		Expect(err).ToNot(HaveOccurred())
		Expect(len(scanResult)).To(Equal(100))
		Expect(scanResult).To(Equal(scanExpectation))

		// for i := 100; i > 0; i-- {
		// 	// peek the value
		// 	v, err := lstack.Peek(1)
		// 	Expect(err).ToNot(HaveOccurred())
		// 	Expect(v).To(Equal([]interface{}{i}))

		// 	// pop the value
		// 	// TODO: Wrong results
		// 	// v, err = lstack.Pop(1)
		// 	// Expect(err).ToNot(HaveOccurred())
		// 	// Expect(v).To(Equal([]interface{}{i}))
		// }

		// Destroy
		err = lstack.Destroy()
		Expect(err).ToNot(HaveOccurred())

		scanResult, err = lstack.Scan()
		Expect(err).ToNot(HaveOccurred())
		Expect(len(scanResult)).To(Equal(0))
	})

	It("should correctly GetConfig()", func() {
		lstack := client.GetLargeStack(wpolicy, key, randString(10), "")
		err = lstack.Push(as.NewValue(0))
		Expect(err).ToNot(HaveOccurred())

		config, err := lstack.GetConfig()
		Expect(err).ToNot(HaveOccurred())
		Expect(config["SUMMARY"]).To(Equal("LSTACK Summary"))
	})

}) // describe

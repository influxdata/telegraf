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
	"math"
	"math/rand"
	"time"

	. "github.com/aerospike/aerospike-client-go"
	// . "github.com/aerospike/aerospike-client-go/logger"
	// . "github.com/aerospike/aerospike-client-go/types"

	// . "github.com/aerospike/aerospike-client-go/utils/buffer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// ALL tests are isolated by SetName and Key, which are 50 random characters
var _ = Describe("Truncate operations test", func() {
	initTestVars()

	Context("Truncate", func() {
		var err error
		var ns = "test"
		var set = randString(50)
		var key *Key
		var rec *Record
		var wpolicy = NewWritePolicy(0, 0)
		wpolicy.SendKey = true

		const keyCount = 1000
		bin1 := NewBin("Aerospike1", rand.Intn(math.MaxInt16))
		bin2 := NewBin("Aerospike2", randString(100))

		BeforeEach(func() {
			for i := 0; i < keyCount; i++ {
				key, err = NewKey(ns, set, i)
				Expect(err).ToNot(HaveOccurred())

				rec, err = client.Operate(wpolicy, key, PutOp(bin1), PutOp(bin2), GetOp())
				Expect(err).ToNot(HaveOccurred())
			}
		})

		var countRecords = func(namespace, setName string) int {
			stmt := NewStatement(namespace, setName)
			res, err := client.Query(nil, stmt)
			Expect(err).ToNot(HaveOccurred())

			cnt := 0
			for rec := range res.Results() {
				Expect(rec.Err).ToNot(HaveOccurred())
				cnt++
			}

			return cnt
		}

		It("must truncate only the current set", func() {
			Expect(countRecords(ns, set)).To(Equal(keyCount))

			err := client.Truncate(nil, ns, set, nil)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(time.Second)
			Expect(countRecords(ns, set)).To(Equal(0))
		})

		It("must truncate the whole namespace", func() {
			Expect(countRecords(ns, "")).ToNot(Equal(0))

			err := client.Truncate(nil, ns, "", nil)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(time.Second)
			Expect(countRecords(ns, "")).To(Equal(0))
		})

		It("must truncate only older records", func() {
			time.Sleep(3 * time.Second)
			t := time.Now()

			Expect(countRecords(ns, set)).To(Equal(keyCount))

			for i := keyCount; i < 2*keyCount; i++ {
				key, err = NewKey(ns, set, i)
				Expect(err).ToNot(HaveOccurred())

				rec, err = client.Operate(wpolicy, key, PutOp(bin1), PutOp(bin2), GetOp())
				Expect(err).ToNot(HaveOccurred())
			}
			Expect(countRecords(ns, set)).To(Equal(2 * keyCount))

			err := client.Truncate(nil, ns, set, &t)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(3 * time.Second)
			Expect(countRecords(ns, set)).To(Equal(keyCount))
		})

	})
})

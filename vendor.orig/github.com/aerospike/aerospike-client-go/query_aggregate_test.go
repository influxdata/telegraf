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
	"os"

	as "github.com/aerospike/aerospike-client-go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func registerUDF(client *as.Client, path, filename string) error {
	regTask, err := client.RegisterUDFFromFile(nil, path+filename+".lua", filename+".lua", as.LUA)
	if err != nil {
		return err
	}

	// wait until UDF is created
	return <-regTask.OnComplete()
}

// ALL tests are isolated by SetName and Key, which are 50 random characters
var _ = Describe("Query Aggregate operations", func() {
	initTestVars()

	// connection data
	var ns = "test"
	var set = randString(50)
	var wpolicy = as.NewWritePolicy(0, 0)
	wpolicy.SendKey = true

	// Set LuaPath
	luaPath, _ := os.Getwd()
	luaPath += "/test/resources/"
	as.SetLuaPath(luaPath)

	const keyCount = 10

	BeforeSuite(func() {
		err := registerUDF(client, luaPath, "sum_single_bin")
		Expect(err).ToNot(HaveOccurred())

		err = registerUDF(client, luaPath, "average")
		Expect(err).ToNot(HaveOccurred())
	})

	BeforeEach(func() {
		set = randString(50)
		for i := 1; i <= keyCount; i++ {
			key, err := as.NewKey(ns, set, randString(50))
			Expect(err).ToNot(HaveOccurred())

			bin1 := as.NewBin("bin1", i)
			client.PutBins(nil, key, bin1)
		}

		// // queries only work on indices
		// idxTask, err := client.CreateIndex(wpolicy, ns, set, set+bin3.Name, bin3.Name, NUMERIC)
		// Expect(err).ToNot(HaveOccurred())

		// wait until index is created
		// Expect(<-idxTask.OnComplete()).ToNot(HaveOccurred())
	})

	It("must return the sum of specified bin to the client", func() {
		stm := as.NewStatement(ns, set)
		res, err := client.QueryAggregate(nil, stm, "sum_single_bin", "sum_single_bin", "bin1")
		Expect(err).ToNot(HaveOccurred())

		Expect(res.TaskId()).To(Equal(stm.TaskId))
		Expect(res.TaskId()).To(BeNumerically(">", 0))

		for rec := range res.Results() {
			Expect(rec.Err).ToNot(HaveOccurred())
			Expect(rec.Record.Bins["SUCCESS"]).To(Equal(float64(55)))
		}
	})

	It("must return Sum and Count to the client", func() {
		stm := as.NewStatement(ns, set)
		res, err := client.QueryAggregate(nil, stm, "average", "average", "bin1")
		Expect(err).ToNot(HaveOccurred())

		for rec := range res.Results() {
			Expect(rec.Err).ToNot(HaveOccurred())
			Expect(rec.Record.Bins["SUCCESS"]).To(Equal(map[interface{}]interface{}{"sum": float64(55), "count": float64(10)}))
		}
	})
})

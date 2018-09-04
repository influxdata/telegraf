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

const udfCDTTests = `
function add(rec, bin, key, value)
	if not aerospike:exists(rec) then
		--create record
		local m1 = map()
		local m2 = map()
		m2[value] = key

		m1[key] = m2			-- set map to map
		rec[bin] = m1		-- set map to record
		--create record
		return aerospike:create(rec)
	end

	--record existed, let's see if key exists
	local m1 = rec[bin]		--get map
	local m2 = m1[key]		--get second map

	--did list already exist?
	if m2 == nil then
		m2 = map() 		--map didn't exist yet, let's create it
	end
	local doesExist = m2[value]
	if doesExist ~= nil then
		return 0 --value already existed, no need to update record
	end

	m2[value] 	= key
	--done setting values, let's store information back to set
	m1[key] 		= m2		--map back to map
	rec[bin] 	= m1	--map back to record

	return aerospike:update(rec) --..and update aerospike :)
end

--remove from map[]map[]
function remove(rec, bin, key, value)
	if not aerospike:exists(rec) then
		return 0 --record does not exist, cannot remove so no error
	end

	local m1 = rec[bin]
	local m2 = m1[key]
	if m2 == nil then
		return 0 --key does not exist, cannot remove
	end

	--remove key from map
	local doesExist = m2[value]
	if doesExist == nil then
		return 0 --value does not exist
	end

	map.remove(m2, value) --remove value from map
	--done, let's update record with modified maps
	m1[key] 		= m2 --back to map
	rec[bin]		= m1 --back to record

	return aerospike:update(rec) --and update!
end
`

var _ = Describe("CDT Map Test", func() {
	initTestVars()

	if !featureEnabled("cdt-map") {
		By("CDT Map Tests will not run since feature is not supported by the server.")
		return
	}

	// connection data
	var err error
	var ns = "test"
	var set = randString(50)
	var key *as.Key
	var wpolicy = as.NewWritePolicy(0, 0)
	var cdtBinName string
	// var list []interface{}

	putMode := as.DefaultMapPolicy()
	addMode := as.NewMapPolicy(as.MapOrder.UNORDERED, as.MapWriteMode.CREATE_ONLY)

	BeforeEach(func() {
		key, err = as.NewKey(ns, set, randString(50))
		Expect(err).ToNot(HaveOccurred())

		cdtBinName = randString(10)
	})

	Describe("Simple Usecases", func() {

		It("should create a valid CDT Map using MapPutOp", func() {
			cdtMap, err := client.Operate(wpolicy, key,
				as.MapPutOp(putMode, cdtBinName, 1, 1),
				as.MapPutOp(putMode, cdtBinName, 2, 2),
				as.MapPutOp(addMode, cdtBinName, 3, 3),
				as.MapPutOp(addMode, cdtBinName, 4, 4),
				as.MapPutOp(addMode, cdtBinName, 6, 6),
				as.MapPutOp(addMode, cdtBinName, 7, 7),
				as.MapPutOp(addMode, cdtBinName, 8, 8),
			)
			Expect(err).ToNot(HaveOccurred())

			cdtMap, err = client.Get(nil, key, cdtBinName)
			Expect(err).ToNot(HaveOccurred())
			Expect(cdtMap.Bins[cdtBinName]).To(Equal(map[interface{}]interface{}{1: 1, 2: 2, 3: 3, 4: 4, 6: 6, 7: 7, 8: 8}))
		})

		It("should unpack an empty Non-Ordered CDT map correctly", func() {
			cdtMap, err := client.Operate(wpolicy, key,
				as.MapPutOp(putMode, cdtBinName, 1, 1),
				as.MapRemoveByKeyOp(cdtBinName, 1, as.MapReturnType.NONE),
			)
			Expect(err).ToNot(HaveOccurred())

			cdtMap, err = client.Get(nil, key, cdtBinName)
			Expect(err).ToNot(HaveOccurred())
			Expect(cdtMap.Bins[cdtBinName]).To(Equal(map[interface{}]interface{}{}))
		})

		It("should unpack an empty Ordered CDT map correctly", func() {
			_, err := client.Operate(wpolicy, key,
				as.MapPutOp(as.NewMapPolicy(as.MapOrder.KEY_ORDERED, as.MapWriteMode.UPDATE), cdtBinName, 1, 1),
				// MapRemoveByKeyOp(cdtBinName, 1, as.MapReturnType.NONE),
			)
			Expect(err).ToNot(HaveOccurred())

			rs, err := client.ScanAll(nil, ns, set)
			Expect(err).ToNot(HaveOccurred())
			for rs := range rs.Results() {
				Expect(rs.Err).ToNot(HaveOccurred())

			}

		})

		It("should create a valid CDT Map using MapPutOp", func() {
			cdtMap, err := client.Operate(wpolicy, key,
				as.MapPutOp(putMode, cdtBinName, 1, 1),
				as.MapPutOp(putMode, cdtBinName, 2, 2),
				as.MapPutOp(addMode, cdtBinName, 3, 3),
				as.MapPutOp(addMode, cdtBinName, 4, 4),

				as.GetOpForBin(cdtBinName),
			)
			// Expect(err).ToNot(HaveOccurred())
			Expect(cdtMap).NotTo(Equal([]interface{}{1, 2, 3, 4, 4, 4, map[interface{}]interface{}{1: 1, 2: 2, 3: 3, 4: 4}}))

			cdtMap, err = client.Get(nil, key, cdtBinName)
			Expect(err).ToNot(HaveOccurred())
			Expect(cdtMap.Bins[cdtBinName]).To(Equal(map[interface{}]interface{}{1: 1, 2: 2, 3: 3, 4: 4}))
		})

		It("should create a valid CDT Map using MapPutItemsOp", func() {
			addMap := map[interface{}]interface{}{
				12:    "myValue",
				-8734: "str2",
				1:     "my default",
			}

			putMap := map[interface{}]interface{}{
				12: "myval12222",
				13: "str13",
			}

			updateMap := map[interface{}]interface{}{
				13: "myval2",
			}

			replaceMap := map[interface{}]interface{}{
				12:    23,
				-8734: "changed",
			}

			putMode := as.DefaultMapPolicy()
			addMode := as.NewMapPolicy(as.MapOrder.KEY_ORDERED, as.MapWriteMode.CREATE_ONLY)
			updateMode := as.NewMapPolicy(as.MapOrder.KEY_ORDERED, as.MapWriteMode.UPDATE_ONLY)

			cdtMap, err := client.Operate(wpolicy, key,
				as.MapPutItemsOp(addMode, cdtBinName, addMap),
				as.MapPutItemsOp(putMode, cdtBinName, putMap),
				as.MapPutItemsOp(updateMode, cdtBinName, updateMap),
				as.MapPutItemsOp(updateMode, cdtBinName, replaceMap),
				as.MapGetByKeyOp(cdtBinName, 1, as.MapReturnType.VALUE),
				as.MapGetByKeyOp(cdtBinName, -8734, as.MapReturnType.VALUE),
				as.MapGetByKeyRangeOp(cdtBinName, 12, 15, as.MapReturnType.KEY_VALUE),
				// as.GetOpForBin(cdtBinName),
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(cdtMap.Bins[cdtBinName]).To(Equal([]interface{}{3, 4, 4, 4, "my default", "changed", []as.MapPair{{Key: 12, Value: 23}, {Key: 13, Value: "myval2"}}}))

			cdtMap, err = client.Get(nil, key, cdtBinName)
			Expect(err).ToNot(HaveOccurred())
			Expect(cdtMap.Bins[cdtBinName]).To(Equal([]as.MapPair{{Key: -8734, Value: "changed"}, {Key: 1, Value: "my default"}, {Key: 12, Value: 23}, {Key: 13, Value: "myval2"}}))
		})

		It("should create a valid CDT Map using mixed MapPutOp and MapPutItemsOp", func() {

			items := map[interface{}]interface{}{
				12:    "myval",
				-8734: "str2",
				1:     "my default",
				7:     1,
			}

			otherBinName := "other_bin"

			cdtMap, err := client.Operate(wpolicy, key,
				as.MapPutItemsOp(as.NewMapPolicy(as.MapOrder.KEY_VALUE_ORDERED, as.MapWriteMode.UPDATE), cdtBinName, items),
				as.PutOp(as.NewBin(otherBinName, "head")),
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(cdtMap.Bins[cdtBinName]).To(Equal(4))
			Expect(cdtMap.Bins).To(HaveKey(otherBinName))

			cdtMap, err = client.Operate(wpolicy, key,
				as.MapGetByKeyOp(cdtBinName, 12, as.MapReturnType.INDEX),
				as.AppendOp(as.NewBin(otherBinName, "...tail")),
				as.GetOpForBin(otherBinName),
			)

			cdtMap, err = client.Get(nil, key)
			Expect(err).ToNot(HaveOccurred())
			Expect(cdtMap.Bins).To(Equal(as.BinMap{cdtBinName: []as.MapPair{{Key: -8734, Value: "str2"}, {Key: 1, Value: "my default"}, {Key: 7, Value: 1}, {Key: 12, Value: "myval"}}, "other_bin": "head...tail"}))
		})

		It("should create a valid CDT Map and then Switch Policy For Order", func() {

			cdtMap, err := client.Operate(wpolicy, key,
				as.MapPutOp(as.DefaultMapPolicy(), cdtBinName, 4, 1),
				as.MapPutOp(as.DefaultMapPolicy(), cdtBinName, 3, 2),
				as.MapPutOp(as.DefaultMapPolicy(), cdtBinName, 2, 3),
				as.MapPutOp(as.DefaultMapPolicy(), cdtBinName, 1, 4),

				as.MapGetByIndexOp(cdtBinName, 2, as.MapReturnType.KEY_VALUE),
				as.MapGetByIndexRangeCountOp(cdtBinName, 0, 10, as.MapReturnType.KEY_VALUE),
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(cdtMap.Bins).To(Equal(as.BinMap{cdtBinName: []interface{}{1, 2, 3, 4, []as.MapPair{{Key: 3, Value: 2}}, []as.MapPair{{Key: 1, Value: 4}, {Key: 2, Value: 3}, {Key: 3, Value: 2}, {Key: 4, Value: 1}}}}))

			cdtMap, err = client.Operate(wpolicy, key,
				as.MapSetPolicyOp(as.NewMapPolicy(as.MapOrder.KEY_ORDERED, as.MapWriteMode.UPDATE), cdtBinName),

				as.MapGetByKeyRangeOp(cdtBinName, 3, 5, as.MapReturnType.COUNT),
				as.MapGetByKeyRangeOp(cdtBinName, -5, 2, as.MapReturnType.KEY_VALUE),
				as.MapGetByIndexRangeCountOp(cdtBinName, 0, 10, as.MapReturnType.KEY_VALUE),
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(cdtMap.Bins).To(Equal(as.BinMap{cdtBinName: []interface{}{interface{}(nil), 2, []as.MapPair{{Key: 1, Value: 4}}, []as.MapPair{{Key: 1, Value: 4}, {Key: 2, Value: 3}, {Key: 3, Value: 2}, {Key: 4, Value: 1}}}}))

		})

		It("should create a valid CDT Map and then apply Inc/Dec operations and Get Correct Values", func() {

			items := map[interface{}]interface{}{
				"Charlie": 55,
				"Jim":     98,
				"John":    76,
				"Harry":   82,
			}

			// Write values to empty map.
			cdtMap, err := client.Operate(wpolicy, key,
				as.MapPutItemsOp(as.DefaultMapPolicy(), cdtBinName, items),
			)

			Expect(err).ToNot(HaveOccurred())

			cdtMap, err = client.Get(nil, key)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(cdtMap.Bins)).To(Equal(1))

			cdtMap, err = client.Operate(wpolicy, key,
				as.MapIncrementOp(as.DefaultMapPolicy(), cdtBinName, "John", 5),
				as.MapDecrementOp(as.DefaultMapPolicy(), cdtBinName, "Jim", 4),
			)
			Expect(err).ToNot(HaveOccurred())

			cdtMap, err = client.Get(nil, key)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(cdtMap.Bins)).To(Equal(1))

			cdtMap, err = client.Operate(nil, key,
				as.MapGetByRankRangeCountOp(cdtBinName, -2, 2, as.MapReturnType.KEY),
				as.MapGetByRankRangeCountOp(cdtBinName, 0, 2, as.MapReturnType.KEY_VALUE),
				as.MapGetByRankOp(cdtBinName, 0, as.MapReturnType.VALUE),
				as.MapGetByRankOp(cdtBinName, 2, as.MapReturnType.KEY),
				as.MapGetByValueRangeOp(cdtBinName, 90, 95, as.MapReturnType.RANK),
				as.MapGetByValueRangeOp(cdtBinName, 90, 95, as.MapReturnType.COUNT),
				as.MapGetByValueRangeOp(cdtBinName, 90, 95, as.MapReturnType.KEY_VALUE),
				as.MapGetByValueRangeOp(cdtBinName, 81, 82, as.MapReturnType.KEY),
				as.MapGetByValueOp(cdtBinName, 77, as.MapReturnType.KEY),
				as.MapGetByValueOp(cdtBinName, 81, as.MapReturnType.RANK),
				as.MapGetByKeyOp(cdtBinName, "Charlie", as.MapReturnType.RANK),
				as.MapGetByKeyOp(cdtBinName, "Charlie", as.MapReturnType.REVERSE_RANK),
			)

			Expect(err).ToNot(HaveOccurred())

			Expect(cdtMap.Bins).To(Equal(as.BinMap{cdtBinName: []interface{}{"Harry", "Jim", []as.MapPair{{Key: "Charlie", Value: 55}, {Key: "John", Value: 81}}, 55, "Harry", []interface{}{3}, 1, []as.MapPair{{Key: "Jim", Value: 94}}, []interface{}{"John"}, []interface{}{}, []interface{}{1}, 0, 3}}))
		})

		It("should create a valid CDT Map and then execute Remove operations", func() {

			items := map[interface{}]interface{}{
				"Charlie": 55,
				"Jim":     98,
				"John":    76,
				"Harry":   82,
				"Sally":   79,
				"Lenny":   84,
				"Abe":     88,
			}

			itemsToRemove := []interface{}{
				"Sally",
				"UNKNOWN",
				"Lenny",
			}

			// Write values to empty map.
			cdtMap, err := client.Operate(wpolicy, key,
				as.MapPutItemsOp(as.DefaultMapPolicy(), cdtBinName, items),
				as.MapRemoveByKeyOp(cdtBinName, "NOTFOUND", as.MapReturnType.VALUE),
				as.MapRemoveByKeyOp(cdtBinName, "Jim", as.MapReturnType.VALUE),
				as.MapRemoveByKeyListOp(cdtBinName, itemsToRemove, as.MapReturnType.VALUE),
				as.MapRemoveByValueOp(cdtBinName, 55, as.MapReturnType.KEY),
			)

			Expect(err).ToNot(HaveOccurred())

			Expect(cdtMap.Bins).To(Equal(as.BinMap{cdtBinName: []interface{}{7, nil, 98, []interface{}{79, 84}, []interface{}{"Charlie"}}}))
		})

	})

	It("should create a valid CDT Map and then execute RemoveRange operations", func() {

		items := map[interface{}]interface{}{
			"Charlie": 55,
			"Jim":     98,
			"John":    76,
			"Harry":   82,
			"Sally":   79,
			"Lenny":   84,
			"Abe":     88,
		}

		// Write values to empty map.
		cdtMap, err := client.Operate(wpolicy, key,
			as.MapPutItemsOp(as.DefaultMapPolicy(), cdtBinName, items),
			as.MapRemoveByKeyRangeOp(cdtBinName, "J", "K", as.MapReturnType.COUNT),
			as.MapRemoveByValueRangeOp(cdtBinName, 80, 85, as.MapReturnType.COUNT),
			as.MapRemoveByIndexRangeCountOp(cdtBinName, 0, 2, as.MapReturnType.COUNT),
			as.MapRemoveByRankRangeCountOp(cdtBinName, 0, 2, as.MapReturnType.COUNT),
		)

		Expect(err).ToNot(HaveOccurred())

		Expect(cdtMap.Bins).To(Equal(as.BinMap{cdtBinName: []interface{}{7, 2, 2, 2, 1}}))
	})

	It("should create a valid CDT Map and then execute Clear operations", func() {

		items := map[interface{}]interface{}{
			"Charlie": 55,
			"Jim":     98,
			"John":    76,
			"Harry":   82,
			"Sally":   79,
			"Lenny":   84,
			"Abe":     88,
		}

		// Write values to empty map.
		cdtMap, err := client.Operate(wpolicy, key,
			as.MapPutItemsOp(as.DefaultMapPolicy(), cdtBinName, items),
		)

		Expect(err).ToNot(HaveOccurred())

		cdtMap, err = client.Get(nil, key)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(cdtMap.Bins)).To(Equal(1))

		cdtMap, err = client.Operate(wpolicy, key,
			as.MapClearOp(cdtBinName),
			as.MapSizeOp(cdtBinName),
		)

		Expect(err).ToNot(HaveOccurred())
		Expect(cdtMap.Bins).To(Equal(as.BinMap{cdtBinName: []interface{}{nil, 0}}))
	})

	It("should create a valid CDT Map and then execute RANK operations", func() {

		items := map[interface{}]interface{}{
			"p1": 0,
			"p2": 0,
			"p3": 0,
			"p4": 0,
		}

		mapPolicy := as.NewMapPolicy(as.MapOrder.KEY_VALUE_ORDERED, as.MapWriteMode.UPDATE)

		// Write values to empty map.
		cdtMap, err := client.Operate(wpolicy, key,
			as.MapPutItemsOp(mapPolicy, cdtBinName, items),
		)

		Expect(err).ToNot(HaveOccurred())

		cdtMap, err = client.Get(nil, key)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(cdtMap.Bins)).To(Equal(1))

		cdtMap, err = client.Operate(wpolicy, key,
			as.MapIncrementOp(mapPolicy, cdtBinName, "p1", 10),
			as.MapIncrementOp(mapPolicy, cdtBinName, "p2", 20),
			as.MapIncrementOp(mapPolicy, cdtBinName, "p3", 1),
			as.MapIncrementOp(mapPolicy, cdtBinName, "p4", 20),
		)

		cdtMap, err = client.Get(nil, key)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(cdtMap.Bins)).To(Equal(1))

		cdtMap, err = client.Operate(wpolicy, key,
			as.MapGetByRankRangeCountOp(cdtBinName, -3, 3, as.MapReturnType.KEY),
		)

		Expect(err).ToNot(HaveOccurred())
		Expect(cdtMap.Bins).To(Equal(as.BinMap{cdtBinName: []interface{}{"p1", "p2", "p4"}}))

		cdtMap, err = client.Operate(wpolicy, key,
			as.MapRemoveByValueOp(cdtBinName, 10, as.MapReturnType.KEY),
			as.MapGetByRankRangeCountOp(cdtBinName, -3, 3, as.MapReturnType.KEY),
		)

		Expect(err).ToNot(HaveOccurred())
		Expect(cdtMap.Bins).To(Equal(as.BinMap{cdtBinName: []interface{}{"p1", []interface{}{"p3", "p2", "p4"}}}))
	})

	It("should handle CDTs in UDFs", func() {

		regTsk, err := client.RegisterUDF(nil, []byte(udfCDTTests), "cdt_tests.lua", as.LUA)
		Expect(err).ToNot(HaveOccurred())

		Expect(<-regTsk.OnComplete()).ToNot(HaveOccurred())

		_, err = client.Execute(nil, key, "cdt_tests", "add", as.NewValue("b"), as.NewValue("k"), as.NewValue(1))
		Expect(err).ToNot(HaveOccurred())

		sets, err := client.ScanAll(nil, "test", "skill")
		Expect(err).ToNot(HaveOccurred())

		rec, err := client.Get(nil, key)
		Expect(err).ToNot(HaveOccurred())
		Expect(rec.Bins).To(Equal(as.BinMap{"b": map[interface{}]interface{}{"k": map[interface{}]interface{}{1: "k"}}}))

		for res := range sets.Results() {
			Expect(res.Err).ToNot(HaveOccurred())
			_, err = client.Delete(nil, res.Record.Key)
			Expect(err).ToNot(HaveOccurred())
		}

		_, err = client.Execute(nil, key, "cdt_tests", "add", as.NewValue("b"), as.NewValue("k"), as.NewValue(1))
		Expect(err).ToNot(HaveOccurred())

		_, err = client.Execute(nil, key, "cdt_tests", "remove", as.NewValue("b"), as.NewValue("k"), as.NewValue(1))
		Expect(err).ToNot(HaveOccurred())

		rec, err = client.Get(nil, key)
		Expect(err).ToNot(HaveOccurred())
		Expect(rec.Bins).To(Equal(as.BinMap{"b": map[interface{}]interface{}{"k": map[interface{}]interface{}{}}}))

		sets, err = client.ScanAll(nil, ns, set)
		Expect(err).ToNot(HaveOccurred())
		for res := range sets.Results() {
			Expect(res.Err).ToNot(HaveOccurred())
			_, err = client.Delete(nil, res.Record.Key)
			Expect(err).ToNot(HaveOccurred())
		}
	})

}) // describe

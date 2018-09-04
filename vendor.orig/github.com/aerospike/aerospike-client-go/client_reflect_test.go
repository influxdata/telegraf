// +build !as_performance

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
	"strings"

	as "github.com/aerospike/aerospike-client-go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// ALL tests are isolated by SetName and Key, which are 50 random characters
var _ = Describe("Aerospike", func() {
	initTestVars()

	Describe("Data operations on complex types with reflection", func() {
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

		BeforeEach(func() {
			key, err = as.NewKey(ns, set, randString(50))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("Put operations", func() {

			Context("Bins with complex types", func() {

				Context("Bins with LIST type", func() {

					It("must save a key with Array Types", func() {
						bin1 := as.NewBin("Aerospike1", []int8{math.MinInt8, 0, 1, 2, 3, math.MaxInt8})
						bin2 := as.NewBin("Aerospike2", []int16{math.MinInt16, 0, 1, 2, 3, math.MaxInt16})
						bin3 := as.NewBin("Aerospike3", []int32{math.MinInt32, 0, 1, 2, 3, math.MaxInt32})
						bin4 := as.NewBin("Aerospike4", []int64{math.MinInt64, 0, 1, 2, 3, math.MaxInt64})
						bin5 := as.NewBin("Aerospike5", []uint8{0, 1, 2, 3, math.MaxUint8})
						bin6 := as.NewBin("Aerospike6", []uint16{0, 1, 2, 3, math.MaxUint16})
						bin7 := as.NewBin("Aerospike7", []uint32{0, 1, 2, 3, math.MaxUint32})
						bin8 := as.NewBin("Aerospike8", []string{"", "\n", "string"})
						bin9 := as.NewBin("Aerospike9", []interface{}{"", 1, nil, true, false, uint64(math.MaxUint64), math.MaxFloat32, math.MaxFloat64, as.NewGeoJSONValue(`{ "type": "Point", "coordinates": [0.00, 0.00] }"`), [3]int{1, 2, 3}})

						// complex type, consisting different arrays
						bin10 := as.NewBin("Aerospike10", []interface{}{
							nil,
							bin1.Value.GetObject(),
							bin2.Value.GetObject(),
							bin3.Value.GetObject(),
							bin4.Value.GetObject(),
							bin5.Value.GetObject(),
							bin6.Value.GetObject(),
							bin7.Value.GetObject(),
							bin8.Value.GetObject(),
							bin9.Value.GetObject(),
							map[interface{}]interface{}{
								1: [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
								[16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}: []interface{}{"string", 12, nil},
								// [3]int{0, 1, 2}:          []interface{}{"string", 12, nil},
								// [3]string{"0", "1", "2"}: []interface{}{"string", 12, nil},
								15:                        nil,
								int8(math.MaxInt8):        int8(math.MaxInt8),
								int64(math.MinInt64):      int64(math.MinInt64),
								int64(math.MaxInt64):      int64(math.MaxInt64),
								uint64(math.MaxUint64):    uint64(math.MaxUint64),
								float32(-math.MaxFloat32): float32(-math.MaxFloat32),
								float64(-math.MaxFloat64): float64(-math.MaxFloat64),
								float32(math.MaxFloat32):  float32(math.MaxFloat32),
								float64(math.MaxFloat64):  float64(math.MaxFloat64),
								"true":    true,
								"false":   false,
								"string":  map[interface{}]interface{}{nil: "string", "string": 19},                // map to complex array
								nil:       []int{18, 41},                                                           // array to complex map
								"GeoJSON": as.NewGeoJSONValue(`{ "type": "Point", "coordinates": [0.00, 0.00] }"`), // bit-sign test
							},
						})

						err = client.PutBins(wpolicy, key, bin1, bin2, bin3, bin4, bin5, bin6, bin7, bin8, bin9, bin10)
						Expect(err).ToNot(HaveOccurred())

						rec, err = client.Get(rpolicy, key)
						Expect(err).ToNot(HaveOccurred())

						arraysEqual(rec.Bins[bin1.Name], bin1.Value.GetObject())
						arraysEqual(rec.Bins[bin2.Name], bin2.Value.GetObject())
						arraysEqual(rec.Bins[bin3.Name], bin3.Value.GetObject())
						arraysEqual(rec.Bins[bin4.Name], bin4.Value.GetObject())
						arraysEqual(rec.Bins[bin5.Name], bin5.Value.GetObject())
						arraysEqual(rec.Bins[bin6.Name], bin6.Value.GetObject())
						arraysEqual(rec.Bins[bin7.Name], bin7.Value.GetObject())
						arraysEqual(rec.Bins[bin8.Name], bin8.Value.GetObject())
						arraysEqual(rec.Bins[bin9.Name], bin9.Value.GetObject())
						arraysEqual(rec.Bins[bin10.Name], bin10.Value.GetObject())
					})

				}) // context list

				Context("Bins with MAP type", func() {

					It("must save a key with Array Types", func() {
						// complex type, consisting different maps
						bin1 := as.NewBin("Aerospike1", map[int32]string{
							0:                    "",
							int32(math.MaxInt32): randString(100),
							int32(math.MinInt32): randString(100),
						})

						bin2 := as.NewBin("Aerospike2", map[interface{}]interface{}{
							15:                        nil,
							"true":                    true,
							"false":                   false,
							int8(math.MaxInt8):        int8(math.MaxInt8),
							int64(math.MinInt64):      int64(math.MinInt64),
							int64(math.MaxInt64):      int64(math.MaxInt64),
							uint64(math.MaxUint64):    uint64(math.MaxUint64),
							float32(-math.MaxFloat32): float32(-math.MaxFloat32),
							float64(-math.MaxFloat64): float64(-math.MaxFloat64),
							float32(math.MaxFloat32):  float32(math.MaxFloat32),
							float64(math.MaxFloat64):  float64(math.MaxFloat64),
							"string":                  map[interface{}]interface{}{nil: "string", "string": 19},                // map to complex array
							nil:                       []int{18, 41},                                                           // array to complex map
							"longString":              strings.Repeat("s", 32911),                                              // bit-sign test
							"GeoJSON":                 as.NewGeoJSONValue(`{ "type": "Point", "coordinates": [0.00, 0.00] }"`), // bit-sign test
						})

						err = client.PutBins(wpolicy, key, bin1, bin2)
						Expect(err).ToNot(HaveOccurred())

						rec, err = client.Get(rpolicy, key)
						Expect(err).ToNot(HaveOccurred())

						mapsEqual(rec.Bins[bin1.Name], bin1.Value.GetObject())
						mapsEqual(rec.Bins[bin2.Name], bin2.Value.GetObject())
					})

				}) // context map

				Context("Bins with LIST type", func() {

					It("must save a key with Array Types", func() {
						bin1 := as.NewBin("Aerospike1", []interface{}{math.MinInt8, 0, 1, 2, 3, math.MaxInt8})
						bin2 := as.NewBin("Aerospike2", []interface{}{math.MinInt16, 0, 1, 2, 3, math.MaxInt16})
						bin3 := as.NewBin("Aerospike3", []interface{}{math.MinInt32, 0, 1, 2, 3, math.MaxInt32})
						bin4 := as.NewBin("Aerospike4", []interface{}{math.MinInt64, 0, 1, 2, 3, math.MaxInt64})
						bin5 := as.NewBin("Aerospike5", []interface{}{0, 1, 2, 3, math.MaxUint8})
						bin6 := as.NewBin("Aerospike6", []interface{}{0, 1, 2, 3, math.MaxUint16})
						bin7 := as.NewBin("Aerospike7", []interface{}{0, 1, 2, 3, math.MaxUint32})
						bin8 := as.NewBin("Aerospike8", []interface{}{"", "\n", "string"})
						bin9 := as.NewBin("Aerospike9", []interface{}{"", 1, nil, true, false, uint64(math.MaxUint64), math.MaxFloat32, math.MaxFloat64, as.NewGeoJSONValue(`{ "type": "Point", "coordinates": [0.00, 0.00] }"`), [3]int{1, 2, 3}})

						// complex type, consisting different arrays
						bin10 := as.NewBin("Aerospike10", []interface{}{
							nil,
							bin1.Value.GetObject(),
							bin2.Value.GetObject(),
							bin3.Value.GetObject(),
							bin4.Value.GetObject(),
							bin5.Value.GetObject(),
							bin6.Value.GetObject(),
							bin7.Value.GetObject(),
							bin8.Value.GetObject(),
							bin9.Value.GetObject(),
							map[interface{}]interface{}{
								1: [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
								[16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}: []interface{}{"string", 12, nil},
								// [3]int{0, 1, 2}:          []interface{}{"string", 12, nil},
								// [3]string{"0", "1", "2"}: []interface{}{"string", 12, nil},
								15:                        nil,
								int8(math.MaxInt8):        int8(math.MaxInt8),
								int64(math.MinInt64):      int64(math.MinInt64),
								int64(math.MaxInt64):      int64(math.MaxInt64),
								uint64(math.MaxUint64):    uint64(math.MaxUint64),
								float32(-math.MaxFloat32): float32(-math.MaxFloat32),
								float64(-math.MaxFloat64): float64(-math.MaxFloat64),
								float32(math.MaxFloat32):  float32(math.MaxFloat32),
								float64(math.MaxFloat64):  float64(math.MaxFloat64),
								"true":    true,
								"false":   false,
								"string":  map[interface{}]interface{}{nil: "string", "string": 19},                // map to complex array
								nil:       []interface{}{18, 41},                                                   // array to complex map
								"GeoJSON": as.NewGeoJSONValue(`{ "type": "Point", "coordinates": [0.00, 0.00] }"`), // bit-sign test
							},
						})

						err = client.PutBins(wpolicy, key, bin1, bin2, bin3, bin4, bin5, bin6, bin7, bin8, bin9, bin10)
						Expect(err).ToNot(HaveOccurred())

						rec, err = client.Get(rpolicy, key)
						Expect(err).ToNot(HaveOccurred())

						arraysEqual(rec.Bins[bin1.Name], bin1.Value.GetObject())
						arraysEqual(rec.Bins[bin2.Name], bin2.Value.GetObject())
						arraysEqual(rec.Bins[bin3.Name], bin3.Value.GetObject())
						arraysEqual(rec.Bins[bin4.Name], bin4.Value.GetObject())
						arraysEqual(rec.Bins[bin5.Name], bin5.Value.GetObject())
						arraysEqual(rec.Bins[bin6.Name], bin6.Value.GetObject())
						arraysEqual(rec.Bins[bin7.Name], bin7.Value.GetObject())
						arraysEqual(rec.Bins[bin8.Name], bin8.Value.GetObject())
						arraysEqual(rec.Bins[bin9.Name], bin9.Value.GetObject())
						arraysEqual(rec.Bins[bin10.Name], bin10.Value.GetObject())
					})

				}) // context list

				Context("Bins with MAP type", func() {

					It("must save a key with Array Types", func() {
						// complex type, consisting different maps
						bin1 := as.NewBin("Aerospike1", map[int32]string{
							0:                    "",
							int32(math.MaxInt32): randString(100),
							int32(math.MinInt32): randString(100),
						})

						bin2 := as.NewBin("Aerospike2", map[interface{}]interface{}{
							15:                        nil,
							"true":                    true,
							"false":                   false,
							int8(math.MaxInt8):        int8(math.MaxInt8),
							int64(math.MinInt64):      int64(math.MinInt64),
							int64(math.MaxInt64):      int64(math.MaxInt64),
							uint64(math.MaxUint64):    uint64(math.MaxUint64),
							float32(-math.MaxFloat32): float32(-math.MaxFloat32),
							float64(-math.MaxFloat64): float64(-math.MaxFloat64),
							float32(math.MaxFloat32):  float32(math.MaxFloat32),
							float64(math.MaxFloat64):  float64(math.MaxFloat64),
							"string":                  map[interface{}]interface{}{nil: "string", "string": 19},                // map to complex array
							nil:                       []int{18, 41},                                                           // array to complex map
							"longString":              strings.Repeat("s", 32911),                                              // bit-sign test
							"GeoJSON":                 as.NewGeoJSONValue(`{ "type": "Point", "coordinates": [0.00, 0.00] }"`), // bit-sign test
						})

						err = client.PutBins(wpolicy, key, bin1, bin2)
						Expect(err).ToNot(HaveOccurred())

						rec, err = client.Get(rpolicy, key)
						Expect(err).ToNot(HaveOccurred())

						mapsEqual(rec.Bins[bin1.Name], bin1.Value.GetObject())
						mapsEqual(rec.Bins[bin2.Name], bin2.Value.GetObject())
					})

				}) // context map

			}) // context complex types

		}) // put context

	})
})

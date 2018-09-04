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
	"bytes"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	as "github.com/aerospike/aerospike-client-go"
	. "github.com/aerospike/aerospike-client-go/types"
	. "github.com/aerospike/aerospike-client-go/utils/buffer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// ALL tests are isolated by SetName and Key, which are 50 random characters
var _ = Describe("Aerospike", func() {
	initTestVars()

	Describe("Client Management", func() {

		It("must open and close the client without a problem", func() {
			// use the same client for all
			client, err := as.NewClientWithPolicy(clientPolicy, *host, *port)
			Expect(err).ToNot(HaveOccurred())
			Expect(client.IsConnected()).To(BeTrue())

			client.Close()
			Expect(client.IsConnected()).To(BeFalse())
		})

		It("must return an error if supplied cluster-name is wrong", func() {
			// use the same client for all
			cpolicy := *clientPolicy
			cpolicy.ClusterName = "haha"
			cpolicy.Timeout = 10 * time.Second
			nclient, err := as.NewClientWithPolicy(&cpolicy, *host, *port)
			aerr, ok := err.(AerospikeError)
			Expect(ok).To(BeTrue())
			Expect(err).To(HaveOccurred())
			Expect(aerr.ResultCode()).To(Equal(CLUSTER_NAME_MISMATCH_ERROR))
			Expect(nclient).To(BeNil())
		})

		It("must return a client even if cluster-name is wrong, but failIfConnected is false", func() {
			// use the same client for all
			cpolicy := *clientPolicy
			cpolicy.ClusterName = "haha"
			cpolicy.Timeout = 10 * time.Second
			cpolicy.FailIfNotConnected = false
			nclient, err := as.NewClientWithPolicy(&cpolicy, *host, *port)
			aerr, ok := err.(AerospikeError)
			Expect(ok).To(BeTrue())
			Expect(err).To(HaveOccurred())
			Expect(aerr.ResultCode()).To(Equal(CLUSTER_NAME_MISMATCH_ERROR))
			Expect(nclient).NotTo(BeNil())
			Expect(nclient.IsConnected()).To(BeFalse())
		})

		It("must connect to the cluster when cluster-name is correct", func() {
			nodeCount := len(client.GetNodes())

			// use the same client for all
			cpolicy := *clientPolicy
			cpolicy.ClusterName = "null"
			cpolicy.Timeout = 10 * time.Second
			nclient, err := as.NewClientWithPolicy(&cpolicy, *host, *port)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(nclient.GetNodes())).To(Equal(nodeCount))
		})

	})

	Describe("Data operations on native types", func() {
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

			Context("Expiration values", func() {

				It("must return 30d if set to TTLServerDefault", func() {
					wpolicy := as.NewWritePolicy(0, as.TTLServerDefault)
					bin := as.NewBin("Aerospike", "value")
					rec, err = client.Operate(wpolicy, key, as.PutOp(bin), as.GetOp())
					Expect(err).ToNot(HaveOccurred())

					defaultTTL, err := strconv.Atoi(nsInfo(ns, "default-ttl"))
					Expect(err).ToNot(HaveOccurred())

					Expect(rec.Expiration).To(Equal(uint32(defaultTTL)))
				})

				It("must return TTLDontExpire if set to TTLDontExpire", func() {
					wpolicy := as.NewWritePolicy(0, as.TTLDontExpire)
					bin := as.NewBin("Aerospike", "value")
					err = client.PutBins(wpolicy, key, bin)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(rec.Expiration).To(Equal(uint32(as.TTLDontExpire)))
				})

				It("must not change the TTL if set to TTLDontUpdate", func() {
					wpolicy := as.NewWritePolicy(0, as.TTLServerDefault)
					bin := as.NewBin("Aerospike", "value")
					err = client.PutBins(wpolicy, key, bin)
					Expect(err).ToNot(HaveOccurred())

					time.Sleep(3 * time.Second)

					wpolicy = as.NewWritePolicy(0, as.TTLDontUpdate)
					bin = as.NewBin("Aerospike", "value")
					err = client.PutBins(wpolicy, key, bin)
					Expect(err).ToNot(HaveOccurred())

					defaultTTL, err := strconv.Atoi(nsInfo(ns, "default-ttl"))
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(rec.Expiration).To(BeNumerically("<=", uint32(defaultTTL-3))) // default expiration on server is set to 30d
				})
			})

			Context("Bins with `nil` values should be deleted", func() {
				It("must save a key with SINGLE bin", func() {
					bin := as.NewBin("Aerospike", "value")
					bin1 := as.NewBin("Aerospike1", "value2") // to avoid deletion of key
					err = client.PutBins(wpolicy, key, bin, bin1)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(rec.Bins[bin.Name]).To(Equal(bin.Value.GetObject()))

					bin2 := as.NewBin("Aerospike", nil)
					err = client.PutBins(wpolicy, key, bin2)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())

					// Key should not exist
					_, exists := rec.Bins[bin.Name]
					Expect(exists).To(Equal(false))
				})

				It("must save a key with MULTIPLE bins", func() {
					bin1 := as.NewBin("Aerospike1", "nil")
					bin2 := as.NewBin("Aerospike2", "value")
					bin3 := as.NewBin("Aerospike3", "value")
					err = client.PutBins(wpolicy, key, bin1, bin2, bin3)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())

					bin2nil := as.NewBin("Aerospike2", nil)
					bin3nil := as.NewBin("Aerospike3", nil)
					err = client.PutBins(wpolicy, key, bin2nil, bin3nil)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())

					// Key should not exist
					_, exists := rec.Bins[bin2.Name]
					Expect(exists).To(Equal(false))
					_, exists = rec.Bins[bin3.Name]
					Expect(exists).To(Equal(false))
				})

				It("must save a key with MULTIPLE bins using a BinMap", func() {
					bin1 := as.NewBin("Aerospike1", "nil")
					bin2 := as.NewBin("Aerospike2", "value")
					bin3 := as.NewBin("Aerospike3", "value")
					err = client.Put(wpolicy, key, as.BinMap{bin1.Name: bin1.Value, bin2.Name: bin2.Value, bin3.Name: bin3.Value})
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())

					bin2nil := as.NewBin("Aerospike2", nil)
					bin3nil := as.NewBin("Aerospike3", nil)
					err = client.Put(wpolicy, key, as.BinMap{bin2nil.Name: bin2nil.Value, bin3nil.Name: bin3nil.Value})
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())

					// Key should not exist
					_, exists := rec.Bins[bin2.Name]
					Expect(exists).To(Equal(false))
					_, exists = rec.Bins[bin3.Name]
					Expect(exists).To(Equal(false))
				})
			})

			Context("Bins with `string` values", func() {
				It("must save a key with SINGLE bin", func() {
					bin := as.NewBin("Aerospike", "Awesome")
					err = client.PutBins(wpolicy, key, bin)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(rec.Bins[bin.Name]).To(Equal(bin.Value.GetObject()))
				})

				It("must save a key with MULTIPLE bins", func() {
					bin1 := as.NewBin("Aerospike1", "Awesome1")
					bin2 := as.NewBin("Aerospike2", "")
					err = client.PutBins(wpolicy, key, bin1, bin2)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(rec.Bins[bin1.Name]).To(Equal(bin1.Value.GetObject()))
					Expect(rec.Bins[bin2.Name]).To(Equal(bin2.Value.GetObject()))
				})
			})

			Context("Bins with `int8` and `uint8` values", func() {
				It("must save a key with SINGLE bin", func() {
					bin := as.NewBin("Aerospike", int8(rand.Intn(math.MaxInt8)))
					err = client.PutBins(wpolicy, key, bin)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(rec.Bins[bin.Name]).To(Equal(bin.Value.GetObject()))
				})

				It("must save a key with MULTIPLE bins", func() {
					bin1 := as.NewBin("Aerospike1", int8(math.MaxInt8))
					bin2 := as.NewBin("Aerospike2", int8(math.MinInt8))
					bin3 := as.NewBin("Aerospike3", uint8(math.MaxUint8))
					err = client.PutBins(wpolicy, key, bin1, bin2, bin3)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(rec.Bins[bin1.Name]).To(Equal(bin1.Value.GetObject()))
					Expect(rec.Bins[bin2.Name]).To(Equal(bin2.Value.GetObject()))
					Expect(rec.Bins[bin3.Name]).To(Equal(bin3.Value.GetObject()))
				})
			})

			Context("Bins with `int16` and `uint16` values", func() {
				It("must save a key with SINGLE bin", func() {
					bin := as.NewBin("Aerospike", int16(rand.Intn(math.MaxInt16)))
					err = client.PutBins(wpolicy, key, bin)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(rec.Bins[bin.Name]).To(Equal(bin.Value.GetObject()))
				})

				It("must save a key with MULTIPLE bins", func() {
					bin1 := as.NewBin("Aerospike1", int16(math.MaxInt16))
					bin2 := as.NewBin("Aerospike2", int16(math.MinInt16))
					bin3 := as.NewBin("Aerospike3", uint16(math.MaxUint16))
					err = client.PutBins(wpolicy, key, bin1, bin2, bin3)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(rec.Bins[bin1.Name]).To(Equal(bin1.Value.GetObject()))
					Expect(rec.Bins[bin2.Name]).To(Equal(bin2.Value.GetObject()))
					Expect(rec.Bins[bin3.Name]).To(Equal(bin3.Value.GetObject()))
				})
			})

			Context("Bins with `int` and `uint` values", func() {
				It("must save a key with SINGLE bin", func() {
					bin := as.NewBin("Aerospike", rand.Int())
					err = client.PutBins(wpolicy, key, bin)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(rec.Bins[bin.Name]).To(Equal(bin.Value.GetObject()))
				})

				It("must save a key with MULTIPLE bins; uint of > MaxInt32 will always result in LongValue", func() {
					bin1 := as.NewBin("Aerospike1", math.MaxInt32)
					bin2, bin3 := func() (*as.Bin, *as.Bin) {
						if Arch32Bits {
							return as.NewBin("Aerospike2", int(math.MinInt32)),
								as.NewBin("Aerospike3", uint(math.MaxInt32))
						}
						return as.NewBin("Aerospike2", int(math.MinInt64)),
							as.NewBin("Aerospike3", uint(math.MaxInt64))

					}()

					err = client.PutBins(wpolicy, key, bin1, bin2, bin3)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(rec.Bins[bin1.Name]).To(Equal(bin1.Value.GetObject()))
					if Arch64Bits {
						Expect(rec.Bins[bin2.Name].(int)).To(Equal(bin2.Value.GetObject()))
						Expect(int64(rec.Bins[bin3.Name].(int))).To(Equal(bin3.Value.GetObject()))
					} else {
						Expect(rec.Bins[bin2.Name]).To(Equal(bin2.Value.GetObject()))
						Expect(rec.Bins[bin3.Name]).To(Equal(bin3.Value.GetObject()))
					}
				})
			})

			Context("Bins with `int64` only values (uint64 is supported via type cast to int64) ", func() {
				It("must save a key with SINGLE bin", func() {
					bin := as.NewBin("Aerospike", rand.Int63())
					err = client.PutBins(wpolicy, key, bin)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())

					if Arch64Bits {
						Expect(int64(rec.Bins[bin.Name].(int))).To(Equal(bin.Value.GetObject()))
					} else {
						Expect(rec.Bins[bin.Name]).To(Equal(bin.Value.GetObject()))
					}
				})

				It("must save a key with MULTIPLE bins", func() {
					bin1 := as.NewBin("Aerospike1", math.MaxInt64)
					bin2 := as.NewBin("Aerospike2", math.MinInt64)
					err = client.PutBins(wpolicy, key, bin1, bin2)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())

					Expect(rec.Bins[bin1.Name]).To(Equal(bin1.Value.GetObject()))
					Expect(rec.Bins[bin2.Name]).To(Equal(bin2.Value.GetObject()))
				})
			})

			Context("Bins with `float32` only values", func() {
				It("must save a key with SINGLE bin", func() {
					bin := as.NewBin("Aerospike", rand.Float32())
					err = client.PutBins(wpolicy, key, bin)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(float64(rec.Bins[bin.Name].(float64))).To(Equal(bin.Value.GetObject()))
				})

				It("must save a key with MULTIPLE bins", func() {
					bin1 := as.NewBin("Aerospike1", math.MaxFloat32)
					bin2 := as.NewBin("Aerospike2", -math.MaxFloat32)
					err = client.PutBins(wpolicy, key, bin1, bin2)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())

					Expect(rec.Bins[bin1.Name]).To(Equal(bin1.Value.GetObject()))
					Expect(rec.Bins[bin2.Name]).To(Equal(bin2.Value.GetObject()))
				})
			})

			Context("Bins with `float64` only values", func() {
				It("must save a key with SINGLE bin", func() {
					bin := as.NewBin("Aerospike", rand.Float64())
					err = client.PutBins(wpolicy, key, bin)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())
					Expect(float64(rec.Bins[bin.Name].(float64))).To(Equal(bin.Value.GetObject()))
				})

				It("must save a key with MULTIPLE bins", func() {
					bin1 := as.NewBin("Aerospike1", math.MaxFloat64)
					bin2 := as.NewBin("Aerospike2", -math.MaxFloat64)
					err = client.PutBins(wpolicy, key, bin1, bin2)
					Expect(err).ToNot(HaveOccurred())

					rec, err = client.Get(rpolicy, key)
					Expect(err).ToNot(HaveOccurred())

					Expect(rec.Bins[bin1.Name]).To(Equal(bin1.Value.GetObject()))
					Expect(rec.Bins[bin2.Name]).To(Equal(bin2.Value.GetObject()))
				})
			})

			Context("Bins with complex types", func() {

				Context("Bins with BLOB type", func() {
					It("must save and retrieve Bins with AerospikeBlobs type", func() {
						person := &testBLOB{name: "SomeDude"}
						bin := as.NewBin("Aerospike1", person)
						err = client.PutBins(wpolicy, key, bin)
						Expect(err).ToNot(HaveOccurred())

						rec, err = client.Get(rpolicy, key)
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("Bins with LIST type", func() {

					It("must save a key with Array Types", func() {
						// All int types and sizes should be encoded into an int64,
						// unless if they are of type uint64, which always encodes to uint64
						// regardless of the values inside
						intList := []interface{}{math.MinInt64, math.MinInt64 + 1}
						for i := uint(0); i < 64; i++ {
							intList = append(intList, -(1 << i))
							intList = append(intList, -(1<<i)-1)
							intList = append(intList, -(1<<i)+1)
							intList = append(intList, 1<<i)
							intList = append(intList, (1<<i)-1)
							intList = append(intList, (1<<i)+1)
						}
						intList = append(intList, -1)
						intList = append(intList, 0)
						intList = append(intList, uint64(1))
						intList = append(intList, math.MaxInt64-1)
						intList = append(intList, math.MaxInt64)
						intList = append(intList, uint64(math.MaxInt64+1))
						intList = append(intList, uint64(math.MaxUint64-1))
						intList = append(intList, uint64(math.MaxUint64))
						bin0 := as.NewBin("Aerospike0", intList)

						bin1 := as.NewBin("Aerospike1", []interface{}{math.MinInt8, 0, 1, 2, 3, math.MaxInt8})
						bin2 := as.NewBin("Aerospike2", []interface{}{math.MinInt16, 0, 1, 2, 3, math.MaxInt16})
						bin3 := as.NewBin("Aerospike3", []interface{}{math.MinInt32, 0, 1, 2, 3, math.MaxInt32})
						bin4 := as.NewBin("Aerospike4", []interface{}{math.MinInt64, 0, 1, 2, 3, math.MaxInt64})
						bin5 := as.NewBin("Aerospike5", []interface{}{0, 1, 2, 3, math.MaxUint8})
						bin6 := as.NewBin("Aerospike6", []interface{}{0, 1, 2, 3, math.MaxUint16})
						bin7 := as.NewBin("Aerospike7", []interface{}{0, 1, 2, 3, math.MaxUint32})
						bin8 := as.NewBin("Aerospike8", []interface{}{"", "\n", "string"})
						bin9 := as.NewBin("Aerospike9", []interface{}{"", 1, nil, true, false, uint64(math.MaxUint64), math.MaxFloat32, math.MaxFloat64, as.NewGeoJSONValue(`{ "type": "Point", "coordinates": [0.00, 0.00] }"`), []interface{}{1, 2, 3}})

						// complex type, consisting different arrays
						bin10 := as.NewBin("Aerospike10", []interface{}{
							nil,
							bin0.Value.GetObject(),
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
								1: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
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
								"intList": intList,
							},
						})

						err = client.PutBins(wpolicy, key, bin0, bin1, bin2, bin3, bin4, bin5, bin6, bin7, bin8, bin9, bin10)
						Expect(err).ToNot(HaveOccurred())

						rec, err = client.Get(rpolicy, key)
						Expect(err).ToNot(HaveOccurred())

						arraysEqual(rec.Bins[bin0.Name], bin0.Value.GetObject())
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
						bin1 := as.NewBin("Aerospike1", map[interface{}]interface{}{
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
							nil:                       []interface{}{18, 41},                                                   // array to complex map
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

		Context("Append operations", func() {
			bin := as.NewBin("Aerospike", randString(rand.Intn(100)))

			BeforeEach(func() {
				err = client.PutBins(wpolicy, key, bin)
				Expect(err).ToNot(HaveOccurred())
			})

			It("must append to a SINGLE bin", func() {
				appbin := as.NewBin(bin.Name, randString(rand.Intn(100)))
				err = client.AppendBins(wpolicy, key, appbin)
				Expect(err).ToNot(HaveOccurred())

				rec, err = client.Get(rpolicy, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(rec.Bins[bin.Name]).To(Equal(bin.Value.GetObject().(string) + appbin.Value.GetObject().(string)))
			})

			It("must append to a SINGLE bin using a BinMap", func() {
				appbin := as.NewBin(bin.Name, randString(rand.Intn(100)))
				err = client.Append(wpolicy, key, as.BinMap{bin.Name: appbin.Value})
				Expect(err).ToNot(HaveOccurred())

				rec, err = client.Get(rpolicy, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(rec.Bins[bin.Name]).To(Equal(bin.Value.GetObject().(string) + appbin.Value.GetObject().(string)))
			})

		}) // append context

		Context("Prepend operations", func() {
			bin := as.NewBin("Aerospike", randString(rand.Intn(100)))

			BeforeEach(func() {
				err = client.PutBins(wpolicy, key, bin)
				Expect(err).ToNot(HaveOccurred())
			})

			It("must Prepend to a SINGLE bin", func() {
				appbin := as.NewBin(bin.Name, randString(rand.Intn(100)))
				err = client.PrependBins(wpolicy, key, appbin)
				Expect(err).ToNot(HaveOccurred())

				rec, err = client.Get(rpolicy, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(rec.Bins[bin.Name]).To(Equal(appbin.Value.GetObject().(string) + bin.Value.GetObject().(string)))
			})

			It("must Prepend to a SINGLE bin using a BinMap", func() {
				appbin := as.NewBin(bin.Name, randString(rand.Intn(100)))
				err = client.Prepend(wpolicy, key, as.BinMap{bin.Name: appbin.Value})
				Expect(err).ToNot(HaveOccurred())

				rec, err = client.Get(rpolicy, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(rec.Bins[bin.Name]).To(Equal(appbin.Value.GetObject().(string) + bin.Value.GetObject().(string)))
			})

		}) // prepend context

		Context("Add operations", func() {
			bin := as.NewBin("Aerospike", rand.Intn(math.MaxInt16))

			BeforeEach(func() {
				err = client.PutBins(wpolicy, key, bin)
				Expect(err).ToNot(HaveOccurred())
			})

			It("must Add to a SINGLE bin", func() {
				addBin := as.NewBin(bin.Name, rand.Intn(math.MaxInt16))
				err = client.AddBins(wpolicy, key, addBin)
				Expect(err).ToNot(HaveOccurred())

				rec, err = client.Get(rpolicy, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(rec.Bins[bin.Name]).To(Equal(addBin.Value.GetObject().(int) + bin.Value.GetObject().(int)))
			})

			It("must Add to a SINGLE bin using a BinMap", func() {
				addBin := as.NewBin(bin.Name, rand.Intn(math.MaxInt16))
				err = client.Add(wpolicy, key, as.BinMap{addBin.Name: addBin.Value})
				Expect(err).ToNot(HaveOccurred())

				rec, err = client.Get(rpolicy, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(rec.Bins[bin.Name]).To(Equal(addBin.Value.GetObject().(int) + bin.Value.GetObject().(int)))
			})

		}) // add context

		Context("Delete operations", func() {
			bin := as.NewBin("Aerospike", rand.Intn(math.MaxInt16))

			BeforeEach(func() {
				err = client.PutBins(wpolicy, key, bin)
				Expect(err).ToNot(HaveOccurred())
			})

			It("must Delete to a non-existing key", func() {
				var nxkey *as.Key
				nxkey, err = as.NewKey(ns, set, randString(50))
				Expect(err).ToNot(HaveOccurred())

				var existed bool
				existed, err = client.Delete(wpolicy, nxkey)
				Expect(err).ToNot(HaveOccurred())
				Expect(existed).To(Equal(false))
			})

			It("must Delete to an existing key", func() {
				var existed bool
				existed, err = client.Delete(wpolicy, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(existed).To(Equal(true))

				existed, err = client.Exists(rpolicy, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(existed).To(Equal(false))
			})

		}) // Delete context

		Context("Touch operations", func() {
			bin := as.NewBin("Aerospike", rand.Intn(math.MaxInt16))

			BeforeEach(func() {
				err = client.PutBins(wpolicy, key, bin)
				Expect(err).ToNot(HaveOccurred())
			})

			It("must Touch to a non-existing key", func() {
				var nxkey *as.Key
				nxkey, err = as.NewKey(ns, set, randString(50))
				Expect(err).ToNot(HaveOccurred())

				err = client.Touch(wpolicy, nxkey)
				Expect(err).To(HaveOccurred())
			})

			It("must Touch to an existing key", func() {
				rec, err = client.Get(rpolicy, key)
				Expect(err).ToNot(HaveOccurred())
				generation := rec.Generation

				wpolicy := as.NewWritePolicy(0, 0)
				wpolicy.SendKey = true
				err = client.Touch(wpolicy, key)
				Expect(err).ToNot(HaveOccurred())

				rec, err = client.Get(rpolicy, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(rec.Generation).To(Equal(generation + 1))

				recordset, err := client.ScanAll(nil, key.Namespace(), key.SetName())
				Expect(err).ToNot(HaveOccurred())

				// make sure the
				for r := range recordset.Results() {
					Expect(r.Err).ToNot(HaveOccurred())
					if bytes.Equal(key.Digest(), r.Record.Key.Digest()) {
						Expect(r.Record.Key.Value()).To(Equal(key.Value()))
						Expect(r.Record.Bins).To(Equal(rec.Bins))
					}
				}
			})

		}) // Touch context

		Context("Exists operations", func() {
			bin := as.NewBin("Aerospike", rand.Intn(math.MaxInt16))

			BeforeEach(func() {
				err = client.PutBins(wpolicy, key, bin)
				Expect(err).ToNot(HaveOccurred())
			})

			It("must check Existence of a non-existing key", func() {
				var nxkey *as.Key
				nxkey, err = as.NewKey(ns, set, randString(50))
				Expect(err).ToNot(HaveOccurred())

				var exists bool
				exists, err = client.Exists(rpolicy, nxkey)
				Expect(err).ToNot(HaveOccurred())
				Expect(exists).To(Equal(false))
			})

			It("must checks Existence of an existing key", func() {
				var exists bool
				exists, err = client.Exists(rpolicy, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(exists).To(Equal(true))
			})

		}) // Exists context

		Context("Batch Exists operations", func() {
			bin := as.NewBin("Aerospike", rand.Intn(math.MaxInt16))
			const keyCount = 2048

			BeforeEach(func() {
			})

			It("must return the result with same ordering", func() {
				var exists []bool
				keys := []*as.Key{}

				for i := 0; i < keyCount; i++ {
					key, err := as.NewKey(ns, set, randString(50))
					Expect(err).ToNot(HaveOccurred())
					keys = append(keys, key)

					// if key shouldExist == true, put it in the DB
					if i%2 == 0 {
						err = client.PutBins(wpolicy, key, bin)
						Expect(err).ToNot(HaveOccurred())

						// make sure they exists in the DB
						exists, err := client.Exists(rpolicy, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(exists).To(Equal(true))
					}
				}

				exists, err = client.BatchExists(rpolicy, keys)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(exists)).To(Equal(len(keys)))
				for idx, keyExists := range exists {
					Expect(keyExists).To(Equal(idx%2 == 0))
				}
			})

		}) // Batch Exists context

		Context("Batch Get operations", func() {
			bin := as.NewBin("Aerospike", rand.Int())
			const keyCount = 2048

			BeforeEach(func() {
			})

			It("must return the records with same ordering as keys", func() {
				binRedundant := as.NewBin("Redundant", "Redundant")

				var records []*as.Record
				type existence struct {
					key         *as.Key
					shouldExist bool // set randomly and checked against later
				}

				exList := make([]existence, 0, keyCount)
				keys := make([]*as.Key, 0, keyCount)

				for i := 0; i < keyCount; i++ {
					key, err := as.NewKey(ns, set, randString(50))
					Expect(err).ToNot(HaveOccurred())
					e := existence{key: key, shouldExist: rand.Intn(100) > 50}
					exList = append(exList, e)
					keys = append(keys, key)

					// if key shouldExist == true, put it in the DB
					if e.shouldExist {
						err = client.PutBins(wpolicy, key, bin, binRedundant)
						Expect(err).ToNot(HaveOccurred())

						// make sure they exists in the DB
						rec, err := client.Get(rpolicy, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(rec.Bins[bin.Name]).To(Equal(bin.Value.GetObject()))
						Expect(rec.Bins[binRedundant.Name]).To(Equal(binRedundant.Value.GetObject()))
					} else {
						// make sure they exists in the DB
						exists, err := client.Exists(rpolicy, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(exists).To(Equal(false))
					}
				}

				records, err = client.BatchGet(rpolicy, keys)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(records)).To(Equal(len(keys)))
				for idx, rec := range records {
					if exList[idx].shouldExist {
						Expect(rec.Bins[bin.Name]).To(Equal(bin.Value.GetObject()))
					} else {
						Expect(rec).To(BeNil())
					}
				}

				records, err = client.BatchGet(rpolicy, keys, bin.Name)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(records)).To(Equal(len(keys)))
				for idx, rec := range records {
					if exList[idx].shouldExist {
						// only bin1 has been requested
						Expect(rec.Bins[binRedundant.Name]).To(BeNil())
						Expect(rec.Bins[bin.Name]).To(Equal(bin.Value.GetObject()))
					} else {
						Expect(rec).To(BeNil())
					}
				}
			})

		}) // Batch Get context

		Context("GetHeader operations", func() {
			bin := as.NewBin("Aerospike", rand.Intn(math.MaxInt16))

			BeforeEach(func() {
				err = client.PutBins(wpolicy, key, bin)
				Expect(err).ToNot(HaveOccurred())
			})

			It("must Get the Header of an existing key after touch", func() {
				rec, err = client.Get(rpolicy, key)
				Expect(err).ToNot(HaveOccurred())
				generation := rec.Generation

				err = client.Touch(wpolicy, key)
				Expect(err).ToNot(HaveOccurred())

				rec, err = client.GetHeader(rpolicy, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(rec.Generation).To(Equal(generation + 1))
				Expect(rec.Bins[bin.Name]).To(BeNil())
			})

		}) // GetHeader context

		Context("Batch Get Header operations", func() {
			bin := as.NewBin("Aerospike", rand.Int())
			const keyCount = 1024

			BeforeEach(func() {
			})

			It("must return the records with same ordering as keys", func() {
				var records []*as.Record
				type existence struct {
					key         *as.Key
					shouldExist bool // set randomly and checked against later
				}

				exList := []existence{}
				keys := []*as.Key{}

				for i := 0; i < keyCount; i++ {
					key, err := as.NewKey(ns, set, randString(50))
					Expect(err).ToNot(HaveOccurred())
					e := existence{key: key, shouldExist: rand.Intn(100) > 50}
					exList = append(exList, e)
					keys = append(keys, key)

					// if key shouldExist == true, put it in the DB
					if e.shouldExist {
						err = client.PutBins(wpolicy, key, bin)
						Expect(err).ToNot(HaveOccurred())

						// update generation
						err = client.Touch(wpolicy, key)
						Expect(err).ToNot(HaveOccurred())

						// make sure they exists in the DB
						exists, err := client.Exists(rpolicy, key)
						Expect(err).ToNot(HaveOccurred())
						Expect(exists).To(Equal(true))
					}
				}

				records, err = client.BatchGetHeader(rpolicy, keys)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(records)).To(Equal(len(keys)))
				for idx, rec := range records {
					if exList[idx].shouldExist {
						Expect(rec.Bins[bin.Name]).To(BeNil())
						Expect(rec.Generation).To(Equal(uint32(2)))
					} else {
						Expect(rec).To(BeNil())
					}
				}
			})

		}) // Batch Get Header context

		Context("Operate operations", func() {
			bin1 := as.NewBin("Aerospike1", rand.Intn(math.MaxInt16))
			bin2 := as.NewBin("Aerospike2", randString(100))

			BeforeEach(func() {
				// err = client.PutBins(wpolicy, key, bin)
				// Expect(err).ToNot(HaveOccurred())
			})

			It("must work correctly when no BinOps are passed as argument", func() {
				key, err := as.NewKey(ns, set, randString(50))
				Expect(err).ToNot(HaveOccurred())

				ops1 := []*as.Operation{}

				wpolicy := as.NewWritePolicy(0, 0)
				rec, err = client.Operate(wpolicy, key, ops1...)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("No operations were passed."))
			})

			It("must send key on Put operations", func() {
				key, err := as.NewKey(ns, set, randString(50))
				Expect(err).ToNot(HaveOccurred())

				ops1 := []*as.Operation{
					as.PutOp(bin1),
					as.PutOp(bin2),
					as.GetOp(),
				}

				wpolicy := as.NewWritePolicy(0, 0)
				wpolicy.SendKey = true
				rec, err = client.Operate(wpolicy, key, ops1...)
				Expect(err).ToNot(HaveOccurred())

				recordset, err := client.ScanAll(nil, key.Namespace(), key.SetName())
				Expect(err).ToNot(HaveOccurred())

				// make sure the result is what we put in
				for r := range recordset.Results() {
					Expect(r.Err).ToNot(HaveOccurred())
					if bytes.Equal(key.Digest(), r.Record.Key.Digest()) {
						Expect(r.Record.Key.Value()).To(Equal(key.Value()))
						Expect(r.Record.Bins).To(Equal(rec.Bins))
					}
				}
			})

			It("must send key on Touch operations", func() {
				key, err := as.NewKey(ns, set, randString(50))
				Expect(err).ToNot(HaveOccurred())

				ops1 := []*as.Operation{
					as.GetOp(),
					as.PutOp(bin2),
				}

				wpolicy := as.NewWritePolicy(0, 0)
				wpolicy.SendKey = false
				rec, err = client.Operate(wpolicy, key, ops1...)
				Expect(err).ToNot(HaveOccurred())

				recordset, err := client.ScanAll(nil, key.Namespace(), key.SetName())
				Expect(err).ToNot(HaveOccurred())

				// make sure the key is not saved
				for r := range recordset.Results() {
					Expect(r.Err).ToNot(HaveOccurred())
					if bytes.Equal(key.Digest(), r.Record.Key.Digest()) {
						Expect(r.Record.Key.Value()).To(BeNil())
					}
				}

				ops2 := []*as.Operation{
					as.GetOp(),
					as.TouchOp(),
				}
				wpolicy.SendKey = true
				rec, err = client.Operate(wpolicy, key, ops2...)
				Expect(err).ToNot(HaveOccurred())

				recordset, err = client.ScanAll(nil, key.Namespace(), key.SetName())
				Expect(err).ToNot(HaveOccurred())

				// make sure the
				for r := range recordset.Results() {
					Expect(r.Err).ToNot(HaveOccurred())
					if bytes.Equal(key.Digest(), r.Record.Key.Digest()) {
						Expect(r.Record.Key.Value()).To(Equal(key.Value()))
						Expect(r.Record.Bins).To(Equal(rec.Bins))
					}
				}
			})

			It("must apply all operations, and result should match expectation", func() {
				key, err := as.NewKey(ns, set, randString(50))
				Expect(err).ToNot(HaveOccurred())

				ops1 := []*as.Operation{
					as.PutOp(bin1),
					as.PutOp(bin2),
					as.GetOp(),
				}

				rec, err = client.Operate(nil, key, ops1...)
				Expect(err).ToNot(HaveOccurred())

				Expect(rec.Bins[bin1.Name]).To(Equal(bin1.Value.GetObject().(int)))
				Expect(rec.Bins[bin2.Name]).To(Equal(bin2.Value.GetObject().(string)))
				Expect(rec.Generation).To(Equal(uint32(1)))

				ops2 := []*as.Operation{
					as.AddOp(bin1),    // double the value of the bin
					as.AppendOp(bin2), // with itself
					as.GetOp(),
				}

				rec, err = client.Operate(nil, key, ops2...)
				Expect(err).ToNot(HaveOccurred())

				Expect(rec.Bins[bin1.Name]).To(Equal(bin1.Value.GetObject().(int) * 2))
				Expect(rec.Bins[bin2.Name]).To(Equal(strings.Repeat(bin2.Value.GetObject().(string), 2)))
				Expect(rec.Generation).To(Equal(uint32(2)))

				ops3 := []*as.Operation{
					as.AddOp(bin1),
					as.PrependOp(bin2),
					as.TouchOp(),
					as.GetOp(),
				}

				rec, err = client.Operate(nil, key, ops3...)
				Expect(err).ToNot(HaveOccurred())

				Expect(rec.Bins[bin1.Name]).To(Equal(bin1.Value.GetObject().(int) * 3))
				Expect(rec.Bins[bin2.Name]).To(Equal(strings.Repeat(bin2.Value.GetObject().(string), 3)))
				Expect(rec.Generation).To(Equal(uint32(3)))

				ops4 := []*as.Operation{
					as.TouchOp(),
					as.GetHeaderOp(),
				}

				rec, err = client.Operate(nil, key, ops4...)
				Expect(err).ToNot(HaveOccurred())

				Expect(rec.Generation).To(Equal(uint32(4)))
				Expect(len(rec.Bins)).To(Equal(0))

				// GetOp should override GetHEaderOp
				ops5 := []*as.Operation{
					as.GetOp(),
					as.GetHeaderOp(),
				}

				rec, err = client.Operate(nil, key, ops5...)
				Expect(err).ToNot(HaveOccurred())

				Expect(rec.Generation).To(Equal(uint32(4)))
				Expect(len(rec.Bins)).To(Equal(2))
			})

			It("must re-apply the same operations, and result should match expectation", func() {
				const listSize = 10
				const cdtBinName = "cdtBin"

				// First Part: For CDTs
				list := []interface{}{}
				opAppend := as.ListAppendOp(cdtBinName, 1)
				for i := 1; i <= listSize; i++ {
					list = append(list, i)

					sz, err := client.Operate(wpolicy, key, opAppend)
					Expect(err).ToNot(HaveOccurred())
					Expect(sz.Bins[cdtBinName]).To(Equal(i))
				}

				op := as.ListGetOp(cdtBinName, -1)
				cdtListRes, err := client.Operate(wpolicy, key, op)
				Expect(err).ToNot(HaveOccurred())
				Expect(cdtListRes.Bins[cdtBinName]).To(Equal(1))

				cdtListRes, err = client.Operate(wpolicy, key, op)
				Expect(err).ToNot(HaveOccurred())
				Expect(cdtListRes.Bins[cdtBinName]).To(Equal(1))

				// Second Part: For other normal Ops
				bin1 := as.NewBin("Aerospike1", 1)
				bin2 := as.NewBin("Aerospike2", "a")

				key, err := as.NewKey(ns, set, randString(50))
				Expect(err).ToNot(HaveOccurred())

				ops1 := []*as.Operation{
					as.PutOp(bin1),
					as.PutOp(bin2),
					as.GetOp(),
				}

				rec, err = client.Operate(nil, key, ops1...)
				Expect(err).ToNot(HaveOccurred())

				Expect(rec.Bins[bin1.Name]).To(Equal(bin1.Value.GetObject().(int)))
				Expect(rec.Bins[bin2.Name]).To(Equal(bin2.Value.GetObject().(string)))
				Expect(rec.Generation).To(Equal(uint32(1)))

				ops2 := []*as.Operation{
					as.AddOp(bin1),
					as.AppendOp(bin2),
					as.GetOp(),
				}

				rec, err = client.Operate(nil, key, ops2...)
				Expect(err).ToNot(HaveOccurred())

				Expect(rec.Bins[bin1.Name]).To(Equal(bin1.Value.GetObject().(int) + 1))
				Expect(rec.Bins[bin2.Name]).To(Equal(bin2.Value.GetObject().(string) + "a"))
				Expect(rec.Generation).To(Equal(uint32(2)))

				rec, err = client.Operate(nil, key, ops2...)
				Expect(err).ToNot(HaveOccurred())

				Expect(rec.Bins[bin1.Name]).To(Equal(bin1.Value.GetObject().(int) + 2))
				Expect(rec.Bins[bin2.Name]).To(Equal(bin2.Value.GetObject().(string) + "aa"))
				Expect(rec.Generation).To(Equal(uint32(3)))

			})

		}) // GetHeader context

	})
})

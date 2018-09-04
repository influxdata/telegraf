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

package aerospike

import (
	"math"
	"reflect"

	. "github.com/onsi/ginkgo"
	gm "github.com/onsi/gomega"

	// . "github.com/aerospike/aerospike-client-go"
	ParticleType "github.com/aerospike/aerospike-client-go/types/particle_type"
	. "github.com/aerospike/aerospike-client-go/utils/buffer"
)

type testBLOB struct {
	name string
}

func (b *testBLOB) EncodeBlob() ([]byte, error) {
	return append([]byte(b.name)), nil
}

func isValidIntegerValue(i int, v Value) bool {
	gm.Expect(reflect.TypeOf(v)).To(gm.Equal(reflect.TypeOf(NewIntegerValue(0))))
	gm.Expect(v.GetObject()).To(gm.Equal(i))
	gm.Expect(v.estimateSize()).To(gm.Equal(int(SizeOfInt)))
	gm.Expect(v.GetType()).To(gm.Equal(ParticleType.INTEGER))

	return true
}

func isValidLongValue(i int64, v Value) bool {
	gm.Expect(reflect.TypeOf(v)).To(gm.Equal(reflect.TypeOf(NewLongValue(0))))
	gm.Expect(v.GetObject().(int64)).To(gm.Equal(i))
	gm.Expect(v.estimateSize()).To(gm.Equal(int(SizeOfInt64)))
	gm.Expect(v.GetType()).To(gm.Equal(ParticleType.INTEGER))

	return true
}

func isValidFloatValue(i float64, v Value) bool {
	gm.Expect(reflect.TypeOf(v)).To(gm.Equal(reflect.TypeOf(NewFloatValue(0))))
	gm.Expect(v.GetObject().(float64)).To(gm.Equal(i))
	gm.Expect(v.estimateSize()).To(gm.Equal(8))
	gm.Expect(v.GetType()).To(gm.Equal(ParticleType.FLOAT))

	return true
}

var _ = Describe("Value Test", func() {

	Context("NullValue", func() {
		It("should create a valid NullValue", func() {
			v := NewValue(nil)

			gm.Expect(v.GetObject()).To(gm.BeNil())
			gm.Expect(v.estimateSize()).To(gm.Equal(0))
			gm.Expect(v.GetType()).To(gm.Equal(ParticleType.NULL))
		})
	})

	Context("StringValues", func() {
		It("should create a valid string value", func() {
			str := "string value"
			v := NewValue(str)

			gm.Expect(v.GetObject()).To(gm.Equal(str))
			gm.Expect(v.estimateSize()).To(gm.Equal(len(str)))
			gm.Expect(v.GetType()).To(gm.Equal(ParticleType.STRING))
		})

		It("should create a valid empty string value", func() {
			str := ""
			v := NewValue(str)

			gm.Expect(v.GetObject()).To(gm.Equal(str))
			gm.Expect(v.estimateSize()).To(gm.Equal(len(str)))
			gm.Expect(v.GetType()).To(gm.Equal(ParticleType.STRING))
		})
	})

	Context("Blob Values", func() {

		It("should create a BytesValue on valid types, and encode", func() {
			person := &testBLOB{name: "SomeDude"}

			bval := NewValue(person)
			gm.Expect(bval.GetType()).To(gm.Equal(ParticleType.BLOB))
			gm.Expect(bval).To(gm.BeAssignableToTypeOf(BytesValue{}))
			gm.Expect(bval.GetObject()).To(gm.Equal([]byte(person.name)))
		})
	})

	Context("Numeric Values", func() {

		It("should create a valid IntegerValue on boundries of int8", func() {
			i := int8(math.MinInt8)
			v := NewValue(i)
			isValidIntegerValue(int(i), v)

			i = int8(math.MaxInt8)
			v = NewValue(i)
			isValidIntegerValue(int(i), v)
		})

		It("should create a valid IntegerValue on boundries of uint8", func() {
			i := uint8(0)
			v := NewValue(i)
			isValidIntegerValue(int(i), v)

			i = uint8(math.MaxUint8)
			v = NewValue(i)
			isValidIntegerValue(int(i), v)
		})

		It("should create a valid IntegerValue on boundries of int16", func() {
			i := int16(math.MinInt16)
			v := NewValue(i)
			isValidIntegerValue(int(i), v)

			i = int16(math.MaxInt16)
			v = NewValue(i)
			isValidIntegerValue(int(i), v)
		})

		It("should create a valid IntegerValue on boundries of uint16", func() {
			i := uint16(0)
			v := NewValue(i)
			isValidIntegerValue(int(i), v)

			i = uint16(math.MaxUint16)
			v = NewValue(i)
			isValidIntegerValue(int(i), v)
		})

		It("should create a valid IntegerValue on boundries of int32", func() {
			i := int32(math.MinInt32)
			v := NewValue(i)
			isValidIntegerValue(int(i), v)

			i = int32(math.MaxInt32)
			v = NewValue(i)
			isValidIntegerValue(int(i), v)
		})

		It("should create a valid IntegerValue on boundries of native int on 32 bit machines", func() {
			if Arch32Bits {
				i := math.MinInt32
				v := NewValue(i)
				isValidIntegerValue(i, v)

				i = math.MaxInt32
				v = NewValue(i)
				isValidIntegerValue(i, v)
			}
		})

		It("should create a valid LongValue after boundries of int32 is passed on 32 bit machines", func() {
			if Arch32Bits {
				i := math.MinInt32 - 1
				v := NewValue(i)
				isValidLongValue(int64(i), v)

				i = math.MaxInt32 + 1
				v = NewValue(i)
				isValidLongValue(int64(i), v)
			}
		})

		It("should create a valid IntegerValue on boundries of native int on 64 bit machines", func() {
			if Arch64Bits {
				i := math.MinInt64
				v := NewValue(i)
				isValidIntegerValue(i, v)

				i = math.MaxInt64
				v = NewValue(i)
				isValidIntegerValue(i, v)
			}
		})

		It("should create a valid LongValue on boundries of int64", func() {
			i := int64(math.MinInt64)
			v := NewValue(i)
			isValidLongValue(i, v)

			i = int64(math.MaxInt64)
			v = NewValue(i)
			isValidLongValue(i, v)
		})

		It("should create a valid FloatValue on boundries of float64", func() {
			i := float64(-math.MaxFloat64)
			v := NewValue(i)
			isValidFloatValue(i, v)

			i = float64(math.MaxFloat64)
			v = NewValue(i)
			isValidFloatValue(i, v)
		})

	}) // numeric values context
})

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
	"math/rand"
	"reflect"

	. "github.com/onsi/gomega"
)

type testBLOB struct {
	name string
}

func (tb *testBLOB) EncodeBlob() ([]byte, error) {
	return append([]byte(tb.name)), nil
}

// generates a random string of specified length
func randString(size int) string {
	const random_alpha_num = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const l = 62
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = random_alpha_num[rand.Intn(l)]
	}
	return string(buf)
}

func normalizeValue(v interface{}, isMapKey bool) interface{} {
	if v != nil {
		switch v.(type) {
		case int8, int16, int32, int, int64:
			return reflect.ValueOf(v).Int()
		case uint8, uint16, uint32, uint:
			return int64(reflect.ValueOf(v).Uint())
		}

		// check for array and map
		switch reflect.TypeOf(v).Kind() {
		case reflect.Array:
			if isMapKey {
				return v
			}
			return sliceToIfcSlice(v)
		case reflect.Slice:
			return sliceToIfcSlice(v)
		case reflect.Map:
			return mapToIfcMap(v)
		}

		return v
	}

	return nil
}

func mapToIfcMap(v interface{}) map[interface{}]interface{} {
	s := reflect.ValueOf(v)
	l := s.Len()
	res := make(map[interface{}]interface{}, l)
	for _, k := range s.MapKeys() {
		v := s.MapIndex(k).Interface()
		res[normalizeValue(k.Interface(), true)] = normalizeValue(v, false)
	}

	return res
}

func sliceToIfcSlice(v interface{}) []interface{} {
	s := reflect.ValueOf(v)
	l := s.Len()
	res := make([]interface{}, l)
	for i := 0; i < l; i++ {
		t := s.Index(i).Interface()
		res[i] = normalizeValue(t, false)
	}

	return res
}

func arraysEqual(ia, ib interface{}) {
	a := sliceToIfcSlice(ia)
	b := sliceToIfcSlice(ib)

	Expect(len(a)).To(Equal(len(b)))
	// Expect(a).To(BeEquivalentTo(b))

	for i := range a {
		switch reflect.ValueOf(a[i]).Kind() {
		case reflect.Map:
			mapsEqual(a[i], b[i])
		case reflect.Slice:
			arraysEqual(a[i], b[i])
		default:
			if a[i] != nil {
				Expect(a[i]).To(BeEquivalentTo(b[i]))
			} else {
				Expect(b[i]).To(BeNil())
			}
		}
	}
}

func mapsEqual(ia, ib interface{}) {
	a := mapToIfcMap(ia)
	b := mapToIfcMap(ib)

	Expect(len(a)).To(Equal(len(b)))
	Expect(a).To(BeEquivalentTo(b))
}

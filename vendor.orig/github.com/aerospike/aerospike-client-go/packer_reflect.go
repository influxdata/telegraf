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

package aerospike

import (
	"fmt"
	"reflect"
)

func init() {
	__packObjectReflect = __concretePackObjectReflect
}

func __concretePackObjectReflect(cmd BufferEx, obj interface{}, mapKey bool) (int, error) {
	// check for array and map
	rv := reflect.ValueOf(obj)
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Array, reflect.Slice:
		if mapKey && reflect.TypeOf(obj).Kind() == reflect.Slice {
			panic(fmt.Sprintf("Maps, Slices, and bounded arrays other than Bounded Byte Arrays are not supported as Map keys. Value: %#v", obj))
		}
		// pack bounded array of bytes differently
		if reflect.TypeOf(obj).Kind() == reflect.Array && reflect.TypeOf(obj).Elem().Kind() == reflect.Uint8 {
			l := rv.Len()
			arr := make([]byte, l)
			for i := 0; i < l; i++ {
				arr[i] = rv.Index(i).Interface().(uint8)
			}
			return __PackBytes(cmd, arr)
		}

		l := rv.Len()
		arr := make([]interface{}, l)
		for i := 0; i < l; i++ {
			arr[i] = rv.Index(i).Interface()
		}
		return __PackIfcList(cmd, arr)
	case reflect.Map:
		if mapKey {
			panic(fmt.Sprintf("Maps, Slices, and bounded arrays other than Bounded Byte Arrays are not supported as Map keys. Value: %#v", obj))
		}
		l := rv.Len()
		amap := make(map[interface{}]interface{}, l)
		for _, i := range rv.MapKeys() {
			amap[i.Interface()] = rv.MapIndex(i).Interface()
		}
		return __PackIfcMap(cmd, amap)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return __PackObject(cmd, rv.Int(), false)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return __PackObject(cmd, rv.Uint(), false)
	case reflect.Bool:
		return __PackObject(cmd, rv.Bool(), false)
	case reflect.String:
		return __PackObject(cmd, rv.String(), false)
	case reflect.Float32, reflect.Float64:
		return __PackObject(cmd, rv.Float(), false)
	}

	panic(fmt.Sprintf("Type `%#v` not supported to pack.", obj))
}

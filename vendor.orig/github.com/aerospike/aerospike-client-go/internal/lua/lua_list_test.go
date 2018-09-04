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

package lua_test

import (
	"github.com/yuin/gopher-lua"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/aerospike/aerospike-client-go/internal/lua"
)

var _ = Describe("Lua List API Test", func() {

	// code vs result
	testMatrix := map[string]interface{}{
		"l = List()\n return l":           []interface{}{},
		"l = List.create()\n return l":    []interface{}{},
		"l = List.create(100)\n return l": make([]interface{}, 0, 100),

		"l = list()\n return l":           []interface{}{},
		"l = list.create()\n return l":    []interface{}{},
		"l = list.create(100)\n return l": make([]interface{}, 0, 100),
		"l = list({1,2})\n return l":      []interface{}{float64(1), float64(2)},

		"l = list({1,2})\n l[1] = 5\n return l": []interface{}{float64(5), float64(2)},
		"l = list({1,2})\n return l[1]":         float64(1),

		"l = list()\n return list.size(l)":           float64(0),
		"l = list.create()\n return list.size(l)":    float64(0),
		"l = list.create(100)\n return list.size(l)": float64(0),
		"l = list({1,2})\n return list.size(l)":      float64(2),

		"l = list{1,2}\n list.insert(l, 1, 0)\n return l": []interface{}{float64(0), float64(1), float64(2)},
		"l = list{1,2}\n list.insert(l, 2, 0)\n return l": []interface{}{float64(1), float64(0), float64(2)},
		"l = list{1,2}\n list.insert(l, 3, 0)\n return l": []interface{}{float64(1), float64(2), float64(0)},

		"l = list{1,2}\n list.append(l, 3)\n return l":                    []interface{}{float64(1), float64(2), float64(3)},
		"l = list{1,2}\n list.append(l, 3)\nlist.append(l, 4)\n return l": []interface{}{float64(1), float64(2), float64(3), float64(4)},

		"l = list{1,2}\n list.prepend(l, 0)\n return l":                     []interface{}{float64(0), float64(1), float64(2)},
		"l = list{1,2}\n list.prepend(l, 3)\nlist.prepend(l, 4)\n return l": []interface{}{float64(4), float64(3), float64(1), float64(2)},

		"l = list{1,2}\n return list.take(l, 1)":                      []interface{}{float64(1)},
		"l = list{1,2}\n return list.take(l, 2)":                      []interface{}{float64(1), float64(2)},
		"l = list{1,2}\n return list.take(l, 3)":                      []interface{}{float64(1), float64(2)},
		"l = list{1,2}\n list.take(l, 1)\nlist.take(l, 2)\n return l": []interface{}{float64(1), float64(2)},

		"l = list{1,2}\n list.remove(l, 1)\n return l":                    []interface{}{float64(2)},
		"l = list{1,2}\n list.remove(l, 2)\n return l":                    []interface{}{float64(1)},
		"l = list{1,2}\n list.remove(l, 1)\nlist.remove(l, 1)\n return l": []interface{}{},

		"l = list{1,2}\n list.drop(l, 1)\n return l":              []interface{}{float64(1), float64(2)},
		"l = list{1,2}\n return list.drop(l, 1)":                  []interface{}{float64(2)},
		"l = list{1,2}\n return list.drop(l, 2)":                  []interface{}{},
		"l = list{1,2}\n return list.drop(l, 5)":                  []interface{}{},
		"l = list{1,2}\n list.drop(l, 1)\nreturn list.drop(l, 1)": []interface{}{float64(2)},

		"l = list{1,2}\n list.trim(l, 1)\n return l": []interface{}{},
		"l = list{1,2}\n list.trim(l, 2)\n return l": []interface{}{float64(1)},

		"l = list{1,2}\n return list.clone(l)": []interface{}{float64(1), float64(2)},

		"l1 = list{1,2}\n l2 = list{3,4}\n list.concat(l1, l2)\n return l1": []interface{}{float64(1), float64(2), float64(3), float64(4)},
		"l1 = list{3,4}\n l2 = list{1,2}\n list.concat(l1, l2)\n return l1": []interface{}{float64(3), float64(4), float64(1), float64(2)},

		"l1 = list{1,2}\n l2 = list{3,4}\n return list.merge(l1, l2)": []interface{}{float64(1), float64(2), float64(3), float64(4)},
		"l1 = list{3,4}\n l2 = list{1,2}\n return list.merge(l1, l2)": []interface{}{float64(3), float64(4), float64(1), float64(2)},

		"l = list{1,2,3,4,5}\n cnt = 0\nfor value in list.iterator(l) do\n\t cnt = cnt + value\n end\n return cnt": float64(15),

		"l = list{1,2,3,4,5}\n return tostring(l)": "[1 2 3 4 5]",
	}

	// following expressions should return an error
	errMatrix := []string{
		"l = list{1,2}\n list.remove(l, 3)",
	}

	It("must run all code blocks", func() {
		instance := LuaPool.Get().(*lua.LState)
		defer instance.Close()
		for source, expected := range testMatrix {

			err := instance.DoString(source)
			Expect(err).NotTo(HaveOccurred())

			By(source)
			Expect(LValueToInterface(instance.CheckAny(-1))).To(Equal(expected))
			instance.Pop(1) // remove received value
		}

	})

	It("must fail all code blocks", func() {
		instance := LuaPool.Get().(*lua.LState)
		defer instance.Close()
		for _, source := range errMatrix {
			By(source)

			err := instance.DoString(source)
			Expect(err).To(HaveOccurred())

		}

	})

})

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

var _ = Describe("Lua Map API Test", func() {

	// code vs result
	testMatrix := map[string]interface{}{
		"m = Map()\n return m": map[interface{}]interface{}{},

		"m = map()\n return m":             map[interface{}]interface{}{},
		"m = map{x = 1, y = 2}\n return m": map[interface{}]interface{}{"x": float64(1), "y": float64(2)},
		"m = map.create(100)\n return m":   make(map[interface{}]interface{}, 100),

		"m = map({x=1,y=2})\n return m['x']":         float64(1),
		"m = map({x=1,y=2})\n m['x'] = 5\n return m": map[interface{}]interface{}{"x": float64(5), "y": float64(2)},

		"m = map()\n return map.size(m)":           float64(0),
		"m = map.create(100)\n return map.size(m)": float64(0),
		"m = map({x=1,y=2})\n return map.size(m)":  float64(2),

		"m = map{x=1,y=2,z=3}\n cnt = 0\nfor k, v in map.pairs(m) do\n\t cnt = cnt + v\n end\n return cnt": float64(6),

		"m = map{x=1,y=2,z=3}\n str = ''\nfor k in map.keys(m) do\n\t str = str .. k\n end\n return string.len(str)": float64(3),
		"m = map{x=1,y=2,z=3}\n cnt = 0\nfor v in map.values(m) do\n\t cnt = cnt + v\n end\n return cnt":             float64(6),

		"m = map{x=1,y=2}\n map.remove(m, 'x')\n return m":                     map[interface{}]interface{}{"y": float64(2)},
		"m = map{x=1,y=2}\n map.remove(m, 'y')\n return m":                     map[interface{}]interface{}{"x": float64(1)},
		"m = map{x=1,y=2}\n map.remove(m, 'x')\nmap.remove(m, 'y')\n return m": map[interface{}]interface{}{},
		"m = map{x=1,y=2}\n map.remove(m, 'z')\nmap.remove(m, 't')\n return m": map[interface{}]interface{}{"x": float64(1), "y": float64(2)},

		"m1 = map({x=1,y=2})\n m2 = map.clone(m1)\n return map.size(m2)": float64(2),

		"m1 = map{x=1,y=2}\n m2 = map{a=3,b=4}\n return map.merge(m1, m2)":                                          map[interface{}]interface{}{"x": float64(1), "y": float64(2), "a": float64(3), "b": float64(4)},
		"m1 = map{x=1,y=2}\n m2 = map{x=3,y=4}\n return map.merge(m1, m2, function(v1, v2)\n return v1 + v2\n end)": map[interface{}]interface{}{"x": float64(4), "y": float64(6)},
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

})

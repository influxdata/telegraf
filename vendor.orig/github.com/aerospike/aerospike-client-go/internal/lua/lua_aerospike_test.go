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

var _ = Describe("Lua Aerospike API Test", func() {

	// code vs result
	testMatrix := map[string]interface{}{
		"aerospike.log(1, 'Warn')": nil,
		"warn('Warn %d', 1)":       nil,

		"aerospike.log(2, 'Info')": nil,
		"info('Info %d', 2)":       nil,

		"aerospike.log(3, 'Debug')": nil,
		"trace('Trace %d', 3)":      nil,

		"aerospike.log(4, 'Debug')": nil,
		"debug('Debug %d', 4)":      nil,
	}

	It("must run all code blocks", func() {
		instance := LuaPool.Get().(*lua.LState)
		defer instance.Close()
		for source := range testMatrix {
			err := instance.DoString(source)
			Expect(err).NotTo(HaveOccurred())
		}

	})

})

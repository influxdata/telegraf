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
	"encoding/hex"

	as "github.com/aerospike/aerospike-client-go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// ALL tests are isolated by SetName and Key, which are 50 random characters
var _ = Describe("Key Test Reflection", func() {
	initTestVars()

	Context("Digests should be the same", func() {

		It("for Arrays", func() {

			// The following two cases should be in exact order
			key, _ := as.NewKey("namespace", "set", []int{1, 2, 3})
			Expect(hex.EncodeToString(key.Digest())).To(Equal("a8b63a8208ebebb49d027d51899121fd0d03d2f7"))

			keyInterfaceArrayOfTheSameValues, _ := as.NewKey("namespace", "set", []interface{}{1, 2, 3})
			Expect(hex.EncodeToString(keyInterfaceArrayOfTheSameValues.Digest())).To(Equal(hex.EncodeToString(key.Digest())))

		})

	})

})

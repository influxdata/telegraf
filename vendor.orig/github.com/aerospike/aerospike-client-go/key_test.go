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
	"math"
	"strings"

	as "github.com/aerospike/aerospike-client-go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// ALL tests are isolated by SetName and Key, which are 50 random characters
var _ = Describe("Key Test", func() {
	initTestVars()

	Context("Digests should be the same", func() {

		It("for Integers", func() {

			key, _ := as.NewKey("namespace", "set", math.MinInt64)
			Expect(hex.EncodeToString(key.Digest())).To(Equal("7185c2a47fb02c996daed26b4e01b83240aee9d4"))

			key, _ = as.NewKey("namespace", "set", math.MaxInt64)
			Expect(hex.EncodeToString(key.Digest())).To(Equal("1698328974afa62c8e069860c1516f780d63dbb8"))

			key, _ = as.NewKey("namespace", "set", math.MinInt32)
			Expect(hex.EncodeToString(key.Digest())).To(Equal("d635a867b755f8f54cdc6275e6fb437df82a728c"))

			key, _ = as.NewKey("namespace", "set", math.MaxInt32)
			Expect(hex.EncodeToString(key.Digest())).To(Equal("fa8c47b8b898af1bbcb20af0d729ca68359a2645"))

			key, _ = as.NewKey("namespace", "set", math.MinInt16)
			Expect(hex.EncodeToString(key.Digest())).To(Equal("7f41e9dd1f3fe3694be0430e04c8bfc7d51ec2af"))

			key, _ = as.NewKey("namespace", "set", math.MaxInt16)
			Expect(hex.EncodeToString(key.Digest())).To(Equal("309fc9c2619c4f65ff7f4cd82085c3ee7a31fc7c"))

			key, _ = as.NewKey("namespace", "set", math.MinInt8)
			Expect(hex.EncodeToString(key.Digest())).To(Equal("93191e549f8f3548d7e2cfc958ddc8c65bcbe4c6"))

			key, _ = as.NewKey("namespace", "set", math.MaxInt8)
			Expect(hex.EncodeToString(key.Digest())).To(Equal("a58f7d98bf60e10fe369c82030b1c9dee053def9"))

			key, _ = as.NewKey("namespace", "set", -1)
			Expect(hex.EncodeToString(key.Digest())).To(Equal("22116d253745e29fc63fdf760b6e26f7e197e01d"))

			key, _ = as.NewKey("namespace", "set", 0)
			Expect(hex.EncodeToString(key.Digest())).To(Equal("93d943aae37b017ad7e011b0c1d2e2143c2fb37d"))

		})

		It("for Strings", func() {

			key, _ := as.NewKey("namespace", "set", "")
			Expect(hex.EncodeToString(key.Digest())).To(Equal("2819b1ff6e346a43b4f5f6b77a88bc3eaac22a83"))

			key, _ = as.NewKey("namespace", "set", strings.Repeat("s", 1))
			Expect(hex.EncodeToString(key.Digest())).To(Equal("607cddba7cd111745ef0a3d783d57f0e83c8f311"))

			key, _ = as.NewKey("namespace", "set", strings.Repeat("a", 10))
			Expect(hex.EncodeToString(key.Digest())).To(Equal("5979fb32a80da070ff356f7695455592272e36c2"))

			key, _ = as.NewKey("namespace", "set", strings.Repeat("m", 100))
			Expect(hex.EncodeToString(key.Digest())).To(Equal("f00ad7dbcb4bd8122d9681bca49b8c2ffd4beeed"))

			key, _ = as.NewKey("namespace", "set", strings.Repeat("t", 1000))
			Expect(hex.EncodeToString(key.Digest())).To(Equal("07ac412d4c33b8628ab147b8db244ce44ae527f8"))

			key, _ = as.NewKey("namespace", "set", strings.Repeat("-", 10000))
			Expect(hex.EncodeToString(key.Digest())).To(Equal("b42e64afbfccb05912a609179228d9249ea1c1a0"))

			key, _ = as.NewKey("namespace", "set", strings.Repeat("+", 100000))
			Expect(hex.EncodeToString(key.Digest())).To(Equal("0a3e888c20bb8958537ddd4ba835e4070bd51740"))

		})

		It("for []byte", func() {

			key, _ := as.NewKey("namespace", "set", []byte{})
			Expect(hex.EncodeToString(key.Digest())).To(Equal("327e2877b8815c7aeede0d5a8620d4ef8df4a4b4"))

			key, _ = as.NewKey("namespace", "set", []byte(strings.Repeat("s", 1)))
			Expect(hex.EncodeToString(key.Digest())).To(Equal("ca2d96dc9a184d15a7fa2927565e844e9254e001"))

			key, _ = as.NewKey("namespace", "set", []byte(strings.Repeat("a", 10)))
			Expect(hex.EncodeToString(key.Digest())).To(Equal("d10982327b2b04c7360579f252e164a75f83cd99"))

			key, _ = as.NewKey("namespace", "set", []byte(strings.Repeat("m", 100)))
			Expect(hex.EncodeToString(key.Digest())).To(Equal("475786aa4ee664532a7d1ea69cb02e4695fcdeed"))

			key, _ = as.NewKey("namespace", "set", []byte(strings.Repeat("t", 1000)))
			Expect(hex.EncodeToString(key.Digest())).To(Equal("5a32b507518a49bf47fdaa3deca53803f5b2e8c3"))

			key, _ = as.NewKey("namespace", "set", []byte(strings.Repeat("-", 10000)))
			Expect(hex.EncodeToString(key.Digest())).To(Equal("ed65c63f7a1f8c6697eb3894b6409a95461fd982"))

			key, _ = as.NewKey("namespace", "set", []byte(strings.Repeat("+", 100000)))
			Expect(hex.EncodeToString(key.Digest())).To(Equal("fe19770c371774ba1a1532438d4851b8a773a9e6"))

		})

		It("for Arrays", func() {

			key, _ := as.NewKey("namespace", "set", []interface{}{})
			Expect(hex.EncodeToString(key.Digest())).To(Equal("2af0111192df4ca297232d1641ff52c2ce51ce2d"))

			key, _ = as.NewKey("namespace", "set", []interface{}{1, []byte{1, 17}, "str"})
			Expect(hex.EncodeToString(key.Digest())).To(Equal("8f5129e079cf66333a8372192d93072a4c661be2"))

		})

		It("for custom digest", func() {
			key, _ := as.NewKey("namespace", "set", []interface{}{})
			Expect(hex.EncodeToString(key.Digest())).To(Equal("2af0111192df4ca297232d1641ff52c2ce51ce2d"))
			err := key.SetDigest([]byte("01234567890123456789"))
			Expect(err, nil)
			Expect(key.Digest()).To(Equal([]byte("01234567890123456789")))

			key, _ = as.NewKeyWithDigest("namespace", "set", []interface{}{}, []byte("01234567890123456789"))
			Expect(key.Digest()).To(Equal([]byte("01234567890123456789")))
		})

	})

})

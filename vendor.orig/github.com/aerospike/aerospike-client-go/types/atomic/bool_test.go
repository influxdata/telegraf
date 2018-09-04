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

package atomic_test

import (
	"runtime"

	. "github.com/aerospike/aerospike-client-go/types/atomic"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Atomic Bool", func() {
	// atomic tests require actual parallelism
	runtime.GOMAXPROCS(runtime.NumCPU())

	var ab *AtomicBool

	BeforeEach(func() {
		ab = NewAtomicBool(true)
	})

	It("must CompareAndToggle correctly", func() {
		Expect(ab.CompareAndToggle(true)).To(BeTrue())
		Expect(ab.CompareAndToggle(true)).To(BeFalse())
	})

})

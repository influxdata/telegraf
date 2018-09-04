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
	"errors"

	. "github.com/onsi/ginkgo"
	gm "github.com/onsi/gomega"
)

// ALL tests are isolated by SetName and Key, which are 50 random characters
var _ = Describe("Recordset test", func() {

	It("must avoid panic on sendError", func() {
		rs := newRecordset(100, 1, 1)

		rs.sendError(errors.New("Error"))
		rs.wgGoroutines.Done()
		rs.Close()
		rs.sendError(errors.New("Error"))

		gm.Expect(<-rs.Errors).NotTo(gm.BeNil())
		// Expect(<-rs.Errors).To(BeNil())
	})

})

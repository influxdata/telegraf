// Copyright 2016 The Linux Foundation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package schema_test

import (
	"strings"
	"testing"

	"github.com/opencontainers/image-spec/schema"
)

func TestImageLayout(t *testing.T) {
	for i, tt := range []struct {
		imageLayout string
		fail        bool
	}{
		// expected faulure:  imageLayoutVersion does not match pattern
		{
			imageLayout: `
{
  "imageLayoutVersion": 1.0.0
}
`,
			fail: true,
		},

		// validate layout
		{
			imageLayout: `
{
  "imageLayoutVersion": "1.0.0"
}
`,
			fail: false,
		},
	} {
		r := strings.NewReader(tt.imageLayout)
		err := schema.ValidatorMediaTypeLayoutHeader.Validate(r)

		if got := err != nil; tt.fail != got {
			t.Errorf("test %d: expected validation failure %t but got %t, err %v", i, tt.fail, got, err)
		}
	}
}

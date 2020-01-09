//Copyright (c) 2019, Groupon, Inc.
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without
//modification, are permitted provided that the following conditions are
//met:
//
//Redistributions of source code must retain the above copyright notice,
//this list of conditions and the following disclaimer.
//
//Redistributions in binary form must reproduce the above copyright
//notice, this list of conditions and the following disclaimer in the
//documentation and/or other materials provided with the distribution.
//
//Neither the name of GROUPON nor the names of its contributors may be
//used to endorse or promote products derived from this software without
//specific prior written permission.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS
//IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
//TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
//PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
//HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
//SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
//TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
//PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
//LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
//NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package tdigestagg

import (
	"github.com/influxdata/tdigest"
	"github.com/influxdata/telegraf/plugins/aggregators/tdigestagg/bucketing"
	"github.com/influxdata/telegraf/plugins/aggregators/tdigestagg/constants"
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)


var mLocal, _ = metric.New("m1",
	map[string]string{
		tagKeyAZ:                valueSeattle,
		tagKeyEnv:               valueDevelopment,
		constants.TagKeyHost:    valueHost,
		constants.TagKeyService: valueService,
		constants.TagKeyRollup:  "local:*",
	},
	map[string]interface{}{
		"a": float64(9),
		"b": float64(8),
		"c": float64(7),
	},
	time.Now(),
)

func TestLocalTag(t *testing.T) {
	acc := testutil.Accumulator{}
	histogram := NewTDigestAgg()
	histogram.Bucketing = []bucketing.BucketConfig{{
		ExcludeTags:           nil,
		SourceTagKey:          constants.TagKeyService,
		AtomReplacementTagKey: constants.TagKeyHost,
	}}

	histogram.Add(mLocal)
	histogram.Push(&acc)

	var expected = tdigest.NewWithCompression(0)
	expected.Add(9, 1)
	//out := expected.Centroids()

	expectedFields := map[string]interface{}{
		constants.FieldMaximum: 9.,
		constants.FieldMinimum: 9.,
		constants.FieldCount:   1.,
		constants.FieldMedian:  9.,
	}
	expectedTags := map[string]string{
		tagKeyAZ:                valueSeattle,
		tagKeyEnv:               valueDevelopment,
		constants.TagKeyHost:    valueHost,
		constants.TagKeyService: valueService,
		constants.TagKeySource:  valueHost,
		constants.TagKeyAtom:    valueHost,
	}
	acc.AssertContainsTaggedFields(t, "m1_a", expectedFields, expectedTags)

	expectedTags[constants.TagKeySource] = valueService
	acc.AssertDoesNotContainsTaggedFields(t, "m1_a", expectedFields, expectedTags)
}

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

package bucketing

import (
	"github.com/docker/docker/pkg/testutil/assert"
	"github.groupondev.com/metrics/telegraf-tdigest-plugin/plugins/aggregators/tdigestagg/constants"
	"testing"
)

func TestAggregates(t *testing.T) {
	bucketOutput := BucketOutputTags{
		aggregationDimensions: nil,
		otherTags: map[string]string{
			constants.TagKeyAggregates: "commaSeparatedList",
		},
	}

	assert.Equal(t, bucketOutput.Aggregates(), "commaSeparatedList")
}

func TestBucketKey(t *testing.T) {
	bucketOutput := BucketOutputTags{
		aggregationDimensions: nil,
		otherTags: map[string]string{
			constants.TagKeyBucketKey: "underscoreSeparatedList",
		},
	}

	assert.Equal(t, bucketOutput.BucketKey(), "underscoreSeparatedList")
}

func TestOutputTags(t *testing.T) {
	bucketOutput := BucketOutputTags{
		aggregationDimensions: map[string]string{
			"aggregationKey": "aggregationValue",
		},
		otherTags: map[string]string{
			"otherKey": "otherValue",
		},
	}

	outputTags := bucketOutput.OutputTags()
	assert.Equal(t, outputTags["aggregationKey"], "aggregationValue")
	assert.Equal(t, outputTags["otherKey"], "otherValue")
}

func TestSetTag(t *testing.T) {
	bucketOutput := BucketOutputTags{
		aggregationDimensions: make(map[string]string),
		otherTags:             make(map[string]string),
	}

	bucketOutput.aggregationDimensions["k1"] = "v1"
	bucketOutput.otherTags["k2"] = "v2"
	assert.Equal(t, bucketOutput.aggregationDimensions["k1"], "v1")
	assert.Equal(t, bucketOutput.otherTags["k2"], "v2")
}

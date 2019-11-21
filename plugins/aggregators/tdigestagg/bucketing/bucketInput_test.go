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
	"github.com/influxdata/telegraf/plugins/aggregators/tdigestagg/constants"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractAtomTag(t *testing.T) {
	bucketInput := BucketInput{
		config: BucketConfig{
			ExcludeTags:  nil,
			SourceTagKey: "k1",
		},
		allTags: map[string]string{
			"k1":   "v1",
			"k2":   "v2",
			"k3":   "v3",
			"atom": "atomValue",
		},
	}

	bucketOutput := BucketOutputTags{
		aggregationDimensions: make(map[string]string),
		otherTags:             make(map[string]string),
	}
	bucketInput.extractAtomTag(&bucketOutput)

	assert.Equal(t, bucketOutput.OutputTags()["atom"], "atomValue")
}

func TestExtractReplacementAtomTag(t *testing.T) {
	bucketInput := BucketInput{
		config: BucketConfig{
			ExcludeTags:           nil,
			SourceTagKey:          "k1",
			AtomReplacementTagKey: "k2",
		},
		allTags: map[string]string{
			"k1": "v1",
			"k2": "v2",
			"k3": "v3",
		},
	}

	bucketOutput := BucketOutputTags{
		aggregationDimensions: make(map[string]string),
		otherTags:             make(map[string]string),
	}
	bucketInput.extractAtomTag(&bucketOutput)

	assert.Equal(t, bucketOutput.OutputTags()["atom"], "v2")
}

func TestExtractAtomTagAtomOverridesReplacement(t *testing.T) {
	bucketInput := BucketInput{
		config: BucketConfig{
			ExcludeTags:           nil,
			SourceTagKey:          "k1",
			AtomReplacementTagKey: "k2",
		},
		allTags: map[string]string{
			"k1":   "v1",
			"k2":   "v2",
			"k3":   "v3",
			"atom": "atomValue",
		},
	}

	bucketOutput := BucketOutputTags{
		aggregationDimensions: make(map[string]string),
		otherTags:             make(map[string]string),
	}
	bucketInput.extractAtomTag(&bucketOutput)

	assert.Equal(t, bucketOutput.OutputTags()["atom"], "atomValue")
}

func TestExcludeAtomTag(t *testing.T) {
	bucketInput := BucketInput{
		config: BucketConfig{
			ExcludeTags:           []string{"atom"},
			SourceTagKey:          "k1",
			AtomReplacementTagKey: "k2",
		},
		allTags: map[string]string{
			"k1":   "v1",
			"k2":   "v2",
			"k3":   "v3",
			"atom": "atomValue",
		},
	}

	bucketOutput := BucketOutputTags{
		aggregationDimensions: make(map[string]string),
		otherTags:             make(map[string]string),
	}
	bucketInput.extractAtomTag(&bucketOutput)

	assert.Equal(t, bucketOutput.OutputTags()["atom"], "v2")
}

func TestExtractMissingAtomTagNoOverride(t *testing.T) {
	bucketInput := BucketInput{
		config: BucketConfig{
			ExcludeTags:  nil,
			SourceTagKey: "k1",
		},
		allTags: map[string]string{
			"k1": "v1",
			"k2": "v2",
			"k3": "v3",
		},
	}

	bucketOutput := BucketOutputTags{
		aggregationDimensions: make(map[string]string),
		otherTags:             make(map[string]string),
	}
	bucketInput.extractAtomTag(&bucketOutput)

	assert.Equal(t, bucketOutput.OutputTags()["atom"], "MISSING_atom")
}

func TestExtractMissingAtomTagWithOverride(t *testing.T) {
	bucketInput := BucketInput{
		config: BucketConfig{
			ExcludeTags:           nil,
			SourceTagKey:          "k1",
			AtomReplacementTagKey: "k2",
		},
		allTags: map[string]string{
			"k1": "v1",
			//"k2":   "v2",
			"k3": "v3",
		},
	}

	bucketOutput := BucketOutputTags{
		aggregationDimensions: make(map[string]string),
		otherTags:             make(map[string]string),
	}
	bucketInput.extractAtomTag(&bucketOutput)

	assert.Equal(t, bucketOutput.OutputTags()["atom"], "MISSING_k2")
}

func TestExtractSourceTag(t *testing.T) {
	bucketInput := BucketInput{
		config: BucketConfig{
			ExcludeTags:  nil,
			SourceTagKey: "k1",
		},
		allTags: map[string]string{
			"k1":   "v1",
			"k2":   "v2",
			"k3":   "v3",
			"atom": "atomValue",
		},
	}

	bucketOutput := BucketOutputTags{
		aggregationDimensions: make(map[string]string),
		otherTags:             make(map[string]string),
	}

	assert.Equal(t, bucketInput.extractSourceTag(&bucketOutput), "v1")
}

func TestExtractExcludedAtomTag(t *testing.T) {
	bucketInput := BucketInput{
		config: BucketConfig{
			ExcludeTags:           []string{"k2"},
			SourceTagKey:          "k1",
			AtomReplacementTagKey: "k2",
		},
		allTags: map[string]string{
			"k1": "v1",
			"k2": "v2",
			"k3": "v3",
		},
	}

	bucketOutput := BucketOutputTags{
		aggregationDimensions: make(map[string]string),
		otherTags:             make(map[string]string),
	}
	bucketInput.extractAtomTag(&bucketOutput)

	assert.Equal(t, bucketOutput.OutputTags()["atom"], "v2")
	assert.Equal(t, bucketOutput.OutputTags()["k2"], "")
}

func TestExtractMissingSourceTag(t *testing.T) {
	bucketInput := BucketInput{
		config: BucketConfig{
			ExcludeTags:  nil,
			SourceTagKey: "k1",
		},
		allTags: map[string]string{
			//"k1":   "v1",
			"k2":   "v2",
			"k3":   "v3",
			"atom": "atomValue",
		},
	}

	bucketOutput := BucketOutputTags{
		aggregationDimensions: make(map[string]string),
		otherTags:             make(map[string]string),
	}

	assert.Equal(t, bucketInput.extractSourceTag(&bucketOutput), "MISSING_k1")
}

func TestExtractExcludedSourceTag(t *testing.T) {
	bucketInput := BucketInput{
		config: BucketConfig{
			ExcludeTags:  []string{"k1"},
			SourceTagKey: "k1",
		},
		allTags: map[string]string{
			"k1":   "v1",
			"k2":   "v2",
			"k3":   "v3",
			"atom": "atomValue",
		},
	}

	bucketOutput := BucketOutputTags{
		aggregationDimensions: make(map[string]string),
		otherTags:             make(map[string]string),
	}

	assert.Equal(t, bucketInput.extractSourceTag(&bucketOutput), "v1")
}

func TestExtractBucketKey(t *testing.T) {
	bucketInput := BucketInput{}

	bucketOutput := BucketOutputTags{
		aggregationDimensions: map[string]string{
			constants.TagKeyAtom: "fred",
			"k1":                 "v1",
		},
		otherTags: map[string]string{
			"NotAggregationTag": "extra",
		},
	}

	bucketInput.extractBucketKey("name", "source", &bucketOutput)

	assert.Equal(t, bucketOutput.OutputTags()[constants.TagKeyBucketKey], "name_source_fred_v1")
}

func TestExtractBucketKeyExcludeTag(t *testing.T) {
	bucketInput := BucketInput{
		config: BucketConfig{
			ExcludeTags: []string{"k1"},
		},
	}

	bucketOutput := BucketOutputTags{
		aggregationDimensions: map[string]string{
			constants.TagKeyAtom: "fred",
			"k1":                 "v1",
		},
		otherTags: map[string]string{
			"NotAggregationTag": "extra",
		},
	}

	bucketInput.extractBucketKey("name", "source", &bucketOutput)

	assert.Equal(t, bucketOutput.OutputTags()[constants.TagKeyBucketKey], "name_source_fred")
}

func TestExtractBucketKeyExcludeAtom(t *testing.T) {
	bucketInput := BucketInput{
		config: BucketConfig{
			ExcludeTags: []string{"atom"},
		},
	}

	bucketOutput := BucketOutputTags{
		aggregationDimensions: map[string]string{
			constants.TagKeyAtom: "fred",
			"k1":                 "v1",
		},
		otherTags: map[string]string{
			"NotAggregationTag": "extra",
		},
	}

	bucketInput.extractBucketKey("name", "source", &bucketOutput)

	assert.Equal(t, bucketOutput.OutputTags()[constants.TagKeyBucketKey], "name_source_v1")
}

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
	"fmt"
	"github.groupondev.com/metrics/telegraf-tdigest-plugin/plugins/aggregators/tdigestagg/constants"
	"github.groupondev.com/metrics/telegraf-tdigest-plugin/plugins/aggregators/tdigestagg/utility"
)

var alwaysSkipTags = []string{constants.TagKeyRollup, constants.TagKeyAggregates}

type BucketInput struct {
	config  BucketConfig
	allTags map[string]string
}

func (bi *BucketInput) extractAtomTag(bot *BucketOutputTags) {
	atomKey := bi.config.AtomReplacementTagKey
	atomValue, hasAtomTag := bi.allTags[constants.TagKeyAtom]
	if !hasAtomTag {
		if "" != bi.config.AtomReplacementTagKey {
			atomReplacementValue, hasAtomReplacement := bi.allTags[bi.config.AtomReplacementTagKey]
			if !hasAtomReplacement {
				fmt.Println(constants.TagKeyAtom + " or " + bi.config.AtomReplacementTagKey + " tag is required for aggregation.  " +
					constants.TagKeyAtom + " and " + bi.config.AtomReplacementTagKey + " tags will be set to " + constants.MissingValueForRequiredTagPrefix + bi.config.AtomReplacementTagKey)
				atomValue = constants.MissingValueForRequiredTagPrefix + bi.config.AtomReplacementTagKey
				bot.otherTags[constants.AtomSlaViolation] = constants.SlaViolationMissing
				bot.aggregationDimensions[atomKey] = atomValue
			} else {
				atomValue = atomReplacementValue
			}
		} else {
			fmt.Println("No atom_replacement_tag_key defined.  " + constants.TagKeyAtom + " tag is required for aggregation.  " +
				constants.TagKeyAtom + " tag will be set to " + constants.MissingValueForRequiredTagPrefix + constants.TagKeyAtom)
			atomValue = constants.MissingValueForRequiredTagPrefix + constants.TagKeyAtom
			bot.otherTags[constants.AtomSlaViolation] = constants.SlaViolationMissing
		}
	}
	bot.otherTags[constants.TagKeyAtom] = atomValue
}

func (bi *BucketInput) extractSourceTag(bot *BucketOutputTags) string {
	source, hasSource := bi.allTags[bi.config.SourceTagKey]
	if !hasSource {
		source = constants.MissingValueForRequiredTagPrefix + bi.config.SourceTagKey
		fmt.Println(bi.config.SourceTagKey + " tag is required for aggregation.  Source and " + bi.config.SourceTagKey +
			" tags will be set to " + source)
		bot.aggregationDimensions[bi.config.SourceTagKey] = source
		bot.otherTags[constants.SourceSlaViolation] = constants.SlaViolationMissing
	}
	bot.otherTags[constants.TagKeySource] = source

	return source
}

func (bi *BucketInput) extractBucketKey(name string, source string, bot *BucketOutputTags) {
	bucketKey := name + constants.BucketKeyDelimiter + source

	// Remove tags excluded by config or always exclude list
	for _, excluded := range bi.config.ExcludeTags {
		delete(bot.aggregationDimensions, excluded)
	}
	for _, excluded := range alwaysSkipTags {
		delete(bot.aggregationDimensions, excluded)
	}

	// Gather all the tag keys used to create the config key so they can be sorted
	var bucketKeyKeySet []string
	for key := range bot.aggregationDimensions {
		bucketKeyKeySet = append(bucketKeyKeySet, key)
	}
	for _, key := range utility.SortedSet(bucketKeyKeySet) {
		bucketKey += constants.BucketKeyDelimiter + bot.aggregationDimensions[key]
	}

	bot.otherTags[constants.TagKeyBucketKey] = bucketKey
}

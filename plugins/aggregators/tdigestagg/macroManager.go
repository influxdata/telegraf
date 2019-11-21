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
	"fmt"
	"github.com/influxdata/telegraf/plugins/aggregators/tdigestagg/bucketing"
	"github.com/influxdata/telegraf/plugins/aggregators/tdigestagg/constants"
	"strings"
)

func translateMacroToAggregates(macro string) string {
	switch macro {
	case constants.MacroTimer:
		return constants.AggregationsTimer
	case constants.MacroCounter:
		return constants.AggregationsCounter
	case constants.MacroGauge:
		return constants.AggregationsGauge
	case constants.MacroDefault:
		return constants.AggregationsGauge
	case constants.MacroLocal:
		return constants.AggregationsLocal
	default:
		return "unsupported_macro"
	}
}

func getAggregatesAndMacroDimensions(tags map[string]string) (string, string) {
	rollupMacro, hasRollup := tags[constants.TagKeyRollup]
	if !hasRollup {
		fmt.Println("No rollup tag present.  Rollup tag will be set to " + constants.MacroValueBadData)
		rollupMacro = constants.MacroValueBadData
	}

	// _rollup tag no longer needed as data tag
	delete(tags, constants.TagKeyRollup)

	primarySplitIndex := strings.Index(rollupMacro, constants.MacroDelimiterPrimary)
	if primarySplitIndex <= 0 {
		fmt.Println("Malformed rollup macro: \"" + rollupMacro + "\".  Rollup tag will be set to " + constants.MacroValueBadData)
		rollupMacro = constants.MacroValueBadData
		primarySplitIndex = strings.Index(rollupMacro, constants.MacroDelimiterPrimary)
	}

	var aggregates string
	if rollupMacro[:primarySplitIndex] != constants.MacroLocal {
		aggregates = translateMacroToAggregates(rollupMacro[:primarySplitIndex])
	} else {
		aggregates = noValuePresent
	}

	return aggregates, rollupMacro[primarySplitIndex+1:]
}

func reduceToRollupTags(allTags map[string]string, macroDimensionString string) map[string]string {
	// Get the full tag key set
	var keySet []string
	for key := range allTags {
		keySet = append(keySet, key)
	}

	addative := true
	if strings.HasPrefix(macroDimensionString, "*-") {
		addative = false
		macroDimensionString = strings.TrimPrefix(macroDimensionString, "*-")
	}

	var rollupTagKeys []string
	if macroDimensionString == constants.MacroWildcard {
		// for *, add all tag keys
		rollupTagKeys = append(rollupTagKeys, keySet...)
	} else {
		// Get macro dimensions
		unExpandedDimensions := strings.Split(macroDimensionString, constants.MacroDelimiterSecondary)

		// Expand wildcards when present
		for _, macroDimension := range unExpandedDimensions {
			rollupTagKeys = append(rollupTagKeys, expandWildcard(macroDimension, keySet)...)
		}
	}

	var tagsIncludedByRollup = make(map[string]string)
	if addative {
		for _, tagKey := range rollupTagKeys {
			if tagValue, hasTag := allTags[tagKey]; hasTag {
				tagsIncludedByRollup[tagKey] = tagValue
			}
		}
	} else {
		for key, value := range allTags {
			tagsIncludedByRollup[key] = value
		}
		for _, tagKey := range rollupTagKeys {
			delete(tagsIncludedByRollup, tagKey)
		}
	}

	return tagsIncludedByRollup
}

func generateBucketData(name string, tags map[string]string, agg *TDigestAgg, bucketDataChannel chan bucketing.BucketData) {

	aggregates, macroDimensionString := getAggregatesAndMacroDimensions(tags)
	tagsIncludedByRollup := reduceToRollupTags(tags, macroDimensionString)

	for _, bucket := range agg.Bucketing {
		bucketDataChannel <- bucketing.NewBucketData(name, bucket, tags, aggregates, tagsIncludedByRollup)
	}
}

func expandWildcard(macroDim string, keySet []string) []string {
	var result []string
	if strings.Contains(macroDim, constants.MacroWildcard) {
		prefix := macroDim[:strings.Index(macroDim, constants.MacroWildcard)]

		for _, tagKey := range keySet {
			if strings.HasPrefix(tagKey, prefix) {
				result = append(result, tagKey)
			}
		}
	} else {
		result = append(result, macroDim)
	}

	return result
}

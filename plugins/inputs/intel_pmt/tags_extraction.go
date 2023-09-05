//go:build linux && amd64

package intel_pmt

import "regexp"

var (
	// core in sample name - like C5_
	coreRegex = regexp.MustCompile("^C([0-9]+)_")
	// CHA in sample name - like CHA43_
	chaRegex = regexp.MustCompile("^CHA([0-9]+)_")
)

func (a *aggregatorInterface) extractTagsFromSample() {
	newAggSample := aggregatorSamples{}
	for _, sample := range a.AggregatorSamples.AggregatorSample {
		matches := coreRegex.FindStringSubmatch(sample.SampleName)
		if len(matches) == 2 {
			// matches[0] is the exact match in the code
			// matches[1] is the captured number (in parentheses)
			sample.core = matches[1]
			sample.SampleName = coreRegex.ReplaceAllString(sample.SampleName, "")
			newAggSample.AggregatorSample = append(newAggSample.AggregatorSample, sample)
			continue
		}
		matches = chaRegex.FindStringSubmatch(sample.SampleName)
		if len(matches) == 2 {
			sample.cha = matches[1]
			sample.SampleName = chaRegex.ReplaceAllString(sample.SampleName, "")
		}
		newAggSample.AggregatorSample = append(newAggSample.AggregatorSample, sample)
	}
	a.AggregatorSamples = newAggSample
}

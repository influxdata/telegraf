//go:build linux && amd64

package intel_pmt

import (
	"errors"
	"regexp"

	"golang.org/x/exp/slices"
)

var metricPatternRegex = regexp.MustCompile(`(?P<class>(C|CHA))\d+_(?P<var>[A-Z0-9_]+)$`)

// verifyNoEmpty checks if all pmt XMLs are not empty.
//
// Data for different GUIDs can be empty
// but data for at least one GUID cannot be empty.
//
// Returns:
//   - nil if at least one pair of XMLs for GUID is not empty.
//   - an error if all XMLs are empty.
func (p *IntelPMT) verifyNoEmpty() error {
	emptyAggInterface := true
	for guid := range p.pmtTelemetryFiles {
		if len(p.pmtAggregatorInterface[guid].AggregatorSamples.AggregatorSample) != 0 {
			emptyAggInterface = false
			break
		}
	}
	if emptyAggInterface {
		return errors.New("all aggregator interface XMLs are empty")
	}
	emptyAgg := true
	for guid := range p.pmtTelemetryFiles {
		if len(p.pmtAggregator[guid].SampleGroup) != 0 {
			emptyAgg = false
			break
		}
	}
	if emptyAgg {
		return errors.New("all aggregator XMLs are empty")
	}
	return nil
}

// filterAggregatorByDatatype filters Aggregator XML by provided datatypes.
//
// Every sample group in aggregator XML consists of several samples.
// Every sample in the group has a datatype assigned.
// This function filters the samples based on their datatype.
//
// Parameters:
//
//	datatypes: string slice of datatypes to include in filtered XML.
func (a *aggregator) filterAggregatorByDatatype(datatypes []string) {
	var newSampleGroup []sampleGroup
	for _, group := range a.SampleGroup {
		var tmpAgg []sample
		for _, aggSample := range group.Sample {
			if slices.Contains(datatypes, aggSample.DatatypeIDRef) {
				tmpAgg = append(tmpAgg, aggSample)
			}
		}
		if len(tmpAgg) > 0 {
			// groupSample can have samples with different datatypeIDRef inside
			// so new groupSample needs to be created
			// containing all needed information and filtered samples only.
			newGroup := sampleGroup{}
			newGroup.SampleID = group.SampleID
			newGroup.Sample = tmpAgg
			newSampleGroup = append(newSampleGroup, newGroup)
		}
	}
	a.SampleGroup = newSampleGroup
}

// filterAggregatorBySampleName filters Aggregator XML by provided sample names.
//
// Every sample has a name specified in the XML.
// This function filters the samples based on their names.
// The match can be exact or can be based on regex match.
//
// Parameters:
//
//	sampleNames: string slice of sample names to include in filtered XML.
func (a *aggregator) filterAggregatorBySampleName(sampleNames []string) {
	var newSampleGroup []sampleGroup
	for _, group := range a.SampleGroup {
		var tmpAgg []sample
		for _, aggSample := range group.Sample {
			if shouldAddSample(aggSample, sampleNames) {
				tmpAgg = append(tmpAgg, aggSample)
			}
		}

		if len(tmpAgg) > 0 {
			newGroup := sampleGroup{}
			newGroup.SampleID = group.SampleID
			newGroup.Sample = tmpAgg
			newSampleGroup = append(newSampleGroup, newGroup)
		}
	}
	a.SampleGroup = newSampleGroup
}

// shouldAddSample is a helper function for filterAggregatorBySampleName
// that checks if the sample should  be added to the sample group.
func shouldAddSample(s sample, sampleNames []string) bool {
	matches := metricPatternRegex.FindStringSubmatch(s.SampleName)
	for _, v := range sampleNames {
		if s.SampleName == v {
			return true
		}
		if len(matches) == 4 {
			if matches[3] == v {
				return true
			}
		}
	}
	return false
}

// filterAggInterfaceByDatatype filter aggregator interface XML by provided datatypes.
//
// Aggregator interface XML contains many aggregator samples inside, each with datatype assigned.
// This function filters aggregator samples based on their datatype.
//
// Parameters:
//
//	datatypes: string slice of datatypes to include in filtered XML.
//	dtMetricsFound: a map of found datatypes for all GUIDs.
func (a *aggregatorInterface) filterAggInterfaceByDatatype(datatypes []string, dtMetricsFound map[string]bool) {
	newAggSample := aggregatorSamples{}
	for _, s := range a.AggregatorSamples.AggregatorSample {
		if slices.Contains(datatypes, s.DatatypeIDRef) {
			dtMetricsFound[s.DatatypeIDRef] = true
			newAggSample.AggregatorSample = append(newAggSample.AggregatorSample, s)
		}
	}
	a.AggregatorSamples = newAggSample
}

// filterAggInterfaceBySampleName filters aggregator interface XML by sample names.
//
// This function filters aggregator samples based on the provided sampleNames.
// When the name for the sample is unique the match is exact.
// When the name is per resource (i.e. Cx_) the match is regex based.
//
// Parameters:
//
//	sampleNames: string slice of sample names to include in filtered XML.
//	sMetricsFound: a map of found metric names for all GUIDs.
func (a *aggregatorInterface) filterAggInterfaceBySampleName(sampleNames []string, sMetricsFound map[string]bool) {
	newAggSample := aggregatorSamples{}
	for _, s := range a.AggregatorSamples.AggregatorSample {
		if shouldAddAggregatorSample(s, sampleNames, sMetricsFound) {
			newAggSample.AggregatorSample = append(newAggSample.AggregatorSample, s)
		}
	}
	a.AggregatorSamples = newAggSample
}

// shouldAddAggregatorSample is a helper function for filterAggInterfaceBySampleName
// that checks if if the sample should be added to the aggregator samples.
func shouldAddAggregatorSample(s aggregatorSample, sampleNames []string, sMetricsFound map[string]bool) bool {
	matches := metricPatternRegex.FindStringSubmatch(s.SampleName)
	for _, userMetricInput := range sampleNames {
		if s.SampleName == userMetricInput {
			sMetricsFound[userMetricInput] = true
			return true
		}
		if len(matches) == 4 {
			if matches[3] == userMetricInput {
				sMetricsFound[userMetricInput] = true
				return true
			}
		}
	}
	return false
}

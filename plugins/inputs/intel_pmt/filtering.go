//go:build linux && amd64

package intel_pmt

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"
)

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
		return fmt.Errorf("all aggregator interface XMLs are empty")
	}
	emptyAgg := true
	for guid := range p.pmtTelemetryFiles {
		if len(p.pmtAggregator[guid].SampleGroup) != 0 {
			emptyAgg = false
			break
		}
	}
	if emptyAgg {
		return fmt.Errorf("all aggregator XMLs are empty")
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
// The match can be exact or can be based on regex match from metricPattern map.
//
// Parameters:
//
//	sampleNames: string slice of sample names to include in filtered XML.
func (a *aggregator) filterAggregatorBySampleName(sampleNames []string) {
	var newSampleGroup []sampleGroup
	for _, group := range a.SampleGroup {
		var tmpAgg []sample
	sampleLoop:
		for _, aggSample := range group.Sample {
			for _, userMetricInput := range sampleNames {
				if strings.Contains(aggSample.SampleName, userMetricInput) {
					tmpAgg = append(tmpAgg, aggSample)
					continue sampleLoop
				}
				for patternInput, re := range metricPatterns {
					if strings.EqualFold(userMetricInput, patternInput) {
						if re.MatchString(aggSample.SampleName) {
							tmpAgg = append(tmpAgg, aggSample)
							continue sampleLoop
						}
					}
				}
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
aggInterfaceLoop:
	for _, s := range a.AggregatorSamples.AggregatorSample {
		for _, userMetricInput := range sampleNames {
			if strings.Contains(s.SampleName, userMetricInput) {
				newAggSample.AggregatorSample = append(newAggSample.AggregatorSample, s)
				sMetricsFound[s.SampleName] = true
				continue aggInterfaceLoop
			}
			for patternInput, re := range metricPatterns {
				if strings.EqualFold(userMetricInput, patternInput) {
					if re.MatchString(s.SampleName) {
						newAggSample.AggregatorSample = append(newAggSample.AggregatorSample, s)
						sMetricsFound[s.SampleName] = true
						continue aggInterfaceLoop
					}
				}
			}
		}
	}
	a.AggregatorSamples = newAggSample
}

// Patterns for some metrics, as they have CHAx or Cx in their name
// with "x" being the core/cha number.
// Inline with metrics described in readme.
var metricPatterns = map[string]*regexp.Regexp{
	// datatype: ttemperature
	"Cx_TEMP": regexp.MustCompile(`C\d+_TEMP$`),

	// datatype: tcore_state
	"Cx_EN": regexp.MustCompile(`C\d+_EN$`),

	// datatype: thist_counter
	"Cx_FREQ_HIST_R0":  regexp.MustCompile(`C\d+_FREQ_HIST_R0$`),
	"Cx_FREQ_HIST_R1":  regexp.MustCompile(`C\d+_FREQ_HIST_R1$`),
	"Cx_FREQ_HIST_R2":  regexp.MustCompile(`C\d+_FREQ_HIST_R2$`),
	"Cx_FREQ_HIST_R3":  regexp.MustCompile(`C\d+_FREQ_HIST_R3$`),
	"Cx_FREQ_HIST_R4":  regexp.MustCompile(`C\d+_FREQ_HIST_R4$`),
	"Cx_FREQ_HIST_R5":  regexp.MustCompile(`C\d+_FREQ_HIST_R5$`),
	"Cx_FREQ_HIST_R6":  regexp.MustCompile(`C\d+_FREQ_HIST_R6$`),
	"Cx_FREQ_HIST_R7":  regexp.MustCompile(`C\d+_FREQ_HIST_R7$`),
	"Cx_FREQ_HIST_R8":  regexp.MustCompile(`C\d+_FREQ_HIST_R8$`),
	"Cx_FREQ_HIST_R9":  regexp.MustCompile(`C\d+_FREQ_HIST_R9$`),
	"Cx_FREQ_HIST_R10": regexp.MustCompile(`C\d+_FREQ_HIST_R10$`),
	"Cx_FREQ_HIST_R11": regexp.MustCompile(`C\d+_FREQ_HIST_R11$`),
	"Cx_VOLT_HIST_R0":  regexp.MustCompile(`C\d+_VOLT_HIST_R0$`),
	"Cx_VOLT_HIST_R1":  regexp.MustCompile(`C\d+_VOLT_HIST_R1$`),
	"Cx_VOLT_HIST_R2":  regexp.MustCompile(`C\d+_VOLT_HIST_R2$`),
	"Cx_VOLT_HIST_R3":  regexp.MustCompile(`C\d+_VOLT_HIST_R3$`),
	"Cx_VOLT_HIST_R4":  regexp.MustCompile(`C\d+_VOLT_HIST_R4$`),
	"Cx_VOLT_HIST_R5":  regexp.MustCompile(`C\d+_VOLT_HIST_R5$`),
	"Cx_VOLT_HIST_R6":  regexp.MustCompile(`C\d+_VOLT_HIST_R6$`),
	"Cx_VOLT_HIST_R7":  regexp.MustCompile(`C\d+_VOLT_HIST_R7$`),
	"Cx_VOLT_HIST_R8":  regexp.MustCompile(`C\d+_VOLT_HIST_R8$`),
	"Cx_VOLT_HIST_R9":  regexp.MustCompile(`C\d+_VOLT_HIST_R9$`),
	"Cx_VOLT_HIST_R10": regexp.MustCompile(`C\d+_VOLT_HIST_R10$`),
	"Cx_VOLT_HIST_R11": regexp.MustCompile(`C\d+_VOLT_HIST_R11$`),
	"Cx_TEMP_HIST_R0":  regexp.MustCompile(`C\d+_TEMP_HIST_R0$`),
	"Cx_TEMP_HIST_R1":  regexp.MustCompile(`C\d+_TEMP_HIST_R1$`),
	"Cx_TEMP_HIST_R2":  regexp.MustCompile(`C\d+_TEMP_HIST_R2$`),
	"Cx_TEMP_HIST_R3":  regexp.MustCompile(`C\d+_TEMP_HIST_R3$`),
	"Cx_TEMP_HIST_R4":  regexp.MustCompile(`C\d+_TEMP_HIST_R4$`),
	"Cx_TEMP_HIST_R5":  regexp.MustCompile(`C\d+_TEMP_HIST_R5$`),
	"Cx_TEMP_HIST_R6":  regexp.MustCompile(`C\d+_TEMP_HIST_R6$`),
	"Cx_TEMP_HIST_R7":  regexp.MustCompile(`C\d+_TEMP_HIST_R7$`),
	"Cx_TEMP_HIST_R8":  regexp.MustCompile(`C\d+_TEMP_HIST_R8$`),
	"Cx_TEMP_HIST_R9":  regexp.MustCompile(`C\d+_TEMP_HIST_R9$`),
	"Cx_TEMP_HIST_R10": regexp.MustCompile(`C\d+_TEMP_HIST_R10$`),
	"Cx_TEMP_HIST_R11": regexp.MustCompile(`C\d+_TEMP_HIST_R11$`),

	// datatype: tpvp_throttle_counter
	"Cx_PVP_THROTTLE_64":   regexp.MustCompile(`C\d+_Cx_PVP_THROTTLE_64$`),
	"Cx_PVP_THROTTLE_1024": regexp.MustCompile(`C\d+_PVP_THROTTLE_1024$`),

	// datatype: tpvp_level_res
	"Cx_PVP_LEVEL_RES_128_L0":  regexp.MustCompile(`C\d+_PVP_LEVEL_RES_128_L0$`),
	"Cx_PVP_LEVEL_RES_128_L1":  regexp.MustCompile(`C\d+_PVP_LEVEL_RES_128_L1$`),
	"Cx_PVP_LEVEL_RES_128_L2":  regexp.MustCompile(`C\d+_PVP_LEVEL_RES_128_L2$`),
	"Cx_PVP_LEVEL_RES_128_L3":  regexp.MustCompile(`C\d+_PVP_LEVEL_RES_128_L3$`),
	"Cx_PVP_LEVEL_RES_256_L0":  regexp.MustCompile(`C\d+_PVP_LEVEL_RES_256_L0$`),
	"Cx_PVP_LEVEL_RES_256_L1":  regexp.MustCompile(`C\d+_PVP_LEVEL_RES_256_L1$`),
	"Cx_PVP_LEVEL_RES_256_L2":  regexp.MustCompile(`C\d+_PVP_LEVEL_RES_256_L2$`),
	"Cx_PVP_LEVEL_RES_256_L3":  regexp.MustCompile(`C\d+_PVP_LEVEL_RES_256_L3$`),
	"Cx_PVP_LEVEL_RES_512_L0":  regexp.MustCompile(`C\d+_PVP_LEVEL_RES_512_L0$`),
	"Cx_PVP_LEVEL_RES_512_L1":  regexp.MustCompile(`C\d+_PVP_LEVEL_RES_512_L1$`),
	"Cx_PVP_LEVEL_RES_512_L2":  regexp.MustCompile(`C\d+_PVP_LEVEL_RES_512_L2$`),
	"Cx_PVP_LEVEL_RES_512_L3":  regexp.MustCompile(`C\d+_PVP_LEVEL_RES_512_L3$`),
	"Cx_PVP_LEVEL_RES_TMUL_L0": regexp.MustCompile(`C\d+_PVP_LEVEL_RES_TMUL_L0$`),
	"Cx_PVP_LEVEL_RES_TMUL_L1": regexp.MustCompile(`C\d+_PVP_LEVEL_RES_TMUL_L1$`),
	"Cx_PVP_LEVEL_RES_TMUL_L2": regexp.MustCompile(`C\d+_PVP_LEVEL_RES_TMUL_L2$`),
	"Cx_PVP_LEVEL_RES_TMUL_L3": regexp.MustCompile(`C\d+_PVP_LEVEL_RES_TMUL_L3$`),

	// datatype: trmid_usage_counter
	"CHAx_RMID0_RDT_CMT":       regexp.MustCompile(`CHA\d+_RMID0_RDT_CMT$`),
	"CHAx_RMID1_RDT_CMT":       regexp.MustCompile(`CHA\d+_RMID1_RDT_CMT$`),
	"CHAx_RMID2_RDT_CMT":       regexp.MustCompile(`CHA\d+_RMID2_RDT_CMT$`),
	"CHAx_RMID3_RDT_CMT":       regexp.MustCompile(`CHA\d+_RMID3_RDT_CMT$`),
	"CHAx_RMID4_RDT_CMT":       regexp.MustCompile(`CHA\d+_RMID4_RDT_CMT$`),
	"CHAx_RMID5_RDT_CMT":       regexp.MustCompile(`CHA\d+_RMID5_RDT_CMT$`),
	"CHAx_RMID6_RDT_CMT":       regexp.MustCompile(`CHA\d+_RMID6_RDT_CMT$`),
	"CHAx_RMID7_RDT_CMT":       regexp.MustCompile(`CHA\d+_RMID7_RDT_CMT$`),
	"CHAx_RMID0_RDT_MBM_TOTAL": regexp.MustCompile(`CHA\d+_RMID0_RDT_MBM_TOTAL$`),
	"CHAx_RMID1_RDT_MBM_TOTAL": regexp.MustCompile(`CHA\d+_RMID1_RDT_MBM_TOTAL$`),
	"CHAx_RMID2_RDT_MBM_TOTAL": regexp.MustCompile(`CHA\d+_RMID2_RDT_MBM_TOTAL$`),
	"CHAx_RMID3_RDT_MBM_TOTAL": regexp.MustCompile(`CHA\d+_RMID3_RDT_MBM_TOTAL$`),
	"CHAx_RMID4_RDT_MBM_TOTAL": regexp.MustCompile(`CHA\d+_RMID4_RDT_MBM_TOTAL$`),
	"CHAx_RMID5_RDT_MBM_TOTAL": regexp.MustCompile(`CHA\d+_RMID5_RDT_MBM_TOTAL$`),
	"CHAx_RMID6_RDT_MBM_TOTAL": regexp.MustCompile(`CHA\d+_RMID6_RDT_MBM_TOTAL$`),
	"CHAx_RMID7_RDT_MBM_TOTAL": regexp.MustCompile(`CHA\d+_RMID7_RDT_MBM_TOTAL$`),
	"CHAx_RMID0_RDT_MBM_LOCAL": regexp.MustCompile(`CHA\d+_RMID0_RDT_MBM_LOCAL$`),
	"CHAx_RMID1_RDT_MBM_LOCAL": regexp.MustCompile(`CHA\d+_RMID1_RDT_MBM_LOCAL$`),
	"CHAx_RMID2_RDT_MBM_LOCAL": regexp.MustCompile(`CHA\d+_RMID2_RDT_MBM_LOCAL$`),
	"CHAx_RMID3_RDT_MBM_LOCAL": regexp.MustCompile(`CHA\d+_RMID3_RDT_MBM_LOCAL$`),
	"CHAx_RMID4_RDT_MBM_LOCAL": regexp.MustCompile(`CHA\d+_RMID4_RDT_MBM_LOCAL$`),
	"CHAx_RMID5_RDT_MBM_LOCAL": regexp.MustCompile(`CHA\d+_RMID5_RDT_MBM_LOCAL$`),
	"CHAx_RMID6_RDT_MBM_LOCAL": regexp.MustCompile(`CHA\d+_RMID6_RDT_MBM_LOCAL$`),
	"CHAx_RMID7_RDT_MBM_LOCAL": regexp.MustCompile(`CHA\d+_RMID7_RDT_MBM_LOCAL$`),

	// datatype: tcore_stress_level
	"Cx_STRESS_LEVEL": regexp.MustCompile(`C\d+_STRESS_LEVEL`),
}

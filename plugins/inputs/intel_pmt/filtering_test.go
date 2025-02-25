//go:build linux && amd64

package intel_pmt

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestFilterAggregatorByDatatype(t *testing.T) {
	t.Run("Filter aggregator, 1 sample group, 2 samples with different DataTypes", func(t *testing.T) {
		agg := aggregator{
			SampleGroup: []sampleGroup{
				{
					SampleID: uint64(0),
					Sample: []sample{
						{
							DatatypeIDRef: "test-datatype",
							Msb:           4,
							Lsb:           4,
							SampleID:      "test-sample-ref",
						},
						{
							DatatypeIDRef: "missing",
							Msb:           0,
							Lsb:           0,
							SampleID:      "missing",
						},
					},
				},
			},
		}
		p := IntelPMT{
			DatatypeFilter: []string{"test-datatype"},
		}
		expected := aggregator{
			SampleGroup: []sampleGroup{
				{
					SampleID: uint64(0),
					Sample: []sample{
						{
							DatatypeIDRef: "test-datatype",
							Msb:           4,
							Lsb:           4,
							SampleID:      "test-sample-ref",
						},
					},
				},
			},
		}
		agg.filterAggregatorByDatatype(p.DatatypeFilter)
		require.Equal(t, expected, agg)
	})

	t.Run("Filter Aggregator, 2 sample groups, only 1 sample group has expected datatype", func(t *testing.T) {
		agg := aggregator{
			SampleGroup: []sampleGroup{
				{
					SampleID: uint64(0),
					Sample: []sample{
						{
							DatatypeIDRef: "test-datatype",
							Msb:           4,
							Lsb:           4,
							SampleID:      "test-sample-ref",
						},
					},
				},
				{
					SampleID: uint64(2),
					Sample: []sample{
						{
							DatatypeIDRef: "missing",
							Msb:           0,
							Lsb:           0,
							SampleID:      "missing",
						},
					},
				},
			},
		}
		p := IntelPMT{
			DatatypeFilter: []string{"test-datatype"},
		}
		expected := aggregator{
			SampleGroup: []sampleGroup{
				{
					SampleID: uint64(0),
					Sample: []sample{
						{
							DatatypeIDRef: "test-datatype",
							Msb:           4,
							Lsb:           4,
							SampleID:      "test-sample-ref",
						},
					},
				},
			},
		}
		agg.filterAggregatorByDatatype(p.DatatypeFilter)
		require.Equal(t, expected, agg)
	})
}

func TestFilterAggregatorInterfaceByDatatype(t *testing.T) {
	t.Run("Filter agg interface, 2 Agg samples, only 1 should remain", func(t *testing.T) {
		aggInterface := aggregatorInterface{
			AggregatorSamples: aggregatorSamples{
				AggregatorSample: []aggregatorSample{
					{
						SampleName:    "test-sample",
						SampleGroup:   "test-group",
						DatatypeIDRef: "test-datatype",
						TransformREF:  "test-transform-ref",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "testvar",
									SampleIDREF: "test-sample-ref",
								},
							},
						},
					},
					{
						SampleName:    "missing",
						SampleGroup:   "missing",
						DatatypeIDRef: "missing",
						TransformREF:  "missing",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "missing",
									SampleIDREF: "missing",
								},
							},
						},
					},
				},
			},
		}
		expected := aggregatorInterface{
			AggregatorSamples: aggregatorSamples{
				AggregatorSample: []aggregatorSample{
					{
						SampleName:    "test-sample",
						SampleGroup:   "test-group",
						DatatypeIDRef: "test-datatype",
						TransformREF:  "test-transform-ref",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "testvar",
									SampleIDREF: "test-sample-ref",
								},
							},
						},
					},
				},
			},
		}

		p := IntelPMT{
			DatatypeFilter: []string{"test-datatype"},
			Log:            testutil.Logger{},
		}
		aggInterface.filterAggInterfaceByDatatype(p.DatatypeFilter, make(map[string]bool))
		require.Equal(t, expected, aggInterface)
	})
}

func TestFilterAggregatorBySampleName(t *testing.T) {
	t.Run("Filter aggregator, 2 sample names, with the same datatype, 1 sample name matches exactly", func(t *testing.T) {
		agg := aggregator{
			SampleGroup: []sampleGroup{
				{
					SampleID: uint64(0),
					Sample: []sample{
						{
							SampleName:    "exists",
							DatatypeIDRef: "test-datatype",
							Msb:           4,
							Lsb:           4,
							SampleID:      "test-sample-ref",
						},
						{
							SampleName:    "missing",
							DatatypeIDRef: "test-datatype",
							Msb:           0,
							Lsb:           0,
							SampleID:      "missing",
						},
					},
				},
			},
		}
		p := IntelPMT{
			SampleFilter: []string{"exists"},
		}
		expected := aggregator{
			SampleGroup: []sampleGroup{
				{
					SampleID: uint64(0),
					Sample: []sample{
						{
							SampleName:    "exists",
							DatatypeIDRef: "test-datatype",
							Msb:           4,
							Lsb:           4,
							SampleID:      "test-sample-ref",
						},
					},
				},
			},
		}
		agg.filterAggregatorBySampleName(p.SampleFilter)
		require.Equal(t, expected, agg)
	})

	t.Run("Filter aggregator, 2 sample names, with the same datatype, 1 sample name matches by regex", func(t *testing.T) {
		agg := aggregator{
			SampleGroup: []sampleGroup{
				{
					SampleID: uint64(0),
					Sample: []sample{
						{
							SampleName:    "C61_TEMP",
							DatatypeIDRef: "test-datatype",
							Msb:           4,
							Lsb:           4,
							SampleID:      "test-sample-ref",
						},
						{
							SampleName:    "C61_TEMP_test",
							DatatypeIDRef: "test-datatype",
							Msb:           0,
							Lsb:           0,
							SampleID:      "missing",
						},
					},
				},
			},
		}
		p := IntelPMT{
			SampleFilter: []string{"TEMP"},
		}
		expected := aggregator{
			SampleGroup: []sampleGroup{
				{
					SampleID: uint64(0),
					Sample: []sample{
						{
							SampleName:    "C61_TEMP",
							DatatypeIDRef: "test-datatype",
							Msb:           4,
							Lsb:           4,
							SampleID:      "test-sample-ref",
						},
					},
				},
			},
		}
		agg.filterAggregatorBySampleName(p.SampleFilter)
		require.Equal(t, expected, agg)
	})
}

func TestFilterAggregatorInterfaceBySampleName(t *testing.T) {
	t.Run("Filter agg interface, 2 Agg samples, 1 sample name matches exactly", func(t *testing.T) {
		aggInterface := aggregatorInterface{
			AggregatorSamples: aggregatorSamples{
				AggregatorSample: []aggregatorSample{
					{
						SampleName:    "C36_PVP_LEVEL_RES_128_L1",
						SampleGroup:   "test-group",
						DatatypeIDRef: "test-datatype",
						TransformREF:  "test-transform-ref",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "testvar",
									SampleIDREF: "test-sample-ref",
								},
							},
						},
					},
					{
						SampleName:    "missing",
						SampleGroup:   "missing",
						DatatypeIDRef: "test-datatype",
						TransformREF:  "missing",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "missing",
									SampleIDREF: "missing",
								},
							},
						},
					},
				},
			},
		}
		expected := aggregatorInterface{
			AggregatorSamples: aggregatorSamples{
				AggregatorSample: []aggregatorSample{
					{
						SampleName:    "C36_PVP_LEVEL_RES_128_L1",
						SampleGroup:   "test-group",
						DatatypeIDRef: "test-datatype",
						TransformREF:  "test-transform-ref",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "testvar",
									SampleIDREF: "test-sample-ref",
								},
							},
						},
					},
				},
			},
		}

		p := IntelPMT{
			SampleFilter: []string{"PVP_LEVEL_RES_128_L1"},
			Log:          testutil.Logger{},
		}
		aggInterface.filterAggInterfaceBySampleName(p.SampleFilter, make(map[string]bool))
		require.Equal(t, expected, aggInterface)
	})

	t.Run("Filter agg interface, 2 Agg samples, 1 sample name matches by regex", func(t *testing.T) {
		aggInterface := aggregatorInterface{
			AggregatorSamples: aggregatorSamples{
				AggregatorSample: []aggregatorSample{
					{
						SampleName:    "test-sample",
						SampleGroup:   "test-group",
						DatatypeIDRef: "test-datatype",
						TransformREF:  "test-transform-ref",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "testvar",
									SampleIDREF: "test-sample-ref",
								},
							},
						},
					},
					{
						SampleName:    "missing",
						SampleGroup:   "missing",
						DatatypeIDRef: "test-datatype",
						TransformREF:  "missing",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "missing",
									SampleIDREF: "missing",
								},
							},
						},
					},
				},
			},
		}
		expected := aggregatorInterface{
			AggregatorSamples: aggregatorSamples{
				AggregatorSample: []aggregatorSample{
					{
						SampleName:    "test-sample",
						SampleGroup:   "test-group",
						DatatypeIDRef: "test-datatype",
						TransformREF:  "test-transform-ref",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "testvar",
									SampleIDREF: "test-sample-ref",
								},
							},
						},
					},
				},
			},
		}

		p := IntelPMT{
			SampleFilter: []string{"test-sample"},
			Log:          testutil.Logger{},
		}
		aggInterface.filterAggInterfaceBySampleName(p.SampleFilter, make(map[string]bool))
		require.Equal(t, expected, aggInterface)
	})
}

func TestVerifyNoEmpty(t *testing.T) {
	t.Run("Correct XMLs, no filtering by user", func(t *testing.T) {
		p := &IntelPMT{
			pmtAggregator: map[string]aggregator{
				"test-guid": {
					SampleGroup: []sampleGroup{{}},
				},
			},
			pmtAggregatorInterface: map[string]aggregatorInterface{
				"test-guid": {
					AggregatorSamples: aggregatorSamples{
						AggregatorSample: []aggregatorSample{{}},
					},
				},
			},
		}
		p.pmtTelemetryFiles = map[string]pmtFileInfo{
			"test-guid": []fileInfo{{}},
		}
		require.NoError(t, p.verifyNoEmpty())
	})

	t.Run("Incorrect XMLs, filtering by datatype that doesn't exist", func(t *testing.T) {
		p := &IntelPMT{
			pmtAggregator: map[string]aggregator{
				"test-guid": {},
			},
			pmtAggregatorInterface: map[string]aggregatorInterface{
				"test-guid": {},
			},
			DatatypeFilter:    []string{"doesn't-exist"},
			Log:               testutil.Logger{},
			pmtTelemetryFiles: map[string]pmtFileInfo{"test-guid": []fileInfo{{}}},
		}
		aggInterface := aggregatorInterface{
			AggregatorSamples: aggregatorSamples{
				AggregatorSample: []aggregatorSample{
					{
						SampleName:    "test-sample",
						SampleGroup:   "test-group",
						DatatypeIDRef: "test-datatype",
						TransformREF:  "test-transform-ref",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "testvar",
									SampleIDREF: "test-sample-ref",
								},
							},
						},
					},
				},
			},
		}
		aggInterface.filterAggInterfaceByDatatype(p.DatatypeFilter, make(map[string]bool))
		p.pmtAggregatorInterface["test-guid"] = aggInterface
		require.ErrorContains(t, p.verifyNoEmpty(), "all aggregator interface XMLs are empty")
	})

	t.Run("Incorrect XMLs, user provided sample names that don't exist", func(t *testing.T) {
		p := &IntelPMT{
			pmtAggregatorInterface: map[string]aggregatorInterface{
				"test-guid": {},
			},
			SampleFilter:      []string{"doesn't-exist"},
			Log:               testutil.Logger{},
			pmtTelemetryFiles: map[string]pmtFileInfo{"test-guid": []fileInfo{{}}},
		}

		aggInterface := aggregatorInterface{
			AggregatorSamples: aggregatorSamples{
				AggregatorSample: []aggregatorSample{
					{
						SampleName:    "test-sample",
						SampleGroup:   "test-group",
						DatatypeIDRef: "test-datatype",
						TransformREF:  "test-transform-ref",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "testvar",
									SampleIDREF: "test-sample-ref",
								},
							},
						},
					},
				},
			},
		}
		aggInterface.filterAggInterfaceBySampleName(p.SampleFilter, make(map[string]bool))
		p.pmtAggregatorInterface["test-guid"] = aggInterface
		require.ErrorContains(t, p.verifyNoEmpty(), "XMLs are empty")
	})
	t.Run("Correct XMLs, user provided correct sample names", func(t *testing.T) {
		p := &IntelPMT{
			pmtAggregator: map[string]aggregator{
				"test-guid": {},
			},
			pmtAggregatorInterface: map[string]aggregatorInterface{
				"test-guid": {},
			},
			SampleFilter:      []string{"test-sample"},
			pmtTelemetryFiles: map[string]pmtFileInfo{"test-guid": []fileInfo{{}}},
		}
		agg := aggregator{
			SampleGroup: []sampleGroup{
				{
					SampleID: uint64(0),
					Sample: []sample{
						{
							SampleName:    "test-sample",
							DatatypeIDRef: "test-datatype",
							Msb:           4,
							Lsb:           4,
							SampleID:      "test-sample-ref",
						},
					},
				},
			},
		}
		aggInterface := aggregatorInterface{
			AggregatorSamples: aggregatorSamples{
				AggregatorSample: []aggregatorSample{
					{
						SampleName:    "test-sample",
						SampleGroup:   "test-group",
						DatatypeIDRef: "test-datatype",
						TransformREF:  "test-transform-ref",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "testvar",
									SampleIDREF: "test-sample-ref",
								},
							},
						},
					},
				},
			},
		}
		agg.filterAggregatorBySampleName(p.SampleFilter)
		aggInterface.filterAggInterfaceBySampleName(p.SampleFilter, make(map[string]bool))
		p.pmtAggregator["test-guid"] = agg
		p.pmtAggregatorInterface["test-guid"] = aggInterface
		require.NoError(t, p.verifyNoEmpty())
	})

	t.Run("Correct XMLs, user provided correct datatype names", func(t *testing.T) {
		p := &IntelPMT{
			pmtAggregator: map[string]aggregator{
				"test-guid": {},
			},
			pmtAggregatorInterface: map[string]aggregatorInterface{
				"test-guid": {},
			},
			DatatypeFilter:    []string{"test-datatype"},
			pmtTelemetryFiles: map[string]pmtFileInfo{"test-guid": []fileInfo{{}}},
		}
		agg := aggregator{
			SampleGroup: []sampleGroup{
				{
					SampleID: uint64(0),
					Sample: []sample{
						{
							SampleName:    "test-sample",
							DatatypeIDRef: "test-datatype",
							Msb:           4,
							Lsb:           4,
							SampleID:      "test-sample-ref",
						},
					},
				},
			},
		}
		aggInterface := aggregatorInterface{
			AggregatorSamples: aggregatorSamples{
				AggregatorSample: []aggregatorSample{
					{
						SampleName:    "test-sample",
						SampleGroup:   "test-group",
						DatatypeIDRef: "test-datatype",
						TransformREF:  "test-transform-ref",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "testvar",
									SampleIDREF: "test-sample-ref",
								},
							},
						},
					},
				},
			},
		}
		agg.filterAggregatorByDatatype(p.DatatypeFilter)
		aggInterface.filterAggInterfaceByDatatype(p.DatatypeFilter, make(map[string]bool))
		p.pmtAggregator["test-guid"] = agg
		p.pmtAggregatorInterface["test-guid"] = aggInterface
		require.NoError(t, p.verifyNoEmpty())
	})

	t.Run("Incorrect XMLs, no datatype metrics found in aggregator sample XML", func(t *testing.T) {
		p := &IntelPMT{
			pmtAggregator: map[string]aggregator{
				"test-guid": {},
			},
			pmtAggregatorInterface: map[string]aggregatorInterface{
				"test-guid": {},
			},
			DatatypeFilter:    []string{"test-datatype"},
			pmtTelemetryFiles: map[string]pmtFileInfo{"test-guid": []fileInfo{{}}},
		}
		agg := aggregator{
			SampleGroup: []sampleGroup{
				{
					SampleID: uint64(0),
					Sample: []sample{
						{
							SampleName: "test-sample",
							// DatatypeIDREF is wrong
							DatatypeIDRef: "wrong",
							Msb:           4,
							Lsb:           4,
							SampleID:      "test-sample-ref",
						},
					},
				},
			},
		}
		aggInterface := aggregatorInterface{
			AggregatorSamples: aggregatorSamples{
				AggregatorSample: []aggregatorSample{
					{
						SampleName:    "test-sample",
						SampleGroup:   "test-group",
						DatatypeIDRef: "test-datatype",
						TransformREF:  "test-transform-ref",
						TransformInputs: transformInputs{
							TransformInput: []transformInput{
								{
									VarName:     "testvar",
									SampleIDREF: "test-sample-ref",
								},
							},
						},
					},
				},
			},
		}
		agg.filterAggregatorByDatatype(p.DatatypeFilter)
		aggInterface.filterAggInterfaceByDatatype(p.DatatypeFilter, make(map[string]bool))
		p.pmtAggregator["test-guid"] = agg
		p.pmtAggregatorInterface["test-guid"] = aggInterface
		require.ErrorContains(t, p.verifyNoEmpty(), "all aggregator XMLs are empty")
	})
}

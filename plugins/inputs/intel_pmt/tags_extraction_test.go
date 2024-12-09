//go:build linux && amd64

package intel_pmt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractTagsFromSample(t *testing.T) {
	tests := []struct {
		name     string
		input    aggregatorInterface
		expected aggregatorInterface
	}{
		{
			name: "Extract core number from sampleName",
			input: aggregatorInterface{
				AggregatorSamples: aggregatorSamples{
					AggregatorSample: []aggregatorSample{
						{
							SampleName:    "C34_test",
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
			},
			expected: aggregatorInterface{
				AggregatorSamples: aggregatorSamples{
					AggregatorSample: []aggregatorSample{
						{
							SampleName:    "test",
							SampleGroup:   "test-group",
							DatatypeIDRef: "test-datatype",
							TransformREF:  "test-transform-ref",
							core:          "34",
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
			},
		},
		{
			name: "Extract cha number from sample",
			input: aggregatorInterface{
				AggregatorSamples: aggregatorSamples{
					AggregatorSample: []aggregatorSample{
						{
							SampleName:    "CHA34_test",
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
			},
			expected: aggregatorInterface{
				AggregatorSamples: aggregatorSamples{
					AggregatorSample: []aggregatorSample{
						{
							SampleName:    "test",
							SampleGroup:   "test-group",
							DatatypeIDRef: "test-datatype",
							TransformREF:  "test-transform-ref",
							cha:           "34",
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
			},
		},
		{
			name: "SampleName doesn't contain any matched patterns, no change in sample",
			input: aggregatorInterface{
				AggregatorSamples: aggregatorSamples{
					AggregatorSample: []aggregatorSample{
						{
							SampleName:    "test",
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
			},
			expected: aggregatorInterface{
				AggregatorSamples: aggregatorSamples{
					AggregatorSample: []aggregatorSample{
						{
							SampleName:    "test",
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.extractTagsFromSample()
			require.Equal(t, tt.expected, tt.input)
		})
	}
}

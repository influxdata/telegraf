//go:build linux && amd64

package intel_pmt

import (
	_ "embed"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func createTempFile(t *testing.T, dir, pattern string, data []byte) (*os.File, os.FileInfo) {
	tempFile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		t.Fatalf("error creating a temporary file %v: %v", tempFile.Name(), err)
	}
	_, err = tempFile.Write(data)
	if err != nil {
		t.Fatalf("error writing buffer to file %v: %v", tempFile.Name(), err)
	}
	fileInfo, err := tempFile.Stat()
	if err != nil {
		t.Fatalf("failed to stat a temporary file %v: %v", tempFile.Name(), err)
	}

	return tempFile, fileInfo
}

func TestTransformEquation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No changes",
			input:    "abc",
			expected: "abc",
		},
		{
			name:     "Remove $ sign",
			input:    "a$b$c",
			expected: "abc",
		},
		{
			name:     "Decode HTML entities",
			input:    "a&amp;b",
			expected: "a&b",
		},
		{
			name:     "Remove $ and decode HTML entities",
			input:    "$a&amp;b$c",
			expected: "a&bc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := transformEquation(tt.input)
			require.Equal(t, tt.expected, output)
		})
	}
}

func TestEval(t *testing.T) {
	tests := []struct {
		name     string
		eq       string
		params   map[string]interface{}
		expected interface{}
		err      bool
	}{
		{
			name:     "empty equation",
			eq:       "",
			params:   nil,
			expected: nil,
			err:      true,
		},
		{
			name:     "Valid equation",
			eq:       "2 + 2",
			params:   nil,
			expected: float64(4),
			err:      false,
		},
		{
			name: "Valid equation with params, valid params",
			eq:   "a + b",
			params: map[string]interface{}{
				"a": 2,
				"b": 3,
			},
			expected: float64(5),
			err:      false,
		},
		{
			name: "Valid equation with params, invalid params",
			eq:   "a + b",
			params: map[string]interface{}{
				"a": 2,
				// "b" is missing
			},
			expected: nil,
			err:      true,
		},
		{
			name:     "Invalid equation",
			eq:       "2 +",
			params:   nil,
			expected: nil,
			err:      true,
		},
		{
			name: "Real equation from PMT - temperature of unused core",
			eq:   "( ( parameter_0 >> 8 ) & 0xff ) + ( ( parameter_0 & 0xff ) / ( 2 ** 8 ) ) - 64",
			params: map[string]interface{}{
				"parameter_0": 0,
			},
			expected: float64(-64),
			err:      false,
		},
		{
			name: "Real equation from PMT - temperature of working core",
			eq:   "( ( parameter_0 >> 8 ) & 0xff ) + ( ( parameter_0 & 0xff ) / ( 2 ** 8 ) ) - 64",
			params: map[string]interface{}{
				"parameter_0": 23600,
			},
			expected: float64(28.1875),
			err:      false,
		},
		{
			name: "Badly parsed real equation from PMT - temperature of working core",
			eq:   "( ( parameter_0 &gt;&gt; 8 ) & 0xff ) + ( ( parameter_0 & 0xff ) / ( 2 ** 8 ) ) - 64",
			params: map[string]interface{}{
				"parameter_0": 23600,
			},
			expected: nil,
			err:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval(tt.eq, tt.params)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetTelemSample(t *testing.T) {
	tests := []struct {
		name     string
		s        sample
		buf      []byte
		offset   uint64
		expected uint64
		err      bool
	}{
		{
			name:     "All bits set",
			s:        sample{Msb: 7, Lsb: 0, mask: 255},
			buf:      []byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			offset:   0,
			expected: 255,
		},
		{
			name:     "Middle bits set",
			s:        sample{Msb: 5, Lsb: 2, mask: 60},
			buf:      []byte{0x3c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // 0x3c = 00111100 in binary
			offset:   0,
			expected: 15,
		},
		{
			name:     "Non-zero offset",
			s:        sample{Msb: 7, Lsb: 0, mask: 255},
			buf:      []byte{0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			offset:   3,
			expected: 255,
		},
		{
			name:     "Single bit set",
			s:        sample{Msb: 4, Lsb: 4, mask: 16},
			buf:      []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // 0x10 = 00010000 in binary
			offset:   0,
			expected: 1,
		},
		{
			name:     "Two bytes set",
			s:        sample{Msb: 14, Lsb: 0, mask: 32767},
			buf:      []byte{0x30, 0x5c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // 0x5c30 = 23600 in decimal
			offset:   0,
			expected: 23600,
		},
		{
			name:     "Offset larger than buffer size",
			s:        sample{Msb: 7, Lsb: 0, mask: 255},
			buf:      []byte{0x00},
			offset:   5,
			expected: 0,
			err:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getTelemSample(tt.s, tt.buf, tt.offset)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestInit(t *testing.T) {
	t.Run("No PmtSpec", func(t *testing.T) {
		p := &IntelPMT{
			PmtSpec: "",
		}
		err := p.Init()
		require.ErrorContains(t, err, "pmt spec is empty")
	})

	t.Run("Incorrect filepath PmtSpec", func(t *testing.T) {
		p := &IntelPMT{
			PmtSpec: "/this/path/doesntexist",
		}
		err := p.Init()
		require.ErrorContains(t, err, "provided pmt spec is not readable")
	})

	t.Run("Incorrect PmtSpec, random letters", func(t *testing.T) {
		p := &IntelPMT{
			PmtSpec: "loremipsum",
		}
		err := p.Init()
		require.ErrorContains(t, err, "provided pmt spec is not readable")
	})

	t.Run("Correct filepath PmtSpec, no pmt/can't read pmt in sysfs", func(t *testing.T) {
		tmp := t.TempDir()
		testFile, _ := createTempFile(t, tmp, "test-file", []byte("<pmt><mappings><mapping></mapping></mappings></pmt>"))
		defer testFile.Close()

		p := &IntelPMT{
			PmtSpec: testFile.Name(),
			Log:     testutil.Logger{},
		}
		err := p.Init()
		require.ErrorContains(t, err, "error while exploring pmt sysfs")
	})
}

func TestGather(t *testing.T) {
	type fields struct {
		PmtSpec                string
		Log                    telegraf.Logger
		pmtTelemetryFiles      map[string]pmtFileInfo
		pmtAggregator          map[string]aggregator
		pmtAggregatorInterface map[string]aggregatorInterface
		pmtTransformations     map[string]map[string]transformation
	}
	type testFile struct {
		guid     string
		content  []byte
		numaNode string
		pciBdf   string
	}
	tests := []struct {
		name     string
		fields   fields
		files    []testFile
		expected []telegraf.Metric
		wantErr  bool
	}{
		{
			name: "Incorrect gather, results map has no value for sample",
			fields: fields{
				pmtAggregator: map[string]aggregator{
					"test-guid": {
						SampleGroup: []sampleGroup{
							{
								SampleID: uint64(0),
								Sample: []sample{
									{
										DatatypeIDRef: "test-datatype",
										Msb:           4,
										Lsb:           4,
										mask:          16,
										SampleID:      "test-sample-ref",
									},
								},
							},
						},
					},
				},
				pmtAggregatorInterface: map[string]aggregatorInterface{
					"test-guid": {
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
												VarName: "testvar",
												// missing sampleIDREF
											},
										},
									},
								},
							},
						},
					},
				},
			},
			files: []testFile{
				{guid: "test-guid", content: []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, numaNode: "0"},
			},
			wantErr: true,
		},
		{
			name: "Failed Gather, no equation for gathered sample",
			fields: fields{
				pmtAggregatorInterface: map[string]aggregatorInterface{
					"test-guid": {
						AggregatorSamples: aggregatorSamples{
							AggregatorSample: []aggregatorSample{
								{SampleName: "test-sample"},
							},
						},
					},
				},
			},
			files: []testFile{
				{guid: "test-guid"},
			},
			wantErr: true,
		},
		{
			name: "Correct gather, 2 guids, 2 metrics returned",
			fields: fields{
				pmtAggregator: map[string]aggregator{
					"test-guid": {
						SampleGroup: []sampleGroup{
							{
								SampleID: uint64(0),
								Sample: []sample{
									{
										DatatypeIDRef: "test-datatype",
										Msb:           4,
										Lsb:           4,
										mask:          16,
										SampleID:      "test-sample-ref",
									},
								},
							},
						},
					},
					"test-guid2": {
						SampleGroup: []sampleGroup{
							{
								SampleID: uint64(0),
								Sample: []sample{
									{
										DatatypeIDRef: "test-datatype2",
										Msb:           14,
										Lsb:           0,
										mask:          32767,
										SampleID:      "test-sample-ref2",
									},
								},
							},
						},
					},
				},
				pmtAggregatorInterface: map[string]aggregatorInterface{
					"test-guid": {
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
					},
					"test-guid2": {
						AggregatorSamples: aggregatorSamples{
							AggregatorSample: []aggregatorSample{
								{
									SampleName:    "test-sample2",
									SampleGroup:   "test-group2",
									DatatypeIDRef: "test-datatype2",
									TransformREF:  "test-transform-ref2",
									TransformInputs: transformInputs{
										TransformInput: []transformInput{
											{
												VarName:     "testv",
												SampleIDREF: "test-sample-ref2",
											},
										},
									},
								},
							},
						},
					},
				},
				pmtTransformations: map[string]map[string]transformation{
					"test-guid": {
						"test-transform-ref": {
							Transform: "testvar + 2",
						},
					},
					"test-guid2": {
						"test-transform-ref2": {
							Transform: "( ( $testv &gt;&gt; 8 ) &amp; 0xff ) + ( ( $testv &amp; 0xff ) / ( 2 ** 8 ) ) - 64",
						},
					},
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"intel_pmt",
					map[string]string{
						"guid":           "test-guid",
						"numa_node":      "0",
						"pci_bdf":        "0000:00:0a.0",
						"sample_name":    "test-sample",
						"sample_group":   "test-group",
						"datatype_idref": "test-datatype",
					},
					map[string]interface{}{
						// 1 from buffer, 2 from equation
						"value": float64(3),
					},
					time.Time{},
				),
				testutil.MustMetric(
					"intel_pmt",
					map[string]string{
						"guid":           "test-guid2",
						"numa_node":      "1",
						"pci_bdf":        "0001:00:0a.0",
						"sample_name":    "test-sample2",
						"sample_group":   "test-group2",
						"datatype_idref": "test-datatype2",
					},
					map[string]interface{}{
						"value": float64(28.1875),
					},
					time.Time{},
				),
			},
			files: []testFile{
				{guid: "test-guid", content: []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, numaNode: "0", pciBdf: "0000:00:0a.0"},
				{guid: "test-guid2", content: []byte{0x30, 0x5c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, numaNode: "1", pciBdf: "0001:00:0a.0"},
			},
			wantErr: false,
		},
		{
			name: "Correct gather, 1 value returned",
			fields: fields{
				pmtAggregator: map[string]aggregator{
					"test-guid": {
						SampleGroup: []sampleGroup{
							{
								SampleID: uint64(0),
								Sample: []sample{
									{
										DatatypeIDRef: "test-datatype",
										Msb:           4,
										Lsb:           4,
										mask:          16,
										SampleID:      "test-sample-ref",
									},
								},
							},
						},
					},
				},
				pmtAggregatorInterface: map[string]aggregatorInterface{
					"test-guid": {
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
					},
				},
				pmtTransformations: map[string]map[string]transformation{
					"test-guid": {
						"test-transform-ref": {
							Transform: "testvar + 2",
						},
					},
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"intel_pmt",
					map[string]string{
						"guid":           "test-guid",
						"numa_node":      "0",
						"pci_bdf":        "0000:00:0a.0",
						"sample_name":    "test-sample",
						"sample_group":   "test-group",
						"datatype_idref": "test-datatype",
					},
					map[string]interface{}{
						// 1 from buffer, 2 from equation
						"value": float64(3),
					},
					time.Time{},
				),
			},
			files: []testFile{
				{guid: "test-guid", content: []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, numaNode: "0", pciBdf: "0000:00:0a.0"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &IntelPMT{
				PmtSpec:                tt.fields.PmtSpec,
				Log:                    testutil.Logger{},
				pmtAggregator:          tt.fields.pmtAggregator,
				pmtTelemetryFiles:      tt.fields.pmtTelemetryFiles,
				pmtAggregatorInterface: tt.fields.pmtAggregatorInterface,
				pmtTransformations:     tt.fields.pmtTransformations,
			}
			var acc testutil.Accumulator
			telemetryFiles := make(map[string]pmtFileInfo)
			tmp := t.TempDir()
			for _, file := range tt.files {
				testFile, _ := createTempFile(t, tmp, "test-file", file.content)
				telemetryFiles[file.guid] = append(telemetryFiles[file.guid], fileInfo{
					path:     testFile.Name(),
					numaNode: file.numaNode,
					pciBdf:   file.pciBdf,
				})
			}
			p.pmtTelemetryFiles = telemetryFiles
			if tt.wantErr {
				require.Error(t, acc.GatherError(p.Gather))
			} else {
				require.NoError(t, acc.GatherError(p.Gather))
				testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
			}
		})
	}
}

//go:build linux && amd64

package intel_pmt

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func createTempFile(t *testing.T, dir string, pattern string, data []byte) (*os.File, os.FileInfo) {
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
			s:        sample{Msb: 7, Lsb: 0},
			buf:      []byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			offset:   0,
			expected: 255,
		},
		{
			name:     "Middle bits set",
			s:        sample{Msb: 5, Lsb: 2},
			buf:      []byte{0x3c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // 0x3c = 00111100 in binary
			offset:   0,
			expected: 15,
		},
		{
			name:     "Non-zero offset",
			s:        sample{Msb: 7, Lsb: 0},
			buf:      []byte{0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			offset:   3,
			expected: 255,
		},
		{
			name:     "Single bit set",
			s:        sample{Msb: 4, Lsb: 4},
			buf:      []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // 0x10 = 00010000 in binary
			offset:   0,
			expected: 1,
		},
		{
			name:     "Two bytes set",
			s:        sample{Msb: 14, Lsb: 0},
			buf:      []byte{0x30, 0x5c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // 0x5c30 = 23600 in decimal
			offset:   0,
			expected: 23600,
		},
		{
			name:     "Offset larger than buffer size",
			s:        sample{Msb: 7, Lsb: 0},
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

type mockReader struct {
	data []byte
	err  error
}

func (mr mockReader) getReadCloser(_ string) (io.ReadCloser, error) {
	if mr.err != nil {
		return nil, mr.err
	}
	return io.NopCloser(bytes.NewReader(mr.data)), nil
}

func TestParseXML(t *testing.T) {
	type Person struct {
		Name string `xml:"name"`
		Age  int    `xml:"age"`
	}

	tests := []struct {
		name   string
		source string
		sr     sourceReader
		v      interface{}
		err    bool
	}{
		{
			name:   "Valid XML",
			source: "test",
			sr:     mockReader{data: []byte(`<Person><name>John</name><age>30</age></Person>`), err: nil},
			v:      &Person{},
			err:    false,
		},
		{
			name:   "Empty XML",
			source: "test",
			sr:     mockReader{data: []byte(``), err: nil},
			v:      &Person{},
			err:    true,
		},
		{
			name:   "Nil interface parameter",
			source: "test",
			sr:     mockReader{data: []byte(`<Person><name>John</name><age>30</age></Person>`), err: nil},
			v:      nil,
			err:    true,
		},
		{
			name:   "Error from SourceReader",
			source: "test",
			sr:     mockReader{data: nil, err: fmt.Errorf("mock error")},
			v:      &Person{},
			err:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseXML(tt.source, tt.sr, tt.v)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGather(t *testing.T) {
	t.Run("Correct gather, 1 value returned", func(t *testing.T) {
		p := &IntelPMT{
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
		}

		expected := []telegraf.Metric{
			testutil.MustMetric(
				"intel_pmt",
				map[string]string{
					"guid":           "test-guid",
					"numa_node":      "0",
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
		}
		tmp := t.TempDir()
		testFile, _ := createTempFile(t, tmp, "test-file", []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		defer testFile.Close()

		p.pmtTelemetryFiles = map[string]pmtFileInfo{
			"test-guid": []fileInfo{
				{path: testFile.Name(),
					numaNode: "0"},
			},
		}

		var acc testutil.Accumulator
		require.NoError(t, acc.GatherError(p.Gather))
		testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
	})
	t.Run("Correct gather, 2 guids, 2 metrics returned", func(t *testing.T) {
		p := &IntelPMT{
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
		}

		expected := []telegraf.Metric{
			testutil.MustMetric(
				"intel_pmt",
				map[string]string{
					"guid":           "test-guid",
					"numa_node":      "0",
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
					"sample_name":    "test-sample2",
					"sample_group":   "test-group2",
					"datatype_idref": "test-datatype2",
				},
				map[string]interface{}{
					"value": float64(28.1875),
				},
				time.Time{},
			),
		}
		tmp := t.TempDir()
		testFile, _ := createTempFile(t, tmp, "test-file", []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		defer testFile.Close()
		testFile2, _ := createTempFile(t, tmp, "test-file2", []byte{0x30, 0x5c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		defer testFile.Close()

		p.pmtTelemetryFiles = map[string]pmtFileInfo{
			"test-guid": []fileInfo{
				{path: testFile.Name(),
					numaNode: "0"},
			},
			"test-guid2": []fileInfo{
				{path: testFile2.Name(),
					numaNode: "1"},
			},
		}

		var acc testutil.Accumulator
		require.NoError(t, acc.GatherError(p.Gather))
		testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
	})

	t.Run("Failed Gather, no equation for gathered sample", func(t *testing.T) {
		p := &IntelPMT{
			pmtAggregatorInterface: map[string]aggregatorInterface{
				"test-guid": {
					AggregatorSamples: aggregatorSamples{
						AggregatorSample: []aggregatorSample{
							{SampleName: "test-sample"},
						},
					},
				},
			},
			Log: testutil.Logger{},
		}
		testFile, err := os.CreateTemp(t.TempDir(), "test-file")
		if err != nil {
			t.Fatalf("error creating a temporary file: %v %v", testFile.Name(), err)
		}
		defer testFile.Close()

		p.pmtTelemetryFiles = map[string]pmtFileInfo{
			"test-guid": []fileInfo{{path: testFile.Name()}},
		}

		var acc testutil.Accumulator
		require.Error(t, acc.GatherError(p.Gather))
	})

	t.Run("Incorrect gather, results map has no value for sample", func(t *testing.T) {
		p := &IntelPMT{
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
			Log: testutil.Logger{},
		}
		tmp := t.TempDir()
		testFile, _ := createTempFile(t, tmp, "test-file", []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		defer testFile.Close()

		p.pmtTelemetryFiles = map[string]pmtFileInfo{
			"test-guid": []fileInfo{
				{path: testFile.Name(),
					numaNode: "0"},
			},
		}

		var acc testutil.Accumulator
		require.Error(t, acc.GatherError(p.Gather))
	})
}

func TestInit(t *testing.T) {
	t.Run("No pmtSource", func(t *testing.T) {
		p := &IntelPMT{
			PmtSource: "",
		}
		err := p.Init()
		require.ErrorContains(t, err, "no source of XMLs provided")
	})

	t.Run("Incorrect filepath pmtSource", func(t *testing.T) {
		p := &IntelPMT{
			PmtSource: "/this/path/doesntexist",
		}
		err := p.Init()
		require.ErrorContains(t, err, "provided pmt source is invalid")
	})

	t.Run("Incorrect http pmtSource", func(t *testing.T) {
		p := &IntelPMT{
			PmtSource: "http://abcdtest.doesntexist",
		}
		err := p.Init()
		require.ErrorContains(t, err, "error reading source")
	})

	t.Run("Incorrect pmtSource, random letters", func(t *testing.T) {
		p := &IntelPMT{
			PmtSource: "loremipsum",
		}
		err := p.Init()
		require.ErrorContains(t, err, "provided pmt source is invalid")
	})

	t.Run("Correct filepath pmtSource, emptyfile", func(t *testing.T) {
		testFile, err := os.CreateTemp(t.TempDir(), "test-file")
		if err != nil {
			t.Fatalf("error creating a temporary file: %v %v", testFile.Name(), err)
		}
		defer testFile.Close()

		p := &IntelPMT{
			PmtSource: testFile.Name(),
		}
		err = p.Init()
		require.ErrorContains(t, err, "error decoding an XML")
	})

	t.Run("Correct filepath pmtSource, no pmt/can't read pmt in sysfs", func(t *testing.T) {
		tmp := t.TempDir()
		testFile, _ := createTempFile(t, tmp, "test-file", []byte("<pmt><mappings><mapping></mapping></mappings></pmt>"))
		defer testFile.Close()

		p := &IntelPMT{
			PmtSource: testFile.Name(),
			Log:       testutil.Logger{},
		}
		err := p.Init()
		require.ErrorContains(t, err, "error while exploring pmt sysfs")
	})
}

func TestReadXMLs(t *testing.T) {
	t.Run("Test single PMT GUID, no XMLs found", func(t *testing.T) {
		p := &IntelPMT{
			pmtMetadata: &pmt{
				Mappings: mappings{
					Mapping: []mapping{
						{GUID: "abc"},
					},
				},
			},
			pmtTelemetryFiles: map[string]pmtFileInfo{
				"abc": []fileInfo{{path: "doesn't-exist"}},
			},
			reader: fileReader{},
		}
		err := p.readXMLs()
		require.Error(t, err)
		require.ErrorContains(t, err, "failed reading XMLs")
	})

	t.Run("Test single PMT GUID, aggregator interface empty", func(t *testing.T) {
		tmp := t.TempDir()

		bufAgg := []byte("<TELEM:Aggregator><TELEM:SampleGroup></TELEM:SampleGroup></TELEM:Aggregator>")
		testAgg, aggName := createTempFile(t, tmp, "test-agg", bufAgg)
		defer testAgg.Close()

		bufAggInterface := []byte("<TELI:AggregatorInterface></TELI:AggregatorInterface>")
		testAggInterface, aggInterfaceName := createTempFile(t, tmp, "test-aggInterface", bufAggInterface)
		defer testAggInterface.Close()

		p := &IntelPMT{
			pmtBasePath: tmp,
			pmtMetadata: &pmt{
				Mappings: mappings{
					Mapping: []mapping{
						{
							GUID: "abc",
							XMLSet: xmlset{
								Aggregator:          aggName.Name(),
								AggregatorInterface: aggInterfaceName.Name(),
							},
						},
					},
				},
			},
			// This is done just so we enter the loop
			pmtTelemetryFiles: map[string]pmtFileInfo{
				"abc": []fileInfo{{path: testAgg.Name()}},
			},
			isFilePath: true,
			reader:     fileReader{},
		}

		err := p.readXMLs()
		require.ErrorContains(t, err, "all aggregator interface XMLs are empty")
	})
}

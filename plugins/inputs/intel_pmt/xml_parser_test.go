//go:build linux && amd64

package intel_pmt

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

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

func TestParseXMLs(t *testing.T) {
	t.Run("Correct filepath PmtSpec, empty spec", func(t *testing.T) {
		testFile, err := os.CreateTemp(t.TempDir(), "test-file")
		if err != nil {
			t.Fatalf("error creating a temporary file: %v %v", testFile.Name(), err)
		}
		defer testFile.Close()

		p := &IntelPMT{
			PmtSpec: testFile.Name(),
			reader:  fileReader{},
		}
		err = p.parseXMLs()
		require.ErrorContains(t, err, "error decoding an XML")
	})
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
			sr:     mockReader{data: nil, err: errors.New("mock error")},
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
			reader: fileReader{},
		}

		err := p.readXMLs()
		require.ErrorContains(t, err, "all aggregator interface XMLs are empty")
	})
}

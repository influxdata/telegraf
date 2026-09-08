package opentelemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	service "go.opentelemetry.io/proto/otlp/collector/profiles/v1development"
	common "go.opentelemetry.io/proto/otlp/common/v1"
	otlp "go.opentelemetry.io/proto/otlp/profiles/v1development"
	resource "go.opentelemetry.io/proto/otlp/resource/v1"

	"github.com/influxdata/telegraf/testutil"
)

// requestWithSample builds a profile request whose tables are all empty apart
// from the entries a well-formed sample needs, so a test can point a single
// index past the end of its table and change nothing else.
func requestWithSample(sample *otlp.Sample, dictionary *otlp.ProfilesDictionary) *service.ExportProfilesServiceRequest {
	return &service.ExportProfilesServiceRequest{
		Dictionary: dictionary,
		ResourceProfiles: []*otlp.ResourceProfiles{{
			Resource: &resource.Resource{},
			ScopeProfiles: []*otlp.ScopeProfiles{{
				Profiles: []*otlp.Profile{{
					ProfileId:  []byte{1, 2, 3, 4},
					SampleType: &otlp.ValueType{TypeStrindex: 0, UnitStrindex: 0},
					PeriodType: &otlp.ValueType{TypeStrindex: 0, UnitStrindex: 0},
					Samples:    []*otlp.Sample{sample},
				}},
			}},
		}},
	}
}

func validDictionary() *otlp.ProfilesDictionary {
	return &otlp.ProfilesDictionary{
		StringTable:   []string{"", "main.go", "main"},
		MappingTable:  []*otlp.Mapping{{}},
		FunctionTable: []*otlp.Function{{NameStrindex: 2, FilenameStrindex: 1}},
		LocationTable: []*otlp.Location{{Lines: []*otlp.Line{{FunctionIndex: 0}}}},
		StackTable:    []*otlp.Stack{{LocationIndices: []int32{0}}},
		AttributeTable: []*otlp.KeyValueAndUnit{
			{KeyStrindex: 1, Value: &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: "v"}}},
		},
	}
}

func TestProfileExportOutOfRangeIndexes(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(sample *otlp.Sample, dictionary *otlp.ProfilesDictionary)
		expected int
	}{
		{
			name:     "well formed",
			mutate:   func(*otlp.Sample, *otlp.ProfilesDictionary) {},
			expected: 1,
		},
		{
			name: "stack index",
			mutate: func(sample *otlp.Sample, _ *otlp.ProfilesDictionary) {
				sample.StackIndex = 999
			},
		},
		{
			name: "location index",
			mutate: func(_ *otlp.Sample, dictionary *otlp.ProfilesDictionary) {
				dictionary.StackTable[0].LocationIndices = []int32{999}
			},
		},
		{
			name: "function index",
			mutate: func(_ *otlp.Sample, dictionary *otlp.ProfilesDictionary) {
				dictionary.LocationTable[0].Lines[0].FunctionIndex = 999
			},
		},
		{
			name: "function name string index",
			mutate: func(_ *otlp.Sample, dictionary *otlp.ProfilesDictionary) {
				dictionary.FunctionTable[0].NameStrindex = 999
			},
		},
		{
			name: "mapping index",
			mutate: func(_ *otlp.Sample, dictionary *otlp.ProfilesDictionary) {
				dictionary.LocationTable[0].MappingIndex = 999
			},
		},
		{
			name: "sample attribute index",
			mutate: func(sample *otlp.Sample, _ *otlp.ProfilesDictionary) {
				sample.AttributeIndices = []int32{999}
			},
		},
		{
			name: "negative index",
			mutate: func(sample *otlp.Sample, _ *otlp.ProfilesDictionary) {
				sample.StackIndex = -1
			},
		},
		{
			name: "missing timestamp for value",
			mutate: func(sample *otlp.Sample, _ *otlp.ProfilesDictionary) {
				sample.TimestampsUnixNano = nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample := &otlp.Sample{
				StackIndex:         0,
				Values:             []int64{1},
				TimestampsUnixNano: []uint64{1234},
				AttributeIndices:   []int32{0},
			}
			dictionary := validDictionary()
			tt.mutate(sample, dictionary)

			var acc testutil.Accumulator
			svc, err := newProfileService(&acc, testutil.Logger{}, nil)
			require.NoError(t, err)

			require.NotPanics(t, func() {
				_, err = svc.Export(context.Background(), requestWithSample(sample, dictionary))
			})
			require.NoError(t, err)
			require.Len(t, acc.Metrics, tt.expected)
		})
	}
}

func TestProfileExportMissingDictionary(t *testing.T) {
	var acc testutil.Accumulator
	svc, err := newProfileService(&acc, testutil.Logger{}, nil)
	require.NoError(t, err)

	request := requestWithSample(&otlp.Sample{}, nil)

	require.NotPanics(t, func() {
		_, err = svc.Export(context.Background(), request)
	})
	require.Error(t, err)
	require.Empty(t, acc.Metrics)
}

func TestProfileExportNilResource(t *testing.T) {
	sample := &otlp.Sample{
		Values:             []int64{1},
		TimestampsUnixNano: []uint64{1234},
	}
	request := requestWithSample(sample, validDictionary())
	request.ResourceProfiles[0].Resource = nil

	var acc testutil.Accumulator
	svc, err := newProfileService(&acc, testutil.Logger{}, nil)
	require.NoError(t, err)

	require.NotPanics(t, func() {
		_, err = svc.Export(context.Background(), request)
	})
	require.NoError(t, err)
	require.Len(t, acc.Metrics, 1)
}

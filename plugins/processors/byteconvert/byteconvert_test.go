package byteconvert

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestByteConvert(t *testing.T) {
	tests := []struct {
		name     string
		bc       *ByteConvert
		input    telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "B to KiB",
			bc: &ByteConvert{
				FieldSrc:    "foo",
				FieldName:   "bar",
				ConvertUnit: "KiB",
			},
			input: testutil.MustMetric(
				"test_metric",
				map[string]string{},
				map[string]interface{}{
					"foo": float64(1024),
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"test_metric",
					map[string]string{},
					map[string]interface{}{
						"foo": float64(1024),
						"bar": float64(1),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "B to MiB",
			bc: &ByteConvert{
				FieldSrc:    "foo",
				FieldName:   "bar",
				ConvertUnit: "MiB",
			},
			input: testutil.MustMetric(
				"test_metric",
				map[string]string{},
				map[string]interface{}{
					"foo": float64(1048576),
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"test_metric",
					map[string]string{},
					map[string]interface{}{
						"foo": float64(1048576),
						"bar": float64(1),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "B to GiB",
			bc: &ByteConvert{
				FieldSrc:    "foo",
				FieldName:   "bar",
				ConvertUnit: "GiB",
			},
			input: testutil.MustMetric(
				"test_metric",
				map[string]string{},
				map[string]interface{}{
					"foo": float64(1073741824),
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"test_metric",
					map[string]string{},
					map[string]interface{}{
						"foo": float64(1073741824),
						"bar": float64(1),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "invalid source field",
			bc: &ByteConvert{
				FieldSrc:    "baz",
				FieldName:   "bar",
				ConvertUnit: "GiB",
			},
			input: testutil.MustMetric(
				"test_metric",
				map[string]string{},
				map[string]interface{}{
					"foo": float64(1073741824),
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"test_metric",
					map[string]string{},
					map[string]interface{}{
						"foo": float64(1073741824),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "invalid convert unit",
			bc: &ByteConvert{
				FieldSrc:    "foo",
				FieldName:   "bar",
				ConvertUnit: "TiB",
			},
			input: testutil.MustMetric(
				"test_metric",
				map[string]string{},
				map[string]interface{}{
					"foo": float64(1073741824),
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"test_metric",
					map[string]string{},
					map[string]interface{}{
						"foo": float64(1073741824),
						"bar": float64(0),
					},
					time.Unix(0, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bc.Init()
			require.NoError(t, err)
			actual := tt.bc.Apply(tt.input)

			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

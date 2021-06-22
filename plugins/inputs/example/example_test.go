package example

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

// This file should contain a set of unit-tests to cover your plugin. This will ease
// spotting bugs and mistakes when later modifying or extending the functionality.
// To do so, please write one 'TestXYZ' function per 'case' e.g. default init,
// things that should fail or expected values from a mockup.

func TestInitDefault(t *testing.T) {
	// This test should succeed with the default initialization.

	// Use whatever you use in the init() function plus the mandatory options.
	// ATTENTION: Always initialze the "Log" as you will get SIGSEGV otherwise.
	plugin := &Example{
		DeviceName: "test",
		Timeout:    config.Duration(100 * time.Millisecond),
		Log:        testutil.Logger{},
	}

	// Test the initialization succeeds
	require.NoError(t, plugin.Init())

	// Also test that default values are set correctly
	require.Equal(t, config.Duration(100*time.Millisecond), plugin.Timeout)
	require.Equal(t, "test", plugin.DeviceName)
	require.Equal(t, int64(2), plugin.NumberFields)
}

func TestInitFail(t *testing.T) {
	// You should also test for your safety nets to work i.e. you get errors for
	// invalid configuration-option values. So check your error paths in Init()
	// and check if you reach them

	// We setup a table-test here to specify "setting" - "expected error" values.
	// Eventhough it seems overkill here for the example plugin, we reuse this structure
	// later for checking the metrics
	tests := []struct {
		name     string
		plugin   *Example
		expected string
	}{
		{
			name:     "all empty",
			plugin:   &Example{},
			expected: "device name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Always initialze the logger to avoid SIGSEGV. This is done automatically by
			// telegraf during normal operation.
			tt.plugin.Log = testutil.Logger{}
			err := tt.plugin.Init()
			require.Error(t, err)
			require.EqualError(t, err, tt.expected)
		})
	}
}

func TestFixedValue(t *testing.T) {
	// You can organize the test e.g. by operation mode (like we do here random vs. fixed), by features or
	// by different metrics gathered. Please choose the partitioning most suited for your plugin

	// We again setup a table-test here to specify "setting" - "expected output metric" pairs.
	tests := []struct {
		name     string
		plugin   *Example
		expected []telegraf.Metric
	}{
		{
			name: "count only",
			plugin: &Example{
				DeviceName:   "test",
				NumberFields: 1,
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"example",
					map[string]string{
						"device": "test",
					},
					map[string]interface{}{
						"count": 1,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"example",
					map[string]string{
						"device": "test",
					},
					map[string]interface{}{
						"count": 2,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"example",
					map[string]string{
						"device": "test",
					},
					map[string]interface{}{
						"count": 3,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "default settings",
			plugin: &Example{
				DeviceName: "test",
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"example",
					map[string]string{
						"device": "test",
					},
					map[string]interface{}{
						"count":  1,
						"field1": float64(0),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"example",
					map[string]string{
						"device": "test",
					},
					map[string]interface{}{
						"count":  2,
						"field1": float64(0),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"example",
					map[string]string{
						"device": "test",
					},
					map[string]interface{}{
						"count":  3,
						"field1": float64(0),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "more fields",
			plugin: &Example{
				DeviceName:   "test",
				NumberFields: 4,
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"example",
					map[string]string{
						"device": "test",
					},
					map[string]interface{}{
						"count":  1,
						"field1": float64(0),
						"field2": float64(0),
						"field3": float64(0),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"example",
					map[string]string{
						"device": "test",
					},
					map[string]interface{}{
						"count":  2,
						"field1": float64(0),
						"field2": float64(0),
						"field3": float64(0),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"example",
					map[string]string{
						"device": "test",
					},
					map[string]interface{}{
						"count":  3,
						"field1": float64(0),
						"field2": float64(0),
						"field3": float64(0),
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator

			tt.plugin.Log = testutil.Logger{}
			require.NoError(t, tt.plugin.Init())

			// Call gather and check no error occurs. In case you use acc.AddError() somewhere
			// in your code, it is not sufficient to only check the return value of Gather().
			require.NoError(t, tt.plugin.Gather(&acc))
			require.Len(t, acc.Errors, 0, "found errors accumulated by acc.AddError()")

			// Wait for the expected number of metrics to avoid flaky tests due to
			// race conditions.
			acc.Wait(len(tt.expected))

			// Compare the metrics in a convenient way. Here we ignore
			// the metric time during comparision as we cannot inject the time
			// during test. For more comparision options check testutil package.
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestRandomValue(t *testing.T) {
	// Sometimes, you cannot know the exact outcome of the gather cycle e.g. if the gathering involves random data.
	// However, you should check the result nevertheless, applying as many conditions as you can.

	// We again setup a table-test here to specify "setting" - "expected output metric" pairs.
	tests := []struct {
		name     string
		plugin   *Example
		template telegraf.Metric
	}{
		{
			name: "count only",
			plugin: &Example{
				DeviceName:           "test",
				NumberFields:         1,
				EnableRandomVariable: true,
			},
			template: testutil.MustMetric(
				"example",
				map[string]string{
					"device": "test",
				},
				map[string]interface{}{
					"count": 1,
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "default settings",
			plugin: &Example{
				DeviceName:           "test",
				EnableRandomVariable: true,
			},
			template: testutil.MustMetric(
				"example",
				map[string]string{
					"device": "test",
				},
				map[string]interface{}{
					"count":  1,
					"field1": float64(0),
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "more fields",
			plugin: &Example{
				DeviceName:           "test",
				NumberFields:         4,
				EnableRandomVariable: true,
			},
			template: testutil.MustMetric(
				"example",
				map[string]string{
					"device": "test",
				},
				map[string]interface{}{
					"count":  1,
					"field1": float64(0),
					"field2": float64(0),
					"field3": float64(0),
				},
				time.Unix(0, 0),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator

			tt.plugin.Log = testutil.Logger{}
			require.NoError(t, tt.plugin.Init())

			// Call gather and check no error occurs. In case you use acc.AddError() somewhere
			// in your code, it is not sufficient to only check the return value of Gather().
			require.NoError(t, tt.plugin.Gather(&acc))
			require.Len(t, acc.Errors, 0, "found errors accumulated by acc.AddError()")

			// Wait for the expected number of metrics to avoid flaky tests due to
			// race conditions.
			acc.Wait(3)

			// Compare all aspects of the metric that are known to you
			for i, m := range acc.GetTelegrafMetrics() {
				require.Equal(t, m.Name(), tt.template.Name())
				require.Equal(t, m.Tags(), tt.template.Tags())

				// Check if all expected fields are there
				fields := m.Fields()
				for k := range tt.template.Fields() {
					if k == "count" {
						require.Equal(t, fields["count"], int64(i+1))
						continue
					}
					_, found := fields[k]
					require.Truef(t, found, "field %q not found", k)
				}
			}
		})
	}
}

func TestGatherFail(t *testing.T) {
	// You should also test for error conditions in your Gather() method. Try to cover all error paths.

	// We again setup a table-test here to specify "setting" - "expected error" pair.
	tests := []struct {
		name     string
		plugin   *Example
		expected string
	}{
		{
			name: "too many fields",
			plugin: &Example{
				DeviceName:   "test",
				NumberFields: 11,
			},
			expected: "too many fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator

			tt.plugin.Log = testutil.Logger{}
			require.NoError(t, tt.plugin.Init())

			err := tt.plugin.Gather(&acc)
			require.Error(t, err)
			require.EqualError(t, err, tt.expected)
		})
	}
}

func TestRandomValueFailPartial(t *testing.T) {
	// You should also test for error conditions in your Gather() with partial output. This is required when
	// using acc.AddError() as Gather() might succeed (return nil) but there are some metrics missing.

	// We again setup a table-test here to specify "setting" - "expected output metric" and "errors".
	tests := []struct {
		name        string
		plugin      *Example
		expected    []telegraf.Metric
		expectedErr string
	}{
		{
			name: "flappy gather",
			plugin: &Example{
				DeviceName:           "flappy",
				NumberFields:         1,
				EnableRandomVariable: true,
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"example",
					map[string]string{
						"device": "flappy",
					},
					map[string]interface{}{
						"count": 1,
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"example",
					map[string]string{
						"device": "flappy",
					},
					map[string]interface{}{
						"count": 2,
					},
					time.Unix(0, 0),
				),
			},
			expectedErr: "too many runs for random values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator

			tt.plugin.Log = testutil.Logger{}
			require.NoError(t, tt.plugin.Init())

			// Call gather and check no error occurs. However, we expect an error accumulated by acc.AddError()
			require.NoError(t, tt.plugin.Gather(&acc))

			// Wait for the expected number of metrics to avoid flaky tests due to
			// race conditions.
			acc.Wait(len(tt.expected))

			// Check the accumulated errors
			require.Len(t, acc.Errors, 1)
			require.EqualError(t, acc.Errors[0], tt.expectedErr)

			// Compare the expected partial metrics.
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

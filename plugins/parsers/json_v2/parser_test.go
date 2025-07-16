package json_v2_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/file"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/parsers/json_v2"
	"github.com/influxdata/telegraf/testutil"
)

func TestMultipleConfigs(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testdata")
	require.NoError(t, err)
	// Make sure testdata contains data
	require.NotEmpty(t, folders)

	// Setup influx parser for parsing the expected metrics
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	inputs.Add("file", func() telegraf.Input {
		return &file.File{}
	})

	for _, f := range folders {
		// Only use directories as those contain test-cases
		if !f.IsDir() {
			continue
		}
		testdataPath := filepath.Join("testdata", f.Name())
		configFilename := filepath.Join(testdataPath, "telegraf.conf")
		expectedFilename := filepath.Join(testdataPath, "expected.out")
		expectedErrorFilename := filepath.Join(testdataPath, "expected.err")

		t.Run(f.Name(), func(t *testing.T) {
			// Read the expected output
			expected, err := testutil.ParseMetricsFromFile(expectedFilename, parser)
			require.NoError(t, err)

			// Read the expected errors if any
			var expectedErrors []string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				var err error
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, expectedErrors)
			}

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))

			// Gather the metrics from the input file configure
			var acc testutil.Accumulator
			var actualErrorMsgs []string
			for _, input := range cfg.Inputs {
				require.NoError(t, input.Init())
				if err := input.Gather(&acc); err != nil {
					actualErrorMsgs = append(actualErrorMsgs, err.Error())
				}
			}

			// If the test has expected error(s) then compare them
			if len(expectedErrors) > 0 {
				sort.Strings(actualErrorMsgs)
				sort.Strings(expectedErrors)
				for i, msg := range expectedErrors {
					require.Contains(t, actualErrorMsgs[i], msg)
				}
			} else {
				require.Empty(t, actualErrorMsgs)
			}

			// Process expected metrics and compare with resulting metrics
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())

			// Folder with timestamp prefixed will also check for matching timestamps to make sure they are parsed correctly
			// The milliseconds weren't matching, seemed like a rounding difference between the influx parser
			// Compares each metrics times separately and ignores milliseconds
			if strings.HasPrefix(f.Name(), "timestamp") {
				require.Len(t, actual, len(expected))
				for i, m := range actual {
					require.Equal(t, expected[i].Time().Truncate(time.Second), m.Time().Truncate(time.Second))
				}
			}
		})
	}
}

func TestParserEmptyConfig(t *testing.T) {
	plugin := &json_v2.Parser{}
	require.ErrorContains(t, plugin.Init(), "no configuration provided")
}

// Test edge cases
func TestJSONV2StringTypeEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		config         json_v2.Config
		expectedFields map[string]string
		expectError    bool
	}{
		{
			name:  "empty object as string",
			input: `{"empty_obj": {}}`,
			config: json_v2.Config{
				MeasurementName: "test",
				Fields: []json_v2.DataSet{
					{Path: "empty_obj", Type: "string"},
				},
			},
			expectedFields: map[string]string{
				"empty_obj": "{}",
			},
		},
		{
			name:  "empty array as string",
			input: `{"empty_arr": []}`,
			config: json_v2.Config{
				MeasurementName: "test",
				Fields: []json_v2.DataSet{
					{Path: "empty_arr", Type: "string"},
				},
			},
			expectedFields: map[string]string{
				"empty_arr": "[]",
			},
		},
		{
			name:  "mixed types - some string, some expanded",
			input: `{"obj1": {"a": 1}, "obj2": {"b": 2}, "simple": "value"}`,
			config: json_v2.Config{
				MeasurementName: "test",
				Fields: []json_v2.DataSet{
					{Path: "obj1", Type: "string"},   // Should be string
					{Path: "simple", Type: "string"}, // Should be string
				},
			},
			expectedFields: map[string]string{
				"obj1":   `{"a": 1}`,
				"simple": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &json_v2.Parser{Configs: []json_v2.Config{tt.config}}
			err := parser.Init()
			require.NoError(t, err)

			metrics, err := parser.Parse([]byte(tt.input))
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, metrics, 1)

			metric := metrics[0]
			for fieldName, expectedValue := range tt.expectedFields {
				actualValue, ok := metric.GetField(fieldName)
				require.True(t, ok, "Field %s should exist", fieldName)
				require.Equal(t, expectedValue, actualValue, "Field %s value mismatch", fieldName)
			}
		})
	}
}

// Integration test that mirrors the original issue from the bug report(#16602)
func TestJSONV2StringTypeFieldsIntegration(t *testing.T) {
	input := `{"controller_id":"C_abcd1234","timestamp":"2025-03-10T08:20:26.506Z",` +
		`"event":{"type":"METER_CONNECTION_QUALITY","state":"BAD",` +
		`"metadata":{"failures":59,"failureRate":"0.98"},"severity":"CRITICAL"},` +
		`"active_issues":[{"type":"METER_CONNECTION_QUALITY",` +
		`"since":"2025-03-10T05:59:22.146Z","last_update":"2025-03-10T08:20:26.506Z",` +
		`"state":"BAD","severity":"CRITICAL","metadata":{"failures":59,"failureRate":"0.98"}}]}`

	parser := &json_v2.Parser{
		Configs: []json_v2.Config{
			{
				MeasurementName: "events",
				Fields: []json_v2.DataSet{
					{Path: "event.metadata", Rename: "metadata", Type: "string"},
					{Path: "active_issues", Rename: "active_issues", Type: "string"},
					{Path: "event.type", Rename: "event_type"},
					{Path: "event.state", Rename: "state"},
					{Path: "event.severity", Rename: "severity"},
				},
				Tags: []json_v2.DataSet{
					{Path: "controller_id", Rename: "controller_id"},
				},
			},
		},
	}

	err := parser.Init()
	require.NoError(t, err)

	metrics, err := parser.Parse([]byte(input))
	require.NoError(t, err)
	require.Len(t, metrics, 1)

	metric := metrics[0]

	// Check that complex structures are stored as strings
	metadataField, ok := metric.GetField("metadata")
	require.True(t, ok, "metadata field should exist")
	metadataStr, ok := metadataField.(string)
	require.True(t, ok, "metadata should be a string, got %T", metadataField)
	require.Contains(t, metadataStr, "failures", "metadata should contain the JSON content")
	require.Contains(t, metadataStr, "failureRate", "metadata should contain the JSON content")

	activeIssuesField, ok := metric.GetField("active_issues")
	require.True(t, ok, "active_issues field should exist")
	activeIssuesStr, ok := activeIssuesField.(string)
	require.True(t, ok, "active_issues should be a string, got %T", activeIssuesField)
	require.Contains(t, activeIssuesStr, "METER_CONNECTION_QUALITY", "active_issues should contain the array content")
	require.Contains(t, activeIssuesStr, "since", "active_issues should contain the array content")

	// Verify no flattened fields exist (this was the bug)
	_, hasMetadataFailures := metric.GetField("metadata_failures")
	require.False(t, hasMetadataFailures, "Should not have flattened metadata_failures field")

	_, hasActiveIssuesType := metric.GetField("active_issues_type")
	require.False(t, hasActiveIssuesType, "Should not have flattened active_issues_type field")

	// Check that simple fields work normally
	eventType, ok := metric.GetField("event_type")
	require.True(t, ok)
	require.Equal(t, "METER_CONNECTION_QUALITY", eventType)

	state, ok := metric.GetField("state")
	require.True(t, ok)
	require.Equal(t, "BAD", state)

	// Check tag
	controllerID, ok := metric.GetTag("controller_id")
	require.True(t, ok)
	require.Equal(t, "C_abcd1234", controllerID)
}

func BenchmarkParsingSequential(b *testing.B) {
	inputFilename := filepath.Join("testdata", "benchmark", "input.json")

	// Configure the plugin
	plugin := &json_v2.Parser{
		Configs: []json_v2.Config{
			{
				MeasurementName: "benchmark",
				JSONObjects: []json_v2.Object{
					{
						Path:               "metrics",
						DisablePrependKeys: true,
					},
				},
			},
		},
	}
	require.NoError(b, plugin.Init())

	// Read the input data
	input, err := os.ReadFile(inputFilename)
	require.NoError(b, err)

	// Do the benchmarking
	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse(input)
	}
}

func BenchmarkParsingParallel(b *testing.B) {
	inputFilename := filepath.Join("testdata", "benchmark", "input.json")

	// Configure the plugin
	plugin := &json_v2.Parser{
		Configs: []json_v2.Config{
			{
				MeasurementName: "benchmark",
				JSONObjects: []json_v2.Object{
					{
						Path:               "metrics",
						DisablePrependKeys: true,
					},
				},
			},
		},
	}
	require.NoError(b, plugin.Init())

	// Read the input data
	input, err := os.ReadFile(inputFilename)
	require.NoError(b, err)

	// Do the benchmarking
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
			plugin.Parse(input)
		}
	})
}

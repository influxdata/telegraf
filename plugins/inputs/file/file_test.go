// +build !windows

// TODO: Windows - should be enabled for Windows when super asterisk is fixed on Windows
// https://github.com/influxdata/telegraf/issues/6248

package file

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/csv"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshFilePaths(t *testing.T) {
	wd, err := os.Getwd()
	r := File{
		Files: []string{filepath.Join(wd, "dev/testfiles/**.log")},
	}
	err = r.Init()
	require.NoError(t, err)

	err = r.refreshFilePaths()
	require.NoError(t, err)
	assert.Equal(t, 2, len(r.filenames))
}

func TestFileTag(t *testing.T) {
	acc := testutil.Accumulator{}
	wd, err := os.Getwd()
	require.NoError(t, err)
	r := File{
		Files:   []string{filepath.Join(wd, "dev/testfiles/json_a.log")},
		FileTag: "filename",
	}
	err = r.Init()
	require.NoError(t, err)

	parserConfig := parsers.Config{
		DataFormat: "json",
	}
	nParser, err := parsers.NewParser(&parserConfig)
	assert.NoError(t, err)
	r.parser = nParser

	err = r.Gather(&acc)
	require.NoError(t, err)

	for _, m := range acc.Metrics {
		for key, value := range m.Tags {
			assert.Equal(t, r.FileTag, key)
			assert.Equal(t, filepath.Base(r.Files[0]), value)
		}
	}
}

func TestJSONParserCompile(t *testing.T) {
	var acc testutil.Accumulator
	wd, _ := os.Getwd()
	r := File{
		Files: []string{filepath.Join(wd, "dev/testfiles/json_a.log")},
	}
	err := r.Init()
	require.NoError(t, err)
	parserConfig := parsers.Config{
		DataFormat: "json",
		TagKeys:    []string{"parent_ignored_child"},
	}
	nParser, err := parsers.NewParser(&parserConfig)
	assert.NoError(t, err)
	r.parser = nParser

	r.Gather(&acc)
	assert.Equal(t, map[string]string{"parent_ignored_child": "hi"}, acc.Metrics[0].Tags)
	assert.Equal(t, 5, len(acc.Metrics[0].Fields))
}

func TestGrokParser(t *testing.T) {
	wd, _ := os.Getwd()
	var acc testutil.Accumulator
	r := File{
		Files: []string{filepath.Join(wd, "dev/testfiles/grok_a.log")},
	}
	err := r.Init()
	require.NoError(t, err)

	parserConfig := parsers.Config{
		DataFormat:   "grok",
		GrokPatterns: []string{"%{COMMON_LOG_FORMAT}"},
	}

	nParser, err := parsers.NewParser(&parserConfig)
	r.parser = nParser
	assert.NoError(t, err)

	err = r.Gather(&acc)
	assert.Equal(t, len(acc.Metrics), 2)
}

func TestCharacterEncoding(t *testing.T) {
	expected := []telegraf.Metric{
		testutil.MustMetric("file",
			map[string]string{
				"dest": "example.org",
				"hop":  "1",
				"ip":   "12.122.114.5",
			},
			map[string]interface{}{
				"avg":    21.55,
				"best":   19.34,
				"loss":   0.0,
				"snt":    10,
				"status": "OK",
				"stdev":  2.05,
				"worst":  26.83,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("file",
			map[string]string{
				"dest": "example.org",
				"hop":  "2",
				"ip":   "192.205.32.238",
			},
			map[string]interface{}{
				"avg":    25.11,
				"best":   20.8,
				"loss":   0.0,
				"snt":    10,
				"status": "OK",
				"stdev":  6.03,
				"worst":  38.85,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("file",
			map[string]string{
				"dest": "example.org",
				"hop":  "3",
				"ip":   "152.195.85.133",
			},
			map[string]interface{}{
				"avg":    20.18,
				"best":   19.75,
				"loss":   0.0,
				"snt":    10,
				"status": "OK",
				"stdev":  0.0,
				"worst":  20.78,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("file",
			map[string]string{
				"dest": "example.org",
				"hop":  "4",
				"ip":   "93.184.216.34",
			},
			map[string]interface{}{
				"avg":    24.02,
				"best":   19.75,
				"loss":   0.0,
				"snt":    10,
				"status": "OK",
				"stdev":  4.67,
				"worst":  32.41,
			},
			time.Unix(0, 0),
		),
	}

	tests := []struct {
		name   string
		plugin *File
		csv    *csv.Config
		file   string
	}{
		{
			name: "empty character_encoding with utf-8",
			plugin: &File{
				Files:             []string{"testdata/mtr-utf-8.csv"},
				CharacterEncoding: "",
			},
			csv: &csv.Config{
				MetricName:  "file",
				SkipRows:    1,
				ColumnNames: []string{"", "", "status", "dest", "hop", "ip", "loss", "snt", "", "", "avg", "best", "worst", "stdev"},
				TagColumns:  []string{"dest", "hop", "ip"},
			},
		},
		{
			name: "utf-8 character_encoding with utf-8",
			plugin: &File{
				Files:             []string{"testdata/mtr-utf-8.csv"},
				CharacterEncoding: "utf-8",
			},
			csv: &csv.Config{
				MetricName:  "file",
				SkipRows:    1,
				ColumnNames: []string{"", "", "status", "dest", "hop", "ip", "loss", "snt", "", "", "avg", "best", "worst", "stdev"},
				TagColumns:  []string{"dest", "hop", "ip"},
			},
		},
		{
			name: "utf-16le character_encoding with utf-16le",
			plugin: &File{
				Files:             []string{"testdata/mtr-utf-16le.csv"},
				CharacterEncoding: "utf-16le",
			},
			csv: &csv.Config{
				MetricName:  "file",
				SkipRows:    1,
				ColumnNames: []string{"", "", "status", "dest", "hop", "ip", "loss", "snt", "", "", "avg", "best", "worst", "stdev"},
				TagColumns:  []string{"dest", "hop", "ip"},
			},
		},
		{
			name: "utf-16be character_encoding with utf-16be",
			plugin: &File{
				Files:             []string{"testdata/mtr-utf-16be.csv"},
				CharacterEncoding: "utf-16be",
			},
			csv: &csv.Config{
				MetricName:  "file",
				SkipRows:    1,
				ColumnNames: []string{"", "", "status", "dest", "hop", "ip", "loss", "snt", "", "", "avg", "best", "worst", "stdev"},
				TagColumns:  []string{"dest", "hop", "ip"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Init()
			require.NoError(t, err)

			parser, err := csv.NewParser(tt.csv)
			require.NoError(t, err)
			tt.plugin.SetParser(parser)

			var acc testutil.Accumulator
			err = tt.plugin.Gather(&acc)
			require.NoError(t, err)

			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

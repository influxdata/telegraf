package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshFilePaths(t *testing.T) {
	wd, err := os.Getwd()
	r := File{
		Files: []string{filepath.Join(wd, "dev/testfiles/**.log")},
	}

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

package reader

import (
	"runtime"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestRefreshFilePaths(t *testing.T) {
	testDir := getPluginDir()
	r := Reader{
		Files: []string{testDir + "/reader/dev/testfiles/**.log"},
	}

	r.refreshFilePaths()
	assert.Equal(t, len(r.filenames), 2)
}
func TestJSONParserCompile(t *testing.T) {
	testDir := getPluginDir()
	var acc testutil.Accumulator
	r := Reader{
		Files: []string{testDir + "/reader/dev/testfiles/json_a.log"},
	}
	parserConfig := parsers.Config{
		DataFormat: "json",
		TagKeys:    []string{"parent_ignored_child"},
	}
	nParser, err := parsers.NewParser(&parserConfig)
	r.parser = nParser
	assert.NoError(t, err)

	r.Gather(&acc)
	assert.Equal(t, map[string]string{"parent_ignored_child": "hi"}, acc.Metrics[0].Tags)
	assert.Equal(t, 5, len(acc.Metrics[0].Fields))
}

func TestGrokParser(t *testing.T) {
	testDir := getPluginDir()
	var acc testutil.Accumulator
	r := Reader{
		Files: []string{testDir + "/reader/dev/testfiles/grok_a.log"},
	}

	parserConfig := parsers.Config{
		DataFormat:   "grok",
		GrokPatterns: []string{"%{COMMON_LOG_FORMAT}"},
	}

	nParser, err := parsers.NewParser(&parserConfig)
	r.parser = nParser
	assert.NoError(t, err)

	err = r.Gather(&acc)
	assert.Equal(t, 2, len(acc.Metrics))
}

func getPluginDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "/reader/reader_test.go", "", 1)
}

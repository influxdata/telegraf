package reader

import (
	"log"
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
		Filepaths: []string{testDir + "/logparser/grok/testdata/**.log"},
	}

	r.refreshFilePaths()
	//log.Printf("filenames: %v", filenames)
	assert.Equal(t, len(r.Filenames), 2)
}
func TestJSONParserCompile(t *testing.T) {
	testDir := getPluginDir()
	var acc testutil.Accumulator
	r := Reader{
		Filepaths: []string{testDir + "/reader/testfiles/json_a.log"},
	}
	parserConfig := parsers.Config{
		DataFormat: "json",
		TagKeys:    []string{"parent_ignored_child"},
	}
	nParser, err := parsers.NewParser(&parserConfig)
	r.parser = nParser
	assert.NoError(t, err)

	r.Gather(&acc)
	log.Printf("acc: %v", acc.Metrics[0].Tags)
	assert.Equal(t, map[string]string{"parent_ignored_child": "hi"}, acc.Metrics[0].Tags)
	assert.Equal(t, 5, len(acc.Metrics[0].Fields))
}

func TestGrokParser(t *testing.T) {
	testDir := getPluginDir()
	var acc testutil.Accumulator
	r := Reader{
		Filepaths: []string{testDir + "/reader/testfiles/grok_a.log"},
	}

	parserConfig := parsers.Config{
		DataFormat: "grok",
		Patterns:   []string{"%{COMMON_LOG_FORMAT}"},
	}

	nParser, err := parsers.NewParser(&parserConfig)
	r.parser = nParser
	assert.NoError(t, err)

	log.Printf("path: %v", r.Filepaths[0])
	err = r.Gather(&acc)
	log.Printf("err: %v", err)
	log.Printf("metric[0]_tags: %v, metric[0]_fields: %v", acc.Metrics[0].Tags, acc.Metrics[0].Fields)
	log.Printf("metric[1]_tags: %v, metric[1]_fields: %v", acc.Metrics[1].Tags, acc.Metrics[1].Fields)
	assert.Equal(t, 2, len(acc.Metrics))
}

func getPluginDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "/reader/reader_test.go", "", 1)
}

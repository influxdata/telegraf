package test

import (
	"testing"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

// All test taken from file_test.go I tried to keep only the ones I thought would be useful

func TestJSONParserCompile(t *testing.T) {
	var acc testutil.Accumulator
	r := Test{
		Metrics: []string{`{
			"parent": {
				"child": 3.0,
				"ignored_child": "hi"
			},
			"ignored_null": null,
			"integer": 4,
			"list": [3, 4],
			"ignored_parent": {
				"another_ignored_null": null,
				"ignored_string": "hello, world!"
			},
			"another_list": [4]
		}`},
	}
	err := r.Init()
	require.NoError(t, err)
	parserConfig := parsers.Config{
		DataFormat: "json",
		TagKeys:    []string{"parent_ignored_child"},
	}
	nParser, err := parsers.NewParser(&parserConfig)
	require.NoError(t, err)
	r.parser = nParser

	require.NoError(t, r.Gather(&acc))
	require.Equal(t, map[string]string{"parent_ignored_child": "hi"}, acc.Metrics[0].Tags)
	require.Equal(t, 5, len(acc.Metrics[0].Fields))
}

func TestGrokParser(t *testing.T) {
	r := Test{
		Metrics: []string{
			`127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`,
			`128.0.0.1 user-identifier tony [10/Oct/2000:13:55:36 -0800] "GET /apache_pb.gif HTTP/1.0" 300 45`,
		},
	}
	require.NoError(t, r.Init())

	parserConfig := parsers.Config{
		DataFormat:   "grok",
		GrokPatterns: []string{"%{COMMON_LOG_FORMAT}"},
	}

	nParser, err := parsers.NewParser(&parserConfig)
	r.parser = nParser
	require.NoError(t, err)

	var acc testutil.Accumulator
	require.NoError(t, r.Gather(&acc))

	require.Equal(t, 2, len(acc.Metrics))
	require.Equal(t, 6, len(acc.Metrics[0].Fields))
}


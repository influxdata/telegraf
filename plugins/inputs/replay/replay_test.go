package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplay(t *testing.T) {
	var acc testutil.Accumulator
	wd, err := os.Getwd()
	require.NoError(t, err)
	r := Replay{
		Files:      []string{filepath.Join(wd, "testdata/test.ilp")},
		Iterations: 2,
	}

	parserConfig := parsers.Config{
		DataFormat: "influx",
	}

	parser, err := parsers.NewParser(&parserConfig)
	assert.NoError(t, err)
	r.SetParser(parser)

	err = r.Start(&acc)
	require.NoError(t, err)
	defer r.Stop()
}

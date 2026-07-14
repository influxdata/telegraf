package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/internal"
)

func TestPrintSampleConfigDefaultOutputIsFile(t *testing.T) {
	var buf bytes.Buffer
	printSampleConfig(&buf, Filters{})

	output := buf.String()
	require.Contains(t, output, "\n[[outputs.file]]")
	require.Contains(t, output, "\n  files = [\"stdout\"]")
}

func TestLoadConfigurationTestModeSkipsDiskOutputBuffer(t *testing.T) {
	savedVersion := internal.Version
	internal.Version = "0.0.0"
	defer func() {
		internal.Version = savedVersion
	}()

	agent := &Telegraf{
		GlobalFlags: GlobalFlags{
			config: []string{"testdata/test_mode_disk_buffer.conf"},
		},
	}
	_, err := agent.loadConfiguration()
	require.ErrorContains(t, err, "creating buffer failed")

	agent.test = true
	cfg, err := agent.loadConfiguration()
	require.NoError(t, err)
	require.Len(t, cfg.Outputs, 1)
	require.Equal(t, "discard", cfg.Outputs[0].Config.BufferStrategy)
}

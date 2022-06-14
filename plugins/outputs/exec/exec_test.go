package exec

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

func TestExec(t *testing.T) {
	t.Skip("Skipping test due to OS/executable dependencies and race condition when ran as part of a test-all")

	tests := []struct {
		name    string
		command []string
		err     bool
		metrics []telegraf.Metric
	}{
		{
			name:    "test success",
			command: []string{"tee"},
			err:     false,
			metrics: testutil.MockMetrics(),
		},
		{
			name:    "test doesn't accept stdin",
			command: []string{"sleep", "5s"},
			err:     true,
			metrics: testutil.MockMetrics(),
		},
		{
			name:    "test command not found",
			command: []string{"/no/exist", "-h"},
			err:     true,
			metrics: testutil.MockMetrics(),
		},
		{
			name:    "test no metrics output",
			command: []string{"tee"},
			err:     false,
			metrics: []telegraf.Metric{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Exec{
				Command: tt.command,
				Timeout: config.Duration(time.Second),
				runner:  &CommandRunner{},
			}

			s, _ := serializers.NewInfluxSerializer()
			e.SetSerializer(s)

			require.NoError(t, e.Connect())
			require.Equal(t, tt.err, e.Write(tt.metrics) != nil)
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name string
		buf  *bytes.Buffer
		len  int
	}{
		{
			name: "long out",
			buf:  bytes.NewBufferString(strings.Repeat("a", maxStderrBytes+100)),
			len:  maxStderrBytes + len("..."),
		},
		{
			name: "multiline out",
			buf:  bytes.NewBufferString("hola\ngato\n"),
			len:  len("hola") + len("..."),
		},
	}
	c := CommandRunner{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := c.truncate(*tt.buf)
			require.Equal(t, tt.len, len(s))
		})
	}
}

func TestExecDocs(t *testing.T) {
	e := &Exec{}
	e.SampleConfig()
	require.NoError(t, e.Close())

	e = &Exec{runner: &CommandRunner{}}
	require.NoError(t, e.Close())
}

package exec

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

func TestExec(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test due to OS/executable dependencies")
	}

	tests := []struct {
		command []string
		err     bool
		metrics []telegraf.Metric
	}{
		{
			command: []string{"tee"},
			err:     false,
			metrics: testutil.MockMetrics(),
		},
		{
			command: []string{"sleep", "5s"},
			err:     true,
			metrics: testutil.MockMetrics(),
		},
		{
			command: []string{"/no/exist", "-h"},
			err:     true,
			metrics: testutil.MockMetrics(),
		},
		{
			command: []string{"tee"},
			err:     false,
			metrics: []telegraf.Metric{},
		},
	}

	for _, tt := range tests {
		e := &Exec{
			Command: tt.command,
			Timeout: internal.Duration{Duration: time.Second},
			runner:  &CommandRunner{},
		}

		s, _ := serializers.NewInfluxSerializer()
		e.SetSerializer(s)

		e.Connect()

		require.Equal(t, tt.err, e.Write(tt.metrics) != nil)
	}
}

func TestExecDocs(t *testing.T) {
	e := &Exec{}
	e.Description()
	e.SampleConfig()
	require.NoError(t, e.Close())

	e = &Exec{runner: &CommandRunner{}}
	require.NoError(t, e.Close())
}

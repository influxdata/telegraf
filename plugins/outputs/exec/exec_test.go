package exec

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

func TestExec(t *testing.T) {
	tests := []struct {
		command []string
		err     bool
	}{
		{
			command: []string{"tee"},
			err:     false,
		},
		{
			command: []string{"sleep", "5s"},
			err:     true,
		},
		{
			command: []string{"/no/exist", "-h"},
			err:     true,
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

		require.Equal(t, tt.err, e.Write(testutil.MockMetrics()) != nil)
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

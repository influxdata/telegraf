package chess

import (
	"bytes"
	"io"
	"log"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func Test_Chess_Leaderboards_Simple(t *testing.T) {
	chess := &Chess{
		Profiles: nil,
		Log:      testutil.Logger{Name: "chess"},
	}
	b := bytes.NewBufferString("")
	log.SetOutput(b)

	var acc testutil.Accumulator
	require.NoError(t, chess.Gather(&acc))
	out, err := io.ReadAll(b)
	require.NoError(t, err)
	require.Empty(t, string(out))

	metric, hasMeas := acc.Get("leaderboards")
	require.True(t, hasMeas)
	require.NotNil(t, metric)

	fields := metric.Fields
	require.NotEmpty(t, fields)
	require.NotEmpty(t, fields["username"])
	require.NotEmpty(t, fields["rank"])
	require.NotEmpty(t, fields["score"])

	tags := metric.Tags
	require.NotEmpty(t, tags)
	require.NotEmpty(t, tags["playerId"])
}

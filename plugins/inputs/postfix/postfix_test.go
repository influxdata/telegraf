//go:build !windows
// +build !windows

package postfix

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestGather(t *testing.T) {
	td := t.TempDir()

	for _, q := range []string{"active", "hold", "incoming", "maildrop", "deferred/0/0", "deferred/F/F"} {
		require.NoError(t, os.MkdirAll(filepath.FromSlash(td+"/"+q), 0755))
	}

	require.NoError(t, os.WriteFile(filepath.FromSlash(td+"/active/01"), []byte("abc"), 0644))
	require.NoError(t, os.WriteFile(filepath.FromSlash(td+"/active/02"), []byte("defg"), 0644))
	require.NoError(t, os.WriteFile(filepath.FromSlash(td+"/hold/01"), []byte("abc"), 0644))
	require.NoError(t, os.WriteFile(filepath.FromSlash(td+"/incoming/01"), []byte("abcd"), 0644))
	require.NoError(t, os.WriteFile(filepath.FromSlash(td+"/deferred/0/0/01"), []byte("abc"), 0644))
	require.NoError(t, os.WriteFile(filepath.FromSlash(td+"/deferred/F/F/F1"), []byte("abc"), 0644))

	p := Postfix{
		QueueDirectory: td,
	}

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))

	metrics := map[string]*testutil.Metric{}
	for _, m := range acc.Metrics {
		metrics[m.Tags["queue"]] = m
	}

	require.Equal(t, int64(2), metrics["active"].Fields["length"])
	require.Equal(t, int64(7), metrics["active"].Fields["size"])
	require.InDelta(t, 0, metrics["active"].Fields["age"], 10)

	require.Equal(t, int64(1), metrics["hold"].Fields["length"])
	require.Equal(t, int64(3), metrics["hold"].Fields["size"])

	require.Equal(t, int64(1), metrics["incoming"].Fields["length"])
	require.Equal(t, int64(4), metrics["incoming"].Fields["size"])

	require.Equal(t, int64(0), metrics["maildrop"].Fields["length"])
	require.Equal(t, int64(0), metrics["maildrop"].Fields["size"])
	require.Equal(t, int64(0), metrics["maildrop"].Fields["age"])

	require.Equal(t, int64(2), metrics["deferred"].Fields["length"])
	require.Equal(t, int64(6), metrics["deferred"].Fields["size"])
}

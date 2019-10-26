package postfix

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGather(t *testing.T) {
	td, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(td)

	for _, q := range []string{"active", "hold", "incoming", "maildrop", "deferred/0/0", "deferred/F/F"} {
		require.NoError(t, os.MkdirAll(filepath.FromSlash(td+"/"+q), 0755))
	}

	require.NoError(t, ioutil.WriteFile(filepath.FromSlash(td+"/active/01"), []byte("abc"), 0644))
	require.NoError(t, ioutil.WriteFile(filepath.FromSlash(td+"/active/02"), []byte("defg"), 0644))
	require.NoError(t, ioutil.WriteFile(filepath.FromSlash(td+"/hold/01"), []byte("abc"), 0644))
	require.NoError(t, ioutil.WriteFile(filepath.FromSlash(td+"/incoming/01"), []byte("abcd"), 0644))
	require.NoError(t, ioutil.WriteFile(filepath.FromSlash(td+"/deferred/0/0/01"), []byte("abc"), 0644))
	require.NoError(t, ioutil.WriteFile(filepath.FromSlash(td+"/deferred/F/F/F1"), []byte("abc"), 0644))

	p := Postfix{
		QueueDirectory: td,
	}

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))

	metrics := map[string]*testutil.Metric{}
	for _, m := range acc.Metrics {
		metrics[m.Tags["queue"]] = m
	}

	assert.Equal(t, int64(2), metrics["active"].Fields["length"])
	assert.Equal(t, int64(7), metrics["active"].Fields["size"])
	assert.InDelta(t, 0, metrics["active"].Fields["age"], 10)

	assert.Equal(t, int64(1), metrics["hold"].Fields["length"])
	assert.Equal(t, int64(3), metrics["hold"].Fields["size"])

	assert.Equal(t, int64(1), metrics["incoming"].Fields["length"])
	assert.Equal(t, int64(4), metrics["incoming"].Fields["size"])

	assert.Equal(t, int64(0), metrics["maildrop"].Fields["length"])
	assert.Equal(t, int64(0), metrics["maildrop"].Fields["size"])
	assert.Equal(t, int64(0), metrics["maildrop"].Fields["age"])

	assert.Equal(t, int64(2), metrics["deferred"].Fields["length"])
	assert.Equal(t, int64(6), metrics["deferred"].Fields["size"])
}

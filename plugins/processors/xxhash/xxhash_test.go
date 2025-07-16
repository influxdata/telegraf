package xxhash

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestXXHash_ExactMatch(t *testing.T) {
	x := &XXHash{
		Keys:      []string{"host", "cpu_usage"},
		KeysMode:  "exact",
		TagHash:   "h",
		FieldHash: "h64",
	}
	require.NoError(t, x.Init())

	m := testutil.MustMetric("cpu",
		map[string]string{
			"host": "node01",
		},
		map[string]interface{}{
			"cpu_usage": 42.5,
			"ignored":   123,
		},
		time.Unix(0, 0),
	)

	x.Apply(m)

	tag := m.Tags()["h"]
	field, ok := m.Fields()["h64"]
	require.True(t, ok)
	require.NotEmpty(t, tag)
	require.IsType(t, int64(0), field)
}

func TestXXHash_RegexMatch(t *testing.T) {
	x := &XXHash{
		Keys:      []string{"^cpu.*", "^host$"},
		KeysMode:  "regex",
		TagHash:   "hash",
		FieldHash: "hash64",
	}
	require.NoError(t, x.Init())

	m := testutil.MustMetric("cpu",
		map[string]string{
			"host": "srv",
			"zone": "z1",
		},
		map[string]interface{}{
			"cpu_user": 1.23,
			"cpu_sys":  2.34,
			"disk":     88,
		},
		time.Now(),
	)

	x.Apply(m)
	require.NotEmpty(t, m.Tags()["hash"])
	require.IsType(t, int64(0), m.Fields()["hash64"])
}

func TestXXHash_EmptyMatch(t *testing.T) {
	x := &XXHash{
		Keys:      []string{"nonexistent"},
		KeysMode:  "exact",
		TagHash:   "h",
		FieldHash: "h64",
	}
	require.NoError(t, x.Init())

	m := testutil.MustMetric("cpu",
		map[string]string{"env": "prod"},
		map[string]interface{}{"load": 0.99},
		time.Now(),
	)

	x.Apply(m)

	_, tagExists := m.Tags()["h"]
	_, fieldExists := m.Fields()["h64"]
	require.False(t, tagExists)
	require.False(t, fieldExists)
}

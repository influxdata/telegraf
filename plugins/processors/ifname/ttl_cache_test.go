package ifname

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTTLCacheExpire(t *testing.T) {
	c := newTTLCache(1*time.Second, 100)

	c.now = func() time.Time {
		return time.Unix(0, 0)
	}

	c.put("ones", nameMap{1: "one"})
	require.Len(t, c.lru.m, 1)

	c.now = func() time.Time {
		return time.Unix(1, 0)
	}

	_, ok, _ := c.get("ones")
	require.False(t, ok)
	require.Empty(t, c.lru.m)
	require.Equal(t, 0, c.lru.l.Len())
}

func TestTTLCache(t *testing.T) {
	c := newTTLCache(1*time.Second, 100)

	c.now = func() time.Time {
		return time.Unix(0, 0)
	}

	expected := nameMap{1: "one"}
	c.put("ones", expected)

	actual, ok, _ := c.get("ones")
	require.True(t, ok)
	require.Equal(t, expected, actual)
}

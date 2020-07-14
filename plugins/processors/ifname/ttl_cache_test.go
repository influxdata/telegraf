package ifname

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTTLCacheExpire(t *testing.T) {
	c := NewTTLCache(1*time.Second, 100)

	c.now = func() time.Time {
		return time.Unix(0, 0)
	}

	c.Put("ones", nameMap{1: "one"})
	require.Len(t, c.lru.m, 1)

	c.now = func() time.Time {
		return time.Unix(1, 0)
	}

	_, ok, _ := c.Get("ones")
	require.False(t, ok)
	require.Len(t, c.lru.m, 0)
	require.Equal(t, c.lru.l.Len(), 0)
}

func TestTTLCache(t *testing.T) {
	c := NewTTLCache(1*time.Second, 100)

	c.now = func() time.Time {
		return time.Unix(0, 0)
	}

	expected := nameMap{1: "one"}
	c.Put("ones", expected)

	actual, ok, _ := c.Get("ones")
	require.True(t, ok)
	require.Equal(t, expected, actual)
}

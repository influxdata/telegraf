package ifname

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	c := NewLRUCache(2)

	c.Put("ones", LRUValType{val: nameMap{1: "one"}})
	twoMap := LRUValType{val: nameMap{2: "two"}}
	c.Put("twos", twoMap)
	c.Put("threes", LRUValType{val: nameMap{3: "three"}})

	_, ok := c.Get("ones")
	require.False(t, ok)

	v, ok := c.Get("twos")
	require.True(t, ok)
	require.Equal(t, twoMap, v)
}

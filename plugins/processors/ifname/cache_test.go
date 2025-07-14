package ifname

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	c := newLRUCache(2)

	c.put("ones", lruValType{val: nameMap{1: "one"}})
	twoMap := lruValType{val: nameMap{2: "two"}}
	c.put("twos", twoMap)
	c.put("threes", lruValType{val: nameMap{3: "three"}})

	_, ok := c.get("ones")
	require.False(t, ok)

	v, ok := c.get("twos")
	require.True(t, ok)
	require.Equal(t, twoMap, v)
}

package kafka

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBackoffFunc(t *testing.T) {
	b := 250 * time.Millisecond
	max := 1100 * time.Millisecond

	f := makeBackoffFunc(b, max)
	require.Equal(t, b, f(0, 0))
	require.Equal(t, b*2, f(1, 0))
	require.Equal(t, b*4, f(2, 0))
	require.Equal(t, max, f(3, 0)) // would be 2000 but that's greater than max

	f = makeBackoffFunc(b, 0)      // max = 0 means no max
	require.Equal(t, b*8, f(3, 0)) // with no max, it's 2000
}

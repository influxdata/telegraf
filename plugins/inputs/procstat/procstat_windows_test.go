package procstat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGather_RealPattern(t *testing.T) {
	pg, err := NewPgrep()
	require.NoError(t, err)
	pids, err := pg.Pattern(`procstat`)
	require.NoError(t, err)
	assert.Equal(t, len(pids) > 0, true)
}

func TestGather_RealFullPattern(t *testing.T) {
	pg, err := NewPgrep()
	require.NoError(t, err)
	pids, err := pg.FullPattern(`procstat`)
	require.NoError(t, err)
	assert.Equal(t, len(pids) > 0, true)
}

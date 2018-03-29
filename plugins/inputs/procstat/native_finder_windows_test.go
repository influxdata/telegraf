package procstat

import (
	"fmt"
	"testing"

	"os/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGather_RealPattern(t *testing.T) {
	pg, err := NewNativeFinder()
	require.NoError(t, err)
	pids, err := pg.Pattern(`procstat`)
	require.NoError(t, err)
	fmt.Println(pids)
	assert.Equal(t, len(pids) > 0, true)
}

func TestGather_RealFullPattern(t *testing.T) {
	pg, err := NewNativeFinder()
	require.NoError(t, err)
	pids, err := pg.FullPattern(`%procstat%`)
	require.NoError(t, err)
	fmt.Println(pids)
	assert.Equal(t, len(pids) > 0, true)
}

func TestGather_RealUser(t *testing.T) {
	user, err := user.Current()
	require.NoError(t, err)
	pg, err := NewNativeFinder()
	require.NoError(t, err)
	pids, err := pg.Uid(user.Username)
	require.NoError(t, err)
	fmt.Println(pids)
	assert.Equal(t, len(pids) > 0, true)
}

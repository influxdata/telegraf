package procstat

import (
	"os/user"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGather_RealPatternIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	pg, err := NewNativeFinder()
	require.NoError(t, err)
	pids, err := pg.Pattern(`procstat`)
	require.NoError(t, err)
	require.NotEmpty(t, pids)
}

func TestGather_RealFullPatternIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	pg, err := NewNativeFinder()
	require.NoError(t, err)

	pids, err := pg.FullPattern(`%procstat%`)
	require.NoError(t, err)
	require.NotEmpty(t, pids)
}

func TestGather_RealUserIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	currentUser, err := user.Current()
	require.NoError(t, err)
	pg, err := NewNativeFinder()
	require.NoError(t, err)
	pids, err := pg.UID(currentUser.Username)
	require.NoError(t, err)
	require.NotEmpty(t, pids)
}

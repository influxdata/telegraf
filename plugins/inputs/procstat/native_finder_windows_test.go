package procstat

import (
	"fmt"
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
	fmt.Println(pids)
	require.Equal(t, len(pids) > 0, true)
}

func TestGather_RealFullPatternIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	pg, err := NewNativeFinder()
	require.NoError(t, err)
	pids, err := pg.FullPattern(`%procstat%`)
	require.NoError(t, err)
	fmt.Println(pids)
	require.Equal(t, len(pids) > 0, true)
}

func TestGather_RealUserIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	user, err := user.Current()
	require.NoError(t, err)
	pg, err := NewNativeFinder()
	require.NoError(t, err)
	pids, err := pg.UID(user.Username)
	require.NoError(t, err)
	fmt.Println(pids)
	require.Equal(t, len(pids) > 0, true)
}

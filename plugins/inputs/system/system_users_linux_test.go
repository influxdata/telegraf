//go:build linux

package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// setupUsers configures gopsutil to read from an empty synthetic utmp file
// so that host.Users() returns zero users deterministically. Returns true
// to indicate the call is mocked and always available.
func setupUsers(t *testing.T) bool {
	t.Helper()
	tmpDir := t.TempDir()
	runDir := filepath.Join(tmpDir, "run")
	require.NoError(t, os.MkdirAll(runDir, 0750))
	require.NoError(t, os.WriteFile(filepath.Join(runDir, "utmp"), nil, 0640))
	t.Setenv("HOST_VAR", tmpDir)
	return true
}

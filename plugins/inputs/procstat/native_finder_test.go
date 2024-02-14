package procstat

import (
	"context"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkPattern(b *testing.B) {
	finder := &NativeFinder{}
	err := finder.Init()
	require.NoError(b, err)
	for n := 0; n < b.N; n++ {
		err := finder.Pattern(".*")
		require.NoError(b, err)
	}
}

func BenchmarkFullPattern(b *testing.B) {
	finder := &NativeFinder{}
	err := finder.Init()
	require.NoError(b, err)
	for n := 0; n < b.N; n++ {
		err := finder.FullPattern(".*")
		require.NoError(b, err)
	}
}

func TestChildPattern(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("Skipping test on unsupported platform")
	}

	// Get our own process name
	parentName, err := os.Executable()
	require.NoError(t, err)

	// Spawn two child processes and get their PIDs
	expected := make([]PID, 0, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// First process
	cmd1 := exec.CommandContext(ctx, "/bin/sh")
	require.NoError(t, cmd1.Start(), "starting first command failed")
	expected = append(expected, PID(cmd1.Process.Pid))

	// Second process
	cmd2 := exec.CommandContext(ctx, "/bin/sh")
	require.NoError(t, cmd2.Start(), "starting first command failed")
	expected = append(expected, PID(cmd2.Process.Pid))

	// Use the plugin to find the children
	finder := &NativeFinder{}
	err = finder.Init()
	require.NoError(t, err)
	err = finder.Pattern(parentName)
	require.NoError(t, err)
	parent, err := finder.GetResult()
	require.NoError(t, err)
	require.Len(t, parent, 1)
	childs, err := finder.Children(parent[0])
	require.NoError(t, err)
	require.ElementsMatch(t, expected, childs)
}

func TestGather_RealPatternIntegration(t *testing.T) {
	var err error
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	finder := &NativeFinder{}
	err = finder.Init()
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		err = finder.Pattern(`procstat`)
	} else {
		err = finder.Pattern(`conhost.exe`)
	}
	require.NoError(t, err)
	pids, err := finder.GetResult()
	require.NoError(t, err)
	require.NotEmpty(t, pids)
}

func TestGather_RealFullPatternIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	if runtime.GOOS != "windows" {
		t.Skip("Skipping integration test on Non-Windows OS")
	}
	finder := &NativeFinder{}
	err := finder.Init()
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		err = finder.FullPattern(`procstat`)
	} else {
		err = finder.FullPattern(`conhost.exe`)
	}
	require.NoError(t, err)
	pids, err := finder.GetResult()
	require.NoError(t, err)
	require.NotEmpty(t, pids)
}

func TestGather_RealUserIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	currentUser, err := user.Current()
	require.NoError(t, err)

	finder := &NativeFinder{}
	err = finder.Init()
	require.NoError(t, err)
	err = finder.UID(currentUser.Username)
	require.NoError(t, err)
	pids, err := finder.GetResult()
	require.NoError(t, err)
	require.NotEmpty(t, pids)
}

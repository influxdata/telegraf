package procstat

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkPattern(b *testing.B) {
	finder := &NativeFinder{}
	for n := 0; n < b.N; n++ {
		_, err := finder.Pattern(".*")
		require.NoError(b, err)
	}
}

func BenchmarkFullPattern(b *testing.B) {
	finder := &NativeFinder{}
	for n := 0; n < b.N; n++ {
		_, err := finder.FullPattern(".*")
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
	childs, err := finder.ChildPattern(parentName)
	require.NoError(t, err)
	require.ElementsMatch(t, expected, childs)
}

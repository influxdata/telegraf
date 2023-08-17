package procstat

import (
	"fmt"
	"os/exec"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkPattern(b *testing.B) {
	f, err := NewNativeFinder()
	require.NoError(b, err)
	for n := 0; n < b.N; n++ {
		_, err := f.Pattern(".*")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkFullPattern(b *testing.B) {
	f, err := NewNativeFinder()
	require.NoError(b, err)
	for n := 0; n < b.N; n++ {
		_, err := f.FullPattern(".*")
		if err != nil {
			panic(err)
		}
	}
}

func TestChildPattern(t *testing.T) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		cmd := exec.Command("/bin/bash", "-c", "sleep 10")
		if err := cmd.Start(); err != nil {
			fmt.Printf("Error starting command: %s\n", err)
			return
		}

		f, err := NewNativeFinder()
		require.NoError(t, err)

		childpids, err := f.ChildPattern("TestChildPattern")
		for _, p := range childpids {
			t.Log(string(p))
		}

		require.Equal(t, []PID{PID(cmd.Process.Pid)}, childpids)
		cmd.Process.Kill()
		if err != nil {
			panic(err)
		}

		var nilpids []PID
		childpids, err = f.ChildPattern("TestChildPattern")
		for _, p := range childpids {
			t.Log(string(p))
		}

		require.Equal(t, nilpids, childpids)
		if err != nil {
			panic(err)
		}
	}
}

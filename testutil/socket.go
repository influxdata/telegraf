package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TempSocket(tb testing.TB) string {
	// On MacOS, the maximum path length of Unix domain socket is 104
	// characters. (https://unix.stackexchange.com/a/367012/376279)
	//
	// On MacOS, tb.TempDir() returns e.g.
	// /var/folders/bl/wbxjgtzx7j5_mjsmfr3ynlc00000gp/T/<the-test-name>/001/socket.sock
	//
	// If the name of the test is long, the path length could exceed 104
	// characters, and this would result in listen unix ...: bind: invalid argument
	if runtime.GOOS == "darwin" {
		sock := filepath.Join("/tmp", "sock")

		tb.Cleanup(func() {
			require.NoError(tb, os.RemoveAll(sock))
		})

		return sock
	}

	return filepath.Join(tb.TempDir(), "sock")
}

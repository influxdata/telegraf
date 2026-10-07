package sudo

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBinary(t *testing.T) {
	if runtime.GOOS == "openbsd" {
		require.Equal(t, "doas", Binary())
		return
	}
	require.Equal(t, "sudo", Binary())
}

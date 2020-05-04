// +build !windows

package shim

import (
	"bytes"
	"context"
	"os"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestShimUSR1SignalingWorks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
		return
	}
	stdoutBytes := bytes.NewBufferString("")
	stdout = stdoutBytes

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wait := runInputPlugin(t, 40*time.Second)

	// sleep a bit to avoid a race condition where the input hasn't loaded yet.
	time.Sleep(10 * time.Millisecond)

	// signal USR1 to yourself.
	pid := os.Getpid()
	process, err := os.FindProcess(pid)
	require.NoError(t, err)

	go func() {
		// On slow machines this signal can fire before the service comes up.
		// rather than depend on accurate sleep times, we'll just retry sending
		// the signal every so often until it goes through.
		for {
			select {
			case <-ctx.Done():
				return // test is done
			default:
				// test isn't done, keep going.
				process.Signal(syscall.SIGUSR1)
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	timeout := time.NewTimer(10 * time.Second)

	select {
	case <-wait:
	case <-timeout.C:
		require.Fail(t, "Timeout waiting for metric to arrive")
	}

	for stdoutBytes.Len() == 0 {
		select {
		case <-timeout.C:
			require.Fail(t, "Timeout waiting to read metric from stdout")
			return
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	out := string(stdoutBytes.Bytes())
	require.Contains(t, out, "\n")
	metricLine := strings.Split(out, "\n")[0]
	require.Equal(t, "measurement,tag=tag field=1i 1234000005678", metricLine)
}

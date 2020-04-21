// +build !windows

package shim

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShimUSR1SignalingWorks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
		return
	}
	stdoutBytes := bytes.NewBufferString("")
	stdout = stdoutBytes

	wait := runInputPlugin(40 * time.Second)

	// sleep a bit to avoid a race condition where the input hasn't loaded yet.
	time.Sleep(10 * time.Millisecond)

	// signal USR1 to yourself.
	pid := os.Getpid()
	process, err := os.FindProcess(pid)
	assert.NoError(t, err)
	err = process.Signal(syscall.SIGUSR1)
	assert.NoError(t, err)

	<-wait
	for stdoutBytes.Len() == 0 {
		time.Sleep(10 * time.Millisecond)
	}

	out := string(stdoutBytes.Bytes())
	if assert.Contains(t, out, "\n") {
		metricLine := strings.Split(out, "\n")[0]
		assert.Equal(t, "measurement,tag=tag field=1i 1234000005678", metricLine)
	}
}

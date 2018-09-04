package daemon // import "github.com/docker/docker/integration-cli/daemon"

import (
	"fmt"
	"strconv"

	"github.com/go-check/check"
	"golang.org/x/sys/windows"
)

// SignalDaemonDump sends a signal to the daemon to write a dump file
func SignalDaemonDump(pid int) {
	ev, _ := windows.UTF16PtrFromString("Global\\docker-daemon-" + strconv.Itoa(pid))
	h2, err := windows.OpenEvent(0x0002, false, ev)
	if h2 == 0 || err != nil {
		return
	}
	windows.PulseEvent(h2)
}

func signalDaemonReload(pid int) error {
	return fmt.Errorf("daemon reload not supported")
}

func cleanupExecRoot(c *check.C, execRoot string) {
}

// +build !windows

package agent

import (
	"os"
	"os/signal"
	"syscall"
)

const flushSignal = syscall.SIGUSR1

func watchForFlushSignal(flushRequested chan os.Signal) {
	signal.Notify(flushRequested, flushSignal)
}

func stopListeningForFlushSignal(flushRequested chan os.Signal) {
	defer signal.Stop(flushRequested)
}

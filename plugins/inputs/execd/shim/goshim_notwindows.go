// +build !windows

package shim

import (
	"os"
	"os/signal"
	"syscall"
)

func listenForCollectMetricsSignals(collectMetricsPrompt chan os.Signal) {
	// just listen to all the signals.
	signal.Notify(collectMetricsPrompt, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2)
}

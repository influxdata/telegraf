// +build windows

package shim

import (
	"os"
	"os/signal"
	"syscall"
)

func listenForCollectMetricsSignals(collectMetricsPrompt chan os.Signal) {
	signal.Notify(collectMetricsPrompt, syscall.SIGHUP)
}

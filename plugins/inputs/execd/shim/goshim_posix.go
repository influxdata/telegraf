//go:build !windows
// +build !windows

package shim

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func listenForCollectMetricsSignals(ctx context.Context, collectMetricsPrompt chan os.Signal) {
	// just listen to all the signals.
	signal.Notify(collectMetricsPrompt, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2)

	go func() {
		<-ctx.Done()
		signal.Stop(collectMetricsPrompt)
	}()
}

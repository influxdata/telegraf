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
		select {
		case <-ctx.Done():
			// context done. stop to signals to avoid pushing messages to a closed channel
			signal.Stop(collectMetricsPrompt)
		}
	}()
}

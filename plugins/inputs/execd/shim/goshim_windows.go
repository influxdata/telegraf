// +build windows

package shim

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func listenForCollectMetricsSignals(ctx context.Context, collectMetricsPrompt chan os.Signal) {
	signal.Notify(collectMetricsPrompt, syscall.SIGHUP)

	go func() {
		select {
		case <-ctx.Done():
			// context done. stop to signals to avoid pushing messages to a closed channel
			signal.Stop(collectMetricsPrompt)
		}
	}()
}

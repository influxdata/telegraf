//go:build windows
// +build windows

package agent

import "os"

func watchForFlushSignal(flushRequested chan os.Signal) {
	// not supported
}

func stopListeningForFlushSignal(flushRequested chan os.Signal) {
	// not supported
}

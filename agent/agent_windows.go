//go:build windows

package agent

import "os"

func watchForFlushSignal(_ chan os.Signal) {
	// not supported
}

func stopListeningForFlushSignal(_ chan os.Signal) {
	// not supported
}

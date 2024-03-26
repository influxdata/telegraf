package internal

import "errors"

var ErrNotConnected = errors.New("not connected")

// StartupError indicates an error that occurred during startup of a plugin
// e.g. due to connectivity issues or resources being not yet available.
// In case the 'Retry' flag is set, the startup of the plugin might be retried
// depending on the configured startup-error-behavior. The 'RemovePlugin'
// flag denotes if the agent should remove the plugin from further processing.
type StartupError struct {
	Err     error
	Retry   bool
	Partial bool
}

func (e *StartupError) Error() string {
	return e.Err.Error()
}

func (e *StartupError) Unwrap() error {
	return e.Err
}

// FatalError indicates a not-recoverable error in the plugin. The corresponding
// plugin should be remove by the agent stopping any further processing for that
// plugin instance.
type FatalError struct {
	Err error
}

func (e *FatalError) Error() string {
	return e.Err.Error()
}

func (e *FatalError) Unwrap() error {
	return e.Err
}

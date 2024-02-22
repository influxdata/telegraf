package telegraf

import "errors"

var ErrNotConnected = errors.New("not connected")

// StartupError indicates an error that occurred during startup of a plugin
// e.g. due to connectivity issues or resources being not yet available.
// In case the 'Retry' flag is set, the startup of the plugin might be retried
// depending on the configured startup-error-behavior. The 'RemovePlugin'
// flag denotes if the agent should remove the plugin from further processing.
type StartupError struct {
	Err          error
	Retry        bool
	RemovePlugin bool
}

func (e *StartupError) Error() string {
	return e.Err.Error()
}

func (e *StartupError) Unwrap() error {
	return e.Err
}

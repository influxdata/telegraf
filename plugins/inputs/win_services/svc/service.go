package svc

// State describes service execution state (Stopped, Running and so on).
type State uint32

// Accepted is used to describe commands accepted by the service.
// Note that Interrogate is always accepted.
type Accepted uint32

// Status combines State and Accepted commands to fully describe running service.
type Status struct {
	State                   State
	Accepts                 Accepted
	CheckPoint              uint32 // used to report progress during a lengthy operation
	WaitHint                uint32 // estimated time required for a pending operation, in milliseconds
	ProcessId               uint32 //nolint:revive // if the service is running, the process identifier of it, and otherwise zero
	Win32ExitCode           uint32 // set if the service has exited with a win32 exit code
	ServiceSpecificExitCode uint32 // set if the service has exited with a service-specific exit code
}

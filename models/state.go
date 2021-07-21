package models

import "sync/atomic"

// PluginState describes what the instantiated plugin is currently doing
// needs to stay int32 for use with atomic
type PluginState int32

const (
	PluginStateCreated PluginState = iota
	PluginStateStarting
	PluginStateRunning
	PluginStateStopping
	PluginStateDead
)

func (p PluginState) String() string {
	switch p {
	case PluginStateCreated:
		return "created"
	case PluginStateStarting:
		return "starting"
	case PluginStateRunning:
		return "running"
	case PluginStateStopping:
		return "stopping"
	default:
		return "dead"
	}
}

type State struct {
	state PluginState
}

func (s *State) setState(newState PluginState) {
	atomic.StoreInt32((*int32)(&s.state), int32(newState))
}

func (s *State) GetState() PluginState {
	return PluginState(atomic.LoadInt32((*int32)(&s.state)))
}

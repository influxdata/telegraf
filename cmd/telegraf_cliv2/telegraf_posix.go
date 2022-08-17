//go:build !windows
// +build !windows

package main

func (a *AgentManager) Run() error {
	stop = make(chan struct{})
	return a.reloadLoop()
}

//go:build !windows
// +build !windows

package main

import "github.com/urfave/cli/v2"

func (a *AgentManager) Run() error {
	stop = make(chan struct{})
	return a.reloadLoop()
}

func cliFlags() []cli.Flag {
	return []cli.Flag{}
}

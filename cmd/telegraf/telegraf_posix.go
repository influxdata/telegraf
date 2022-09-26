//go:build !windows

package main

import "github.com/urfave/cli/v2"

func (t *Telegraf) Run() error {
	stop = make(chan struct{})
	return t.reloadLoop()
}

func cliFlags() []cli.Flag {
	return []cli.Flag{}
}

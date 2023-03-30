//go:build !windows

package main

import (
	"fmt"
	"syscall"

	"github.com/urfave/cli/v2"
)

func (t *Telegraf) Run() error {
	stop = make(chan struct{})
	return t.reloadLoop()
}

func cliFlags() []cli.Flag {
	return []cli.Flag{}
}

func getLockedMemoryLimit() uint64 {
	const RLIMIT_MEMLOCK = 8

	var limit syscall.Rlimit
	if err := syscall.Getrlimit(RLIMIT_MEMLOCK, &limit); err != nil {
		panic(fmt.Errorf("Cannot get limit for locked memory: %w", err))
	}
	return limit.Max
}

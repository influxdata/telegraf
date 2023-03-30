//go:build !windows

package main

import (
	"log"
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
	// From https://elixir.bootlin.com/linux/latest/source/include/uapi/asm-generic/resource.h#L35
	const rlimit_memlock = 8

	var limit syscall.Rlimit
	if err := syscall.Getrlimit(rlimit_memlock, &limit); err != nil {
		panic(fmt.Errorf("Cannot get limit for locked memory: %w", err))
	}
	return limit.Max
}

//go:build !windows

package main

import (
	"log"
	"syscall"
)

func (t *Telegraf) Run() error {
	stop = make(chan struct{})
	defer close(stop)

	cfg, err := t.loadConfiguration()
	if err != nil {
		return err
	}
	t.cfg = cfg
	return t.reloadLoop()
}

func getLockedMemoryLimit() uint64 {
	// From https://elixir.bootlin.com/linux/latest/source/include/uapi/asm-generic/resource.h#L35
	const rLimitMemlock = 8

	var limit syscall.Rlimit
	if err := syscall.Getrlimit(rLimitMemlock, &limit); err != nil {
		log.Printf("E! Cannot get limit for locked memory: %v", err)
		return 0
	}
	//nolint:unconvert // required for e.g. FreeBSD that has the field as int64
	return uint64(limit.Max)
}

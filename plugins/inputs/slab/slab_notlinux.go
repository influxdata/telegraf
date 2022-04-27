//go:build !linux
// +build !linux

package slab

import "github.com/influxdata/telegraf"

func (ss *SlabStats) Init() error {
	ss.Log.Warn("Current platform is not supported")
	return nil
}

func (ss *SlabStats) Gather(acc telegraf.Accumulator) error {
	return nil
}

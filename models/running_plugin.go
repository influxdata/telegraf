package models

import (
	"strconv"

	"github.com/influxdata/telegraf"
)

// PluginID is the random id assigned to the plugin so it can be referenced later
type PluginID string

func (id PluginID) Uint64() uint64 {
	result, _ := strconv.ParseUint(string(id), 16, 64)
	return result
}

type RunningPlugin interface {
	Init() error
	GetID() uint64
	GetState() PluginState
}

// ProcessorRunner is an interface common to processors and aggregators so that aggregators can act like processors, including being ordered with and between processors
type ProcessorRunner interface {
	RunningPlugin
	Start(acc telegraf.Accumulator) error
	Add(m telegraf.Metric, acc telegraf.Accumulator) error
	Stop()
	Order() int64
	LogName() string
}

// ProcessorRunners add sorting
type ProcessorRunners []ProcessorRunner

func (rp ProcessorRunners) Len() int           { return len(rp) }
func (rp ProcessorRunners) Swap(i, j int)      { rp[i], rp[j] = rp[j], rp[i] }
func (rp ProcessorRunners) Less(i, j int) bool { return rp[i].Order() < rp[j].Order() }

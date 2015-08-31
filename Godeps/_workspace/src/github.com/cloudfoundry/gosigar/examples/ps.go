// Copyright (c) 2012 VMware, Inc.

package main

import (
	"fmt"
	"github.com/cloudfoundry/gosigar"
)

func main() {
	pids := sigar.ProcList{}
	pids.Get()

	// ps -eo pid,ppid,stime,time,rss,state,comm
	fmt.Print("  PID  PPID STIME     TIME    RSS S COMMAND\n")

	for _, pid := range pids.List {
		state := sigar.ProcState{}
		mem := sigar.ProcMem{}
		time := sigar.ProcTime{}

		if err := state.Get(pid); err != nil {
			continue
		}
		if err := mem.Get(pid); err != nil {
			continue
		}
		if err := time.Get(pid); err != nil {
			continue
		}

		fmt.Printf("%5d %5d %s %s %6d %c %s\n",
			pid, state.Ppid,
			time.FormatStartTime(), time.FormatTotal(),
			mem.Resident/1024, state.State, state.Name)
	}
}

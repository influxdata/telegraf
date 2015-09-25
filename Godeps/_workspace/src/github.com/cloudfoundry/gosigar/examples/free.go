// Copyright (c) 2012 VMware, Inc.

package main

import (
	"fmt"
	"github.com/cloudfoundry/gosigar"
	"os"
)

func format(val uint64) uint64 {
	return val / 1024
}

func main() {
	mem := sigar.Mem{}
	swap := sigar.Swap{}

	mem.Get()
	swap.Get()

	fmt.Fprintf(os.Stdout, "%18s %10s %10s\n",
		"total", "used", "free")

	fmt.Fprintf(os.Stdout, "Mem:    %10d %10d %10d\n",
		format(mem.Total), format(mem.Used), format(mem.Free))

	fmt.Fprintf(os.Stdout, "-/+ buffers/cache: %10d %10d\n",
		format(mem.ActualUsed), format(mem.ActualFree))

	fmt.Fprintf(os.Stdout, "Swap:   %10d %10d %10d\n",
		format(swap.Total), format(swap.Used), format(swap.Free))
}

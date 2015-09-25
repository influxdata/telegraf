package main

import (
	"fmt"
	"time"

	"github.com/cloudfoundry/gosigar"
)

func main() {
	cpus := sigar.CpuList{}
	cpus.Get()
	tcpu := getOverallCpu(cpus)

	for i, cpu := range cpus.List {
		fmt.Printf("CPU%d Ticks: %d\n", i, cpu.Total())
	}

	fmt.Printf("Total CPU Ticks: %d\n", tcpu.Total())
	fmt.Printf("Total CPU Time: %d\n", tcpu.Total()/128)
	fmt.Printf("User CPU Time: %d\n", tcpu.User/128)

	time.Sleep(1 * time.Second)
	tcpu2 := sigar.Cpu{}
	tcpu2.Get()

	dcpu := tcpu2.Delta(tcpu)
	tcpuDelta := tcpu2.Total() - tcpu.Total()
	iPercentage := 100.0 * float64(dcpu.Idle) / float64(tcpuDelta)
	fmt.Printf("Idle percentage: %f\n", iPercentage)
	bPercentage := 100.0 * float64(busy(tcpu2)-busy(tcpu)) / float64(tcpuDelta)
	fmt.Printf("Busy percentage: %f\n", bPercentage)
}

func busy(c sigar.Cpu) uint64 {
	return c.Total() - c.Idle
}

func getOverallCpu(cl sigar.CpuList) sigar.Cpu {
	var overallCpu sigar.Cpu
	for _, c := range cl.List {
		overallCpu.User += c.User
		overallCpu.Nice += c.Nice
		overallCpu.Sys += c.Sys
		overallCpu.Idle += c.Idle
		overallCpu.Wait += c.Wait
		overallCpu.Irq += c.Irq
		overallCpu.SoftIrq += c.SoftIrq
		overallCpu.Stolen += c.Stolen
	}
	return overallCpu
}

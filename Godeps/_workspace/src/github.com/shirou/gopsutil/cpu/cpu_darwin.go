// +build darwin

package cpu

/*
#include <stdlib.h>
#include <sys/sysctl.h>
#include <sys/mount.h>
#include <mach/mach_init.h>
#include <mach/mach_host.h>
#include <mach/host_info.h>
#include <libproc.h>
#include <mach/processor_info.h>
#include <mach/vm_map.h>
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"
)

// sys/resource.h
const (
	CPUser    = 0
	CPNice    = 1
	CPSys     = 2
	CPIntr    = 3
	CPIdle    = 4
	CPUStates = 5
)

// default value. from time.h
var ClocksPerSec = float64(128)

// these CPU times for darwin is borrowed from influxdb/telegraf.

func perCPUTimes() ([]CPUTimesStat, error) {
	var (
		count   C.mach_msg_type_number_t
		cpuload *C.processor_cpu_load_info_data_t
		ncpu    C.natural_t
	)

	status := C.host_processor_info(C.host_t(C.mach_host_self()),
		C.PROCESSOR_CPU_LOAD_INFO,
		&ncpu,
		(*C.processor_info_array_t)(unsafe.Pointer(&cpuload)),
		&count)

	if status != C.KERN_SUCCESS {
		return nil, fmt.Errorf("host_processor_info error=%d", status)
	}

	// jump through some cgo casting hoops and ensure we properly free
	// the memory that cpuload points to
	target := C.vm_map_t(C.mach_task_self_)
	address := C.vm_address_t(uintptr(unsafe.Pointer(cpuload)))
	defer C.vm_deallocate(target, address, C.vm_size_t(ncpu))

	// the body of struct processor_cpu_load_info
	// aka processor_cpu_load_info_data_t
	var cpu_ticks [C.CPU_STATE_MAX]uint32

	// copy the cpuload array to a []byte buffer
	// where we can binary.Read the data
	size := int(ncpu) * binary.Size(cpu_ticks)
	buf := C.GoBytes(unsafe.Pointer(cpuload), C.int(size))

	bbuf := bytes.NewBuffer(buf)

	var ret []CPUTimesStat

	for i := 0; i < int(ncpu); i++ {
		err := binary.Read(bbuf, binary.LittleEndian, &cpu_ticks)
		if err != nil {
			return nil, err
		}

		c := CPUTimesStat{
			CPU:    fmt.Sprintf("cpu%d", i),
			User:   float64(cpu_ticks[C.CPU_STATE_USER]) / ClocksPerSec,
			System: float64(cpu_ticks[C.CPU_STATE_SYSTEM]) / ClocksPerSec,
			Nice:   float64(cpu_ticks[C.CPU_STATE_NICE]) / ClocksPerSec,
			Idle:   float64(cpu_ticks[C.CPU_STATE_IDLE]) / ClocksPerSec,
		}

		ret = append(ret, c)
	}

	return ret, nil
}

func allCPUTimes() ([]CPUTimesStat, error) {
	var count C.mach_msg_type_number_t = C.HOST_CPU_LOAD_INFO_COUNT
	var cpuload C.host_cpu_load_info_data_t

	status := C.host_statistics(C.host_t(C.mach_host_self()),
		C.HOST_CPU_LOAD_INFO,
		C.host_info_t(unsafe.Pointer(&cpuload)),
		&count)

	if status != C.KERN_SUCCESS {
		return nil, fmt.Errorf("host_statistics error=%d", status)
	}

	c := CPUTimesStat{
		CPU:    "cpu-total",
		User:   float64(cpuload.cpu_ticks[C.CPU_STATE_USER]) / ClocksPerSec,
		System: float64(cpuload.cpu_ticks[C.CPU_STATE_SYSTEM]) / ClocksPerSec,
		Nice:   float64(cpuload.cpu_ticks[C.CPU_STATE_NICE]) / ClocksPerSec,
		Idle:   float64(cpuload.cpu_ticks[C.CPU_STATE_IDLE]) / ClocksPerSec,
	}

	return []CPUTimesStat{c}, nil

}

func CPUTimes(percpu bool) ([]CPUTimesStat, error) {
	if percpu {
		return perCPUTimes()
	}

	return allCPUTimes()
}

// Returns only one CPUInfoStat on FreeBSD
func CPUInfo() ([]CPUInfoStat, error) {
	var ret []CPUInfoStat

	out, err := exec.Command("/usr/sbin/sysctl", "machdep.cpu").Output()
	if err != nil {
		return ret, err
	}

	c := CPUInfoStat{}
	for _, line := range strings.Split(string(out), "\n") {
		values := strings.Fields(line)
		if len(values) < 1 {
			continue
		}

		t, err := strconv.ParseInt(values[1], 10, 64)
		// err is not checked here because some value is string.
		if strings.HasPrefix(line, "machdep.cpu.brand_string") {
			c.ModelName = strings.Join(values[1:], " ")
		} else if strings.HasPrefix(line, "machdep.cpu.family") {
			c.Family = values[1]
		} else if strings.HasPrefix(line, "machdep.cpu.model") {
			c.Model = values[1]
		} else if strings.HasPrefix(line, "machdep.cpu.stepping") {
			if err != nil {
				return ret, err
			}
			c.Stepping = int32(t)
		} else if strings.HasPrefix(line, "machdep.cpu.features") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.leaf7_features") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.extfeatures") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.core_count") {
			if err != nil {
				return ret, err
			}
			c.Cores = int32(t)
		} else if strings.HasPrefix(line, "machdep.cpu.cache.size") {
			if err != nil {
				return ret, err
			}
			c.CacheSize = int32(t)
		} else if strings.HasPrefix(line, "machdep.cpu.vendor") {
			c.VendorID = values[1]
		}
	}

	// Use the rated frequency of the CPU. This is a static value and does not
	// account for low power or Turbo Boost modes.
	out, err = exec.Command("/usr/sbin/sysctl", "hw.cpufrequency").Output()
	if err != nil {
		return ret, err
	}

	values := strings.Fields(string(out))
	mhz, err := strconv.ParseFloat(values[1], 64)
	if err != nil {
		return ret, err
	}
	c.Mhz = mhz / 1000000.0

	return append(ret, c), nil
}

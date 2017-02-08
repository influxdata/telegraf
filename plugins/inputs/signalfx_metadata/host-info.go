package signalfxMetadata

import (
	"io/ioutil"
	"log"
	"regexp"
	"strconv"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

// GetCPUInfo - adds information about the host cpu to the supplied map
func GetCPUInfo(info map[string]string) {
	if cpus, err := cpu.Info(); err == nil {
		var numCPUs = len(cpus)
		var numCores = int64(0)
		var logicalCPU = 0
		var CPUModel = ""
		if val, er := cpu.Counts(true); er == nil {
			logicalCPU = val
		}
		for _, cpu := range cpus {
			numCores = int64(cpu.Cores) + numCores
			CPUModel = cpu.ModelName
		}
		info["host_physical_cpus"] = strconv.Itoa(numCPUs)
		info["host_cpu_cores"] = strconv.FormatInt(numCores, 10)
		info["host_cpu_model"] = CPUModel
		info["host_logical_cpus"] = strconv.Itoa(logicalCPU)
		info["host_physical_cpus"] = strconv.Itoa(numCPUs)
	} else {
		log.Println("E! Input [signalfx-metadata] ", err)
	}
}

// GetKernelInfo - adds information about the host kernel to the supplied map
func GetKernelInfo(info map[string]string) {
	if hostInfo, err := host.Info(); err == nil {
		info["host_kernel_name"] = hostInfo.OS
		info["host_kernel_version"] = hostInfo.KernelVersion
		info["host_os_name"] = hostInfo.Platform
		info["host_os_version"] = hostInfo.PlatformVersion
		if hostInfo.OS == "linux" {
			GetLinuxVersion(info)
		}
	} else {
		log.Println("E! Input [signalfx-metadata] ", err)
	}
}

// GetLinuxVersion - adds information about the host linux version to the supplied map
func GetLinuxVersion(info map[string]string) {
	var response string
	var file []byte
	var err error
	response, err = getStringFromFile(`DISTRIB_DESCRIPTION="(.*)"`, "/etc/lsb-release")
	if err != nil {
		response, err = getStringFromFile(`PRETTY_NAME="(.*)"`, "/etc/os-release")
	}
	if err != nil {
		file, err = ioutil.ReadFile("/etc/centos-release")
		if err != nil {
			file, err = ioutil.ReadFile("/etc/redhat-release")
		}
		if err != nil {
			file, err = ioutil.ReadFile("/etc/system-release")
		}
		if err == nil {
			response = string(file)
		}
	}
	if err == nil {
		info["host_linux_version"] = response
	}
}

// GetMemory - adds information about the host memory to the supplied map
func GetMemory(info map[string]string) {
	mem, err := mem.VirtualMemory()
	if err == nil {
		info["host_mem_total"] = strconv.FormatUint(mem.Total/1024, 10)
	}
}

func getStringFromFile(pattern string, path string) (string, error) {
	var file []byte
	var match [][]byte
	var err error
	var reg = regexp.MustCompile(pattern)
	var response string

	if file, err = ioutil.ReadFile(path); err == nil {
		match = reg.FindSubmatch(file)
		if len(match) > 1 {
			response = string(match[1])
		}
	}
	return response, err
}

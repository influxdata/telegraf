package cpufreq

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const defaultHostSys = "/sys"

type CPUFreq struct {
	PathSysfs       string `toml:"host_sys"`
	GatherThrottles bool   `toml:"gather_throttles"`
}

var sampleConfig = `
  ## Path for sysfs filesystem.
  ## See https://www.kernel.org/doc/Documentation/filesystems/sysfs.txt
  ## Defaults:
  # host_sys = "/sys"
  ## Gather CPU throttles per core
  ## Defaults:
  # gather_throttles = false
`

func (g *CPUFreq) SampleConfig() string {
	return sampleConfig
}

func (g *CPUFreq) Description() string {
	return "Read specific statistics per cgroup"
}

func (g *CPUFreq) Gather(acc telegraf.Accumulator) error {

	if g.PathSysfs == "" {
		if os.Getenv("HOST_SYS") != "" {
			g.PathSysfs = os.Getenv("HOST_SYS")
		} else {
			g.PathSysfs = defaultHostSys
		}
	}

	cpus, err := filepath.Glob(path.Join(g.PathSysfs, "devices/system/cpu/cpu[0-9]*"))
	if err != nil {
		acc.AddError(err)
	}

	var value uint64
	// cpu loop
	for _, cpu := range cpus {
		fileds := make(map[string]interface{})
		tags := make(map[string]string)

		_, cpuName := filepath.Split(cpu)
		cpuNum := strings.TrimPrefix(cpuName, "cpu")

		tags["cpu"] = cpuNum

		// sysfs cpufreq values are kHz, thus multiply by 1000 to export base units (hz).
		if _, err := os.Stat(filepath.Join(cpu, "cpufreq")); os.IsNotExist(err) {
			acc.AddError(err)
		} else {
			if value, err = readUintFromFile(filepath.Join(cpu, "cpufreq", "scaling_cur_freq")); err != nil {
				acc.AddError(err)
				continue
			} else {
				fileds["cur_freq"] = uint64(value) * 1000
			}
			if value, err = readUintFromFile(filepath.Join(cpu, "cpufreq", "scaling_min_freq")); err != nil {
				acc.AddError(err)
			} else {
				fileds["min_freq"] = uint64(value) * 1000
			}
			if value, err = readUintFromFile(filepath.Join(cpu, "cpufreq", "scaling_max_freq")); err != nil {
				acc.AddError(err)
			} else {
				fileds["max_freq"] = uint64(value) * 1000
			}
		}

		if g.GatherThrottles {
			if value, err := readUintFromFile(filepath.Join(cpu, "thermal_throttle", "core_throttle_count")); err != nil {
				acc.AddError(err)
			} else {
				fileds["throttle_count"] = uint64(value)
			}
		}

		acc.AddFields("cpufreq", fileds, tags)
	}

	return nil
}

func init() {
	inputs.Add("cpufreq", func() telegraf.Input { return &CPUFreq{} })
}

func readUintFromFile(path string) (uint64, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

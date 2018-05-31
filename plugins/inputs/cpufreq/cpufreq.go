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

const defaultPathSysfs = "/sys"

type CPUFreq struct {
	PathSysfs          string `toml:"path_sysfs"`
	ThrottlesPerSocket bool   `toml:"throttles_per_socket"`
	ThrottlesPerCore   bool   `toml:"throttles_per_core"`
}

var sampleConfig = `
  ## PathSysfs
  # path_sysfs = "/sys"
  ## 
  # throttles_per_socket = false
  ##
  # throttles_per_core = false
`

func (g *CPUFreq) SampleConfig() string {
	return sampleConfig
}

func (g *CPUFreq) Description() string {
	return "Read specific statistics per cgroup"
}

func (g *CPUFreq) Gather(acc telegraf.Accumulator) error {

	if g.PathSysfs == "" {
		g.PathSysfs = defaultPathSysfs
	}

	cpus, err := filepath.Glob(path.Join(g.PathSysfs, "devices/system/cpu/cpu[0-9]*"))
	if err != nil {
		acc.AddError(err)
	}

	var value uint64
	packageThrottles := make(map[uint64]uint64)
	packageCoreThrottles := make(map[uint64]map[uint64]uint64)

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
				fileds["cur_freq"] = float64(value) * 1000.0
			}
			if value, err = readUintFromFile(filepath.Join(cpu, "cpufreq", "scaling_min_freq")); err != nil {
				acc.AddError(err)
			} else {
				fileds["min_freq"] = float64(value) * 1000.0
			}
			if value, err = readUintFromFile(filepath.Join(cpu, "cpufreq", "scaling_max_freq")); err != nil {
				acc.AddError(err)
			} else {
				fileds["max_freq"] = float64(value) * 1000.0
			}

			acc.AddFields("cpufreq", fileds, tags)
		}

		if g.ThrottlesPerSocket || g.ThrottlesPerCore {
			var physicalPackageID, coreID uint64

			// topology/physical_package_id
			if physicalPackageID, err = readUintFromFile(filepath.Join(cpu, "topology", "physical_package_id")); err != nil {
				acc.AddError(err)
				continue
			}
			// topology/core_id
			if coreID, err = readUintFromFile(filepath.Join(cpu, "topology", "core_id")); err != nil {
				acc.AddError(err)
				continue
			}

			// core_throttles
			if _, present := packageCoreThrottles[physicalPackageID]; !present {
				packageCoreThrottles[physicalPackageID] = make(map[uint64]uint64)
			}
			if _, present := packageCoreThrottles[physicalPackageID][coreID]; !present {
				// Read thermal_throttle/core_throttle_count only once
				if coreThrottleCount, err := readUintFromFile(filepath.Join(cpu, "thermal_throttle", "core_throttle_count")); err == nil {
					packageCoreThrottles[physicalPackageID][coreID] = coreThrottleCount
				} else {
					acc.AddError(err)
				}
			}

			// cpu_package_throttles
			if _, present := packageThrottles[physicalPackageID]; !present {
				// Read thermal_throttle/package_throttle_count only once
				if packageThrottleCount, err := readUintFromFile(filepath.Join(cpu, "thermal_throttle", "package_throttle_count")); err == nil {
					packageThrottles[physicalPackageID] = packageThrottleCount
				} else {
					acc.AddError(err)
				}
			}
		}
	}

	if g.ThrottlesPerSocket {
		for physicalPackageID, packageThrottleCount := range packageThrottles {
			acc.AddFields(
				"cpufreq_cpu_throttles",
				map[string]interface{}{
					"count": float64(packageThrottleCount),
				},
				map[string]string{
					"cpu": strconv.FormatUint(physicalPackageID, 10),
				},
			)
		}
	}

	if g.ThrottlesPerCore {
		for physicalPackageID, coreMap := range packageCoreThrottles {
			for coreID, coreThrottleCount := range coreMap {
				acc.AddFields(
					"cpufreq_core_throttles",
					map[string]interface{}{
						"count": float64(coreThrottleCount),
					},
					map[string]string{
						"cpu":  strconv.FormatUint(physicalPackageID, 10),
						"core": strconv.FormatUint(coreID, 10),
					},
				)
			}
		}
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

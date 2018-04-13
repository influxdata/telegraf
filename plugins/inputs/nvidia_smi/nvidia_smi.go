package nvidia_smi

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	measurement = "nvidia_smi"
	metrics     = "fan.speed,memory.total,memory.used,memory.free,pstate,temperature.gpu,name,uuid,compute_mode,utilization.gpu,utilization.memory,index"
	metricNames = [][]string{
		[]string{"fan_speed", "field"},
		[]string{"memory_total", "field"},
		[]string{"memory_used", "field"},
		[]string{"memory_free", "field"},
		[]string{"pstate", "tag"},
		[]string{"temperature_gpu", "field"},
		[]string{"name", "tag"},
		[]string{"uuid", "tag"},
		[]string{"compute_mode", "tag"},
		[]string{"utilization_gpu", "field"},
		[]string{"utilization_memory", "field"},
		[]string{"index", "tag"},
	}
)

// NvidiaSMI holds the methods for this plugin
type NvidiaSMI struct {
	BinPath string

	metrics string
}

// Description returns the description of the NvidiaSMI plugin
func (smi *NvidiaSMI) Description() string {
	return ""
}

// SampleConfig returns the sample configuration for the NvidiaSMI plugin
func (smi *NvidiaSMI) SampleConfig() string {
	return `
## Path to nvidia-smi
bin_path = /usr/bin/nvidia-smi
`
}

func (smi *NvidiaSMI) getGPUCount() (int, error) {
	opts := []string{"--format=noheader,nounits,csv", "--query-gpu=count", "--id=0"}
	ret, err := exec.Command(smi.BinPath, opts...).CombinedOutput()
	if err != nil {
		return 0, err
	}

	retS := strings.TrimSuffix(string(ret), "\n")
	retI, errI := strconv.Atoi(retS)
	if errI != nil {
		return 0, err
	}
	return retI, nil
}

func (smi *NvidiaSMI) getResult(gpuID int) (map[string]string, map[string]interface{}, error) {
	tags := make(map[string]string, 0)
	fields := make(map[string]interface{}, 0)

	// Construct and execute metrics query
	opts := []string{"--format=noheader,nounits,csv", fmt.Sprintf("--query-gpu=%s", smi.metrics), fmt.Sprintf("--id=%d", gpuID)}
	ret, err := exec.Command(smi.BinPath, opts...).CombinedOutput()
	if err != nil {
		return tags, fields, err
	}

	// Format the metrics into tags and fields
	met := strings.Split(string(ret), ", ")
	for i, m := range metricNames {
		if m[1] == "tag" {
			tags[m[0]] = met[i]
			continue
		}

		fields[m[0]] = fmt.Sprintf("%si", met[i])
	}

	return tags, fields, nil
}

// Gather implements the telegraf interface
func (smi *NvidiaSMI) Gather(acc telegraf.Accumulator) error {

	if _, err := os.Stat(smi.BinPath); os.IsNotExist(err) {
		return fmt.Errorf("nvidia-smi binary not at path %s, cannot gather GPU data", smi.BinPath)
	}

	gpuCount, err := smi.getGPUCount()
	if err != nil {
		return err
	}

	for i := 0; i < gpuCount; i++ {
		tags, fields, err := smi.getResult(i)
		if err != nil {
			return fmt.Errorf("Error getting GPU stats: %s", err)
		}
		acc.AddFields(measurement, fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("simple", func() telegraf.Input {
		return &NvidiaSMI{
			BinPath: "/usr/bin/nvidia-smi",
			metrics: metrics,
		}
	})
}

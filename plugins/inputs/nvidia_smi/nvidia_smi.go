package nvidia_smi

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
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
	Timeout time.Duration

	metrics string
}

// Description returns the description of the NvidiaSMI plugin
func (smi *NvidiaSMI) Description() string {
	return "Pulls statistics from nvidia GPUs attached to the host"
}

// SampleConfig returns the sample configuration for the NvidiaSMI plugin
func (smi *NvidiaSMI) SampleConfig() string {
	return `
## Optional: path to nvidia-smi binary, defaults to $PATH via exec.LookPath
# bin_path = /usr/bin/nvidia-smi

## Optional: timeout for GPU polling
# timeout = 5s
`
}

func (smi *NvidiaSMI) pollSMI() (string, error) {
	// Construct and execute metrics query
	opts := []string{"--format=noheader,nounits,csv", fmt.Sprintf("--query-gpu=%s", smi.metrics)}
	ret, err := internal.CombinedOutputTimeout(exec.Command(smi.BinPath, opts...), smi.Timeout)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

func gatherNvidiaSMI(ret string, acc telegraf.Accumulator) error {

	// Format the metrics into tags and fields
	lines := strings.Split(string(ret), "\n")
	for _, line := range lines {
		tags := make(map[string]string, 0)
		fields := make(map[string]interface{}, 0)
		met := strings.Split(line, ", ")
		for i, m := range metricNames {
			if m[1] == "tag" {
				tags[m[0]] = strings.TrimSpace(met[i])
				continue
			}
			out, err := strconv.ParseInt(met[i], 10, 64)
			if err != nil {
				return err
			}

			fields[m[0]] = out
		}
		acc.AddFields(measurement, fields, tags)
	}

	return nil
}

// Gather implements the telegraf interface
func (smi *NvidiaSMI) Gather(acc telegraf.Accumulator) error {

	if _, err := os.Stat(smi.BinPath); os.IsNotExist(err) {
		return fmt.Errorf("nvidia-smi binary not at path %s, cannot gather GPU data", smi.BinPath)
	}

	data, err := smi.pollSMI()
	if err != nil {
		return err
	}

	err = gatherNvidiaSMI(data, acc)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	inputs.Add("nvidia_smi", func() telegraf.Input {
		return &NvidiaSMI{
			BinPath: "/usr/bin/nvidia-smi",
			Timeout: 5 * time.Second,
			metrics: metrics,
		}
	})
}

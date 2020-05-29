package nvidia_smi

import (
	"encoding/xml"
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

const measurement = "nvidia_smi"

// NvidiaSMI holds the methods for this plugin
type NvidiaSMI struct {
	BinPath string
	Timeout internal.Duration
}

// Description returns the description of the NvidiaSMI plugin
func (smi *NvidiaSMI) Description() string {
	return "Pulls statistics from nvidia GPUs attached to the host"
}

// SampleConfig returns the sample configuration for the NvidiaSMI plugin
func (smi *NvidiaSMI) SampleConfig() string {
	return `
  ## Optional: path to nvidia-smi binary, defaults to $PATH via exec.LookPath
  # bin_path = "/usr/bin/nvidia-smi"

  ## Optional: timeout for GPU polling
  # timeout = "5s"
`
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
			Timeout: internal.Duration{Duration: 5 * time.Second},
		}
	})
}

func (smi *NvidiaSMI) pollSMI() ([]byte, error) {
	// Construct and execute metrics query
	ret, err := internal.CombinedOutputTimeout(exec.Command(smi.BinPath, "-q", "-x"), smi.Timeout.Duration)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func gatherNvidiaSMI(ret []byte, acc telegraf.Accumulator) error {
	smi := &SMI{}
	err := xml.Unmarshal(ret, smi)
	if err != nil {
		return err
	}

	metrics := smi.genTagsFields()

	for _, metric := range metrics {
		acc.AddFields(measurement, metric.fields, metric.tags)
	}

	return nil
}

type metric struct {
	tags   map[string]string
	fields map[string]interface{}
}

func (s *SMI) genTagsFields() []metric {
	metrics := []metric{}
	for i, gpu := range s.GPU {
		tags := map[string]string{
			"index": strconv.Itoa(i),
		}
		fields := map[string]interface{}{}

		setTagIfUsed(tags, "pstate", gpu.PState)
		setTagIfUsed(tags, "name", gpu.ProdName)
		setTagIfUsed(tags, "uuid", gpu.UUID)
		setTagIfUsed(tags, "compute_mode", gpu.ComputeMode)

		setIfUsed("int", fields, "fan_speed", gpu.FanSpeed)
		setIfUsed("int", fields, "memory_total", gpu.Memory.Total)
		setIfUsed("int", fields, "memory_used", gpu.Memory.Used)
		setIfUsed("int", fields, "memory_free", gpu.Memory.Free)
		setIfUsed("int", fields, "temperature_gpu", gpu.Temp.GPUTemp)
		setIfUsed("int", fields, "utilization_gpu", gpu.Utilization.GPU)
		setIfUsed("int", fields, "utilization_memory", gpu.Utilization.Memory)
		setIfUsed("int", fields, "pcie_link_gen_current", gpu.PCI.LinkInfo.PCIEGen.CurrentLinkGen)
		setIfUsed("int", fields, "pcie_link_width_current", gpu.PCI.LinkInfo.LinkWidth.CurrentLinkWidth)
		setIfUsed("int", fields, "encoder_stats_session_count", gpu.Encoder.SessionCount)
		setIfUsed("int", fields, "encoder_stats_average_fps", gpu.Encoder.AverageFPS)
		setIfUsed("int", fields, "encoder_stats_average_latency", gpu.Encoder.AverageLatency)
		setIfUsed("int", fields, "clocks_current_graphics", gpu.Clocks.Graphics)
		setIfUsed("int", fields, "clocks_current_sm", gpu.Clocks.SM)
		setIfUsed("int", fields, "clocks_current_memory", gpu.Clocks.Memory)
		setIfUsed("int", fields, "clocks_current_video", gpu.Clocks.Video)

		setIfUsed("float", fields, "power_draw", gpu.Power.PowerDraw)
		metrics = append(metrics, metric{tags, fields})
	}
	return metrics
}

func setTagIfUsed(m map[string]string, k, v string) {
	if v != "" {
		m[k] = v
	}
}

func setIfUsed(t string, m map[string]interface{}, k, v string) {
	vals := strings.Fields(v)
	if len(vals) < 1 {
		return
	}

	val := vals[0]
	if k == "pcie_link_width_current" {
		val = strings.TrimSuffix(vals[0], "x")
	}

	switch t {
	case "float":
		if val != "" {
			f, err := strconv.ParseFloat(val, 64)
			if err == nil {
				m[k] = f
			}
		}
	case "int":
		if val != "" {
			i, err := strconv.Atoi(val)
			if err == nil {
				m[k] = i
			}
		}
	}
}

// SMI defines the structure for the output of _nvidia-smi -q -x_.
type SMI struct {
	GPU GPU `xml:"gpu"`
}

// GPU defines the structure of the GPU portion of the smi output.
type GPU []struct {
	FanSpeed    string           `xml:"fan_speed"` // int
	Memory      MemoryStats      `xml:"fb_memory_usage"`
	PState      string           `xml:"performance_state"`
	Temp        TempStats        `xml:"temperature"`
	ProdName    string           `xml:"product_name"`
	UUID        string           `xml:"uuid"`
	ComputeMode string           `xml:"compute_mode"`
	Utilization UtilizationStats `xml:"utilization"`
	Power       PowerReadings    `xml:"power_readings"`
	PCI         PCI              `xml:"pci"`
	Encoder     EncoderStats     `xml:"encoder_stats"`
	Clocks      ClockStats       `xml:"clocks"`
}

// MemoryStats defines the structure of the memory portions in the smi output.
type MemoryStats struct {
	Total string `xml:"total"` // int
	Used  string `xml:"used"`  // int
	Free  string `xml:"free"`  // int
}

// TempStats defines the structure of the temperature portion of the smi output.
type TempStats struct {
	GPUTemp string `xml:"gpu_temp"` // int
}

// UtilizationStats defines the structure of the utilization portion of the smi output.
type UtilizationStats struct {
	GPU    string `xml:"gpu_util"`    // int
	Memory string `xml:"memory_util"` // int
}

// PowerReadings defines the structure of the power_readings portion of the smi output.
type PowerReadings struct {
	PowerDraw string `xml:"power_draw"` // float
}

// PCI defines the structure of the pci portion of the smi output.
type PCI struct {
	LinkInfo struct {
		PCIEGen struct {
			CurrentLinkGen string `xml:"current_link_gen"` // int
		} `xml:"pcie_gen"`
		LinkWidth struct {
			CurrentLinkWidth string `xml:"current_link_width"` // int
		} `xml:"link_widths"`
	} `xml:"pci_gpu_link_info"`
}

// EncoderStats defines the structure of the encoder_stats portion of the smi output.
type EncoderStats struct {
	SessionCount   string `xml:"session_count"`   // int
	AverageFPS     string `xml:"average_fps"`     // int
	AverageLatency string `xml:"average_latency"` // int
}

// ClockStats defines the structure of the clocks portion of the smi output.
type ClockStats struct {
	Graphics string `xml:"graphics_clock"` // int
	SM       string `xml:"sm_clock"`       // int
	Memory   string `xml:"mem_clock"`      // int
	Video    string `xml:"video_clock"`    // int
}

package rocm_smi

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const measurement = "rocm_smi"

type ROCmSMI struct {
	BinPath string
	Timeout config.Duration
}

// Description returns the description of the ROCmSMI plugin
func (rsmi *ROCmSMI) Description() string {
	return "Query statistics from AMD Graphics cards using rocm-smi binary"
}

var ROCmSMIConfig = `
## Optional: path to rocm-smi binary, defaults to $PATH via exec.LookPath
# bin_path = "/opt/rocm/bin/rocm-smi"

## Optional: timeout for GPU polling
# timeout = "5s"
`

// SampleConfig returns the sample configuration for the ROCmSMI plugin
func (rsmi *ROCmSMI) SampleConfig() string {
	return ROCmSMIConfig
}

// Gather implements the telegraf interface
func (rsmi *ROCmSMI) Gather(acc telegraf.Accumulator) error {
	if _, err := os.Stat(rsmi.BinPath); os.IsNotExist(err) {
		return fmt.Errorf("rocm-smi binary not found in path %s, cannot query GPUs statistics", rsmi.BinPath)
	}

	data, err := rsmi.pollROCmSMI()
	if err != nil {
		return err
	}

	err = gatherROCmSMI(data, acc)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	inputs.Add("rocm_smi", func() telegraf.Input {
		return &ROCmSMI{
			BinPath: "/opt/rocm/bin/rocm-smi",
			Timeout: config.Duration(5 * time.Second),
		}
	})
}

func (rsmi *ROCmSMI) pollROCmSMI() ([]byte, error) {
	// Construct and execute metrics query, there currently exist (ROCm v4.3.x) a "-a" option
	// that does not provide all the information, so each needed parameter is set manually
	cmd := exec.Command(rsmi.BinPath,
		"-o",
		"-l",
		"-m",
		"-M",
		"-g",
		"-c",
		"-t",
		"-u",
		"-i",
		"-f",
		"-p",
		"-P",
		"-s",
		"-S",
		"-v",
		"--showreplaycount",
		"--showpids",
		"--showdriverversion",
		"--showmemvendor",
		"--showfwinfo",
		"--showproductname",
		"--showserial",
		"--showuniqueid",
		"--showbus",
		"--showpendingpages",
		"--showpagesinfo",
		"--showmeminfo",
		"all",
		"--showretiredpages",
		"--showunreservablepages",
		"--showmemuse",
		"--showvoltage",
		"--showtopo",
		"--showtopoweight",
		"--showtopohops",
		"--showtopotype",
		"--showtoponuma",
		"--json")

	ret, _ := internal.StdOutputTimeout(cmd,
		time.Duration(rsmi.Timeout))
	return ret, nil
}

func gatherROCmSMI(ret []byte, acc telegraf.Accumulator) error {
	var gpus map[string]GPU
	var sys map[string]sysInfo

	err1 := json.Unmarshal(ret, &gpus)
	if err1 != nil {
		log.Fatal(err1)
	}

	err2 := json.Unmarshal(ret, &sys)
	if err2 != nil {
		log.Fatal(err2)
	}

	metrics := genTagsFields(gpus, sys)

	for _, metric := range metrics {
		acc.AddFields(measurement, metric.fields, metric.tags)
	}

	return nil
}

type metric struct {
	tags   map[string]string
	fields map[string]interface{}
}

func genTagsFields(gpus map[string]GPU, system map[string]sysInfo) []metric {
	metrics := []metric{}
	for cardId, payload := range gpus {
		if strings.Contains(cardId, "card") {
			tags := map[string]string{
				"name": cardId,
			}
			fields := map[string]interface{}{}

			totVRAM, _ := strconv.Atoi(payload.Gpu_VRAM_total_memory)
			usdVRAM, _ := strconv.Atoi(payload.Gpu_VRAM_total_used_memory)
			strFree := strconv.Itoa(totVRAM - usdVRAM)

			setTagIfUsed(tags, "gpu_id", payload.Gpu_id)
			setTagIfUsed(tags, "gpu_unique_id", payload.Gpu_unique_id)

			setIfUsed("int", fields, "driver_version", strings.Replace(system["system"].Driver_version, ".", "", -1))
			setIfUsed("int", fields, "fan_speed", payload.Gpu_fan_speed_percentage)
			setIfUsed("int", fields, "memory_total", payload.Gpu_VRAM_total_memory)
			setIfUsed("int", fields, "memory_used", payload.Gpu_VRAM_total_used_memory)
			setIfUsed("int", fields, "memory_free", strFree)
			setIfUsed("float", fields, "temperature_sensor_edge", payload.Gpu_temperature_sensor_edge)
			setIfUsed("float", fields, "temperature_sensor_sensor_junction", payload.Gpu_temperature_sensor_junction)
			setIfUsed("float", fields, "temperature_sensor_memory", payload.Gpu_temperature_sensor_memory)
			setIfUsed("int", fields, "utilization_gpu", payload.Gpu_use_percentage)
			setIfUsed("int", fields, "utilization_memory", payload.Gpu_memory_use_percentage)
			setIfUsed("int", fields, "clocks_current_sm", strings.Trim(payload.Gpu_sclk_clock_speed, "(Mhz)"))
			setIfUsed("int", fields, "clocks_current_memory", strings.Trim(payload.Gpu_mclk_clock_speed, "(Mhz)"))
			setIfUsed("float", fields, "power_draw", payload.Gpu_average_power)

			metrics = append(metrics, metric{tags, fields})
		}
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
	case "str":
		if val != "" {
			m[k] = val
		}
	}
}

type sysInfo struct {
	Driver_version string `json:"Driver version"`
}

type GPU struct {
	Gpu_id                          string `json:"GPU ID"`
	Gpu_unique_id                   string `json:"Unique ID"`
	Gpu_VBIOS_version               string `json:"VBIOS version"`
	Gpu_temperature_sensor_edge     string `json:"Temperature (Sensor edge) (C)"`
	Gpu_temperature_sensor_junction string `json:"Temperature (Sensor junction) (C)"`
	Gpu_temperature_sensor_memory   string `json:"Temperature (Sensor memory) (C)"`
	Gpu_dcefclk_clock_speed         string `json:"dcefclk clock speed"`
	Gpu_dcefclk_clock_level         string `json:"dcefclk clock level"`
	Gpu_fclk_clock_speed            string `json:"fclk clock speed"`
	Gpu_fclk_clock_level            string `json:"fclk clock level"`
	Gpu_mclk_clock_speed            string `json:"mclk clock speed:"`
	Gpu_mclk_clock_level            string `json:"mclk clock level:"`
	Gpu_sclk_clock_speed            string `json:"sclk clock speed:"`
	Gpu_sclk_clock_level            string `json:"sclk clock level:"`
	Gpu_socclk_clock_speed          string `json:"socclk clock speed"`
	Gpu_socclk_clock_level          string `json:"socclk clock level"`
	Gpu_pcie_clock                  string `json:"pcie clock level"`
	Gpu_fan_speed_level             string `json:"Fan speed (level)"`
	Gpu_fan_speed_percentage        string `json:"Fan speed (%)"` // int
	Gpu_fan_RPM                     string `json:"Fan RPM"`
	Gpu_performance_Level           string `json:"Performance Level"`
	Gpu_overdrive                   string `json:"GPU OverDrive value (%)"`
	Gpu_max_power                   string `json:"Max Graphics Package Power (W)"`
	Gpu_average_power               string `json:"Average Graphics Package Power (W)"`
	Gpu_use_percentage              string `json:"GPU use (%)"`
	Gpu_memory_use_percentage       string `json:"GPU memory use (%)"`
	Gpu_memory_vendor               string `json:"GPU memory vendor"`
	Gpu_PCIe_replay                 string `json:"PCIe Replay Count"`
	Gpu_serial_number               string `json:"Serial Number"`
	Gpu_voltage_mV                  string `json:"Voltage (mV)"`
	Gpu_PCI_bus                     string `json:"PCI Bus"`
	Gpu_ASD_firmware                string `json:"ASD firmware version"`
	Gpu_CE_firmware                 string `json:"CE firmware version"`
	Gpu_DMCU_firmware               string `json:"DMCU firmware version"`
	Gpu_MC_firmware                 string `json:"MC firmware version"`
	Gpu_ME_firmware                 string `json:"ME firmware version"`
	Gpu_MEC_firmware                string `json:"MEC firmware version"`
	Gpu_MEC2_firmware               string `json:"MEC2 firmware version"`
	Gpu_PFP_firmware                string `json:"PFP firmware version"`
	Gpu_RLC_firmware                string `json:"RLC firmware version"`
	Gpu_RLC_SRLC                    string `json:"RLC SRLC firmware version"`
	Gpu_RLC_SRLG                    string `json:"RLC SRLG firmware version"`
	Gpu_RLC_SRLS                    string `json:"RLC SRLS firmware version"`
	Gpu_SDMA_firmware               string `json:"SDMA firmware version"`
	Gpu_SDMA2_firmware              string `json:"SDMA2 firmware version"`
	Gpu_SMC_firmware                string `json:"SMC firmware version"`
	Gpu_SOS_firmware                string `json:"SOS firmware version"`
	Gpu_TA_RAS                      string `json:"TA RAS firmware version"`
	Gpu_TA_XGMI                     string `json:"TA XGMI firmware version"`
	Gpu_UVD_firmware                string `json:"UVD firmware version"`
	Gpu_VCE_firmware                string `json:"VCE firmware version"`
	Gpu_VCN_firmware                string `json:"VCN firmware version"`
	Gpu_card_series                 string `json:"Card series"`
	Gpu_card_model                  string `json:"Card model"`
	Gpu_card_vendor                 string `json:"Card vendor"`
	Gpu_card_SKU                    string `json:"Card SKU"`
	Gpu_NUMA_node                   string `json:"(Topology) Numa Node"`
	Gpu_NUMA_affinity               string `json:"(Topology) Numa Affinity"`
	Gpu_vis_VRAM_total_memory       string `json:"VIS_VRAM Total Memory (B)"`
	Gpu_vis_VRAM_total_used_memory  string `json:"VIS_VRAM Total Used Memory (B)"`
	Gpu_VRAM_total_memory           string `json:"VRAM Total Memory (B)"`
	Gpu_VRAM_total_used_memory      string `json:"VRAM Total Used Memory (B)"`
	Gpu_GTT_total_memory            string `json:"GTT Total Memory (B)"`
	Gpu_GTT_total_used_memory       string `json:"GTT Total Used Memory (B)"`
}

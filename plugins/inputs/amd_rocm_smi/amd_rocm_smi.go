package amd_rocm_smi

import (
	"encoding/json"
	"fmt"
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

const measurement = "amd_rocm_smi"

type ROCmSMI struct {
	BinPath string
	Timeout config.Duration
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
	inputs.Add("amd_rocm_smi", func() telegraf.Input {
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
		return err1
	}

	err2 := json.Unmarshal(ret, &sys)
	if err2 != nil {
		return err2
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
	for cardID, payload := range gpus {
		if strings.Contains(cardID, "card") {
			tags := map[string]string{
				"name": cardID,
			}
			fields := map[string]interface{}{}

			totVRAM, _ := strconv.ParseInt(payload.GpuVRAMTotalMemory, 10, 64)
			usdVRAM, _ := strconv.ParseInt(payload.GpuVRAMTotalUsedMemory, 10, 64)
			strFree := strconv.FormatInt(totVRAM-usdVRAM, 10)

			setTagIfUsed(tags, "gpu_id", payload.GpuID)
			setTagIfUsed(tags, "gpu_unique_id", payload.GpuUniqueID)

			setIfUsed("int", fields, "driver_version", strings.Replace(system["system"].DriverVersion, ".", "", -1))
			setIfUsed("int", fields, "fan_speed", payload.GpuFanSpeedPercentage)
			setIfUsed("int64", fields, "memory_total", payload.GpuVRAMTotalMemory)
			setIfUsed("int64", fields, "memory_used", payload.GpuVRAMTotalUsedMemory)
			setIfUsed("int64", fields, "memory_free", strFree)
			setIfUsed("float", fields, "temperature_sensor_edge", payload.GpuTemperatureSensorEdge)
			setIfUsed("float", fields, "temperature_sensor_junction", payload.GpuTemperatureSensorJunction)
			setIfUsed("float", fields, "temperature_sensor_memory", payload.GpuTemperatureSensorMemory)
			setIfUsed("int", fields, "utilization_gpu", payload.GpuUsePercentage)
			setIfUsed("int", fields, "utilization_memory", payload.GpuMemoryUsePercentage)
			setIfUsed("int", fields, "clocks_current_sm", strings.Trim(payload.GpuSclkClockSpeed, "(Mhz)"))
			setIfUsed("int", fields, "clocks_current_memory", strings.Trim(payload.GpuMclkClockSpeed, "(Mhz)"))
			setIfUsed("float", fields, "power_draw", payload.GpuAveragePower)

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
	case "int64":
		if val != "" {
			i, err := strconv.ParseInt(val, 10, 64)
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
	DriverVersion string `json:"Driver version"`
}

type GPU struct {
	GpuID                        string `json:"GPU ID"`
	GpuUniqueID                  string `json:"Unique ID"`
	GpuVBIOSVersion              string `json:"VBIOS version"`
	GpuTemperatureSensorEdge     string `json:"Temperature (Sensor edge) (C)"`
	GpuTemperatureSensorJunction string `json:"Temperature (Sensor junction) (C)"`
	GpuTemperatureSensorMemory   string `json:"Temperature (Sensor memory) (C)"`
	GpuDcefClkClockSpeed         string `json:"dcefclk clock speed"`
	GpuDcefClkClockLevel         string `json:"dcefclk clock level"`
	GpuFclkClockSpeed            string `json:"fclk clock speed"`
	GpuFclkClockLevel            string `json:"fclk clock level"`
	GpuMclkClockSpeed            string `json:"mclk clock speed:"`
	GpuMclkClockLevel            string `json:"mclk clock level:"`
	GpuSclkClockSpeed            string `json:"sclk clock speed:"`
	GpuSclkClockLevel            string `json:"sclk clock level:"`
	GpuSocclkClockSpeed          string `json:"socclk clock speed"`
	GpuSocclkClockLevel          string `json:"socclk clock level"`
	GpuPcieClock                 string `json:"pcie clock level"`
	GpuFanSpeedLevel             string `json:"Fan speed (level)"`
	GpuFanSpeedPercentage        string `json:"Fan speed (%)"`
	GpuFanRPM                    string `json:"Fan RPM"`
	GpuPerformanceLevel          string `json:"Performance Level"`
	GpuOverdrive                 string `json:"GPU OverDrive value (%)"`
	GpuMaxPower                  string `json:"Max Graphics Package Power (W)"`
	GpuAveragePower              string `json:"Average Graphics Package Power (W)"`
	GpuUsePercentage             string `json:"GPU use (%)"`
	GpuMemoryUsePercentage       string `json:"GPU memory use (%)"`
	GpuMemoryVendor              string `json:"GPU memory vendor"`
	GpuPCIeReplay                string `json:"PCIe Replay Count"`
	GpuSerialNumber              string `json:"Serial Number"`
	GpuVoltagemV                 string `json:"Voltage (mV)"`
	GpuPCIBus                    string `json:"PCI Bus"`
	GpuASDDirmware               string `json:"ASD firmware version"`
	GpuCEFirmware                string `json:"CE firmware version"`
	GpuDMCUFirmware              string `json:"DMCU firmware version"`
	GpuMCFirmware                string `json:"MC firmware version"`
	GpuMEFirmware                string `json:"ME firmware version"`
	GpuMECFirmware               string `json:"MEC firmware version"`
	GpuMEC2Firmware              string `json:"MEC2 firmware version"`
	GpuPFPFirmware               string `json:"PFP firmware version"`
	GpuRLCFirmware               string `json:"RLC firmware version"`
	GpuRLCSRLC                   string `json:"RLC SRLC firmware version"`
	GpuRLCSRLG                   string `json:"RLC SRLG firmware version"`
	GpuRLCSRLS                   string `json:"RLC SRLS firmware version"`
	GpuSDMAFirmware              string `json:"SDMA firmware version"`
	GpuSDMA2Firmware             string `json:"SDMA2 firmware version"`
	GpuSMCFirmware               string `json:"SMC firmware version"`
	GpuSOSFirmware               string `json:"SOS firmware version"`
	GpuTARAS                     string `json:"TA RAS firmware version"`
	GpuTAXGMI                    string `json:"TA XGMI firmware version"`
	GpuUVDFirmware               string `json:"UVD firmware version"`
	GpuVCEFirmware               string `json:"VCE firmware version"`
	GpuVCNFirmware               string `json:"VCN firmware version"`
	GpuCardSeries                string `json:"Card series"`
	GpuCardModel                 string `json:"Card model"`
	GpuCardVendor                string `json:"Card vendor"`
	GpuCardSKU                   string `json:"Card SKU"`
	GpuNUMANode                  string `json:"(Topology) Numa Node"`
	GpuNUMAAffinity              string `json:"(Topology) Numa Affinity"`
	GpuVisVRAMTotalMemory        string `json:"VIS_VRAM Total Memory (B)"`
	GpuVisVRAMTotalUsedMemory    string `json:"VIS_VRAM Total Used Memory (B)"`
	GpuVRAMTotalMemory           string `json:"VRAM Total Memory (B)"`
	GpuVRAMTotalUsedMemory       string `json:"VRAM Total Used Memory (B)"`
	GpuGTTTotalMemory            string `json:"GTT Total Memory (B)"`
	GpuGTTTotalUsedMemory        string `json:"GTT Total Used Memory (B)"`
}

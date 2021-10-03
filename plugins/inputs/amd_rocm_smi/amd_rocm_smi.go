package amd_rocm_smi

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/transport"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/xpath"

	generic "github.com/influxdata/telegraf/plugins/common/receive_parse"
)

func NewAMDSMI() *generic.ReceiveAndParse {
	return &generic.ReceiveAndParse{
		DescriptionText: "Query statistics from AMD Graphics cards using rocm-smi binary",
		Receiver: &transport.Exec{
			BinPath: "/opt/rocm/bin/rocm-smi",
			Timeout: config.Duration(5 * time.Second),
			BinArgs: []string{
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
				"--json"},
		},
		Parser: &xpath.Parser{
			Format: "xpath_json",
			// PrintDocument: true,
			Configs: []xpath.Config{
				{
					MetricDefaultName: "amd_rocm_smi",
					IgnoreNaN:         true,
					Selection:         "//*[starts-with(name(.), 'card')]",
					Tags: map[string]string{
						"name":          "name()",
						"gpu_id":        "child::*[name() = 'GPU ID']",
						"gpu_unique_id": "child::*[name() = 'Unique ID']",
					},
					Fields: map[string]string{
						"temperature_sensor_edge":     "number(child::*[name() = 'Temperature (Sensor edge) (C)'])",
						"temperature_sensor_junction": "number(child::*[name() = 'Temperature (Sensor junction) (C)'])",
						"temperature_sensor_memory":   "number(child::*[name() = 'Temperature (Sensor memory) (C)'])",
						"power_draw":                  "number(child::*[name() = 'Average Graphics Package Power (W)'])",
						"driver_version":              "/system/child::*[name() = 'Driver version']",
					},
					FieldsInt: map[string]string{
						"fan_speed":             "child::*[name() = 'Fan speed (%)']",
						"memory_total":          "child::*[name() = 'VRAM Total Memory (B)']",
						"memory_used":           "child::*[name() = 'VRAM Total Used Memory (B)']",
						"utilization_gpu":       "child::*[name() = 'GPU use (%)']",
						"utilization_memory":    "child::*[name() = 'GPU memory use (%)']",
						"clocks_current_sm":     "substring(child::*[name() = 'sclk clock speed:'], 2, string-length(.)-5)",
						"clocks_current_memory": "substring(child::*[name() = 'mclk clock speed:'], 2, string-length(.)-5)",
					},
				},
			},
		},
		PostProcessors: []generic.PostProcessor{
			{
				Name:    "memory_free computation",
				Process: postProcessMemoryFree,
			},
			{
				Name:    "driver_version conversion",
				Process: postProcessDriverVersion,
			},
		},
	}
}

func postProcessMemoryFree(m telegraf.Metric) error {
	fields := m.Fields()

	iTotal, found := fields["memory_total"]
	if !found {
		return fmt.Errorf("memory_total missing")
	}
	total, ok := iTotal.(int64)
	if !ok {
		return fmt.Errorf("memory_total is not int64 but %T", iTotal)
	}

	iUsed, found := fields["memory_used"]
	if !found {
		return fmt.Errorf("memory_used missing")
	}
	used, ok := iUsed.(int64)
	if !ok {
		return fmt.Errorf("memory_used is not int64 but %T", iUsed)
	}

	m.AddField("memory_free", total-used)
	return nil
}

func postProcessDriverVersion(m telegraf.Metric) error {
	iVersion, found := m.GetField("driver_version")
	if !found {
		return nil
	}

	sVersion, ok := iVersion.(string)
	if !ok {
		return fmt.Errorf("driver_version is not string but %T", iVersion)
	}

	sVersion = strings.Replace(sVersion, ".", "", -1)
	version, err := strconv.ParseInt(sVersion, 10, 64)
	if err != nil {
		return err
	}

	m.AddField("driver_version", version)
	return nil
}

func init() {
	inputs.Add("amd_rocm_smi", func() telegraf.Input { return NewAMDSMI() })
}

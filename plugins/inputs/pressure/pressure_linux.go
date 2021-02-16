// +build linux

package pressure

import (
	"bytes"
	"github.com/influxdata/telegraf"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type pressure struct {
	cpu struct {
		some *pressureFields
	}
	memory struct {
		some *pressureFields
		full *pressureFields
	}
	io struct {
		some *pressureFields
		full *pressureFields
	}
}

type pressureFields struct {
	avg10  float64
	avg60  float64
	avg300 float64
	total  uint64
}

const (
	PSI_ROOT = "/proc/pressure/"
)

var parsingErrors uint8 = 0

func (p *Pressure) Gather(acc telegraf.Accumulator) error {
	if !checkCompatibility() {
		p.Log.Error("System is not compatible with plugin. Probably you running kernel version <4.20 or your kernel compiled without CONFIG_PSI.")
		return nil
	}
	rawCpu, err := ioutil.ReadFile(PSI_ROOT + "cpu")
	if err != nil {
		p.Log.Errorf("Error occurred while reading cpu pressure data: %s\n", err)
	}
	rawMem, err := ioutil.ReadFile(PSI_ROOT + "memory")
	if err != nil {
		p.Log.Errorf("Error occurred while reading memory pressure data: %s\n", err)
	}
	rawIo, err := ioutil.ReadFile(PSI_ROOT + "io")
	if err != nil {
		p.Log.Errorf("Error occurred while reading io pressure data: %s\n", err)
	}
	splitCpu := bytes.Split(rawCpu, []byte("\n"))
	splitMem := bytes.Split(rawMem, []byte("\n"))
	splitIo := bytes.Split(rawIo, []byte("\n"))
	metrics := &pressure{
		cpu: struct {
			some *pressureFields
		}{
			some: parsePressureData(splitCpu[0]),
		},
		memory: struct {
			some *pressureFields
			full *pressureFields
		}{
			some: parsePressureData(splitMem[0]),
			full: parsePressureData(splitMem[1]),
		},
		io: struct {
			some *pressureFields
			full *pressureFields
		}{
			some: parsePressureData(splitIo[0]),
			full: parsePressureData(splitIo[1]),
		},
	}
	if parsingErrors != 0 {
		p.Log.Error("There was parsing errors when collecting data")
		return nil
	}
	acc.AddGauge("pressure", map[string]interface{}{
		"avg10":  metrics.cpu.some.avg10,
		"avg60":  metrics.cpu.some.avg60,
		"avg300": metrics.cpu.some.avg300,
		"total":  metrics.cpu.some.total,
	}, map[string]string{
		"resource": "cpu",
		"type":     "some",
	})
	acc.AddGauge("pressure", map[string]interface{}{
		"avg10":  metrics.memory.some.avg10,
		"avg60":  metrics.memory.some.avg60,
		"avg300": metrics.memory.some.avg300,
		"total":  metrics.memory.some.total,
	}, map[string]string{
		"resource": "memory",
		"type":     "some",
	})
	acc.AddGauge("pressure", map[string]interface{}{
		"avg10":  metrics.memory.full.avg10,
		"avg60":  metrics.memory.full.avg60,
		"avg300": metrics.memory.full.avg300,
		"total":  metrics.memory.full.total,
	}, map[string]string{
		"resource": "memory",
		"type":     "full",
	})
	acc.AddGauge("pressure", map[string]interface{}{
		"avg10":  metrics.io.some.avg10,
		"avg60":  metrics.io.some.avg60,
		"avg300": metrics.io.some.avg300,
		"total":  metrics.io.some.total,
	}, map[string]string{
		"resource": "io",
		"type":     "some",
	})
	acc.AddGauge("pressure", map[string]interface{}{
		"avg10":  metrics.io.full.avg10,
		"avg60":  metrics.io.full.avg60,
		"avg300": metrics.io.full.avg300,
		"total":  metrics.io.full.total,
	}, map[string]string{
		"resource": "io",
		"type":     "full",
	})
	return nil
}

func checkCompatibility() bool {
	_, err := os.Stat("/proc/pressure")
	if err != nil {
		return false
	}
	return true
}

func parsePressureData(line []byte) *pressureFields {
	fields := strings.Fields(string(line))
	avgs := make([]float64, 3)
	for i := 1; i < len(fields)-1; i++ {
		var err error
		avgs[i-1], err = strconv.ParseFloat(strings.Split(fields[i], "=")[1], 64)
		if err != nil {
			parsingErrors++
		}
	}
	tot, _ := strconv.ParseUint(strings.Split(fields[4], "=")[1], 10, 64)
	return &pressureFields{
		avg10:  avgs[0],
		avg60:  avgs[1],
		avg300: avgs[2],
		total:  tot,
	}
}


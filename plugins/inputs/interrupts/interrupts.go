//go:generate ../../../tools/readme_config_includer/generator
package interrupts

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Interrupts struct {
	CPUAsTag bool `toml:"cpu_as_tag"`
}

type irq struct {
	id     string
	typ    string
	device string
	total  int64
	cpus   []int64
}

func (*Interrupts) SampleConfig() string {
	return sampleConfig
}

func (s *Interrupts) Gather(acc telegraf.Accumulator) error {
	for measurement, file := range map[string]string{"interrupts": "/proc/interrupts", "soft_interrupts": "/proc/softirqs"} {
		irqs, err := parseFile(file)
		if err != nil {
			acc.AddError(err)
			continue
		}
		reportMetrics(measurement, irqs, acc, s.CPUAsTag)
	}
	return nil
}

func parseInterrupts(r io.Reader) ([]irq, error) {
	var irqs []irq
	var cpucount int
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		cpus := strings.Fields(scanner.Text())
		if cpus[0] != "CPU0" {
			return nil, fmt.Errorf("expected first line to start with CPU0, but was %s", scanner.Text())
		}
		cpucount = len(cpus)
	}

scan:
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if !strings.HasSuffix(fields[0], ":") {
			continue
		}
		irqid := strings.TrimRight(fields[0], ":")
		irq := newIRQ(irqid)
		irqvals := fields[1:]
		for i := 0; i < cpucount; i++ {
			if i < len(irqvals) {
				irqval, err := strconv.ParseInt(irqvals[i], 10, 64)
				if err != nil {
					continue scan
				}
				irq.cpus = append(irq.cpus, irqval)
			}
		}
		for _, irqval := range irq.cpus {
			irq.total += irqval
		}
		_, err := strconv.ParseInt(irqid, 10, 64)
		if err == nil && len(fields) >= cpucount+2 {
			irq.typ = fields[cpucount+1]
			irq.device = strings.Join(fields[cpucount+2:], " ")
		} else if len(fields) > cpucount {
			irq.typ = strings.Join(fields[cpucount+1:], " ")
		}
		irqs = append(irqs, *irq)
	}
	if scanner.Err() != nil {
		return nil, fmt.Errorf("error scanning file: %w", scanner.Err())
	}
	return irqs, nil
}

func gatherTagsFields(irq irq) (map[string]string, map[string]interface{}) {
	tags := map[string]string{"irq": irq.id, "type": irq.typ, "device": irq.device}
	fields := map[string]interface{}{"total": irq.total}
	for i := 0; i < len(irq.cpus); i++ {
		cpu := fmt.Sprintf("CPU%d", i)
		fields[cpu] = irq.cpus[i]
	}
	return tags, fields
}

func parseFile(file string) ([]irq, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %s", file)
	}
	defer f.Close()

	irqs, err := parseInterrupts(f)
	if err != nil {
		return nil, fmt.Errorf("parsing %q: %w", file, err)
	}
	return irqs, nil
}

func reportMetrics(measurement string, irqs []irq, acc telegraf.Accumulator, cpusAsTags bool) {
	for _, irq := range irqs {
		tags, fields := gatherTagsFields(irq)
		if cpusAsTags {
			for cpu, count := range irq.cpus {
				cpuTags := map[string]string{"cpu": fmt.Sprintf("cpu%d", cpu)}
				for k, v := range tags {
					cpuTags[k] = v
				}
				acc.AddFields(measurement, map[string]interface{}{"count": count}, cpuTags)
			}
		} else {
			acc.AddFields(measurement, fields, tags)
		}
	}
}

func newIRQ(id string) *irq {
	return &irq{id: id}
}

func init() {
	inputs.Add("interrupts", func() telegraf.Input {
		return &Interrupts{}
	})
}

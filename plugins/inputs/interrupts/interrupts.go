package interrupts

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Interrupts struct{}

type IRQ struct {
	ID     string
	Type   string
	Device string
	CPU    int
	Count  int64
}

func NewIRQ(id string, cpu int, count int64) *IRQ {
	return &IRQ{ID: id, CPU: cpu, Count: count}
}

const sampleConfig = `
  ## To filter which IRQs to collect, make use of tagpass / tagdrop, i.e.
  # [inputs.interrupts.tagdrop]
    # irq = [ "NET_RX", "TASKLET" ]
`

func (s *Interrupts) Description() string {
	return "This plugin gathers interrupts data from /proc/interrupts and /proc/softirqs."
}

func (s *Interrupts) SampleConfig() string {
	return sampleConfig
}

func parseInterrupts(r io.Reader) ([]IRQ, error) {
	var irqs []IRQ
	var cpucount int
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		cpus := strings.Fields(scanner.Text())
		if cpus[0] != "CPU0" {
			return nil, fmt.Errorf("Expected first line to start with CPU0, but was %s", scanner.Text())
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
		irqvals := fields[1:]

		_, err := strconv.ParseInt(irqid, 10, 64)
		irqType := ""
		irqDevice := ""
		if err == nil && len(fields) >= cpucount+2 {
			irqType = fields[cpucount+1]
			irqDevice = strings.Join(fields[cpucount+2:], " ")
		} else if len(fields) > cpucount {
			irqType = strings.Join(fields[cpucount+1:], " ")
		}

		for i := 0; i < cpucount; i++ {
			if i < len(irqvals) {
				irqval, err := strconv.ParseInt(irqvals[i], 10, 64)
				if err != nil {
					continue scan
				}
				irq := NewIRQ(irqid, i, irqval)
				irq.Type = irqType
				irq.Device = irqDevice
				irqs = append(irqs, *irq)
			}
		}

	}
	if scanner.Err() != nil {
		return nil, fmt.Errorf("Error scanning file: %s", scanner.Err())
	}
	return irqs, nil
}

func gatherTagsFields(irq IRQ) (map[string]string, map[string]interface{}) {
	tags := map[string]string{"irq": irq.ID, "type": irq.Type, "device": irq.Device, "cpu": "cpu" + strconv.Itoa(irq.CPU)}
	fields := map[string]interface{}{"count": irq.Count}
	return tags, fields
}

func (s *Interrupts) Gather(acc telegraf.Accumulator) error {
	for measurement, file := range map[string]string{"interrupts": "/proc/interrupts", "soft_interrupts": "/proc/softirqs"} {
		f, err := os.Open(file)
		if err != nil {
			acc.AddError(fmt.Errorf("Could not open file: %s", file))
			continue
		}
		defer f.Close()
		irqs, err := parseInterrupts(f)
		if err != nil {
			acc.AddError(fmt.Errorf("Parsing %s: %s", file, err))
			continue
		}
		for _, irq := range irqs {
			tags, fields := gatherTagsFields(irq)
			acc.AddFields(measurement, fields, tags)
		}
	}
	return nil
}

func init() {
	inputs.Add("interrupts", func() telegraf.Input {
		return &Interrupts{}
	})
}

package interrupts

import (
	"bufio"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"strconv"
	"strings"
)

type Interrupts struct{}

type IRQ struct {
	ID     string
	Type   string
	Device string
	Total  int64
	Cpus   []int64
}

func NewIRQ(id string) *IRQ {
	return &IRQ{ID: id, Cpus: []int64{}}
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

func parseInterrupts(irqdata string) ([]IRQ, error) {
	var irqs []IRQ
	var cpucount int
	scanner := bufio.NewScanner(strings.NewReader(irqdata))
	ok := scanner.Scan()
	if ok {
		cpus := strings.Fields(scanner.Text())
		if cpus[0] == "CPU0" {
			cpucount = len(cpus)
		}
	} else if scanner.Err() != nil {
		return irqs, fmt.Errorf("Reading %s: %s", scanner.Text(), scanner.Err())
	}
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if !strings.HasSuffix(fields[0], ":") {
			continue
		}
		irqid := strings.TrimRight(fields[0], ":")
		irq := NewIRQ(irqid)
		irqvals := fields[1:len(fields)]
		for i := 0; i < cpucount; i++ {
			if i < len(irqvals) {
				irqval, err := strconv.ParseInt(irqvals[i], 10, 64)
				if err != nil {
					return irqs, fmt.Errorf("Unable to parse %q from %q: %s", irqvals[i], scanner.Text(), err)
				}
				irq.Cpus = append(irq.Cpus, irqval)
			}
		}
		for _, irqval := range irq.Cpus {
			irq.Total += irqval
		}
		_, err := strconv.ParseInt(irqid, 10, 64)
		if err == nil && len(fields) >= cpucount+2 {
			irq.Type = fields[cpucount+1]
			irq.Device = strings.Join(fields[cpucount+2:], " ")
		} else if len(fields) > cpucount {
			irq.Type = strings.Join(fields[cpucount+1:], " ")
		}
		irqs = append(irqs, *irq)
	}
	return irqs, nil
}

func (s *Interrupts) Gather(acc telegraf.Accumulator) error {
	data, err := ioutil.ReadFile("/proc/interrupts")
	if err != nil {
		acc.AddError(fmt.Errorf("Reading %s: %s", "/proc/interrupts", err))
	}
	irqdata := string(data)
	irqs, err := parseInterrupts(irqdata)
	if err != nil {
		acc.AddError(fmt.Errorf("Parsing %s: %s", "/proc/interrupts", err))
	}
	for _, irq := range irqs {
		tags := map[string]string{"irq": irq.ID, "type": irq.Type, "device": irq.Device}
		fields := map[string]interface{}{"total": irq.Total}
		for i := 0; i < len(irq.Cpus); i++ {
			cpu := fmt.Sprintf("CPU%d", i)
			fields[cpu] = irq.Cpus[i]
		}
		acc.AddFields("interrupts", fields, tags)
	}

	data, err = ioutil.ReadFile("/proc/softirqs")
	if err != nil {
		acc.AddError(fmt.Errorf("Reading %s: %s", "/proc/softirqs", err))
	}
	irqdata = string(data)
	irqs, err = parseInterrupts(irqdata)
	if err != nil {
		acc.AddError(fmt.Errorf("Parsing %s: %s", "/proc/softirqs", err))
	}
	for _, irq := range irqs {
		tags := map[string]string{"irq": irq.ID, "type": irq.Type, "device": irq.Device}
		fields := map[string]interface{}{"total": irq.Total}
		for i := 0; i < len(irq.Cpus); i++ {
			cpu := fmt.Sprintf("CPU%d", i)
			fields[cpu] = irq.Cpus[i]
		}
		acc.AddFields("soft_interrupts", fields, tags)
	}
	return nil
}

func init() {
	inputs.Add("interrupts", func() telegraf.Input {
		return &Interrupts{}
	})
}

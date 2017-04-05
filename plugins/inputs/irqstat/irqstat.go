package irqstat

import (
	"bufio"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"strconv"
	"strings"
)

type Irqstat struct {
	Include []string
}

type IRQ struct {
	ID     string
	Type   string
	Device string
	Values map[string]interface{}
}

func NewIRQ(id string) *IRQ {
	return &IRQ{ID: id, Values: make(map[string]interface{})}
}

const sampleConfig = `
  ## A list of IRQs to include for metric ingestion, if not specified
  ## will default to collecting all IRQs.
  # include = ["0", "1", "30", "NET_RX"]
`

func (s *Irqstat) Description() string {
	return "This plugin gathers interrupts data from /proc/interrupts and /proc/softirqs."
}

func (s *Irqstat) SampleConfig() string {
	return sampleConfig
}

func parseInterrupts(irqdata string, include []string) ([]IRQ, error) {
	var err error
	var irqs []IRQ
	var cpucount int
	scanner := bufio.NewScanner(strings.NewReader(irqdata))
	for scanner.Scan() {
		var irqval, irqtotal int64
		var irqtype, irqdevice string
		fields := strings.Fields(scanner.Text())
		if len(fields) > 0 && fields[0] == "CPU0" {
			cpucount = len(fields)
		}
		if !strings.HasSuffix(fields[0], ":") {
			continue
		}
		irqid := strings.TrimRight(fields[0], ":")
		irq := NewIRQ(irqid)
		_, err := strconv.ParseInt(irqid, 10, 64)
		if err == nil && len(fields) > cpucount {
			irqtype = fields[cpucount+1]
			irqdevice = strings.Join(fields[cpucount+2:], " ")
		} else if len(fields) > cpucount {
			irqtype = strings.Join(fields[cpucount+1:], " ")
		}
		fields = fields[1:len(fields)]
		for i := 0; i < cpucount; i++ {
			cpu := fmt.Sprintf("CPU%d", i)
			if i < len(fields) {
				irqval, err = strconv.ParseInt(fields[i], 10, 64)
				if err != nil {
					return irqs, err
				}
				irq.Values[cpu] = irqval
			}
			irqtotal = irqval + irqtotal
		}
		irq.Type = irqtype
		irq.Device = irqdevice
		irq.Values["total"] = irqtotal
		if len(include) == 0 || stringInSlice(irq.ID, include) {
			irqs = append(irqs, *irq)
		}
	}
	return irqs, err
}

func stringInSlice(x string, list []string) bool {
	for _, y := range list {
		if y == x {
			return true
		}
	}
	return false
}

func (s *Irqstat) Gather(acc telegraf.Accumulator) error {
	files := []string{"/proc/interrupts", "/proc/softirqs"}
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			acc.AddError(err)
		}
		irqdata := string(data)
		irqs, err := parseInterrupts(irqdata, s.Include)
		if err != nil {
			acc.AddError(err)
		}
		for _, irq := range irqs {
			tags := map[string]string{"irq": irq.ID, "type": irq.Type, "device": irq.Device}
			fields := irq.Values
			if file == "/proc/softirqs" {
				acc.AddFields("soft_interrupts", fields, tags)
			} else {
				acc.AddFields("interrupts", fields, tags)
			}
		}
	}
	return nil
}

func init() {
	inputs.Add("irqstat", func() telegraf.Input {
		return &Irqstat{}
	})
}

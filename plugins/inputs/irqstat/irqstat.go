package irqstat

import (
	"bufio"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type Irqstat struct {
	Include []string
}

type IRQ struct {
	ID     string
	Fields map[string]interface{}
	Tags   map[string]string
}

func NewIRQ(id string) *IRQ {
	return &IRQ{ID: id, Fields: make(map[string]interface{}), Tags: make(map[string]string)}
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

func parseInterrupts(irqdata string, include []string) []IRQ {
	var irqs []IRQ
	var cpucount int
	scanner := bufio.NewScanner(strings.NewReader(irqdata))
	for scanner.Scan() {
		var irqval, irqtotal int64
		var irqtype, irqdevice string
		fields := strings.Fields(scanner.Text())
		ff := fields[0]
		if ff == "CPU0" {
			cpucount = len(fields)
		}

		if ff[len(ff)-1:] == ":" {
			fields = fields[1:len(fields)]
			irqid := ff[:len(ff)-1]
			irq := NewIRQ(irqid)
			_, err := strconv.ParseInt(irqid, 10, 64)
			if err == nil {
				irqtype = fields[cpucount]
				irqdevice = strings.Join(fields[cpucount+1:], " ")
			} else {
				if len(fields) > cpucount {
					irqtype = strings.Join(fields[cpucount:], " ")
				}
			}
			for i := 0; i < cpucount; i++ {
				cpu := fmt.Sprintf("CPU%d", i)
				irq.Tags["irq"] = irqid
				if i < len(fields) {
					irqval, err = strconv.ParseInt(fields[i], 10, 64)
					if err != nil {
						log.Fatal(err)
					}
					irq.Fields[cpu] = irqval
				}
				irqtotal = irqval + irqtotal
			}
			irq.Tags["type"] = irqtype
			irq.Tags["device"] = irqdevice
			irq.Fields["total"] = irqtotal
			if len(include) == 0 {
				irqs = append(irqs, *irq)
			} else {
				if stringInSlice(irq.ID, include) {
					irqs = append(irqs, *irq)
				}
			}
		}
	}
	return irqs
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
			log.Fatal(err)
		}
		irqdata := string(data)
		irqs := parseInterrupts(irqdata, s.Include)
		for _, irq := range irqs {
			if file == "/proc/softirqs" {
				acc.AddFields("soft_interrupts", irq.Fields, irq.Tags)
			} else {
				acc.AddFields("interrupts", irq.Fields, irq.Tags)
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

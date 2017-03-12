package irqstat

import (
	"bufio"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"log"
	"os"
	"strconv"
	"strings"
)

type Irqstat struct {
	Include []string
	Irqmap  map[string]map[string]interface{}
}

func NewIrqstat() *Irqstat {
	return &Irqstat{
		Irqmap: make(map[string]map[string]interface{}),
	}
}

const sampleConfig = `
  ## A list of IRQs to include for metric ingestion, if not specified
  ## will default to collecting all IRQs.
  # include = ["0", "1"]
`

func (s *Irqstat) Description() string {
	return "This plugin gathers IRQ types and associated values from /proc/interrupts and /proc/softirqs for each CPU."
}

func (s *Irqstat) SampleConfig() string {
	return sampleConfig
}

func (s *Irqstat) ParseIrqFile(path string) {
	var cpucount int
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var irqval int64
		var irqtotal int64
		irqdesc := "none"
		irqdevice := "none"
		fields := strings.Fields(scanner.Text())
		ff := fields[0]
		if ff == "CPU0" {
			cpucount = len(fields)
		}

		if ff[len(ff)-1:] == ":" {
			fields = fields[1:len(fields)]
			irqtype := ff[:len(ff)-1]
			if path == "/proc/softirqs" {
				irqtype = irqtype + "_softirq"
			}
			_, err := strconv.ParseInt(irqtype, 10, 64)
			if err == nil {
				irqdesc = fields[cpucount]
				irqdevice = strings.Join(fields[cpucount+1:], " ")
			} else {
				if len(fields) > cpucount {
					irqdesc = strings.Join(fields[cpucount:], " ")
				}
			}
			for i := 0; i < cpucount; i++ {
				cpukey := fmt.Sprintf("CPU%d", i)
				if s.Irqmap[irqtype] == nil {
					s.Irqmap[irqtype] = make(map[string]interface{})
				}
				irqval = 0
				if i < len(fields) {
					irqval, err = strconv.ParseInt(fields[i], 10, 64)
					if err != nil {
						log.Fatal(err)
					}
				}
				s.Irqmap[irqtype][cpukey] = irqval
				irqtotal = irqval + irqtotal
			}
			s.Irqmap[irqtype]["type"] = irqdesc
			s.Irqmap[irqtype]["device"] = irqdevice
			s.Irqmap[irqtype]["total"] = irqtotal
		}
	}
	file.Close()
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
	irqtags := make(map[string]string)
	irqfields := make(map[string]interface{})
	files := []string{"/proc/interrupts", "/proc/softirqs"}
	for _, file := range files {
		s.ParseIrqFile(file)
	}
	for irq, fields := range s.Irqmap {
		irqtype := strings.Split(irq, "_softirq")[0]
		irqtags["irq"] = irqtype
		for k, v := range fields {
			switch t := v.(type) {
			case int64:
				irqfields[k] = t
			case string:
				irqtags[k] = t
			}
		}
		for k, _ := range irqtags {
			if irqtags[k] == "none" {
				delete(irqtags, k)
			}
		}
		if len(s.Include) == 0 {
			if strings.HasSuffix(irq, "_softirq") {
				acc.AddFields("soft_interrupts", irqfields, irqtags)
			} else {
				acc.AddFields("interrupts", irqfields, irqtags)
			}
		} else {
			if stringInSlice(irqtype, s.Include) {
				if strings.HasSuffix(irq, "_softirq") {
					acc.AddFields("soft_interrupts", irqfields, irqtags)
				} else {
					acc.AddFields("interrupts", irqfields, irqtags)
				}
			}
		}
	}
	return nil
}

func init() {
	inputs.Add("irqstat", func() telegraf.Input {
		return NewIrqstat()
	})
}

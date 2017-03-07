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
	Path    string
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
  #
  ## The location of the interrupts file, defaults to /proc/interrupts.
  # path = "/some/path/interrupts"
`

func (s *Irqstat) Description() string {
	return "This plugin gathers IRQ types and associated values from /proc/interrupts for each CPU."
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
		fields := strings.Fields(scanner.Text())
		ff := fields[0]

		if ff == "CPU0" {
			cpucount = len(fields)
		}

		if ff[len(ff)-1:] == ":" {
			irqtype := ff[:len(ff)-1]
			fields = fields[1:len(fields)]
			for i := 0; i < cpucount; i++ {
				cpukey := fmt.Sprintf("CPU%d", i)

				if s.Irqmap[cpukey] == nil {
					s.Irqmap[cpukey] = make(map[string]interface{})
				}

				irqval := 0 // Default an IRQ's value to 0
				if i < len(fields) {
					irqval, err = strconv.Atoi(fields[i])
					if err != nil {
						log.Fatal(err)
					}
				}
				s.Irqmap[cpukey][irqtype] = irqval
			}
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

func (s *Irqstat) ParseIrqMap(include []string) {
	for cpukey, irqfields := range s.Irqmap {
		s.Irqmap[cpukey] = make(map[string]interface{})
		for irqtype, irqval := range irqfields {
			if stringInSlice(irqtype, include) {
				s.Irqmap[cpukey][irqtype] = irqval
			}
		}
	}
}

func (s *Irqstat) Gather(acc telegraf.Accumulator) error {
	cputags := make(map[string]string)
	path := s.Path
	include := s.Include

	if len(path) == 0 {
		path = "/proc/interrupts"
	}

	if len(include) == 0 {
		s.ParseIrqFile(path)
	} else {
		s.ParseIrqFile(path)
		s.ParseIrqMap(include)
	}

	for cpukey, irqfields := range s.Irqmap {
		cputags["cpu"] = cpukey
		acc.AddFields("irqstat", irqfields, cputags)
	}
	return nil
}

func init() {
	inputs.Add("irqstat", func() telegraf.Input {
		return NewIrqstat()
	})
}

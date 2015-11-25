package procstat

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/shirou/gopsutil/process"

	"github.com/influxdb/telegraf/plugins"
)

type Specification struct {
	PidFile string `toml:"pid_file"`
	Exe     string
	Prefix  string
	Pattern string
}

type Procstat struct {
	Specifications []*Specification
}

func NewProcstat() *Procstat {
	return &Procstat{}
}

var sampleConfig = `
  [[procstat.specifications]]
  prefix = "" # optional string to prefix measurements
  # Must specify one of: pid_file, exe, or pattern
  # PID file to monitor process
  pid_file = "/var/run/nginx.pid"
  # executable name (ie, pgrep <exe>)
  # exe = "nginx"
  # pattern as argument for pgrep (ie, pgrep -f <pattern>)
  # pattern = "nginx"
`

func (_ *Procstat) SampleConfig() string {
	return sampleConfig
}

func (_ *Procstat) Description() string {
	return "Monitor process cpu and memory usage"
}

func (p *Procstat) Gather(acc plugins.Accumulator) error {
	var wg sync.WaitGroup

	for _, specification := range p.Specifications {
		wg.Add(1)
		go func(spec *Specification, acc plugins.Accumulator) {
			defer wg.Done()
			procs, err := spec.createProcesses()
			if err != nil {
				log.Printf("Error: procstat getting process, exe: [%s] pidfile: [%s] pattern: [%s] %s",
					spec.Exe, spec.PidFile, spec.Pattern, err.Error())
			} else {
				for _, proc := range procs {
					p := NewSpecProcessor(spec.Prefix, acc, proc)
					p.pushMetrics()
				}
			}
		}(specification, acc)
	}
	wg.Wait()

	return nil
}

func (spec *Specification) createProcesses() ([]*process.Process, error) {
	var out []*process.Process
	var errstring string
	var outerr error

	pids, err := spec.getAllPids()
	if err != nil {
		errstring += err.Error() + " "
	}

	for _, pid := range pids {
		p, err := process.NewProcess(int32(pid))
		if err == nil {
			out = append(out, p)
		} else {
			errstring += err.Error() + " "
		}
	}

	if errstring != "" {
		outerr = fmt.Errorf("%s", errstring)
	}

	return out, outerr
}

func (spec *Specification) getAllPids() ([]int32, error) {
	var pids []int32
	var err error

	if spec.PidFile != "" {
		pids, err = pidsFromFile(spec.PidFile)
	} else if spec.Exe != "" {
		pids, err = pidsFromExe(spec.Exe)
	} else if spec.Pattern != "" {
		pids, err = pidsFromPattern(spec.Pattern)
	} else {
		err = fmt.Errorf("Either exe, pid_file or pattern has to be specified")
	}

	return pids, err
}

func pidsFromFile(file string) ([]int32, error) {
	var out []int32
	var outerr error
	pidString, err := ioutil.ReadFile(file)
	if err != nil {
		outerr = fmt.Errorf("Failed to read pidfile '%s'. Error: '%s'", file, err)
	} else {
		pid, err := strconv.Atoi(strings.TrimSpace(string(pidString)))
		if err != nil {
			outerr = err
		} else {
			out = append(out, int32(pid))
		}
	}
	return out, outerr
}

func pidsFromExe(exe string) ([]int32, error) {
	var out []int32
	var outerr error
	pgrep, err := exec.Command("pgrep", exe).Output()
	if err != nil {
		return out, fmt.Errorf("Failed to execute pgrep. Error: '%s'", err)
	} else {
		pids := strings.Fields(string(pgrep))
		for _, pid := range pids {
			ipid, err := strconv.Atoi(pid)
			if err == nil {
				out = append(out, int32(ipid))
			} else {
				outerr = err
			}
		}
	}
	return out, outerr
}

func pidsFromPattern(pattern string) ([]int32, error) {
	var out []int32
	var outerr error
	pgrep, err := exec.Command("pgrep", "-f", pattern).Output()
	if err != nil {
		return out, fmt.Errorf("Failed to execute pgrep. Error: '%s'", err)
	} else {
		pids := strings.Fields(string(pgrep))
		for _, pid := range pids {
			ipid, err := strconv.Atoi(pid)
			if err == nil {
				out = append(out, int32(ipid))
			} else {
				outerr = err
			}
		}
	}
	return out, outerr
}

func init() {
	plugins.Add("procstat", func() plugins.Plugin {
		return NewProcstat()
	})
}

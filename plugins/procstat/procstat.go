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
	# Use one of pid_file or exe to find process
	pid_file = "/var/run/nginx.pid"
	# executable name (used by pgrep)
	# exe = "nginx"
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
				log.Printf("Error: procstat getting process, exe: %s, pidfile: %s, %s",
					spec.Exe, spec.PidFile, err.Error())
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
	var pids []int32

	if spec.PidFile != "" {
		pid, err := pidFromFile(spec.PidFile)
		if err != nil {
			errstring += err.Error() + " "
		} else {
			pids = append(pids, int32(pid))
		}
	} else if spec.Exe != "" {
		exepids, err := pidsFromExe(spec.Exe)
		if err != nil {
			errstring += err.Error() + " "
		}
		pids = append(pids, exepids...)
	} else {
		errstring += fmt.Sprintf("Either exe or pid_file has to be specified")
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

func pidFromFile(file string) (int, error) {
	pidString, err := ioutil.ReadFile(file)
	if err != nil {
		return -1, fmt.Errorf("Failed to read pidfile '%s'. Error: '%s'", file, err)
	} else {
		return strconv.Atoi(strings.TrimSpace(string(pidString)))
	}
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

func init() {
	plugins.Add("procstat", func() plugins.Plugin {
		return NewProcstat()
	})
}

package procstat

import (
	"fmt"
	"github.com/influxdb/telegraf/plugins"
	"github.com/shirou/gopsutil/process"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type Specification struct {
	PidFile string `toml:pid_file`
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
  [[process.specifications]]
	# pid file
	pid_file = "/path/to/foo.pid"
	# executable name (used by pgrep)
	exe = "/path/to/foo"
	name = "foo" # required
`

func (_ *Procstat) SampleConfig() string {
	return sampleConfig
}

func (_ *Procstat) Description() string {
	return "Monitor  process cpu and memory usage"
}

func (p *Procstat) Gather(acc plugins.Accumulator) error {
	var wg sync.WaitGroup
	var outerr error
	for _, specification := range p.Specifications {
		wg.Add(1)
		go func(spec *Specification, acc plugins.Accumulator) {
			defer wg.Done()
			proc, err := spec.createProcess()
			if err != nil {
				outerr = err
			} else {
				outerr = NewSpecProcessor(spec.Prefix, acc, proc).pushMetrics()
			}
		}(specification, acc)
	}
	wg.Wait()
	return outerr
}

func (spec *Specification) createProcess() (*process.Process, error) {
	if spec.PidFile != "" {
		pid, err := pidFromFile(spec.PidFile)
		if err != nil {
			return nil, err
		}
		return process.NewProcess(int32(pid))
	} else if spec.Exe != "" {
		pid, err := pidFromExe(spec.Exe)
		if err != nil {
			return nil, err
		}
		return process.NewProcess(int32(pid))
	} else {
		return nil, fmt.Errorf("Either exe or pid_file has to be specified")
	}
}

func pidFromFile(file string) (int, error) {
	pidString, err := ioutil.ReadFile(file)
	if err != nil {
		return -1, fmt.Errorf("Failed to read pidfile '%s'. Error: '%s'", file, err)
	} else {
		return strconv.Atoi(strings.TrimSpace(string(pidString)))
	}
}

func pidFromExe(exe string) (int, error) {
	pidString, err := exec.Command("pgrep", exe).Output()
	if err != nil {
		return -1, fmt.Errorf("Failed to execute pgrep. Error: '%s'", err)
	} else {
		return strconv.Atoi(strings.TrimSpace(string(pidString)))
	}
}

func init() {
	plugins.Add("process", func() plugins.Plugin {
		return NewProcstat()
	})
}

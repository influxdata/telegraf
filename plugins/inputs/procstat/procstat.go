package procstat

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/process"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Procstat struct {
	PidFile     string `toml:"pid_file"`
	Exe         string
	Pattern     string
	Prefix      string
	ProcessName string
	User        string

	// pidmap maps a pid to a process object, so we don't recreate every gather
	pidmap map[int32]*process.Process
	// tagmap maps a pid to a map of tags for that pid
	tagmap map[int32]map[string]string
}

func NewProcstat() *Procstat {
	return &Procstat{
		pidmap: make(map[int32]*process.Process),
		tagmap: make(map[int32]map[string]string),
	}
}

var sampleConfig = `
  ## Must specify one of: pid_file, exe, or pattern
  ## PID file to monitor process
  pid_file = "/var/run/nginx.pid"
  ## executable name (ie, pgrep <exe>)
  # exe = "nginx"
  ## pattern as argument for pgrep (ie, pgrep -f <pattern>)
  # pattern = "nginx"
  ## user as argument for pgrep (ie, pgrep -u <user>)
  # user = "nginx"

  ## override for process_name
  ## This is optional; default is sourced from /proc/<pid>/status
  # process_name = "bar"
  ## Field name prefix
  prefix = ""
  ## comment this out if you want raw cpu_time stats
  fielddrop = ["cpu_time_*"]
`

func (_ *Procstat) SampleConfig() string {
	return sampleConfig
}

func (_ *Procstat) Description() string {
	return "Monitor process cpu and memory usage"
}

func (p *Procstat) Gather(acc telegraf.Accumulator) error {
	err := p.createProcesses()
	if err != nil {
		log.Printf("Error: procstat getting process, exe: [%s]	pidfile: [%s] pattern: [%s] user: [%s] %s",
			p.Exe, p.PidFile, p.Pattern, p.User, err.Error())
	} else {
		for pid, proc := range p.pidmap {
			p := NewSpecProcessor(p.ProcessName, p.Prefix, pid, acc, proc, p.tagmap[pid])
			p.pushMetrics()
		}
	}

	return nil
}

func (p *Procstat) createProcesses() error {
	var errstring string
	var outerr error

	pids, err := p.getAllPids()
	if err != nil {
		errstring += err.Error() + " "
	}

	for _, pid := range pids {
		_, ok := p.pidmap[pid]
		if !ok {
			proc, err := process.NewProcess(pid)
			if err == nil {
				p.pidmap[pid] = proc
			} else {
				errstring += err.Error() + " "
			}
		}
	}

	if errstring != "" {
		outerr = fmt.Errorf("%s", errstring)
	}

	return outerr
}

func (p *Procstat) getAllPids() ([]int32, error) {
	var pids []int32
	var err error

	if p.PidFile != "" {
		pids, err = p.pidsFromFile()
	} else if p.Exe != "" {
		pids, err = p.pidsFromExe()
	} else if p.Pattern != "" {
		pids, err = p.pidsFromPattern()
	} else if p.User != "" {
		pids, err = p.pidsFromUser()
	} else {
		err = fmt.Errorf("Either exe, pid_file, user, or pattern has to be specified")
	}

	return pids, err
}

func (p *Procstat) pidsFromFile() ([]int32, error) {
	var out []int32
	var outerr error
	pidString, err := ioutil.ReadFile(p.PidFile)
	if err != nil {
		outerr = fmt.Errorf("Failed to read pidfile '%s'. Error: '%s'",
			p.PidFile, err)
	} else {
		pid, err := strconv.Atoi(strings.TrimSpace(string(pidString)))
		if err != nil {
			outerr = err
		} else {
			out = append(out, int32(pid))
			p.tagmap[int32(pid)] = map[string]string{
				"pidfile": p.PidFile,
			}
		}
	}
	return out, outerr
}

func (p *Procstat) pidsFromExe() ([]int32, error) {
	var out []int32
	var outerr error
	bin, err := exec.LookPath("pgrep")
	if err != nil {
		return out, fmt.Errorf("Couldn't find pgrep binary: %s", err)
	}
	pgrep, err := exec.Command(bin, p.Exe).Output()
	if err != nil {
		return out, fmt.Errorf("Failed to execute %s. Error: '%s'", bin, err)
	} else {
		pids := strings.Fields(string(pgrep))
		for _, pid := range pids {
			ipid, err := strconv.Atoi(pid)
			if err == nil {
				out = append(out, int32(ipid))
				p.tagmap[int32(ipid)] = map[string]string{
					"exe": p.Exe,
				}
			} else {
				outerr = err
			}
		}
	}
	return out, outerr
}

func (p *Procstat) pidsFromPattern() ([]int32, error) {
	var out []int32
	var outerr error
	bin, err := exec.LookPath("pgrep")
	if err != nil {
		return out, fmt.Errorf("Couldn't find pgrep binary: %s", err)
	}
	pgrep, err := exec.Command(bin, "-f", p.Pattern).Output()
	if err != nil {
		return out, fmt.Errorf("Failed to execute %s. Error: '%s'", bin, err)
	} else {
		pids := strings.Fields(string(pgrep))
		for _, pid := range pids {
			ipid, err := strconv.Atoi(pid)
			if err == nil {
				out = append(out, int32(ipid))
				p.tagmap[int32(ipid)] = map[string]string{
					"pattern": p.Pattern,
				}
			} else {
				outerr = err
			}
		}
	}
	return out, outerr
}

func (p *Procstat) pidsFromUser() ([]int32, error) {
	var out []int32
	var outerr error
	bin, err := exec.LookPath("pgrep")
	if err != nil {
		return out, fmt.Errorf("Couldn't find pgrep binary: %s", err)
	}
	pgrep, err := exec.Command(bin, "-u", p.User).Output()
	if err != nil {
		return out, fmt.Errorf("Failed to execute %s. Error: '%s'", bin, err)
	} else {
		pids := strings.Fields(string(pgrep))
		for _, pid := range pids {
			ipid, err := strconv.Atoi(pid)
			if err == nil {
				out = append(out, int32(ipid))
				p.tagmap[int32(ipid)] = map[string]string{
					"user": p.User,
				}
			} else {
				outerr = err
			}
		}
	}
	return out, outerr
}

func init() {
	inputs.Add("procstat", func() telegraf.Input {
		return NewProcstat()
	})
}

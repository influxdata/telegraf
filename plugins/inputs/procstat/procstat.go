package procstat

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"

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
	PidTag      bool

	// source is how we got the info - pidfile, exe, etc.
	// we keep it to add a tag
	sourceKey   string
	sourceValue string
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
  ## This is optional; moves pid into a tag instead of a field
  pid_tag = false
`

func (_ *Procstat) SampleConfig() string {
	return sampleConfig
}

func (_ *Procstat) Description() string {
	return "Monitor process cpu and memory usage"
}

func (p *Procstat) Gather(acc telegraf.Accumulator) error {
	pids, err := p.getAllPids()
	if err != nil {
		log.Printf("E! Error: procstat getting process, exe: [%s] pidfile: [%s] pattern: [%s] user: [%s] %s",
			p.Exe, p.PidFile, p.Pattern, p.User, err.Error())
	} else {
		for _, pid := range pids {
			tags := map[string]string{
				p.sourceKey: p.sourceValue}
			if p.PidTag {
				tags["pid"] = fmt.Sprint(pid)
			}
			p := NewSpecProcessor(p.ProcessName, p.Prefix, pid, acc, tags)
			err := p.pushMetrics()
			if err != nil {
				log.Printf("E! Error: procstat: %s", err.Error())
			}
		}
	}

	return nil
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

func pidFromList(pids []string) ([]int32, error) {
	var out []int32
	for _, pid := range pids {
		ipid, err := strconv.Atoi(pid)
		if err == nil {
			out = append(out, int32(ipid))
		} else {
			return nil, err
		}
	}
	return out, nil
}

func (p *Procstat) pidsFromFile() ([]int32, error) {
	p.sourceKey = "pidfile"
	p.sourceValue = p.PidFile
	pidString, err := ioutil.ReadFile(p.PidFile)
	if err != nil {
		outerr := fmt.Errorf("Failed to read pidfile '%s'. Error: '%s'",
			p.PidFile, err)
		return nil, outerr
	}
	pids := strings.Fields(string(pidString))
	out, err := pidFromList(pids)
	return out, err
}

func (p *Procstat) pidsFromExe() ([]int32, error) {
	p.sourceKey = "exe"
	p.sourceValue = p.Exe
	bin, err := exec.LookPath("pgrep")
	if err != nil {
		return nil, fmt.Errorf("Couldn't find pgrep binary: %s", err)
	}
	pgrep, err := exec.Command(bin, p.Exe).Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to execute %s. Error: '%s'", bin, err)
	}
	pids := strings.Fields(string(pgrep))
	out, err := pidFromList(pids)
	return out, err
}

func (p *Procstat) pidsFromPattern() ([]int32, error) {
	p.sourceKey = "pattern"
	p.sourceValue = p.Pattern
	bin, err := exec.LookPath("pgrep")
	if err != nil {
		return nil, fmt.Errorf("Couldn't find pgrep binary: %s", err)
	}
	pgrep, err := exec.Command(bin, "-f", p.Pattern).Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to execute %s. Error: '%s'", bin, err)
	}
	pids := strings.Fields(string(pgrep))
	out, err := pidFromList(pids)
	return out, err
}

func (p *Procstat) pidsFromUser() ([]int32, error) {
	p.sourceKey = "user"
	p.sourceValue = p.User
	bin, err := exec.LookPath("pgrep")
	if err != nil {
		return nil, fmt.Errorf("Couldn't find pgrep binary: %s", err)
	}
	pgrep, err := exec.Command(bin, "-u", p.User).Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to execute %s. Error: '%s'", bin, err)
	}
	pids := strings.Fields(string(pgrep))
	out, err := pidFromList(pids)
	return out, err
}

func init() {
	inputs.Add("procstat", func() telegraf.Input {
		return &Procstat{}
	})
}

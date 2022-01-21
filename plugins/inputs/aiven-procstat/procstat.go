package aiven_procstat

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/shirou/gopsutil/v3/process"
)

var (
	defaultPIDFinder = NewPgrep
	defaultProcess   = NewProc
)

type PID int32

type Procstat struct {
	PidFinder    string `toml:"pid_finder"`
	PidFile      string `toml:"pid_file"`
	Exe          string
	Pattern      string
	Prefix       string
	CmdLineTag   bool `toml:"cmdline_tag"`
	ProcessName  string
	User         string
	SystemdUnit  string
	SystemdUnits []string
	CGroup       string `toml:"cgroup"`
	PidTag       bool
	WinService   string `toml:"win_service"`

	finder PIDFinder

	createPIDFinder func() (PIDFinder, error)
	procs           map[PID]Process
	createProcess   func(PID) (Process, error)
}

var sampleConfig = `
  ## PID file to monitor process
  pid_file = "/var/run/nginx.pid"
  ## executable name (ie, pgrep <exe>)
  # exe = "nginx"
  ## pattern as argument for pgrep (ie, pgrep -f <pattern>)
  # pattern = "nginx"
  ## user as argument for pgrep (ie, pgrep -u <user>)
  # user = "nginx"
  ## Systemd unit name. Use systemd_units when getting metrics
  ## for several units.
  # systemd_unit = "nginx.service"
  ## Systemd unit name array
  # systemd_units = ["nginx.service", "haproxy.service"]
  ## CGroup name or path
  # cgroup = "systemd/system.slice/nginx.service"

  ## Windows service name
  # win_service = ""

  ## override for process_name
  ## This is optional; default is sourced from /proc/<pid>/status
  # process_name = "bar"

  ## Field name prefix
  # prefix = ""

  ## When true add the full cmdline as a tag.
  # cmdline_tag = false

  ## Add the PID as a tag instead of as a field.  When collecting multiple
  ## processes with otherwise matching tags this setting should be enabled to
  ## ensure each process has a unique identity.
  ##
  ## Enabling this option may result in a large number of series, especially
  ## when processes have a short lifetime.
  # pid_tag = false

  ## Method to use when finding process IDs.  Can be one of 'pgrep', or
  ## 'native'.  The pgrep finder calls the pgrep executable in the PATH while
  ## the native finder performs the search directly in a manor dependent on the
  ## platform.  Default is 'pgrep'
  # pid_finder = "pgrep"
`

func (_ *Procstat) SampleConfig() string {
	return sampleConfig
}

func (_ *Procstat) Description() string {
	return "Monitor process cpu and memory usage"
}

func (p *Procstat) Gather(acc telegraf.Accumulator) error {
	if p.createPIDFinder == nil {
		switch p.PidFinder {
		case "native":
			p.createPIDFinder = NewNativeFinder
		case "pgrep":
			p.createPIDFinder = NewPgrep
		default:
			p.PidFinder = "pgrep"
			p.createPIDFinder = defaultPIDFinder
		}

	}
	if p.createProcess == nil {
		p.createProcess = defaultProcess
	}

	procs, err := p.updateProcesses(acc, p.procs)
	if err != nil {
		acc.AddError(fmt.Errorf("E! Error: procstat getting process, exe: [%s] pidfile: [%s] pattern: [%s] user: [%s] %s",
			p.Exe, p.PidFile, p.Pattern, p.User, err.Error()))
	}
	p.procs = procs

	for _, proc := range p.procs {
		p.addMetrics(proc, acc)
	}

	return nil
}

// Add metrics a single Process
func (p *Procstat) addMetrics(proc Process, acc telegraf.Accumulator) {
	var prefix string
	if p.Prefix != "" {
		prefix = p.Prefix + "_"
	}

	fields := map[string]interface{}{}

	//If process_name tag is not already set, set to actual name
	if _, nameInTags := proc.Tags()["process_name"]; !nameInTags {
		name, err := proc.Name()
		if err == nil {
			proc.Tags()["process_name"] = name
		}
	}

	//If user tag is not already set, set to actual name
	if _, ok := proc.Tags()["user"]; !ok {
		user, err := proc.Username()
		if err == nil {
			proc.Tags()["user"] = user
		}
	}

	//If pid is not present as a tag, include it as a field.
	if _, pidInTags := proc.Tags()["pid"]; !pidInTags {
		fields["pid"] = int32(proc.PID())
	}

	//If cmd_line tag is true and it is not already set add cmdline as a tag
	if p.CmdLineTag {
		if _, ok := proc.Tags()["cmdline"]; !ok {
			Cmdline, err := proc.Cmdline()
			if err == nil {
				proc.Tags()["cmdline"] = Cmdline
			}
		}
	}

	numThreads, err := proc.NumThreads()
	if err == nil {
		fields[prefix+"num_threads"] = numThreads
	}

	fds, err := proc.NumFDs()
	if err == nil {
		fields[prefix+"num_fds"] = fds
	}

	ctx, err := proc.NumCtxSwitches()
	if err == nil {
		fields[prefix+"voluntary_context_switches"] = ctx.Voluntary
		fields[prefix+"involuntary_context_switches"] = ctx.Involuntary
	}

	io, err := proc.IOCounters()
	if err == nil {
		fields[prefix+"read_count"] = io.ReadCount
		fields[prefix+"write_count"] = io.WriteCount
		fields[prefix+"read_bytes"] = io.ReadBytes
		fields[prefix+"write_bytes"] = io.WriteBytes
	}

	createdAt, err := proc.CreateTime() //Returns epoch in ms
	if err == nil {
		fields[prefix+"created_at"] = createdAt * 1000000 //Convert ms to ns
	}

	cpu_time, err := proc.Times()
	if err == nil {
		fields[prefix+"cpu_time_user"] = cpu_time.User
		fields[prefix+"cpu_time_system"] = cpu_time.System
		fields[prefix+"cpu_time_idle"] = cpu_time.Idle
		fields[prefix+"cpu_time_nice"] = cpu_time.Nice
		fields[prefix+"cpu_time_iowait"] = cpu_time.Iowait
		fields[prefix+"cpu_time_irq"] = cpu_time.Irq
		fields[prefix+"cpu_time_soft_irq"] = cpu_time.Softirq
		fields[prefix+"cpu_time_steal"] = cpu_time.Steal
		fields[prefix+"cpu_time_guest"] = cpu_time.Guest
		fields[prefix+"cpu_time_guest_nice"] = cpu_time.GuestNice
	}

	cpu_perc, err := proc.Percent(time.Duration(0))
	if err == nil {
		fields[prefix+"cpu_usage"] = cpu_perc
	}

	mem, err := proc.MemoryInfo()
	if err == nil {
		fields[prefix+"memory_rss"] = mem.RSS
		fields[prefix+"memory_vms"] = mem.VMS
		fields[prefix+"memory_swap"] = mem.Swap
		fields[prefix+"memory_data"] = mem.Data
		fields[prefix+"memory_stack"] = mem.Stack
		fields[prefix+"memory_locked"] = mem.Locked
	}

	mem_perc, err := proc.MemoryPercent()
	if err == nil {
		fields[prefix+"memory_usage"] = mem_perc
	}

	rlims, err := proc.RlimitUsage(true)
	if err == nil {
		for _, rlim := range rlims {
			var name string
			switch rlim.Resource {
			case process.RLIMIT_CPU:
				name = "cpu_time"
			case process.RLIMIT_DATA:
				name = "memory_data"
			case process.RLIMIT_STACK:
				name = "memory_stack"
			case process.RLIMIT_RSS:
				name = "memory_rss"
			case process.RLIMIT_NOFILE:
				name = "num_fds"
			case process.RLIMIT_MEMLOCK:
				name = "memory_locked"
			case process.RLIMIT_AS:
				name = "memory_vms"
			case process.RLIMIT_LOCKS:
				name = "file_locks"
			case process.RLIMIT_SIGPENDING:
				name = "signals_pending"
			case process.RLIMIT_NICE:
				name = "nice_priority"
			case process.RLIMIT_RTPRIO:
				name = "realtime_priority"
			default:
				continue
			}

			fields[prefix+"rlimit_"+name+"_soft"] = rlim.Soft
			fields[prefix+"rlimit_"+name+"_hard"] = rlim.Hard
			if name != "file_locks" { // gopsutil doesn't currently track the used file locks count
				fields[prefix+name] = rlim.Used
			}
		}
	}

	acc.AddFields("procstat", fields, proc.Tags())
}

// Update monitored Processes
func (p *Procstat) updateProcesses(acc telegraf.Accumulator, prevInfo map[PID]Process) (map[PID]Process, error) {
	pidsArray, tagsArray, err := p.findPids()
	if err != nil {
		return nil, err
	}

	procs := make(map[PID]Process, len(prevInfo))

	for index, pids := range pidsArray {
		tags := tagsArray[index]

		finderTags := make(map[string]string)
		for k, v := range tags {
			finderTags[k] = v
		}

		// Add metric for the number of matched pids
		fields := make(map[string]interface{})
		finderTags["pid_finder"] = p.PidFinder
		fields["pid_count"] = len(pids)
		acc.AddFields("procstat_lookup", fields, tags)

		for _, pid := range pids {
			info, ok := prevInfo[pid]
			if ok {
				// Assumption: if a process has no name, it probably does not exist
				if name, _ := info.Name(); name == "" {
					continue
				}
				procs[pid] = info
			} else {
				proc, err := p.createProcess(pid)
				if err != nil {
					// No problem; process may have ended after we found it
					continue
				}
				// Assumption: if a process has no name, it probably does not exist
				if name, _ := proc.Name(); name == "" {
					continue
				}
				procs[pid] = proc

				// Add initial tags
				for k, v := range tags {
					proc.Tags()[k] = v
				}

				// Add pid tag if needed
				if p.PidTag {
					proc.Tags()["pid"] = strconv.Itoa(int(pid))
				}
				if p.ProcessName != "" {
					proc.Tags()["process_name"] = p.ProcessName
				}
			}
		}
	}
	return procs, nil
}

// Create and return PIDGatherer lazily
func (p *Procstat) getPIDFinder() (PIDFinder, error) {
	if p.finder == nil {
		f, err := p.createPIDFinder()
		if err != nil {
			return nil, err
		}
		p.finder = f
	}
	return p.finder, nil
}

// Get matching PIDs and their initial tags
func (p *Procstat) findPids() ([][]PID, []map[string]string, error) {
	var pidsArray [][]PID
	var tagsArray []map[string]string
	var pids []PID
	tags := make(map[string]string)
	var err error

	f, err := p.getPIDFinder()
	if err != nil {
		return nil, nil, err
	}

	if p.PidFile != "" {
		pids, err = f.PidFile(p.PidFile)
		tags = map[string]string{"pidfile": p.PidFile}
	} else if p.Exe != "" {
		pids, err = f.Pattern(p.Exe)
		tags = map[string]string{"exe": p.Exe}
	} else if p.Pattern != "" {
		pids, err = f.FullPattern(p.Pattern)
		tags = map[string]string{"pattern": p.Pattern}
	} else if p.User != "" {
		pids, err = f.Uid(p.User)
		tags = map[string]string{"user": p.User}
	} else if p.SystemdUnit != "" {
		pidsArray, tagsArray, err = p.systemdUnitPIDs([]string{p.SystemdUnit})
	} else if p.SystemdUnits != nil {
		pidsArray, tagsArray, err = p.systemdUnitPIDs(p.SystemdUnits)
	} else if p.CGroup != "" {
		pids, err = p.cgroupPIDs()
		tags = map[string]string{"cgroup": p.CGroup}
	} else if p.WinService != "" {
		pids, err = p.winServicePIDs()
		tags = map[string]string{"win_service": p.WinService}
	} else {
		err = fmt.Errorf("Either exe, pid_file, user, pattern, systemd_unit, systemd_units, or cgroup must be specified")
	}

	if pids != nil {
		pidsArray = [][]PID{pids}
		tagsArray = []map[string]string{tags}
	}

	return pidsArray, tagsArray, err
}

// execCommand is so tests can mock out exec.Command usage.
var execCommand = exec.Command

func (p *Procstat) systemdUnitPIDs(units []string) ([][]PID, []map[string]string, error) {
	var pidsArray [][]PID
	var tagsArray []map[string]string
	var pids []PID
	var tags map[string]string
	// Lines with PID look like "  ├─ 123 /usr/bin/foo" or "  └─4567 /usr/bin/bar"
	// (possibly with some non-whitespace leading characters)
	pidMatcher, err := regexp.Compile(`.*?[├└]─\s*(\d+)\s+\S+.*`)
	if err != nil {
		return nil, nil, err
	}
	// Use systemctl status and parse the pids from there. This provides output that is
	// slightly more tedious to parse than "systemctl show <unit>" output but this allows
	// getting all pids for the unit and it works for systemd containers. Also, when
	// several systemd units are defined this performs much better as we can make do with
	// just a single systemctl invocation instead of one per unit.
	cmd := execCommand("systemctl", "status")
	out, err := cmd.Output()
	if err != nil {
		return nil, nil, err
	}
LINES:
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		line := bytes.TrimRight(line, "\t ")
		for _, unit := range units {
			// Lines with unit name look like " ├─foo.service" or "  └─bar.service"
			unitWithSuffix := unit
			if !strings.HasSuffix(unit, ".service") {
				unitWithSuffix = unit + ".service"
			}
			if bytes.HasSuffix(line, []byte("─"+unitWithSuffix)) {
				if tags != nil && pids != nil {
					pidsArray = append(pidsArray, pids)
					tagsArray = append(tagsArray, tags)
					pids = nil
				}
				tags = map[string]string{"systemd_unit": unit}
				continue LINES
			}
		}

		matches := pidMatcher.FindSubmatch(line)
		if matches == nil {
			// This was not a pid line. Presumably new unit but not one we're interested in.
			// If we had matched some pids for a unit record those
			if tags != nil && pids != nil {
				pidsArray = append(pidsArray, pids)
				tagsArray = append(tagsArray, tags)
			}
			pids = nil
			tags = nil
		} else if tags != nil {
			// Previous line started a section for a unit we're tracking or it was another
			// pid for that unit. Get the pid from this line too.
			pid, err := strconv.ParseInt(string(matches[1]), 10, 32)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid pid '%s'", matches[1])
			}
			pids = append(pids, PID(pid))
		}
	}
	return pidsArray, tagsArray, nil
}

func (p *Procstat) cgroupPIDs() ([]PID, error) {
	var pids []PID

	procsPath := p.CGroup
	if procsPath[0] != '/' {
		procsPath = "/sys/fs/cgroup/" + procsPath
	}
	procsPath = filepath.Join(procsPath, "cgroup.procs")
	out, err := ioutil.ReadFile(procsPath)
	if err != nil {
		return nil, err
	}
	for _, pidBS := range bytes.Split(out, []byte{'\n'}) {
		if len(pidBS) == 0 {
			continue
		}
		pid, err := strconv.ParseInt(string(pidBS), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid pid '%s'", pidBS)
		}
		pids = append(pids, PID(pid))
	}

	return pids, nil
}

func (p *Procstat) winServicePIDs() ([]PID, error) {
	var pids []PID

	pid, err := queryPidWithWinServiceName(p.WinService)
	if err != nil {
		return pids, err
	}

	pids = append(pids, PID(pid))

	return pids, nil
}

func init() {
	inputs.Add("aiven-procstat", func() telegraf.Input {
		return &Procstat{}
	})
}

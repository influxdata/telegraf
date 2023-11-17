//go:generate ../../../tools/readme_config_includer/generator
package procstat

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type PID int32

type Procstat struct {
	PidFinder              string   `toml:"pid_finder"`
	PidFile                string   `toml:"pid_file"`
	Exe                    string   `toml:"exe"`
	Pattern                string   `toml:"pattern"`
	Prefix                 string   `toml:"prefix"`
	CmdLineTag             bool     `toml:"cmdline_tag"`
	ProcessName            string   `toml:"process_name"`
	User                   string   `toml:"user"`
	SystemdUnits           string   `toml:"systemd_units"`
	SupervisorUnit         []string `toml:"supervisor_unit"`
	IncludeSystemdChildren bool     `toml:"include_systemd_children"`
	CGroup                 string   `toml:"cgroup"`
	PidTag                 bool     `toml:"pid_tag"`
	WinService             string   `toml:"win_service"`
	Mode                   string

	solarisMode bool
	finder      PIDFinder
	procs       map[PID]Process

	createProcess func(PID) (Process, error)
}

type PidsTags struct {
	PIDS []PID
	Tags map[string]string
	Err  error
}

func (*Procstat) SampleConfig() string {
	return sampleConfig
}

func (p *Procstat) Init() error {
	// Check solaris mode
	p.solarisMode = strings.ToLower(p.Mode) == "solaris"

	// Instantiate the finder
	switch p.PidFinder {
	case "", "pgrep":
		p.PidFinder = "pgrep"
		finder, err := newPgrepFinder()
		if err != nil {
			return fmt.Errorf("creating pgrep finder failed: %w", err)
		}
		p.finder = finder
	case "native":
		// gopsutil relies on pgrep when looking up children on darwin
		// see https://github.com/shirou/gopsutil/blob/v3.23.10/process/process_darwin.go#L235
		requiresChildren := len(p.SupervisorUnit) > 0 && p.Pattern != ""
		if requiresChildren && runtime.GOOS == "darwin" {
			return errors.New("configuration requires the 'pgrep' finder on you OS")
		}
		p.finder = &NativeFinder{}
	default:
		return fmt.Errorf("unknown pid_finder %q", p.PidFinder)
	}

	return nil
}

func (p *Procstat) Gather(acc telegraf.Accumulator) error {
	pidCount := 0
	now := time.Now()
	newProcs := make(map[PID]Process, len(p.procs))
	tags := make(map[string]string)
	pidTags := p.findPids()
	for _, pidTag := range pidTags {
		if len(pidTag.PIDS) < 1 && len(p.SupervisorUnit) > 0 {
			continue
		}
		pids := pidTag.PIDS
		err := pidTag.Err
		pidCount += len(pids)
		for key, value := range pidTag.Tags {
			tags[key] = value
		}
		if err != nil {
			fields := map[string]interface{}{
				"pid_count":   0,
				"running":     0,
				"result_code": 1,
			}
			tags["pid_finder"] = p.PidFinder
			tags["result"] = "lookup_error"
			acc.AddFields("procstat_lookup", fields, tags, now)
			return err
		}

		p.updateProcesses(pids, pidTag.Tags, p.procs, newProcs)
	}

	p.procs = newProcs
	for _, proc := range p.procs {
		p.addMetric(proc, acc, now)
	}

	fields := map[string]interface{}{
		"pid_count":   pidCount,
		"running":     len(p.procs),
		"result_code": 0,
	}

	tags["pid_finder"] = p.PidFinder
	tags["result"] = "success"
	if len(p.SupervisorUnit) > 0 {
		tags["supervisor_unit"] = strings.Join(p.SupervisorUnit, ";")
	}
	acc.AddFields("procstat_lookup", fields, tags, now)

	return nil
}

// Add metrics a single Process
func (p *Procstat) addMetric(proc Process, acc telegraf.Accumulator, t time.Time) {
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
			cmdline, err := proc.Cmdline()
			if err == nil {
				proc.Tags()["cmdline"] = cmdline
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

	faults, err := proc.PageFaults()
	if err == nil {
		fields[prefix+"minor_faults"] = faults.MinorFaults
		fields[prefix+"major_faults"] = faults.MajorFaults
		fields[prefix+"child_minor_faults"] = faults.ChildMinorFaults
		fields[prefix+"child_major_faults"] = faults.ChildMajorFaults
	}

	io, err := proc.IOCounters()
	if err == nil {
		fields[prefix+"read_count"] = io.ReadCount
		fields[prefix+"write_count"] = io.WriteCount
		fields[prefix+"read_bytes"] = io.ReadBytes
		fields[prefix+"write_bytes"] = io.WriteBytes
	}

	createdAt, err := proc.CreateTime() // returns epoch in ms
	if err == nil {
		fields[prefix+"created_at"] = createdAt * 1000000 // ms to ns
	}

	cpuTime, err := proc.Times()
	if err == nil {
		fields[prefix+"cpu_time_user"] = cpuTime.User
		fields[prefix+"cpu_time_system"] = cpuTime.System
		fields[prefix+"cpu_time_iowait"] = cpuTime.Iowait // only reported on Linux
	}

	cpuPerc, err := proc.Percent(time.Duration(0))
	if err == nil {
		if p.solarisMode {
			fields[prefix+"cpu_usage"] = cpuPerc / float64(runtime.NumCPU())
		} else {
			fields[prefix+"cpu_usage"] = cpuPerc
		}
	}

	// This only returns values for RSS and VMS
	mem, err := proc.MemoryInfo()
	if err == nil {
		fields[prefix+"memory_rss"] = mem.RSS
		fields[prefix+"memory_vms"] = mem.VMS
	}

	collectMemmap(proc, prefix, fields)

	memPerc, err := proc.MemoryPercent()
	if err == nil {
		fields[prefix+"memory_usage"] = memPerc
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

	ppid, err := proc.Ppid()
	if err == nil {
		fields[prefix+"ppid"] = ppid
	}

	status, err := proc.Status()
	if err == nil {
		fields[prefix+"status"] = status[0]
	}

	acc.AddFields("procstat", fields, proc.Tags(), t)
}

// Update monitored Processes
func (p *Procstat) updateProcesses(pids []PID, tags map[string]string, prevInfo map[PID]Process, procs map[PID]Process) {
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

// Get matching PIDs and their initial tags
func (p *Procstat) findPids() []PidsTags {
	switch {
	case len(p.SupervisorUnit) > 0:
		return p.findSupervisorUnits()
	case p.SystemdUnits != "":
		return p.systemdUnitPIDs()
	case p.WinService != "":
		pids, err := p.winServicePIDs()
		tags := map[string]string{"win_service": p.WinService}
		return []PidsTags{{pids, tags, err}}
	case p.CGroup != "":
		return p.cgroupPIDs()
	case p.PidFile != "":
		pids, err := p.finder.PidFile(p.PidFile)
		tags := map[string]string{"pidfile": p.PidFile}
		return []PidsTags{{pids, tags, err}}
	case p.Exe != "":
		pids, err := p.finder.Pattern(p.Exe)
		tags := map[string]string{"exe": p.Exe}
		return []PidsTags{{pids, tags, err}}
	case p.Pattern != "":
		pids, err := p.finder.FullPattern(p.Pattern)
		tags := map[string]string{"pattern": p.Pattern}
		return []PidsTags{{pids, tags, err}}
	case p.User != "":
		pids, err := p.finder.UID(p.User)
		tags := map[string]string{"user": p.User}
		return []PidsTags{{pids, tags, err}}
	}
	err := fmt.Errorf("either exe, pid_file, user, pattern, systemd_unit, cgroup, or win_service must be specified")
	return []PidsTags{{nil, nil, err}}
}

func (p *Procstat) findSupervisorUnits() []PidsTags {
	var pidTags []PidsTags
	groups, groupsTags, err := p.supervisorPIDs()
	if err != nil {
		pidTags = append(pidTags, PidsTags{nil, nil, err})
		return pidTags
	}
	// According to the PID, find the system process number and use pgrep to filter to get the number of child processes
	for _, group := range groups {
		p.Pattern = groupsTags[group]["pid"]
		if p.Pattern == "" {
			pidTags = append(pidTags, PidsTags{nil, groupsTags[group], err})
			return pidTags
		}

		// Get all children of the supervisor unit
		pids, err := p.finder.ChildPattern(p.Pattern)
		if err != nil {
			pidTags = append(pidTags, PidsTags{nil, nil, err})
			return pidTags
		}
		tags := map[string]string{"pattern": p.Pattern, "parent_pid": p.Pattern}

		// Handle situations where the PID does not exist
		if len(pids) == 0 {
			pidTags = append(pidTags, PidsTags{nil, groupsTags[group], err})
			continue
		}

		// Merge tags map
		for k, v := range groupsTags[group] {
			_, ok := tags[k]
			if !ok {
				tags[k] = v
			}
		}
		// Remove duplicate pid tags
		delete(tags, "pid")
		pidTags = append(pidTags, PidsTags{pids, tags, err})
	}
	return pidTags
}

// execCommand is so tests can mock out exec.Command usage.
var execCommand = exec.Command

func (p *Procstat) supervisorPIDs() ([]string, map[string]map[string]string, error) {
	out, err := execCommand("supervisorctl", "status", strings.Join(p.SupervisorUnit, " ")).Output()
	if err != nil {
		if !strings.Contains(err.Error(), "exit status 3") {
			return nil, nil, err
		}
	}
	lines := strings.Split(string(out), "\n")
	// Get the PID, running status, running time and boot time of the main process:
	// pid 11779, uptime 17:41:16
	// Exited too quickly (process log may have details)
	mainPids := make(map[string]map[string]string)
	for _, line := range lines {
		if line == "" {
			continue
		}

		kv := strings.Fields(line)
		if len(kv) < 2 {
			// Not a key-value pair
			continue
		}
		name := kv[0]

		statusMap := map[string]string{
			"supervisor_unit": name,
			"status":          kv[1],
		}

		switch kv[1] {
		case "FATAL", "EXITED", "BACKOFF", "STOPPING":
			statusMap["error"] = strings.Join(kv[2:], " ")
		case "RUNNING":
			statusMap["pid"] = strings.ReplaceAll(kv[3], ",", "")
			statusMap["uptimes"] = kv[5]
		case "STOPPED", "UNKNOWN", "STARTING":
			// No additional info
		}
		mainPids[name] = statusMap
	}

	return p.SupervisorUnit, mainPids, nil
}

func (p *Procstat) systemdUnitPIDs() []PidsTags {
	if p.IncludeSystemdChildren {
		p.CGroup = fmt.Sprintf("systemd/system.slice/%s", p.SystemdUnits)
		return p.cgroupPIDs()
	}

	var pidTags []PidsTags
	pids, err := p.simpleSystemdUnitPIDs()
	tags := map[string]string{"systemd_unit": p.SystemdUnits}
	pidTags = append(pidTags, PidsTags{pids, tags, err})
	return pidTags
}

func (p *Procstat) simpleSystemdUnitPIDs() ([]PID, error) {
	out, err := execCommand("systemctl", "show", p.SystemdUnits).Output()
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(out, []byte{'\n'})
	pids := make([]PID, 0, len(lines))
	for _, line := range lines {
		kv := bytes.SplitN(line, []byte{'='}, 2)
		if len(kv) != 2 {
			continue
		}
		if !bytes.Equal(kv[0], []byte("MainPID")) {
			continue
		}
		if len(kv[1]) == 0 || bytes.Equal(kv[1], []byte("0")) {
			return nil, nil
		}
		pid, err := strconv.ParseInt(string(kv[1]), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid pid %q", kv[1])
		}
		pids = append(pids, PID(pid))
	}

	return pids, nil
}

func (p *Procstat) cgroupPIDs() []PidsTags {
	procsPath := p.CGroup
	if procsPath[0] != '/' {
		procsPath = "/sys/fs/cgroup/" + procsPath
	}

	items, err := filepath.Glob(procsPath)
	if err != nil {
		return []PidsTags{{nil, nil, fmt.Errorf("glob failed: %w", err)}}
	}

	pidTags := make([]PidsTags, 0, len(items))
	for _, item := range items {
		pids, err := p.singleCgroupPIDs(item)
		tags := map[string]string{"cgroup": p.CGroup, "cgroup_full": item}
		pidTags = append(pidTags, PidsTags{pids, tags, err})
	}

	return pidTags
}

func (p *Procstat) singleCgroupPIDs(path string) ([]PID, error) {
	ok, err := isDir(path)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("not a directory %s", path)
	}
	procsPath := filepath.Join(path, "cgroup.procs")
	out, err := os.ReadFile(procsPath)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(out, []byte{'\n'})
	pids := make([]PID, 0, len(lines))
	for _, pidBS := range lines {
		if len(pidBS) == 0 {
			continue
		}
		pid, err := strconv.ParseInt(string(pidBS), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid pid %q", pidBS)
		}
		pids = append(pids, PID(pid))
	}

	return pids, nil
}

func isDir(path string) (bool, error) {
	result, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return result.IsDir(), nil
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
	inputs.Add("procstat", func() telegraf.Input {
		return &Procstat{createProcess: NewProc}
	})
}

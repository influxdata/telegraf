//go:generate ../../../tools/readme_config_includer/generator
package procstat

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"net"
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

var (
	defaultPIDFinder = NewPgrep
	defaultProcess   = NewProc
	// defaultCollection is the default group of metrics to gather
	defaultCollection = []string{
		metricsThreads,
		metricsFDs,
		metricsContextSwitches,
		metricsPageFaults,
		metricsIO,
		metricsCreateTime,
		metricsCPU,
		metricsCPUPercent,
		metricsMemory,
		metricsMemoryPercent,
		metricsLimits,
	}
)

const (
	// metricsThreads to enable collection of number of threads
	metricsThreads = "threads"
	// metricsFDs to enable collection of number of file descriptors
	metricsFDs = "fds"
	// metricsContextSwitches to enable collection of context switches
	metricsContextSwitches = "ctx_switches"
	// metricsPageFaults to enable collection of page faults
	metricsPageFaults = "page_faults"
	// metricsIO to enable collection of IO
	metricsIO = "io"
	// metricsCreateTime to enable collection of proc creation time
	metricsCreateTime = "create_time"
	// metricsCPU to enable collection of CPU time used
	metricsCPU = "cpu"
	// metricsCPUPercent to enable collection of percentage of CPU used
	metricsCPUPercent = "cpu_percent"
	// metricsMemory to enable collection of memory used
	metricsMemory = "mem"
	// metricsMemoryPercent to enable collection of memory percentage used
	metricsMemoryPercent = "mem_percent"
	// metricsLimits to enable collection of procs' limits
	metricsLimits = "limits"
	// metricsTCPStats to enable collection of procs' TCP stats
	metricsTCPStats = "tcp_stats"
	// metricsConnectionsEndpoints to enable collection of metric procstat_tcp
	metricsConnectionsEndpoints = "connections_endpoints"

	// metricNameTCPConnections is the measurement name for TCP connections metrics
	metricNameTCPConnections = "procstat_tcp"
	// tcpConnectionKey is the metric value to put all the listen endpoints
	tcpConnectionKey = "conn"
	// tcpListenKey is the metric value to put all the connection endpoints
	tcpListenKey = "listen"
)

type PID int32

type Procstat struct {
	PidFinder              string `toml:"pid_finder"`
	PidFile                string `toml:"pid_file"`
	Exe                    string
	Pattern                string
	Prefix                 string
	CmdLineTag             bool `toml:"cmdline_tag"`
	ProcessName            string
	User                   string
	SystemdUnit            string `toml:"systemd_unit"`
	IncludeSystemdChildren bool   `toml:"include_systemd_children"`
	CGroup                 string `toml:"cgroup"`
	PidTag                 bool
	WinService             string `toml:"win_service"`
	Mode                   string
	MetricsInclude         []string        `toml:"metrics_include"`
	Log                    telegraf.Logger `toml:"-"`

	solarisMode bool

	finder PIDFinder

	createPIDFinder func() (PIDFinder, error)
	procs           map[PID]Process
	createProcess   func(PID) (Process, error)
}

type PidsTags struct {
	PIDS []PID
	Tags map[string]string
	Err  error
}

func (*Procstat) SampleConfig() string {
	return sampleConfig
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

	pidCount := 0
	now := time.Now()
	newProcs := make(map[PID]Process, len(p.procs))
	tags := make(map[string]string)
	pidTags := p.findPids()
	for _, pidTag := range pidTags {
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

	// Initialize the conn object. Gather info about all TCP connections organized per PID
	// Avoid repeating this task for each proc
	netInfo := networkInfo{}
	// Only collect this info if we are going to use it (avoid reading all /proc/N/fd dirs)
	if (p.metricEnabled(metricsTCPStats) || p.metricEnabled(metricsConnectionsEndpoints)) && len(p.procs) > 0 {
		err := netInfo.Fetch()
		if err != nil {
			acc.AddError(fmt.Errorf("getting TCP network info: %w", err))
		}
	}

	for _, proc := range p.procs {
		p.addMetric(proc, acc, now, netInfo)
	}

	fields := map[string]interface{}{
		"pid_count":   pidCount,
		"running":     len(p.procs),
		"result_code": 0,
	}

	tags["pid_finder"] = p.PidFinder
	tags["result"] = "success"
	acc.AddFields("procstat_lookup", fields, tags, now)

	return nil
}

// Add metrics a single Process
func (p *Procstat) addMetric(proc Process, acc telegraf.Accumulator, t time.Time, netInfo networkInfo) {
	var prefix string
	if p.Prefix != "" {
		prefix = p.Prefix + "_"
	}

	fields := map[string]interface{}{}

	// If process_name tag is not already set, set to actual name
	if _, nameInTags := proc.Tags()["process_name"]; !nameInTags {
		name, err := proc.Name()
		if err == nil {
			proc.Tags()["process_name"] = name
		}
	}

	// If user tag is not already set, set to actual name
	if _, ok := proc.Tags()["user"]; !ok {
		user, err := proc.Username()
		if err == nil {
			proc.Tags()["user"] = user
		}
	}

	// If pid is not present as a tag, include it as a field.
	if _, pidInTags := proc.Tags()["pid"]; !pidInTags {
		fields["pid"] = int32(proc.PID())
	}

	// If cmd_line tag is true and it is not already set add cmdline as a tag
	if p.CmdLineTag {
		if _, ok := proc.Tags()["cmdline"]; !ok {
			cmdline, err := proc.Cmdline()
			if err == nil {
				proc.Tags()["cmdline"] = cmdline
			}
		}
	}

	if p.metricEnabled(metricsThreads) {
		numThreads, err := proc.NumThreads()
		if err == nil {
			fields[prefix+"num_threads"] = numThreads
		}
	}

	if p.metricEnabled(metricsFDs) {
		fds, err := proc.NumFDs()
		if err == nil {
			fields[prefix+"num_fds"] = fds
		}
	}

	if p.metricEnabled(metricsContextSwitches) {
		ctx, err := proc.NumCtxSwitches()
		if err == nil {
			fields[prefix+"voluntary_context_switches"] = ctx.Voluntary
			fields[prefix+"involuntary_context_switches"] = ctx.Involuntary
		}
	}

	if p.metricEnabled(metricsPageFaults) {
		faults, err := proc.PageFaults()
		if err == nil {
			fields[prefix+"minor_faults"] = faults.MinorFaults
			fields[prefix+"major_faults"] = faults.MajorFaults
			fields[prefix+"child_minor_faults"] = faults.ChildMinorFaults
			fields[prefix+"child_major_faults"] = faults.ChildMajorFaults
		}
	}

	if p.metricEnabled(metricsIO) {
		io, err := proc.IOCounters()
		if err == nil {
			fields[prefix+"read_count"] = io.ReadCount
			fields[prefix+"write_count"] = io.WriteCount
			fields[prefix+"read_bytes"] = io.ReadBytes
			fields[prefix+"write_bytes"] = io.WriteBytes
		}
	}

	if p.metricEnabled(metricsCreateTime) {
		createdAt, err := proc.CreateTime() // Returns epoch in ms
		if err == nil {
			fields[prefix+"created_at"] = createdAt * 1000000 // Convert ms to ns
		}
	}

	if p.metricEnabled(metricsCPU) {
		cpuTime, err := proc.Times()
		if err == nil {
			fields[prefix+"cpu_time_user"] = cpuTime.User
			fields[prefix+"cpu_time_system"] = cpuTime.System
			fields[prefix+"cpu_time_idle"] = cpuTime.Idle
			fields[prefix+"cpu_time_nice"] = cpuTime.Nice
			fields[prefix+"cpu_time_iowait"] = cpuTime.Iowait
			fields[prefix+"cpu_time_irq"] = cpuTime.Irq
			fields[prefix+"cpu_time_soft_irq"] = cpuTime.Softirq
			fields[prefix+"cpu_time_steal"] = cpuTime.Steal
			fields[prefix+"cpu_time_guest"] = cpuTime.Guest
			fields[prefix+"cpu_time_guest_nice"] = cpuTime.GuestNice
		}
	}

	if p.metricEnabled(metricsCPUPercent) {
		cpuPerc, err := proc.Percent(time.Duration(0))
		if err == nil {
			if p.solarisMode {
				fields[prefix+"cpu_usage"] = cpuPerc / float64(runtime.NumCPU())
			} else {
				fields[prefix+"cpu_usage"] = cpuPerc
			}
		}
	}

	if p.metricEnabled(metricsMemory) {
		mem, err := proc.MemoryInfo()
		if err == nil {
			fields[prefix+"memory_rss"] = mem.RSS
			fields[prefix+"memory_vms"] = mem.VMS
			fields[prefix+"memory_swap"] = mem.Swap
			fields[prefix+"memory_data"] = mem.Data
			fields[prefix+"memory_stack"] = mem.Stack
			fields[prefix+"memory_locked"] = mem.Locked
		}
	}

	if p.metricEnabled(metricsMemoryPercent) {
		memPerc, err := proc.MemoryPercent()
		if err == nil {
			fields[prefix+"memory_usage"] = memPerc
		}
	}

	if p.metricEnabled(metricsLimits) {
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
	}

	ppid, err := proc.Ppid()
	if err == nil {
		fields[prefix+"ppid"] = ppid
	}

	status, err := proc.Status()
	if err == nil {
		fields[prefix+"status"] = status[0]
	}

	if p.metricEnabled(metricsTCPStats) {
		// Add values with the number of connections in each TCP state
		pidConnections, err := netInfo.GetConnectionsByPid(uint32(proc.PID()))
		if err == nil {
			addConnectionStats(pidConnections, fields, prefix)
		} else {
			// Ignore errors because pid was not found. It is normal to have procs without connections
			if !errors.Is(err, errPIDNotFound) {
				p.Log.Debugf("not able to get connections for pid=%v: %v", proc.PID(), err)
			}
		}
	}

	acc.AddFields("procstat", fields, proc.Tags(), t)

	if p.metricEnabled(metricsConnectionsEndpoints) {
		// add measurement procstat_tcp with tcp listeners and connections for each proccess
		err := addConnectionEndpoints(acc, proc, netInfo)
		if err != nil {
			p.Log.Debugf("not able to generate network metrics for pid=%v: %v", proc.PID(), err)
		}
	}
}

// extractIPs extract and return IPs from addresses
func extractIPs(addreses []net.Addr) (ret []net.IP, err error) {
	for _, a := range addreses {
		ip, _, err := net.ParseCIDR(a.String())
		if err != nil {
			return nil, fmt.Errorf("parsing interface address: %w", err)
		}
		ret = append(ret, ip)
	}
	return ret, nil
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
func (p *Procstat) findPids() []PidsTags {
	var pidTags []PidsTags

	if p.SystemdUnit != "" {
		groups := p.systemdUnitPIDs()
		return groups
	} else if p.CGroup != "" {
		groups := p.cgroupPIDs()
		return groups
	} else {
		f, err := p.getPIDFinder()
		if err != nil {
			pidTags = append(pidTags, PidsTags{nil, nil, err})
			return pidTags
		}
		pids, tags, err := p.SimpleFindPids(f)
		pidTags = append(pidTags, PidsTags{pids, tags, err})
	}

	return pidTags
}

// Get matching PIDs and their initial tags
func (p *Procstat) SimpleFindPids(f PIDFinder) ([]PID, map[string]string, error) {
	var pids []PID
	tags := make(map[string]string)
	var err error

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
		pids, err = f.UID(p.User)
		tags = map[string]string{"user": p.User}
	} else if p.WinService != "" {
		pids, err = p.winServicePIDs()
		tags = map[string]string{"win_service": p.WinService}
	} else {
		err = fmt.Errorf("either exe, pid_file, user, pattern, systemd_unit, cgroup, or win_service must be specified")
	}

	return pids, tags, err
}

// execCommand is so tests can mock out exec.Command usage.
var execCommand = exec.Command

func (p *Procstat) systemdUnitPIDs() []PidsTags {
	if p.IncludeSystemdChildren {
		p.CGroup = fmt.Sprintf("systemd/system.slice/%s", p.SystemdUnit)
		return p.cgroupPIDs()
	}

	var pidTags []PidsTags

	pids, err := p.simpleSystemdUnitPIDs()
	tags := map[string]string{"systemd_unit": p.SystemdUnit}
	pidTags = append(pidTags, PidsTags{pids, tags, err})
	return pidTags
}

func (p *Procstat) simpleSystemdUnitPIDs() ([]PID, error) {
	out, err := execCommand("systemctl", "show", p.SystemdUnit).Output()
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

func (p *Procstat) Init() error {
	if strings.ToLower(p.Mode) == "solaris" {
		p.solarisMode = true
	}

	return nil
}

func containsIP(a []net.IP, x net.IP) bool {
	for _, n := range a {
		if x.Equal(n) {
			return true
		}
	}
	return false
}

func isIPV4(ip net.IP) bool {
	return ip.To4() != nil
}

func isIPV6(ip net.IP) bool {
	return ip.To4() == nil
}

// endpointString return the correct representation of ip and port for IPv4 and IPv6
func endpointString(ip net.IP, port uint32) string {
	if isIPV6(ip) {
		return fmt.Sprintf("[%s]:%d", ip, port)
	}
	return fmt.Sprintf("%s:%d", ip, port)
}

// metricEnabled check is some group of metrics are enabled in the config file
func (p *Procstat) metricEnabled(m string) bool {
	for _, n := range p.MetricsInclude {
		if m == n {
			return true
		}
	}
	return false
}

func init() {
	inputs.Add("procstat", func() telegraf.Input {
		return &Procstat{
			// Default metrics to gather
			MetricsInclude: defaultCollection,
		}
	})
}

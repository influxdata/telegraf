package procstat

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/shirou/gopsutil/process"
)

var (
	defaultPIDFinder = NewPgrep
	defaultProcess   = NewProc
	// defaultCollection is the default group of metrics to gather
	defaultCollection = []string{
		MetricsThreads,
		MetricsFDs,
		MetricsContextSwitches,
		MetricsPageFaults,
		MetricsIO,
		MetricsCreateTime,
		MetricsCPU,
		MetricsCPUPercent,
		MetricsMemory,
		MetricsMemoryPercent,
		MetricsLimits,
	}
)

const (
	// MetricsThreads to enable collection of number of threads
	MetricsThreads = "threads"
	// MetricsFDs to enable collection of number of file descriptors
	MetricsFDs = "fds"
	// MetricsContextSwitches to enable collection of context switches
	MetricsContextSwitches = "ctx_switches"
	// MetricsPageFaults to enable collection of page faults
	MetricsPageFaults = "page_faults"
	// MetricsIO to enable collection of IO
	MetricsIO = "io"
	// MetricsCreateTime to enable collection of proc creation time
	MetricsCreateTime = "create_time"
	// MetricsCPU to enable collection of CPU time used
	MetricsCPU = "cpu"
	// MetricsCPUPercent to enable collection of percentage of CPU used
	MetricsCPUPercent = "cpu_percent"
	// MetricsMemory to enable collection of memory used
	MetricsMemory = "mem"
	// MetricsMemoryPercent to enable collection of memory percentage used
	MetricsMemoryPercent = "mem_percent"
	// MetricsLimits to enable collection of procs' limits
	MetricsLimits = "limits"
	// MetricsTCPStats to enable collection of procs' TCP stats
	MetricsTCPStats = "tcp_stats"
	// MetricsConnectionsEndpoints to enable collection of metric procstat_tcp
	MetricsConnectionsEndpoints = "connections_endpoints"

	// MetricNameTCPConnections is the measurement name for TCP connections metrics
	MetricNameTCPConnections = "procstat_tcp"
	// TCPConnectionKey is the metric value to put all the listen endpoints
	TCPConnectionKey = "conn"
	// TCPListenKey is the metric value to put all the connection endpoints
	TCPListenKey = "listen"
)

type PID int32

type Procstat struct {
	PidFinder      string `toml:"pid_finder"`
	PidFile        string `toml:"pid_file"`
	Exe            string
	Pattern        string
	Prefix         string
	CmdLineTag     bool `toml:"cmdline_tag"`
	ProcessName    string
	User           string
	SystemdUnit    string
	CGroup         string `toml:"cgroup"`
	PidTag         bool
	WinService     string `toml:"win_service"`
	Mode           string
	MetricsInclude []string `toml:"metrics_include"`

	solarisMode bool
	finder      PIDFinder

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
  ## Systemd unit name
  # systemd_unit = "nginx.service"
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

  ## Mode to use when calculating CPU usage. Can be one of 'solaris' or 'irix'.
  # mode = "irix"

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

  ## Select wich extra metrics should be added:
  ##  - "threads": to enable collection of number of file descriptors
  ##  - "fds": to enable collection of context switches
  ##  - "ctx_switches": to enable collection of page faults
  ##  - "page_faults": to enable collection of IO
  ##  - "io": to enable collection of proc creation time
  ##  - "create_time": to enable collection of CPU time used
  ##  - "cpu": to enable collection of percentage of CPU used
  ##  - "cpu_percent": to enable collection of memory used
  ##  - "mem": to enable collection of memory percentage used
  ##  - "mem_percent": to enable collection of procs' limits
  ##  - "limits": to enable collection of procs' limits
  ##  - "tcp_stats": tcp_* and upd_socket metrics
  ##  - "connections_endpoints": new metric procstat_tcp with connections and listeners endpoints
  ## Default value:
  # metrics_include = [
  #  "threads",
  #  "fds",
  #  "ctx_switches",
  #  "page_faults",
  #  "io",
  #  "create_time",
  #  "cpu",
  #  "cpu_percent",
  #  "mem",
  #  "mem_percent",
  #  "limits",
  # ]
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

	pids, tags, err := p.findPids(acc)
	now := time.Now()

	if err != nil {
		fields := map[string]interface{}{
			"pid_count":   0,
			"running":     0,
			"result_code": 1,
		}
		tags := map[string]string{
			"pid_finder": p.PidFinder,
			"result":     "lookup_error",
		}
		acc.AddFields("procstat_lookup", fields, tags, now)
		return err
	}

	procs, err := p.updateProcesses(pids, tags, p.procs)
	if err != nil {
		acc.AddError(fmt.Errorf("E! Error: procstat getting process, exe: [%s] pidfile: [%s] pattern: [%s] user: [%s] %s",
			p.Exe, p.PidFile, p.Pattern, p.User, err.Error()))
	}
	p.procs = procs

	// Initialize the conn object. Gather info about all TCP connections organized per PID
	// Avoid repeating this task for each proc
	netInfo := NetworkInfo{}
	err = netInfo.Fetch()
	if err != nil {
		acc.AddError(fmt.Errorf("E! [inputs.procstat] getting TCP network info: %v", err))
	}

	for _, proc := range p.procs {
		p.addMetric(proc, acc, now, netInfo)
	}

	fields := map[string]interface{}{
		"pid_count":   len(pids),
		"running":     len(procs),
		"result_code": 0,
	}
	tags["pid_finder"] = p.PidFinder
	tags["result"] = "success"
	acc.AddFields("procstat_lookup", fields, tags, now)

	return nil
}

// Add metrics a single Process
func (p *Procstat) addMetric(proc Process, acc telegraf.Accumulator, t time.Time, netInfo NetworkInfo) {
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

	if p.metricEnabled(MetricsThreads) {
		numThreads, err := proc.NumThreads()
		if err == nil {
			fields[prefix+"num_threads"] = numThreads
		}
	}

	if p.metricEnabled(MetricsFDs) {
		fds, err := proc.NumFDs()
		if err == nil {
			fields[prefix+"num_fds"] = fds
		}
	}

	if p.metricEnabled(MetricsContextSwitches) {
		ctx, err := proc.NumCtxSwitches()
		if err == nil {
			fields[prefix+"voluntary_context_switches"] = ctx.Voluntary
			fields[prefix+"involuntary_context_switches"] = ctx.Involuntary
		}
	}

	if p.metricEnabled(MetricsPageFaults) {
		faults, err := proc.PageFaults()
		if err == nil {
			fields[prefix+"minor_faults"] = faults.MinorFaults
			fields[prefix+"major_faults"] = faults.MajorFaults
			fields[prefix+"child_minor_faults"] = faults.ChildMinorFaults
			fields[prefix+"child_major_faults"] = faults.ChildMajorFaults
		}
	}

	if p.metricEnabled(MetricsIO) {
		io, err := proc.IOCounters()
		if err == nil {
			fields[prefix+"read_count"] = io.ReadCount
			fields[prefix+"write_count"] = io.WriteCount
			fields[prefix+"read_bytes"] = io.ReadBytes
			fields[prefix+"write_bytes"] = io.WriteBytes
		}
	}

	if p.metricEnabled(MetricsCreateTime) {
		createdAt, err := proc.CreateTime() //Returns epoch in ms
		if err == nil {
			fields[prefix+"created_at"] = createdAt * 1000000 //Convert ms to ns
		}
	}

	if p.metricEnabled(MetricsCPU) {
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
	}

	if p.metricEnabled(MetricsCPUPercent) {
		cpu_perc, err := proc.Percent(time.Duration(0))
		if err == nil {
			if p.solarisMode {
				fields[prefix+"cpu_usage"] = cpu_perc / float64(runtime.NumCPU())
			} else {
				fields[prefix+"cpu_usage"] = cpu_perc
			}
		}
	}

	if p.metricEnabled(MetricsMemory) {
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

	if p.metricEnabled(MetricsMemoryPercent) {
		mem_perc, err := proc.MemoryPercent()
		if err == nil {
			fields[prefix+"memory_usage"] = mem_perc
		}
	}

	if p.metricEnabled(MetricsLimits) {
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

	if p.metricEnabled(MetricsTCPStats) {
		// Add values with the number of connections in each TCP state
		pidConnections, err := netInfo.GetConnectionsByPid(uint32(proc.PID()))
		if err == nil {
			addConnectionStats(pidConnections, fields, prefix)
		} else {
			// Ignore errors because pid was not found. It is normal to have procs without connections
			if !errors.Is(err, ErrorPIDNotFound) {
				acc.AddError(fmt.Errorf("D! [inputs.procstat] not able to get connections for pid=%v: %v", proc.PID(), err))
			}
		}
	}

	acc.AddFields("procstat", fields, proc.Tags(), t)

	if p.metricEnabled(MetricsConnectionsEndpoints) {
		// add measurement procstat_tcp with tcp listeners and connections for each proccess
		err := addConnectionEnpoints(acc, proc, netInfo)
		if err != nil {
			acc.AddError(fmt.Errorf("D! [inputs.procstat] not able to generate network metrics for pid=%v: %v", proc.PID(), err))
		}
	}
}

// appendIPs extract and return IPs from addresses
func extractIPs(addreses []net.Addr) (ret []net.IP, err error) {
	for _, a := range addreses {
		ip, _, err := net.ParseCIDR(a.String())
		if err != nil {
			return nil, fmt.Errorf("parsing interface address: %v", err)
		}
		ret = append(ret, ip)
	}
	return ret, nil
}

// Update monitored Processes
func (p *Procstat) updateProcesses(pids []PID, tags map[string]string, prevInfo map[PID]Process) (map[PID]Process, error) {
	procs := make(map[PID]Process, len(prevInfo))

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
func (p *Procstat) findPids(acc telegraf.Accumulator) ([]PID, map[string]string, error) {
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
		pids, err = p.systemdUnitPIDs()
		tags = map[string]string{"systemd_unit": p.SystemdUnit}
	} else if p.CGroup != "" {
		pids, err = p.cgroupPIDs()
		tags = map[string]string{"cgroup": p.CGroup}
	} else if p.WinService != "" {
		pids, err = p.winServicePIDs()
		tags = map[string]string{"win_service": p.WinService}
	} else {
		err = fmt.Errorf("Either exe, pid_file, user, pattern, systemd_unit, cgroup, or win_service must be specified")
	}

	return pids, tags, err
}

// execCommand is so tests can mock out exec.Command usage.
var execCommand = exec.Command

func (p *Procstat) systemdUnitPIDs() ([]PID, error) {
	var pids []PID
	cmd := execCommand("systemctl", "show", p.SystemdUnit)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	for _, line := range bytes.Split(out, []byte{'\n'}) {
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
			return nil, fmt.Errorf("invalid pid '%s'", kv[1])
		}
		pids = append(pids, PID(pid))
	}
	return pids, nil
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

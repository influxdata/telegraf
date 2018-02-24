package procstat

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	gopsutilmem "github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

var (
	defaultPIDFinder = NewPgrep
	defaultProcess   = NewProc
)

type PID int32

type Procstat struct {
	PidFinder   string `toml:"pid_finder"`
	PidFile     string `toml:"pid_file"`
	Exe         string
	Pattern     string
	Prefix      string
	ProcessName string
	User        string
	SystemdUnit string
	CGroup      string `toml:"cgroup"`
	PidTag      bool

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
  ## Systemd unit name
  # systemd_unit = "nginx.service"
  ## CGroup name or path
  # cgroup = "systemd/system.slice/nginx.service"

  ## override for process_name
  ## This is optional; default is sourced from /proc/<pid>/status
  # process_name = "bar"

  ## Field name prefix
  # prefix = ""

  ## Add PID as a tag instead of a field; useful to differentiate between
  ## processes whose tags are otherwise the same.  Can create a large number
  ## of series, use judiciously.
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
			p.createPIDFinder = defaultPIDFinder
		}

	}
	if p.createProcess == nil {
		p.createProcess = defaultProcess
	}

	procs, err := p.updateProcesses(p.procs)
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

type statStat struct {
	minflt                uint64
	maxflt                uint64
	delayacct_blkio_ticks uint64
	state                 string
}
type additionalIOStat struct {
	cancelled_write_bytes uint64
}

// this is taken from https://github.com/shirou/gopsutil/blob/v2.18.02/internal/common/common.go#L286
// so that HostProc will work

// GetEnv retrieves the environment variable key. If it does not exist it returns the default.
func GetEnv(key string, dfault string, combineWith ...string) string {
	value := os.Getenv(key)
	if value == "" {
		value = dfault
	}

	switch len(combineWith) {
	case 0:
		return value
	case 1:
		return filepath.Join(value, combineWith[0])
	default:
		all := make([]string, len(combineWith)+1)
		all[0] = value
		copy(all[1:], combineWith)
		return filepath.Join(all...)
	}
}

func HostProc(combineWith ...string) string {
	return GetEnv("HOST_PROC", "/proc", combineWith...)
}

func GetStatCounters(proc Process) (*statStat, error) {
	pid := proc.PID()
	contents, err := ioutil.ReadFile(HostProc(strconv.Itoa(int(pid)), "stat"))
	if err != nil {
		return nil, err
	}
	fields := strings.Split(string(contents), " ")
	// fmt.Printf("%v", fields)
	minflt, err := strconv.ParseUint(fields[9], 10, 64)
	if err != nil {
		return nil, err
	}
	maxflt, err := strconv.ParseUint(fields[11], 10, 64)
	if err != nil {
		return nil, err
	}

	delayacct_blkio_ticks, err := strconv.ParseUint(fields[41], 10, 64)
	if err != nil {
		return nil, err
	}
	state := fields[2]

	stat := &statStat{
		minflt:                minflt,
		maxflt:                maxflt,
		delayacct_blkio_ticks: delayacct_blkio_ticks,
		state: state,
	}
	return stat, nil

}

func getAdditionalFromIO(proc Process) (*additionalIOStat, error) {
	pid := proc.PID()

	ioline, err := ioutil.ReadFile(HostProc(strconv.Itoa(int(pid)), "io"))
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(ioline), "\n")
	ret := &additionalIOStat{}

	for _, line := range lines {
		field := strings.Fields(line)
		if len(field) < 2 {
			continue
		}
		t, err := strconv.ParseUint(field[1], 10, 64)
		if err != nil {
			return nil, err
		}
		param := field[0]
		if strings.HasSuffix(param, ":") {
			param = param[:len(param)-1]
		}
		switch param {
		case "cancelled_write_bytes":
			ret.cancelled_write_bytes = t
		}
	}

	return ret, nil
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

	//If pid is not present as a tag, include it as a field.
	if _, pidInTags := proc.Tags()["pid"]; !pidInTags {
		fields["pid"] = int32(proc.PID())
	}

	numThreads, err := proc.NumThreads()
	if err == nil {
		fields[prefix+"num_threads"] = numThreads
	}

	createTime, err := proc.CreateTime()
	if err == nil {
		fields[prefix+"create_time"] = createTime
	}

	// uid gid: Real, effective, saved set, and filesystem UIDs (http://man7.org/linux/man-pages/man5/proc.5.html)
	// so we are using real uid here
	uids, err := proc.Uids()
	if err == nil {
		fields[prefix+"uid"] = uids[0]
	}
	gids, err := proc.Uids()
	if err == nil {
		fields[prefix+"gid"] = gids[0]
	}
	cmdline, err := proc.CmdlineSlice()
	if err == nil {
		fields[prefix+"cmdline"] = strings.Join(cmdline, " ")
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

	stat, err := GetStatCounters(proc)
	if err == nil {
		fields[prefix+"minor_faults"] = stat.minflt
		fields[prefix+"major_faults"] = stat.maxflt
		fields[prefix+"delayacct_blkio_ticks"] = stat.delayacct_blkio_ticks
		fields[prefix+"state"] = stat.state
	}

	io, err := proc.IOCounters()
	if err == nil {
		fields[prefix+"read_count"] = io.ReadCount
		fields[prefix+"write_count"] = io.WriteCount
		fields[prefix+"read_bytes"] = io.ReadBytes
		fields[prefix+"write_bytes"] = io.WriteBytes
	}
	additional_io, err := getAdditionalFromIO(proc)
	if err == nil {
		fields[prefix+"cancelled_write_bytes"] = additional_io.cancelled_write_bytes
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
		fields[prefix+"cpu_time_stolen"] = cpu_time.Stolen
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

	virtual, err := gopsutilmem.VirtualMemory()
	if err == nil {
		fields[prefix+"memory_used_percent"] = float64(mem.RSS) / float64(virtual.Total) * 100
	}

	// total nework traffic
	net, err := proc.NetIOCounters(false)
	if err == nil {
		fields[prefix+"net_bytes_sent"] = net[0].BytesSent
		fields[prefix+"net_bytes_recv"] = net[0].BytesRecv

		fields[prefix+"net_packets_sent"] = net[0].PacketsSent
		fields[prefix+"net_packets_recv"] = net[0].PacketsRecv

		fields[prefix+"net_errors_in"] = net[0].Errin
		fields[prefix+"net_errors_out"] = net[0].Errout
	}
	// per interface network traffic
	net2, err := proc.NetIOCounters(true)
	if err == nil {
		for _, net_stat := range net2 {
			fields_per_interface := map[string]interface{}{}
			var newTags map[string]string
			newTags = map[string]string{"interface": net_stat.Name}

			for k, v := range proc.Tags() {
				newTags[k] = v
			}

			fields_per_interface[prefix+"net_bytes_sent"] = net_stat.BytesSent
			fields_per_interface[prefix+"net_bytes_recv"] = net_stat.BytesRecv

			fields_per_interface[prefix+"net_packets_sent"] = net_stat.PacketsSent
			fields_per_interface[prefix+"net_packets_recv"] = net_stat.PacketsRecv

			fields_per_interface[prefix+"net_errors_in"] = net_stat.Errin
			fields_per_interface[prefix+"net_errors_out"] = net_stat.Errout
			acc.AddFields("procstat_net", fields_per_interface, newTags)
		}
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
func (p *Procstat) updateProcesses(prevInfo map[PID]Process) (map[PID]Process, error) {
	pids, tags, err := p.findPids()
	if err != nil {
		return nil, err
	}

	procs := make(map[PID]Process, len(prevInfo))

	for _, pid := range pids {
		info, ok := prevInfo[pid]
		if ok {
			procs[pid] = info
		} else {
			proc, err := p.createProcess(pid)
			if err != nil {
				// No problem; process may have ended after we found it
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
func (p *Procstat) findPids() ([]PID, map[string]string, error) {
	var pids []PID
	var tags map[string]string
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
	} else {
		err = fmt.Errorf("Either exe, pid_file, user, pattern, systemd_unit, or cgroup must be specified")
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
		if len(kv[1]) == 0 {
			return nil, nil
		}
		pid, err := strconv.Atoi(string(kv[1]))
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
		pid, err := strconv.Atoi(string(pidBS))
		if err != nil {
			return nil, fmt.Errorf("invalid pid '%s'", pidBS)
		}
		pids = append(pids, PID(pid))
	}

	return pids, nil
}

func init() {
	inputs.Add("procstat", func() telegraf.Input {
		return &Procstat{}
	})
}

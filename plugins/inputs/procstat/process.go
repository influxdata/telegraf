package procstat

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/shirou/gopsutil/v3/process"
)

type Process interface {
	PID() PID
	Name() (string, error)
	SetTag(string, string)
	MemoryMaps(bool) (*[]process.MemoryMapsStat, error)
	Metric(prefix string, tagging map[string]bool, solarisMode bool) telegraf.Metric
}

type PIDFinder interface {
	PidFile(path string) ([]PID, error)
	Pattern(pattern string) ([]PID, error)
	UID(user string) ([]PID, error)
	FullPattern(path string) ([]PID, error)
	Children(pid PID) ([]PID, error)
}

type Proc struct {
	hasCPUTimes bool
	tags        map[string]string
	*process.Process
}

func newProc(pid PID) (Process, error) {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, err
	}

	proc := &Proc{
		Process:     p,
		hasCPUTimes: false,
		tags:        make(map[string]string),
	}
	return proc, nil
}

func (p *Proc) PID() PID {
	return PID(p.Process.Pid)
}

func (p *Proc) SetTag(k, v string) {
	p.tags[k] = v
}

func (p *Proc) percent(_ time.Duration) (float64, error) {
	cpuPerc, err := p.Process.Percent(time.Duration(0))
	if !p.hasCPUTimes && err == nil {
		p.hasCPUTimes = true
		return 0, fmt.Errorf("must call Percent twice to compute percent cpu")
	}
	return cpuPerc, err
}

// Add metrics a single Process
func (p *Proc) Metric(prefix string, tagging map[string]bool, solarisMode bool) telegraf.Metric {
	if prefix != "" {
		prefix += "_"
	}

	fields := make(map[string]interface{})
	numThreads, err := p.NumThreads()
	if err == nil {
		fields[prefix+"num_threads"] = numThreads
	}

	fds, err := p.NumFDs()
	if err == nil {
		fields[prefix+"num_fds"] = fds
	}

	ctx, err := p.NumCtxSwitches()
	if err == nil {
		fields[prefix+"voluntary_context_switches"] = ctx.Voluntary
		fields[prefix+"involuntary_context_switches"] = ctx.Involuntary
	}

	faults, err := p.PageFaults()
	if err == nil {
		fields[prefix+"minor_faults"] = faults.MinorFaults
		fields[prefix+"major_faults"] = faults.MajorFaults
		fields[prefix+"child_minor_faults"] = faults.ChildMinorFaults
		fields[prefix+"child_major_faults"] = faults.ChildMajorFaults
	}

	io, err := p.IOCounters()
	if err == nil {
		fields[prefix+"read_count"] = io.ReadCount
		fields[prefix+"write_count"] = io.WriteCount
		fields[prefix+"read_bytes"] = io.ReadBytes
		fields[prefix+"write_bytes"] = io.WriteBytes
	}

	createdAt, err := p.CreateTime() // returns epoch in ms
	if err == nil {
		fields[prefix+"created_at"] = createdAt * 1000000 // ms to ns
	}

	cpuTime, err := p.Times()
	if err == nil {
		fields[prefix+"cpu_time_user"] = cpuTime.User
		fields[prefix+"cpu_time_system"] = cpuTime.System
		fields[prefix+"cpu_time_iowait"] = cpuTime.Iowait // only reported on Linux
	}

	cpuPerc, err := p.percent(time.Duration(0))
	if err == nil {
		if solarisMode {
			fields[prefix+"cpu_usage"] = cpuPerc / float64(runtime.NumCPU())
		} else {
			fields[prefix+"cpu_usage"] = cpuPerc
		}
	}

	// This only returns values for RSS and VMS
	mem, err := p.MemoryInfo()
	if err == nil {
		fields[prefix+"memory_rss"] = mem.RSS
		fields[prefix+"memory_vms"] = mem.VMS
	}

	collectMemmap(p, prefix, fields)

	memPerc, err := p.MemoryPercent()
	if err == nil {
		fields[prefix+"memory_usage"] = memPerc
	}

	rlims, err := p.RlimitUsage(true)
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

	// Add the tags as requested by the user
	cmdline, err := p.Cmdline()
	if err == nil {
		if tagging["cmdline"] {
			p.tags["cmdline"] = cmdline
		} else {
			fields[prefix+"cmdline"] = cmdline
		}
	}

	if tagging["pid"] {
		p.tags["pid"] = strconv.Itoa(int(p.Pid))
	} else {
		fields["pid"] = p.Pid
	}

	ppid, err := p.Ppid()
	if err == nil {
		if tagging["ppid"] {
			p.tags["ppid"] = strconv.Itoa(int(ppid))
		} else {
			fields[prefix+"ppid"] = ppid
		}
	}

	status, err := p.Status()
	if err == nil {
		if tagging["status"] {
			p.tags["status"] = status[0]
		} else {
			fields[prefix+"status"] = status[0]
		}
	}

	user, err := p.Username()
	if err == nil {
		if tagging["user"] {
			p.tags["user"] = user
		} else {
			fields[prefix+"user"] = user
		}
	}

	if _, exists := p.tags["process_name"]; !exists {
		name, err := p.Name()
		if err == nil {
			p.tags["process_name"] = name
		}
	}

	return metric.New("procstat", p.tags, fields, time.Time{})
}

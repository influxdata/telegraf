package procstat

import (
	"time"

	"github.com/shirou/gopsutil/process"

	"github.com/influxdata/telegraf"
)

type SpecProcessor struct {
	Prefix string
	pid    int32
	tags   map[string]string
	fields map[string]interface{}
	acc    telegraf.Accumulator
	proc   *process.Process
}

func NewSpecProcessor(
	processName string,
	prefix string,
	pid int32,
	acc telegraf.Accumulator,
	p *process.Process,
	tags map[string]string,
) *SpecProcessor {
	if processName != "" {
		tags["process_name"] = processName
	} else {
		name, err := p.Name()
		if err == nil {
			tags["process_name"] = name
		}
	}
	return &SpecProcessor{
		Prefix: prefix,
		pid:    pid,
		tags:   tags,
		fields: make(map[string]interface{}),
		acc:    acc,
		proc:   p,
	}
}

func (p *SpecProcessor) pushMetrics() {
	var prefix string
	if p.Prefix != "" {
		prefix = p.Prefix + "_"
	}
	fields := map[string]interface{}{"pid": p.pid}

	numThreads, err := p.proc.NumThreads()
	if err == nil {
		fields[prefix+"num_threads"] = numThreads
	}

	fds, err := p.proc.NumFDs()
	if err == nil {
		fields[prefix+"num_fds"] = fds
	}

	ctx, err := p.proc.NumCtxSwitches()
	if err == nil {
		fields[prefix+"voluntary_context_switches"] = ctx.Voluntary
		fields[prefix+"involuntary_context_switches"] = ctx.Involuntary
	}

	io, err := p.proc.IOCounters()
	if err == nil {
		fields[prefix+"read_count"] = io.ReadCount
		fields[prefix+"write_count"] = io.WriteCount
		fields[prefix+"read_bytes"] = io.ReadBytes
		fields[prefix+"write_bytes"] = io.WriteCount
	}

	cpu_time, err := p.proc.Times()
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

	cpu_perc, err := p.proc.Percent(time.Duration(0))
	if err == nil && cpu_perc != 0 {
		fields[prefix+"cpu_usage"] = cpu_perc
	}

	mem, err := p.proc.MemoryInfo()
	if err == nil {
		fields[prefix+"memory_rss"] = mem.RSS
		fields[prefix+"memory_vms"] = mem.VMS
		fields[prefix+"memory_swap"] = mem.Swap
	}

	p.acc.AddFields("procstat", fields, p.tags)
}

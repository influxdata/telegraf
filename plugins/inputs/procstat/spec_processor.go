package procstat

import (
	"time"

	"github.com/shirou/gopsutil/process"

	"github.com/influxdata/telegraf"
)

type SpecProcessor struct {
	Prefix string
	tags   map[string]string
	fields map[string]interface{}
	acc    telegraf.Accumulator
	proc   *process.Process
}

func NewSpecProcessor(
	prefix string,
	acc telegraf.Accumulator,
	p *process.Process,
	tags map[string]string,
) *SpecProcessor {
	if name, err := p.Name(); err == nil {
		tags["process_name"] = name
	}
	return &SpecProcessor{
		Prefix: prefix,
		tags:   tags,
		fields: make(map[string]interface{}),
		acc:    acc,
		proc:   p,
	}
}

func (p *SpecProcessor) pushMetrics() {
	fields := map[string]interface{}{}

	numThreads, err := p.proc.NumThreads()
	if err == nil {
		fields["num_threads"] = numThreads
	}

	fds, err := p.proc.NumFDs()
	if err == nil {
		fields["num_fds"] = fds
	}

	ctx, err := p.proc.NumCtxSwitches()
	if err == nil {
		fields["voluntary_context_switches"] = ctx.Voluntary
		fields["involuntary_context_switches"] = ctx.Involuntary
	}

	io, err := p.proc.IOCounters()
	if err == nil {
		fields["read_count"] = io.ReadCount
		fields["write_count"] = io.WriteCount
		fields["read_bytes"] = io.ReadBytes
		fields["write_bytes"] = io.WriteCount
	}

	cpu_time, err := p.proc.CPUTimes()
	if err == nil {
		fields["cpu_time_user"] = cpu_time.User
		fields["cpu_time_system"] = cpu_time.System
		fields["cpu_time_idle"] = cpu_time.Idle
		fields["cpu_time_nice"] = cpu_time.Nice
		fields["cpu_time_iowait"] = cpu_time.Iowait
		fields["cpu_time_irq"] = cpu_time.Irq
		fields["cpu_time_soft_irq"] = cpu_time.Softirq
		fields["cpu_time_steal"] = cpu_time.Steal
		fields["cpu_time_stolen"] = cpu_time.Stolen
		fields["cpu_time_guest"] = cpu_time.Guest
		fields["cpu_time_guest_nice"] = cpu_time.GuestNice
	}

	cpu_perc, err := p.proc.CPUPercent(time.Duration(0))
	if err == nil && cpu_perc != 0 {
		fields["cpu_usage"] = cpu_perc
	}

	mem, err := p.proc.MemoryInfo()
	if err == nil {
		fields["memory_rss"] = mem.RSS
		fields["memory_vms"] = mem.VMS
		fields["memory_swap"] = mem.Swap
	}

	p.acc.AddFields("procstat", fields, p.tags)
}

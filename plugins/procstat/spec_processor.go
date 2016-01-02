package procstat

import (
	"fmt"
	"log"

	"github.com/shirou/gopsutil/process"

	"github.com/influxdb/telegraf/plugins"
)

type SpecProcessor struct {
	Prefix string
	tags   map[string]string
	fields map[string]interface{}
	acc    plugins.Accumulator
	proc   *process.Process
}

func (p *SpecProcessor) add(metric string, value interface{}) {
	var mname string
	if p.Prefix == "" {
		mname = metric
	} else {
		mname = p.Prefix + "_" + metric
	}
	p.fields[mname] = value
}

func (p *SpecProcessor) flush() {
	p.acc.AddFields("procstat", p.fields, p.tags)
	p.fields = make(map[string]interface{})
}

func NewSpecProcessor(
	prefix string,
	acc plugins.Accumulator,
	p *process.Process,
) *SpecProcessor {
	tags := make(map[string]string)
	tags["pid"] = fmt.Sprintf("%v", p.Pid)
	if name, err := p.Name(); err == nil {
		tags["name"] = name
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
	if err := p.pushFDStats(); err != nil {
		log.Printf("procstat, fd stats not available: %s", err.Error())
	}
	if err := p.pushCtxStats(); err != nil {
		log.Printf("procstat, ctx stats not available: %s", err.Error())
	}
	if err := p.pushIOStats(); err != nil {
		log.Printf("procstat, io stats not available: %s", err.Error())
	}
	if err := p.pushCPUStats(); err != nil {
		log.Printf("procstat, cpu stats not available: %s", err.Error())
	}
	if err := p.pushMemoryStats(); err != nil {
		log.Printf("procstat, mem stats not available: %s", err.Error())
	}
	p.flush()
}

func (p *SpecProcessor) pushFDStats() error {
	fds, err := p.proc.NumFDs()
	if err != nil {
		return fmt.Errorf("NumFD error: %s\n", err)
	}
	p.add("num_fds", fds)
	return nil
}

func (p *SpecProcessor) pushCtxStats() error {
	ctx, err := p.proc.NumCtxSwitches()
	if err != nil {
		return fmt.Errorf("ContextSwitch error: %s\n", err)
	}
	p.add("voluntary_context_switches", ctx.Voluntary)
	p.add("involuntary_context_switches", ctx.Involuntary)
	return nil
}

func (p *SpecProcessor) pushIOStats() error {
	io, err := p.proc.IOCounters()
	if err != nil {
		return fmt.Errorf("IOCounters error: %s\n", err)
	}
	p.add("read_count", io.ReadCount)
	p.add("write_count", io.WriteCount)
	p.add("read_bytes", io.ReadBytes)
	p.add("write_bytes", io.WriteCount)
	return nil
}

func (p *SpecProcessor) pushCPUStats() error {
	cpu_time, err := p.proc.CPUTimes()
	if err != nil {
		return err
	}
	p.add("cpu_time_user", cpu_time.User)
	p.add("cpu_time_system", cpu_time.System)
	p.add("cpu_time_idle", cpu_time.Idle)
	p.add("cpu_time_nice", cpu_time.Nice)
	p.add("cpu_time_iowait", cpu_time.Iowait)
	p.add("cpu_time_irq", cpu_time.Irq)
	p.add("cpu_time_soft_irq", cpu_time.Softirq)
	p.add("cpu_time_soft_steal", cpu_time.Steal)
	p.add("cpu_time_soft_stolen", cpu_time.Stolen)
	p.add("cpu_time_soft_guest", cpu_time.Guest)
	p.add("cpu_time_soft_guest_nice", cpu_time.GuestNice)

	return nil
}

func (p *SpecProcessor) pushMemoryStats() error {
	mem, err := p.proc.MemoryInfo()
	if err != nil {
		return err
	}
	p.add("memory_rss", mem.RSS)
	p.add("memory_vms", mem.VMS)
	p.add("memory_swap", mem.Swap)
	return nil
}

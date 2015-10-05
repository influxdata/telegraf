package procstat

import (
	"fmt"

	"github.com/shirou/gopsutil/process"

	"github.com/influxdb/telegraf/plugins"
)

type SpecProcessor struct {
	Prefix string
	tags   map[string]string
	acc    plugins.Accumulator
	proc   *process.Process
}

func (p *SpecProcessor) add(metric string, value interface{}) {
	p.acc.Add(p.Prefix+"_"+metric, value, p.tags)
}

func NewSpecProcessor(prefix string, acc plugins.Accumulator, p *process.Process) *SpecProcessor {
	return &SpecProcessor{
		Prefix: prefix,
		tags:   map[string]string{},
		acc:    acc,
		proc:   p,
	}
}

func (p *SpecProcessor) pushMetrics() error {
	if err := p.pushFDStats(); err != nil {
		return err
	}
	if err := p.pushCtxStats(); err != nil {
		return err
	}
	if err := p.pushIOStats(); err != nil {
		return err
	}
	if err := p.pushCPUStats(); err != nil {
		return err
	}
	if err := p.pushMemoryStats(); err != nil {
		return err
	}
	return nil
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
	cpu, err := p.proc.CPUTimes()
	if err != nil {
		return err
	}
	p.add("cpu_user", cpu.User)
	p.add("cpu_system", cpu.System)
	p.add("cpu_idle", cpu.Idle)
	p.add("cpu_nice", cpu.Nice)
	p.add("cpu_iowait", cpu.Iowait)
	p.add("cpu_irq", cpu.Irq)
	p.add("cpu_soft_irq", cpu.Softirq)
	p.add("cpu_soft_steal", cpu.Steal)
	p.add("cpu_soft_stolen", cpu.Stolen)
	p.add("cpu_soft_guest", cpu.Guest)
	p.add("cpu_soft_guest_nice", cpu.GuestNice)
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

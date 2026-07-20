//go:build !windows

package cpu

import (
	"errors"

	"github.com/shirou/gopsutil/v4/cpu"
)

type stats cpu.TimesStat

func cpuTimes(perCPU, totalCPU bool) ([]stats, error) {
	var cpuTimes []stats
	if perCPU {
		perCPUTimes, err := cpu.Times(true)
		if err != nil {
			return nil, err
		}
		for _, s := range perCPUTimes {
			cpuTimes = append(cpuTimes, stats(s))
		}
	}
	if totalCPU {
		totalCPUTimes, err := cpu.Times(false)
		if err != nil {
			return nil, err
		}
		for _, s := range totalCPUTimes {
			cpuTimes = append(cpuTimes, stats(s))
		}
	}
	return cpuTimes, nil
}

func (s *stats) cycleFields(active bool) map[string]interface{} {
	fields := map[string]interface{}{
		"time_user":       s.User,
		"time_system":     s.System,
		"time_idle":       s.Idle,
		"time_nice":       s.Nice,
		"time_iowait":     s.Iowait,
		"time_irq":        s.Irq,
		"time_softirq":    s.Softirq,
		"time_steal":      s.Steal,
		"time_guest":      s.Guest,
		"time_guest_nice": s.GuestNice,
	}
	if active {
		fields["time_active"] = s.User + s.System + s.Nice + s.Iowait + s.Irq + s.Softirq + s.Steal
	}
	return fields
}

func (s *stats) percentageFields(last stats, active bool) (map[string]interface{}, error) {
	total := s.User + s.System + s.Nice + s.Iowait + s.Irq + s.Softirq + s.Steal + s.Idle
	lastTotal := last.User + last.System + last.Nice + last.Iowait + last.Irq + last.Softirq + last.Steal + last.Idle

	if total < lastTotal {
		return nil, errors.New("current total CPU time is less than previous total CPU time")
	}

	if total == lastTotal {
		return nil, nil
	}

	dT := total - lastTotal

	fields := map[string]interface{}{
		"usage_user":       100.0 * (s.User - last.User - (s.Guest - last.Guest)) / dT,
		"usage_system":     100.0 * (s.System - last.System) / dT,
		"usage_idle":       100.0 * (s.Idle - last.Idle) / dT,
		"usage_nice":       100.0 * (s.Nice - last.Nice - (s.GuestNice - last.GuestNice)) / dT,
		"usage_iowait":     100.0 * (s.Iowait - last.Iowait) / dT,
		"usage_irq":        100.0 * (s.Irq - last.Irq) / dT,
		"usage_softirq":    100.0 * (s.Softirq - last.Softirq) / dT,
		"usage_steal":      100.0 * (s.Steal - last.Steal) / dT,
		"usage_guest":      100.0 * (s.Guest - last.Guest) / dT,
		"usage_guest_nice": 100.0 * (s.GuestNice - last.GuestNice) / dT,
	}
	if active {
		dActive := s.User + s.System + s.Nice + s.Iowait + s.Irq + s.Softirq + s.Steal
		dActive -= last.User + last.System + last.Nice + last.Iowait + last.Irq + last.Softirq + last.Steal
		fields["usage_active"] = 100.0 * dActive / dT
	}

	return fields, nil
}

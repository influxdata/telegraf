// +build linux freebsd darwin

package cpu

import "time"

func init() {
	lastCPUTimes, _ = CPUTimes(false)
	lastPerCPUTimes, _ = CPUTimes(true)
}

func CPUPercent(interval time.Duration, percpu bool) ([]float64, error) {
	getAllBusy := func(t CPUTimesStat) (float64, float64) {
		busy := t.User + t.System + t.Nice + t.Iowait + t.Irq +
			t.Softirq + t.Steal + t.Guest + t.GuestNice + t.Stolen
		return busy + t.Idle, busy
	}

	calculate := func(t1, t2 CPUTimesStat) float64 {
		t1All, t1Busy := getAllBusy(t1)
		t2All, t2Busy := getAllBusy(t2)

		if t2Busy <= t1Busy {
			return 0
		}
		if t2All <= t1All {
			return 1
		}
		return (t2Busy - t1Busy) / (t2All - t1All) * 100
	}

	cpuTimes, err := CPUTimes(percpu)
	if err != nil {
		return nil, err
	}

	if interval > 0 {
		if !percpu {
			lastCPUTimes = cpuTimes
		} else {
			lastPerCPUTimes = cpuTimes
		}
		time.Sleep(interval)
		cpuTimes, err = CPUTimes(percpu)
		if err != nil {
			return nil, err
		}
	}

	ret := make([]float64, len(cpuTimes))
	if !percpu {
		ret[0] = calculate(lastCPUTimes[0], cpuTimes[0])
		lastCPUTimes = cpuTimes
	} else {
		for i, t := range cpuTimes {
			ret[i] = calculate(lastPerCPUTimes[i], t)
		}
		lastPerCPUTimes = cpuTimes
	}
	return ret, nil
}

package system

import (
	"fmt"

	"github.com/influxdb/telegraf/plugins"
	"github.com/influxdb/telegraf/plugins/system/ps/cpu"
)

type CPUStats struct {
	ps        PS
	lastStats []cpu.CPUTimesStat
}

func NewCPUStats(ps PS) *CPUStats {
	times, _ := ps.CPUTimes()
	stats := CPUStats{
		ps:        ps,
		lastStats: times,
	}

	return &stats
}

func (_ *CPUStats) Description() string {
	return "Read metrics about cpu usage"
}

func (_ *CPUStats) SampleConfig() string { return "" }

func (s *CPUStats) Gather(acc plugins.Accumulator) error {
	times, err := s.ps.CPUTimes()

	if err != nil {
		return fmt.Errorf("error getting CPU info: %s", err)
	}

	for i, cts := range times {
		tags := map[string]string{
			"cpu": cts.CPU,
		}

		busy, total := busyAndTotalCpuTime(cts)

		// Add total cpu numbers
		add(acc, "user", cts.User, tags)
		add(acc, "system", cts.System, tags)
		add(acc, "idle", cts.Idle, tags)
		add(acc, "nice", cts.Nice, tags)
		add(acc, "iowait", cts.Iowait, tags)
		add(acc, "irq", cts.Irq, tags)
		add(acc, "softirq", cts.Softirq, tags)
		add(acc, "steal", cts.Steal, tags)
		add(acc, "guest", cts.Guest, tags)
		add(acc, "guestNice", cts.GuestNice, tags)
		add(acc, "stolen", cts.Stolen, tags)
		add(acc, "busy", busy, tags)

		// Add in percentage
		lastCts := s.lastStats[i]
		lastBusy, lastTotal := busyAndTotalCpuTime(lastCts)
		busyDelta := busy - lastBusy
		totalDelta := total - lastTotal

		if totalDelta < 0 {
			return fmt.Errorf("Error: current total CPU time is less than previous total CPU time")
		}

		if totalDelta == 0 {
			return nil
		}

		add(acc, "percentageUser", 100*(cts.User-lastCts.User)/totalDelta, tags)
		add(acc, "percentageSystem", 100*(cts.System-lastCts.System)/totalDelta, tags)
		add(acc, "percentageIdle", 100*(cts.Idle-lastCts.Idle)/totalDelta, tags)
		add(acc, "percentageNice", 100*(cts.Nice-lastCts.Nice)/totalDelta, tags)
		add(acc, "percentageIowait", 100*(cts.Iowait-lastCts.Iowait)/totalDelta, tags)
		add(acc, "percentageIrq", 100*(cts.Irq-lastCts.Irq)/totalDelta, tags)
		add(acc, "percentageSoftirq", 100*(cts.Softirq-lastCts.Softirq)/totalDelta, tags)
		add(acc, "percentageSteal", 100*(cts.Steal-lastCts.Steal)/totalDelta, tags)
		add(acc, "percentageGuest", 100*(cts.Guest-lastCts.Guest)/totalDelta, tags)
		add(acc, "percentageGuestNice", 100*(cts.GuestNice-lastCts.GuestNice)/totalDelta, tags)
		add(acc, "percentageStolen", 100*(cts.Stolen-lastCts.Stolen)/totalDelta, tags)

		add(acc, "percentageBusy", 100*busyDelta/totalDelta, tags)

	}

	s.lastStats = times

	return nil
}

func busyAndTotalCpuTime(t cpu.CPUTimesStat) (float64, float64) {
	busy := t.User + t.System + t.Nice + t.Iowait + t.Irq + t.Softirq + t.Steal +
		t.Guest + t.GuestNice + t.Stolen

	return busy, busy + t.Idle
}

func init() {
	plugins.Add("cpu", func() plugins.Plugin {
		realPS := &systemPS{}
		return NewCPUStats(realPS)
	})
}

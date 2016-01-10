// +build linux

package system

import (
	"fmt"

	"github.com/influxdb/telegraf/plugins/inputs"
)

type DockerStats struct {
	ps PS
}

func (_ *DockerStats) Description() string {
	return "Read metrics about docker containers"
}

func (_ *DockerStats) SampleConfig() string { return "" }

func (s *DockerStats) Gather(acc inputs.Accumulator) error {
	containers, err := s.ps.DockerStat()
	if err != nil {
		return fmt.Errorf("error getting docker info: %s", err)
	}

	for _, cont := range containers {
		tags := map[string]string{
			"id":      cont.Id,
			"name":    cont.Name,
			"command": cont.Command,
		}
		for k, v := range cont.Labels {
			tags[k] = v
		}

		cts := cont.CPU

		fields := map[string]interface{}{
			"user":       cts.User,
			"system":     cts.System,
			"idle":       cts.Idle,
			"nice":       cts.Nice,
			"iowait":     cts.Iowait,
			"irq":        cts.Irq,
			"softirq":    cts.Softirq,
			"steal":      cts.Steal,
			"guest":      cts.Guest,
			"guest_nice": cts.GuestNice,

			"cache":                  cont.Mem.Cache,
			"rss":                    cont.Mem.RSS,
			"rss_huge":               cont.Mem.RSSHuge,
			"mapped_file":            cont.Mem.MappedFile,
			"swap_in":                cont.Mem.Pgpgin,
			"swap_out":               cont.Mem.Pgpgout,
			"page_fault":             cont.Mem.Pgfault,
			"page_major_fault":       cont.Mem.Pgmajfault,
			"inactive_anon":          cont.Mem.InactiveAnon,
			"active_anon":            cont.Mem.ActiveAnon,
			"inactive_file":          cont.Mem.InactiveFile,
			"active_file":            cont.Mem.ActiveFile,
			"unevictable":            cont.Mem.Unevictable,
			"memory_limit":           cont.Mem.HierarchicalMemoryLimit,
			"total_cache":            cont.Mem.TotalCache,
			"total_rss":              cont.Mem.TotalRSS,
			"total_rss_huge":         cont.Mem.TotalRSSHuge,
			"total_mapped_file":      cont.Mem.TotalMappedFile,
			"total_swap_in":          cont.Mem.TotalPgpgIn,
			"total_swap_out":         cont.Mem.TotalPgpgOut,
			"total_page_fault":       cont.Mem.TotalPgFault,
			"total_page_major_fault": cont.Mem.TotalPgMajFault,
			"total_inactive_anon":    cont.Mem.TotalInactiveAnon,
			"total_active_anon":      cont.Mem.TotalActiveAnon,
			"total_inactive_file":    cont.Mem.TotalInactiveFile,
			"total_active_file":      cont.Mem.TotalActiveFile,
			"total_unevictable":      cont.Mem.TotalUnevictable,
		}
		acc.AddFields("docker", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("docker", func() inputs.Input {
		return &DockerStats{ps: &systemPS{}}
	})
}

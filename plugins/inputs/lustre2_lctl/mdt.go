//go:build linux

package lustre2_lctl

import (
	"fmt"

	"github.com/influxdata/telegraf"
)

func gatherMDTRecoveryStatus(volumes []string, measurement string, acc telegraf.Accumulator) {
	for _, volume := range volumes {
		result, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("mdt.%s.recovery_status", volume))
		if err != nil {
			acc.AddError(err)
			return
		}

		acc.AddGauge(measurement, map[string]interface{}{
			"recovery_status": parseRecoveryStatus(result),
		}, map[string]string{
			"volume": volume,
		})
	}
}

func gatherMDTJobstats(volumes []string, measurement string, acc telegraf.Accumulator) {
	for _, volume := range volumes {
		result, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("mdt.%s.job_stats", volume))
		if err != nil {
			acc.AddError(err)
			return
		}

		jobstats := parseJobStats(result)

		for jobid, entries := range jobstats {
			for _, entry := range entries {
				acc.AddGauge(measurement, map[string]interface{}{
					fmt.Sprintf("jobstats_%s_samples", entry.Operation): entry.Samples,
				}, map[string]string{
					"volume": volume,
					"jobid":  jobid,
				})
				acc.AddGauge(measurement, map[string]interface{}{
					fmt.Sprintf("jobstats_%s_max", entry.Operation):   entry.Max,
					fmt.Sprintf("jobstats_%s_min", entry.Operation):   entry.Min,
					fmt.Sprintf("jobstats_%s_sum", entry.Operation):   entry.Sum,
					fmt.Sprintf("jobstats_%s_sumsq", entry.Operation): entry.Sumsq,
				}, map[string]string{
					"volume": volume,
					"unit":   entry.Unit,
					"jobid":  jobid,
				})
			}
		}
	}
}

func gatherMDTStats(volumes []string, measurement string, acc telegraf.Accumulator) {
	for _, volume := range volumes {
		result, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("mdt.%s.md_stats", volume))
		if err != nil {
			acc.AddError(err)
			return
		}

		stats := parseStats(result)

		for _, stat := range stats {
			acc.AddGauge(measurement, map[string]interface{}{
				fmt.Sprintf("stats_%s_samples", stat.Operation): stat.Samples,
			}, map[string]string{
				"volume": volume,
			})
			acc.AddGauge(measurement, map[string]interface{}{
				fmt.Sprintf("stats_%s_max", stat.Operation):   stat.Max,
				fmt.Sprintf("stats_%s_min", stat.Operation):   stat.Min,
				fmt.Sprintf("stats_%s_sum", stat.Operation):   stat.Sum,
				fmt.Sprintf("stats_%s_sumsq", stat.Operation): stat.Sumsq,
			}, map[string]string{
				"volume": volume,
				"unit":   stat.Unit,
			})
		}
	}
}

func gatherMDT(collect []string, namespace string, acc telegraf.Accumulator) {
	measurement := namespace + "_mdt"

	// get volumes's name.
	result, _ := executeCommand("lctl", "get_param", "-N", "mdt.*")
	volumes := parserVolumesName(result)
	for _, c := range collect {
		switch c {
		case "mdt.*.md_stats":
			gatherMDTStats(volumes, measurement, acc)
		case "mdt.*.job_stats":
			gatherMDTJobstats(volumes, measurement, acc)
		case "mdt.*.recovery_status":
			gatherMDTRecoveryStatus(volumes, measurement, acc)
		}
	}
}

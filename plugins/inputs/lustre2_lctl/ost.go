//go:build linux

package lustre2_lctl

import (
	"fmt"

	"github.com/influxdata/telegraf"
)

func gatherOSTObdfilterRecoveryStatus(volumes []string, measurement string, acc telegraf.Accumulator) {
	for _, v := range volumes {
		content, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("obdfilter.%s.recovery_status", v))
		if err != nil {
			acc.AddError(err)
			return
		}

		acc.AddGauge(measurement, map[string]interface{}{
			"recovery_status": parseRecoveryStatus(content),
		}, map[string]string{
			"volume": v,
		})
	}
}

func gatherOSTObdfilterJobstats(volumes []string, measurement string, acc telegraf.Accumulator) {
	for _, volume := range volumes {
		result, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("obdfilter.%s.job_stats", volume))
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

func gatherOSTObdfilterStats(volumes []string, measurement string, acc telegraf.Accumulator) {
	for _, volume := range volumes {
		result, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("obdfilter.%s.stats", volume))
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

func gatherOSTObdfilterKbytestotal(volumes []string, measurement string, acc telegraf.Accumulator) {
	var capacity int64
	for _, v := range volumes {
		if result, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("obdfilter.%s.kbytestotal", v)); err != nil {
			acc.AddError(err)
		} else {
			if _, err := fmt.Sscanf(result, "%d", &capacity); err != nil {
				acc.AddError(err)
			} else {
				acc.AddGauge(measurement, map[string]interface{}{
					"capacity_kbytestotal": capacity,
				}, map[string]string{
					"volume": v,
				})
			}
		}
	}
}

func gatherOSTObdfilterKbytesavail(volumes []string, measurement string, acc telegraf.Accumulator) {
	var capacity int64
	for _, v := range volumes {
		if result, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("obdfilter.%s.kbytesavail", v)); err != nil {
			acc.AddError(err)
		} else {
			if _, err := fmt.Sscanf(result, "%d", &capacity); err != nil {
				acc.AddError(err)
			} else {
				acc.AddGauge(measurement, map[string]interface{}{
					"capacity_kbytesavail": capacity,
				}, map[string]string{
					"volume": v,
				})
			}
		}
	}
}

func gatherOSTObdfilterKbytesfree(volumes []string, measurement string, acc telegraf.Accumulator) {
	var capacity int64
	for _, v := range volumes {
		if result, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("obdfilter.%s.kbytesfree", v)); err != nil {
			acc.AddError(err)
		} else {
			if _, err := fmt.Sscanf(result, "%d", &capacity); err != nil {
				acc.AddError(err)
			} else {
				acc.AddGauge(measurement, map[string]interface{}{
					"capacity_kbytesfree": capacity,
				}, map[string]string{
					"volume": v,
				})
			}
		}
	}
}

// gatherOST gathers metrics about ost_collect.
func gatherOST(collect []string, namespace string, acc telegraf.Accumulator) {
	measurement := namespace + "_ost"

	// get volumes's name.
	result, _ := executeCommand("lctl", "get_param", "-N", "obdfilter.*")
	volumes := parserVolumesName(result)

	for _, c := range collect {
		switch c {
		case "obdfilter.*.stats":
			gatherOSTObdfilterStats(volumes, measurement, acc)
		case "obdfilter.*.job_stats":
			gatherOSTObdfilterJobstats(volumes, measurement, acc)
		case "obdfilter.*.recovery_status":
			gatherOSTObdfilterRecoveryStatus(volumes, measurement, acc)
		case "obdfilter.*.kbytesfree":
			gatherOSTObdfilterKbytesfree(volumes, measurement, acc)
		case "obdfilter.*.kbytesavail":
			gatherOSTObdfilterKbytesavail(volumes, measurement, acc)
		case "obdfilter.*.kbytestotal":
			gatherOSTObdfilterKbytestotal(volumes, measurement, acc)
		}
	}
}

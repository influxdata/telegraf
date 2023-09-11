//go:build linux

package lustre2_lctl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
)

var (
	ostObdfilterStatsOps = regexp.MustCompile(`(\w+)\s*(\d*)\s*(\w+)\s*\[(\w+)\]\s*(\d*)\s*(\d*)\s*(\d*)\s*(\d*)`)
)

// gatherOST
//
//	@param ost
//	@param namespace
//	@param acc
func gatherOST(ost OST, namespace string, acc telegraf.Accumulator, log telegraf.Logger) {
	measurement := namespace + "_ost"

	gatherObdfilter(ost.Obdfilter, measurement, acc)
}

// gatherObdfilter gather metrics about obdfilter.
//
//	@param obdfilter
//	@param acc
func gatherObdfilter(obdfilter Obdfilter, measurement string, acc telegraf.Accumulator) {
	// Get volumes' name.
	result, _ := executeCommand("lctl", "get_param", "-N", "obdfilter.*")

	volumes, _ := parserVolumesName(result)

	// gatherOSTObdfilterRecoveryStatus
	gatherOSTObdfilterRecoveryStatus(obdfilter.RecoveryStatus, volumes, measurement, acc)

	// gatherOSTObdfilterJobstats
	gatherOSTObdfilterJobstats(obdfilter.Jobstats, volumes, measurement, acc)

	// gatherOSTObdfilterStats
	gatherOSTObdfilterStats(obdfilter.Stats, volumes, measurement, acc)

	// gatherOSTCapacity
	gatherOSTCapacity(obdfilter.Capacity, volumes, measurement, acc)
}

// gatherOSTObdfilterRecoveryStatus gathers recovery status of volumes of ost.
//
//	@param flag gather flag.
//	@param volumes volumes' name.
//	@param measurement
//	@param acc
func gatherOSTObdfilterRecoveryStatus(flag bool, volumes []string, measurement string, acc telegraf.Accumulator) {
	if !flag {
		return
	}

	for _, v := range volumes {

		content, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("obdfilter.%s.recovery_status", v))
		if err != nil {
			acc.AddError(err)
			return
		}

		acc.AddGauge(measurement, map[string]interface{}{
			"recovery_status": parseRecoveryStatus(string(content)),
		}, map[string]string{
			"volume": v,
		})
	}
}

// gatherOSTObdfilterJobstats gather metrics about jobstats
//
//	@param flag gather flag.
//	@param volumes volumes' name.
//	@param measurement
//	@param acc
func gatherOSTObdfilterJobstats(flag Stats, volumes []string, measurement string, acc telegraf.Accumulator) {
	if !flag.RW && !flag.OP {
		return
	}

	for _, volume := range volumes {

		result, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("obdfilter.%s.job_stats", volume))
		if err != nil {
			acc.AddError(err)
			return
		}

		jobstats := parseJobStats(result)

		for jobid, entries := range jobstats {

			for _, entry := range entries {

				if flag.RW && (strings.Contains(entry.Operation, "read") || strings.Contains(entry.Operation, "write")) {
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

				if flag.OP && !(strings.Contains(entry.Operation, "read") || strings.Contains(entry.Operation, "write")) {
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
}

// gatherOSTObdfilterStats gathers metrics about stats.
//
//	@param flag gather flags.
//	@param volumes volumes' name.
//	@param measurement
//	@param acc
func gatherOSTObdfilterStats(flag Stats, volumes []string, measurement string, acc telegraf.Accumulator) {
	if !flag.RW && !flag.OP {
		return
	}

	for _, volume := range volumes {

		result, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("obdfilter.%s.stats", volume))
		if err != nil {
			acc.AddError(err)
			return
		}

		stats := parseStats(result)

		for _, stat := range stats {

			if flag.RW && (strings.Contains(stat.Operation, "read") || strings.Contains(stat.Operation, "write")) {
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

			if flag.OP && !(strings.Contains(stat.Operation, "read") || strings.Contains(stat.Operation, "write")) {
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
}

// gatherOSTCapacity
//
//	@param flag
//	@param volumes
//	@param measurement
//	@param acc
func gatherOSTCapacity(flag bool, volumes []string, measurement string, acc telegraf.Accumulator) {
	if !flag {
		return
	}

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

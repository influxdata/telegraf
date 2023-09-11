//go:build linux

package lustre2_lctl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
)

var (
	mdtVolPattern = regexp.MustCompile(`mdt.(.+)`)
)

// getMDTVolumes gets the name of volumes of MDT.
//
//	@return []string the name of volumes.
//	@return error
func getMDTVolumes() ([]string, error) {
	volumes := make([]string, 0)
	result, err := executeCommand("lctl", "get_param", "-N", "mdt.*")
	if err != nil {
		return nil, err
	}
	tmp := mdtVolPattern.FindAllStringSubmatch(result, -1)
	for _, value := range tmp {
		volumes = append(volumes, value[len(value)-1])
	}

	return volumes, nil
}

// gatherMDTRecoveryStatus gathers recovery status of MDT.
//
//	@param flag gather flag.
//	@param measurement
//	@param volumes the volumes' name.
//	@param acc
func gatherMDTRecoveryStatus(flag bool, measurement string, volumes []string, acc telegraf.Accumulator) {
	if !flag {
		return
	}

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

// gatherMDTJobstats gathers job status of MDT.
//
//	@param flag gather flag.
//	@param measurement
//	@param volumes the volumes' name.
//	@param acc
func gatherMDTJobstats(flag Stats, measurement string, volumes []string, acc telegraf.Accumulator) {
	if !flag.RW && !flag.OP {
		return
	}

	for _, volume := range volumes {
		result, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("mdt.%s.job_stats", volume))
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

// gatherMDTStats gathers stats of mdt.
//
//	@param flag gather flag.
//	@param measurement
//	@param volumes the volumes' name.
//	@param acc
func gatherMDTStats(flag Stats, measurement string, volumes []string, acc telegraf.Accumulator) {
	if !flag.RW && !flag.OP {
		return
	}

	for _, volume := range volumes {
		result, err := executeCommand("lctl", "get_param", "-n", fmt.Sprintf("mdt.%s.md_stats", volume))
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

func gatherMDT(mdt MDT, namespace string, acc telegraf.Accumulator) {
	measurement := namespace + "_mdt"
	// Get volumes' name.
	result, _ := executeCommand("lctl", "get_param", "-N", "mdt.*")

	volumes, _ := parserVolumesName(result)

	gatherMDTRecoveryStatus(mdt.RecoveryStatus, measurement, volumes, acc)
	gatherMDTJobstats(mdt.Jobstats, measurement, volumes, acc)
	gatherMDTStats(mdt.Stats, measurement, volumes, acc)
}

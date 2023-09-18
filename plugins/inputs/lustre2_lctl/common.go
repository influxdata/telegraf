//go:build linux

package lustre2_lctl

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
)

var (
	execCommand           = exec.Command
	volumesPattern        = regexp.MustCompile(`(\w*).(.+)`)
	recoveryStatusPattern = regexp.MustCompile(`status:(.+)`)
	jobstatJobIDPattern   = regexp.MustCompile(`- job_id:\s*(.*)`)
	jobstatEntryPattern   = regexp.MustCompile(`(\w+):\s*{([^}]+)}`)
	jobStatKVPattern      = regexp.MustCompile(`\s*(\w+):\s*([\w\s]+),?`)
	statPattern           = regexp.MustCompile(`(\S+)\s+(\d+)\s+samples\s+\[(\S+)]((?:\s+\d+)+)?`)
)

// executeCommand wraps os/exec functions.
//nolint:unparam // currently the command is always `lctl`
func executeCommand(name string, arg ...string) (string, error) {
	cmd := execCommand(name, arg...)
	result, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute command `%s`: %w", cmd.String(), err)
	}
	return string(result), nil
}

// parseRecoveryStatus parses the result of recovery_status
func parseRecoveryStatus(content string) int64 {
	status := recoveryStatusPattern.FindStringSubmatch(content)
	if strings.ToLower(strings.TrimSpace(status[1])) == "complete" {
		return 1
	}

	return 0
}

type Jobstat struct {
	Operation string
	Unit      string
	Samples   uint64
	Min       uint64
	Max       uint64
	Sum       uint64
	Sumsq     uint64
}

// parseJobStats parses the result of job_stats.
func parseJobStats(content string) map[string][]*Jobstat {
	result := make(map[string][]*Jobstat)

	scanner := bufio.NewScanner(strings.NewReader(content))

	jobid := ""
	for scanner.Scan() {
		linetext := strings.TrimSpace(scanner.Text())

		if jobstatJobIDPattern.MatchString(linetext) {
			jobid = jobstatJobIDPattern.FindStringSubmatch(linetext)[1]
			result[jobid] = make([]*Jobstat, 0)
			continue
		}

		if jobstatEntryPattern.MatchString(linetext) {
			matches := jobstatEntryPattern.FindStringSubmatch(linetext)
			if len(matches) == 3 {
				jobstats := &Jobstat{}
				jobstats.Operation = matches[1]
				innerKeyValuePairs := matches[2]

				// Find all matches for key-value pairs.
				keyValues := jobStatKVPattern.FindAllStringSubmatch(innerKeyValuePairs, -1)

				// fmt.Println("Outer Key:", outerKey)
				for _, keyValue := range keyValues {
					key := strings.TrimSpace(keyValue[1])
					value := strings.TrimSpace(keyValue[2])
					// fmt.Printf("%s: %s\n", key, value)

					switch key {
					case "samples":
						v, _ := strconv.ParseUint(value, 10, 64)
						jobstats.Samples = v
					case "min":
						v, _ := strconv.ParseUint(value, 10, 64)
						jobstats.Min = v
					case "max":
						v, _ := strconv.ParseUint(value, 10, 64)
						jobstats.Max = v
					case "sum":
						v, _ := strconv.ParseUint(value, 10, 64)
						jobstats.Sum = v
					case "sumsq":
						v, _ := strconv.ParseUint(value, 10, 64)
						jobstats.Sumsq = v
					case "unit":
						jobstats.Unit = value
					}
				}

				result[jobid] = append(result[jobid], jobstats)
			}
		}
	}

	return result
}

type Stat struct {
	Operation string
	Unit      string
	Samples   uint64
	Min       uint64
	Max       uint64
	Sum       uint64
	Sumsq     uint64
}

// parseStats parses the result of stats.
func parseStats(content string) []*Stat {
	stats := make([]*Stat, 0)
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		linetext := strings.TrimSpace(scanner.Text())
		match := statPattern.FindStringSubmatch(linetext)

		if len(match) == 0 {
			continue
		}
		result := [4]uint64{}
		samples, _ := strconv.ParseUint(match[2], 10, 64)
		valuesStr := match[4]

		if valuesStr != "" {
			values := strings.Fields(valuesStr)

			for k, valStr := range values {
				result[k], _ = strconv.ParseUint(valStr, 10, 64)
			}
		}

		stat := &Stat{
			Operation: match[1],
			Unit:      match[3],
			Samples:   samples,
			Min:       result[0],
			Max:       result[1],
			Sum:       result[2],
			Sumsq:     result[3],
		}
		stats = append(stats, stat)
	}

	return stats
}

// gatherHealth gathers health of lustre nodes.
func gatherHealth(measurement string, acc telegraf.Accumulator) {
	content, err := executeCommand("lctl", "get_param", "-n", "health_check")
	if err != nil {
		acc.AddError(err)
		return
	}

	if strings.HasPrefix(strings.ToLower(content), "health") {
		acc.AddGauge(measurement, map[string]interface{}{
			"health_check": 1,
		}, nil)
	} else {
		acc.AddGauge(measurement, map[string]interface{}{
			"health_check": 0,
		}, nil)
	}
}

// parserVolumesName parsers volumes's name from command `lctl get_param -N obdfilter.*`
func parserVolumesName(content string) []string {
	volumes := make([]string, 0)
	vsName := volumesPattern.FindAllStringSubmatch(content, -1)
	for _, value := range vsName {
		volumes = append(volumes, value[len(value)-1])
	}
	return volumes
}

package lustre2

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

func RetrvMDTRecoveryStatus() ([]Fields, []Tags, error) {
	var err error
	var stdout bytes.Buffer

	fields := make([]Fields, 0)
	tags := make([]Tags, 0)

	/* Executing command to get volume. */
	cmd := exec.Command("lctl", "get_param", "-N", "mdt.*.recovery_status")
	cmd.Stdout = &stdout
	if err = cmd.Run(); err != nil {
		return nil, nil, fmt.Errorf("unable to execute `%s`: %w", cmd.String(), err)
	}
	paramPaths := make([]string, 0)
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		paramPaths = append(paramPaths, scanner.Text())
	}

	stdout.Reset()

	/* Executing command to get recovery status. */
	for _, paramPath := range paramPaths {
		cmd := exec.Command("lctl", "get_param", "-n", paramPath)
		cmd.Stdout = &stdout
		if err = cmd.Run(); err != nil {
			return nil, nil, fmt.Errorf("unable to execute `%s`: %w", cmd.String(), err)
		}

		var rs OSTRecoveryStatus
		reader := bufio.NewReader(&stdout)
		err = yaml.NewDecoder(reader).Decode(&rs)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to decode the result of command(%s): %w", cmd.String(), err)
		}

		if strings.ToLower(rs.Status) == "complete" {
			fields = append(fields, Fields{
				"recovery_status": 1,
			})
		} else {
			fields = append(fields, Fields{
				"recovery_status": 0,
			})
		}

		tags = append(tags, Tags{ // tags must be set before setting fields in order for tags and fields to correspond.
			"volume": strings.Split(paramPath, ".")[1],
		})
	}
	return fields, tags, nil
}

// RetrvMDTJobStats retrieve data of jobstat and build data needed by accumulator.
//
//	@return []Fields
//	@return []Tags
//	@return error
func RetrvMDTJobStats() ([]Fields, []Tags, error) {
	var err error
	var stdout bytes.Buffer

	fields := make([]Fields, 0)
	tags := make([]Tags, 0)

	/* Executing command to get volume. */
	cmd := exec.Command("lctl", "get_param", "-N", "mdt.*.job_stats")
	cmd.Stdout = &stdout
	if err = cmd.Run(); err != nil {
		return nil, nil, fmt.Errorf("unable to execute `%s`: %w", cmd.String(), err)
	}
	paramPaths := make([]string, 0)
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		paramPaths = append(paramPaths, scanner.Text())
	}

	stdout.Reset()

	/* Executing command to get recovery status. */
	for _, paramPath := range paramPaths {
		cmd := exec.Command("lctl", "get_param", "-n", paramPath)
		cmd.Stdout = &stdout
		if err = cmd.Run(); err != nil {
			return nil, nil, fmt.Errorf("unable to execute `%s`: %w", cmd.String(), err)
		}

		jobstats, err := DecodeJobStats(&stdout)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to decode the result of command(%s): %w", cmd.String(), err)
		}

		for jobid, job := range jobstats {

			for metric, value := range job {

				tag := Tags{ // tags must be set before setting fields in order for tags and fields to correspond.
					"volume": strings.Split(paramPath, ".")[1],
					"jobid":  jobid,
					"unit":   value.Unit,
				}

				field := Fields{
					fmt.Sprintf("mdt_jobstats_%s_samples", metric): value.Samples,
					fmt.Sprintf("mdt_jobstats_%s_max", metric):     value.Max,
					fmt.Sprintf("mdt_jobstats_%s_min", metric):     value.Min,
					fmt.Sprintf("mdt_jobstats_%s_sum", metric):     value.Sum,
					fmt.Sprintf("mdt_jobstats_%s_sumsq", metric):   value.Sumsq,
				}

				fields = append(fields, field)
				tags = append(tags, tag)
			}
		}

	}

	return fields, tags, nil
}

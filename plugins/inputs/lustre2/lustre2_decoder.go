package lustre2

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v2"
)

// DecodeRecoveryStatus decodes the recovery status
//
//	@param r
//	@return int
//	@return error
func DecodeRecoveryStatus(r io.Reader) (int, error) {

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		linetext := scanner.Text()
		if strings.HasPrefix(linetext, "status") {
			tmp, ok := splitLinetext(linetext)
			if !ok {
				return 0, fmt.Errorf("unable to get recovery status: %s", linetext)
			}

			if strings.ToLower(tmp[1]) == "complete" {
				return 1, nil
			} else {
				return 0, nil
			}
		}
	}

	return 0, fmt.Errorf("unable to get recovery status(not found)")
}

// DecodeMDTJobStats decode result of `lctl get_param -n *.*.job_stats`
//
//	@param r the result of command.
//	@return map[string]map[string]MDTJSOperationInfo is jobid: job information.
//	@return error
func DecodeJobStats(r io.Reader) (map[string]map[string]JobStatsInfo, error) {

	rlt := make(map[string]map[string]JobStatsInfo)

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {

		linetext := scanner.Text()

		if strings.Contains(linetext, "job_id") { // Processing a block which will be stored into jobstats.
		PROCESS_JOB:
			tmp, ok := splitLinetext(linetext)
			if !ok {
				continue
			} else {
				rlt[tmp[1]] = make(map[string]JobStatsInfo)
			}

			for scanner.Scan() {
				linetext = scanner.Text()
				if strings.Contains(linetext, "job_id") { // next block.
					goto PROCESS_JOB
				}

				if operation, ok := splitLinetext(linetext); !ok {
					continue
				} else {
					var info JobStatsInfo
					err := yaml.NewDecoder(strings.NewReader(operation[1])).Decode(&info)
					if err != nil {
						continue
					} else {
						rlt[tmp[1]][operation[0]] = info
					}

				}
			}

		}
	}

	return rlt, nil
}

// splitLinetext separate from the first comma. because the formation of result of lctl get_param
// is usually title: value
//
//	@param text
//	@return []string
//	@return bool
func splitLinetext(text string) ([]string, bool) {

	data := make([]string, 0)
	for _, d := range strings.SplitN(text, ":", 2) {
		data = append(data, strings.TrimSpace(d))
	}

	return data, len(data) == 2
}

type JobStatsInfo struct {
	Samples int    `yaml:"samples,omitempty"`
	Unit    string `yaml:"unit,omitempty"`
	Min     int    `yaml:"min,omitempty"`
	Max     int    `yaml:"max,omitempty"`
	Sum     int    `yaml:"sum,omitempty"`
	Sumsq   int    `yaml:"sumsq,omitempty"`
}

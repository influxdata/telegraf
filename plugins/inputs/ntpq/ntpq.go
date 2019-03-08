package ntpq

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Mapping of ntpq header names to tag keys
var tagHeaders map[string]string = map[string]string{
	"remote": "remote",
	"refid":  "refid",
	"st":     "stratum",
	"t":      "type",
}

// Mapping of the ntpq tag key to the index in the command output
var tagI map[string]int = map[string]int{
	"remote":  -1,
	"refid":   -1,
	"stratum": -1,
	"type":    -1,
}

// Mapping of float metrics to their index in the command output
var floatI map[string]int = map[string]int{
	"delay":  -1,
	"offset": -1,
	"jitter": -1,
}

// Mapping of int metrics to their index in the command output
var intI map[string]int = map[string]int{
	"when":  -1,
	"poll":  -1,
	"reach": -1,
}

type NTPQ struct {
	runQ func() ([]byte, error)

	DNSLookup bool `toml:"dns_lookup"`
}

func (n *NTPQ) Description() string {
	return "Get standard NTP query metrics, requires ntpq executable."
}

func (n *NTPQ) SampleConfig() string {
	return `
  ## If false, set the -n ntpq flag. Can reduce metric gather time.
  dns_lookup = true
`
}

func (n *NTPQ) Gather(acc telegraf.Accumulator) error {
	out, err := n.runQ()
	if err != nil {
		return err
	}

	// Due to problems with a parsing, we have to use regexp expression in order
	// to remove string that starts from '(' and ends with space
	// see: https://github.com/influxdata/telegraf/issues/2386
	reg, err := regexp.Compile("\\s+\\([\\S]*")
	if err != nil {
		return err
	}

	lineCounter := 0
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()

		tags := make(map[string]string)
		// if there is an ntpq state prefix, remove it and make it it's own tag
		// see https://github.com/influxdata/telegraf/issues/1161
		if strings.ContainsAny(string(line[0]), "*#o+x.-") {
			tags["state_prefix"] = string(line[0])
			line = strings.TrimLeft(line, "*#o+x.-")
		}

		line = reg.ReplaceAllString(line, "")

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// If lineCounter == 0, then this is the header line
		if lineCounter == 0 {
			for i, field := range fields {
				// Check if field is a tag:
				if tagKey, ok := tagHeaders[field]; ok {
					tagI[tagKey] = i
					continue
				}

				// check if field is a float metric:
				if _, ok := floatI[field]; ok {
					floatI[field] = i
					continue
				}

				// check if field is an int metric:
				if _, ok := intI[field]; ok {
					intI[field] = i
					continue
				}
			}
		} else {
			mFields := make(map[string]interface{})

			// Get tags from output
			for key, index := range tagI {
				if index == -1 {
					continue
				}
				tags[key] = fields[index]
			}

			// Get integer metrics from output
			for key, index := range intI {
				if index == -1 || index >= len(fields) {
					continue
				}
				if fields[index] == "-" {
					continue
				}

				if key == "when" {
					when := fields[index]
					switch {
					case strings.HasSuffix(when, "h"):
						m, err := strconv.Atoi(strings.TrimSuffix(fields[index], "h"))
						if err != nil {
							acc.AddError(fmt.Errorf("E! Error ntpq: parsing int: %s", fields[index]))
							continue
						}
						// seconds in an hour
						mFields[key] = int64(m) * 3600
						continue
					case strings.HasSuffix(when, "d"):
						m, err := strconv.Atoi(strings.TrimSuffix(fields[index], "d"))
						if err != nil {
							acc.AddError(fmt.Errorf("E! Error ntpq: parsing int: %s", fields[index]))
							continue
						}
						// seconds in a day
						mFields[key] = int64(m) * 86400
						continue
					case strings.HasSuffix(when, "m"):
						m, err := strconv.Atoi(strings.TrimSuffix(fields[index], "m"))
						if err != nil {
							acc.AddError(fmt.Errorf("E! Error ntpq: parsing int: %s", fields[index]))
							continue
						}
						// seconds in a day
						mFields[key] = int64(m) * 60
						continue
					}
				}

				m, err := strconv.Atoi(fields[index])
				if err != nil {
					acc.AddError(fmt.Errorf("E! Error ntpq: parsing int: %s", fields[index]))
					continue
				}
				mFields[key] = int64(m)
			}

			// get float metrics from output
			for key, index := range floatI {
				if index == -1 || index >= len(fields) {
					continue
				}
				if fields[index] == "-" {
					continue
				}

				m, err := strconv.ParseFloat(fields[index], 64)
				if err != nil {
					acc.AddError(fmt.Errorf("E! Error ntpq: parsing float: %s", fields[index]))
					continue
				}
				mFields[key] = m
			}

			acc.AddFields("ntpq", mFields, tags)
		}

		lineCounter++
	}
	return nil
}

func (n *NTPQ) runq() ([]byte, error) {
	bin, err := exec.LookPath("ntpq")
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	if n.DNSLookup {
		cmd = exec.Command(bin, "-p")
	} else {
		cmd = exec.Command(bin, "-p", "-n")
	}

	return cmd.Output()
}

func init() {
	inputs.Add("ntpq", func() telegraf.Input {
		n := &NTPQ{}
		n.runQ = n.runq
		return n
	})
}

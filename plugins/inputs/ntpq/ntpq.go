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

// Mapping of ntpq header names to integer fields
var fieldIHeaders map[string]string = map[string]string{
	"when":  "when",
	"poll":  "poll",
	"reach": "reach",
}

// Mapping of ntpq header names to float fields
var fieldFHeaders map[string]string = map[string]string{
	"delay":  "delay",
	"offset": "offset",
	"jitter": "jitter",
}

type NTPQ struct {
	runQ   func(string) ([]byte, error)
	tagI   map[string]int
	floatI map[string]int
	intI   map[string]int

	binary string

	Servers      []string `toml:"servers"`
	DecimalReach bool     `toml:"decimal_reach"`
	WideMode     bool     `toml:"wide_mode"`
	DNSLookup    bool     `toml:"dns_lookup"`
}

const defaultServer = "localhost"

func (n *NTPQ) Description() string {
	return "Get standard NTP query metrics, requires ntpq executable."
}

func (n *NTPQ) SampleConfig() string {
	return `
  ## Specify servers to measure.
  ## If no servers are specified, then localhost is used as the host.
  # servers = ["localhost"]

  ## NTP displays the reach parameter of a peer as octal number.
  ## If you wish to convert to a regular decimal number set this to true.
  ## Default is false to preserve backwards compability.
  # decimal_reach = false

  ## If false, set the -n ntpq flag. Can reduce metric gather time.
  dns_lookup = true

  ## Try to use wide mode (-w) in output from ntpq if binary supports it.
  ## Default is false to preserve backwards compability.
  # wide_mode = false
`
}

func (n *NTPQ) Gather(acc telegraf.Accumulator) error {
	// Due to problems with a parsing, we have to use regexp expression in order
	// to remove string that starts from '(' and ends with space
	// see: https://github.com/influxdata/telegraf/issues/2386
	reg, err := regexp.Compile("\\s+\\([\\S]*")
	if err != nil {
		return err
	}

	for _, server := range n.Servers {
		out, err := n.runQ(server)
		if err != nil {
			return err
		}

		lineCounter := 0
		numColumns := 0
		scanner := bufio.NewScanner(bytes.NewReader(out))
		for scanner.Scan() {
			line := scanner.Text()
			line = reg.ReplaceAllString(line, "")

			tags := make(map[string]string)
			tags["server"] = string(server)

			fields := strings.Fields(line)

			// Wide mode have broken lines for long hostnames
			if n.WideMode {
				if len(fields) == 1 {
					// broken line, append next line
					scanner.Scan()
					line := scanner.Text()
					line = reg.ReplaceAllString(line, "")
					fields = append(fields, strings.Fields(line)...)
					lineCounter++
				}
			} else {
				// not running wide mode, check if number of fields is valid
				if len(fields) < 2 {
					lineCounter++
					continue
				}
			}

			// If lineCounter == 0, then this is the header line
			if lineCounter == 0 {
				numColumns = len(fields)
				for i, field := range fields {
					// Check if field is a tag:
					if tagKey, ok := tagHeaders[field]; ok {
						n.tagI[tagKey] = i
						continue
					}

					// check if field is a float metric:
					if fieldKey, ok := fieldFHeaders[field]; ok {
						n.floatI[fieldKey] = i
						continue
					}

					// check if field is an int metric:
					if fieldKey, ok := fieldIHeaders[field]; ok {
						n.intI[fieldKey] = i
						continue
					}
				}

				// all done, skip bar next line it's the bar below the header
				// and increase line counter and next for loop
				scanner.Scan()
				lineCounter++
			} else {
				if len(fields) != numColumns {
					lineCounter++
					continue
				}

				mFields := make(map[string]interface{})

				// Get tags from output
				for key, index := range n.tagI {
					if index == -1 {
						continue
					}
					if key == "remote" {
						// if there is an ntpq state prefix, remove it and make it it's own tag
						// see https://github.com/influxdata/telegraf/issues/1161
						if strings.ContainsAny(string(fields[index][0]), "*#o+x.-") {
							tags["state_prefix"] = string(fields[index][0])
							fields[index] = strings.TrimLeft(fields[index], "*#o+x.-")
						}
					}
					tags[key] = fields[index]
				}

				// Get integer metrics from output
				for key, index := range n.intI {
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
							// simply just seconds
							mFields[key] = int64(m) * 60
							continue
						}
					}

					if key == "reach" && n.DecimalReach {
						reach := fields[index]

						m, err := strconv.ParseInt(reach, 8, 64)
						if err != nil {
							acc.AddError(fmt.Errorf("E! Error ntpq: parsing octal int: %s", fields[index]))
							continue
						}

						mFields[key] = int64(m)
						continue
					}

					m, err := strconv.Atoi(fields[index])
					if err != nil {
						acc.AddError(fmt.Errorf("E! Error ntpq: parsing int: %s", fields[index]))
						continue
					}
					mFields[key] = int64(m)
				}

				// get float metrics from output
				for key, index := range n.floatI {
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
	}
	return nil
}

func (n *NTPQ) runq(server string) ([]byte, error) {
	// do we have a valid binary?
	bin, err := exec.LookPath(n.binary)
	if err != nil {
		return nil, err
	} else {
		n.binary = bin
	}

	var cmd *exec.Cmd

	// build args
	cmd_args := []string{"-p"}

	// is wide mode supported?
	if n.WideMode {
		cmd_args = append(cmd_args, "-w")
	}

	// should we not use dns lookups?
	if !n.DNSLookup {
		cmd_args = append(cmd_args, "-n")
	}

	// append server to argument list
	cmd_args = append(cmd_args, server)
	cmd = exec.Command(n.binary, cmd_args...)

	return cmd.Output()
}

func newNTPQ() *NTPQ {
	// Mapping of the ntpq tag key to the index in the command output
	tagI := map[string]int{
		"remote":  -1,
		"refid":   -1,
		"stratum": -1,
		"type":    -1,
	}

	// Mapping of float metrics to their index in the command output
	floatI := map[string]int{
		"delay":  -1,
		"offset": -1,
		"jitter": -1,
	}

	// Mapping of int metrics to their index in the command output
	intI := map[string]int{
		"when":  -1,
		"poll":  -1,
		"reach": -1,
	}

	n := &NTPQ{
		tagI:     tagI,
		floatI:   floatI,
		intI:     intI,
		binary:   "ntpq",
		WideMode: false,
	}

	// no server, default to defaultServer (localhost)
	if len(n.Servers) == 0 {
		n.Servers = append(n.Servers, defaultServer)
	}

	// find binary
	if bin, err := exec.LookPath(n.binary); err == nil {
		// binary found so lets store absolute/relative path to
		// skip checking PATH in each LookPath call
		n.binary = bin

		if n.WideMode {
			// we wish to use wide mode so let's check if binary supports it
			if _, err := exec.Command(n.binary, "-w", "-c quit").Output(); err == nil {
				// GREAT SUCCESS
				n.WideMode = true
			} else {
				// EPIC FAIL, disable it
				n.WideMode = false
			}
		}
	} else {
		// could not find binary so check again at gather
		// this disables wide mode until telegraf is restarted
		n.binary = "ntpq"
	}

	n.runQ = n.runq
	return n
}

func init() {
	inputs.Add("ntpq", func() telegraf.Input {
		return newNTPQ()
	})
}

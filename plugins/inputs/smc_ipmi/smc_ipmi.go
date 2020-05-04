package smc_ipmi

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	execCommand       = exec.Command // execCommand is used to mock commands in tests.
	execCommandPminfo = exec.Command
	re_ipmi           = regexp.MustCompile(`^(?P<status>[^|GS-]*)\|(?P<sensor>[^|]*)\|(?P<reading>[^|]*)\|(?P<low_limit>[^|]*)\|(?P<high_limit>[^|]*)\|`)
	re_sensor_num     = regexp.MustCompile(`^\(\d+\)`)
	re_temp_reading   = regexp.MustCompile(`^\d+C\/\d+F`)
	re_pmb            = regexp.MustCompile(`^(?P<item>[^[|-]*)\|(?P<value>[^|]*)`)
)

// Smcipmi plugin options
type Smcipmi struct {
	Path     string
	Servers  []string
	TempUnit string
	Timeout  internal.Duration
}

// Description for plugin
func (s *Smcipmi) Description() string {
	return "Reads IPMI data via SMCIPMITool"
}

// SampleConfig returns example config for plugin
func (s *Smcipmi) SampleConfig() string {
	return `
  ## Path to SMCIPMITool executable
  ## https://www.supermicro.com/en/solutions/management-software/ipmi-utilities
  # path = "/usr/bin/smcipmitool/SMCIPMITool"

  # servers = ["USERID:PASSW0RD@(192.168.1.1)"]

  ## Retrieve temperature values as celsius "C" or fahrenheit "F"
  ## Defaults to celsius
  # temp_unit = "F"
`
}

// Gather - main function
func (s *Smcipmi) Gather(acc telegraf.Accumulator) error {
	if len(s.Path) == 0 {
		return fmt.Errorf("SMCIPMITool not found: Ensure SMCIPMITool is in your PATH or defined in your config")
	}

	if len(s.Servers) > 0 {
		wg := sync.WaitGroup{}
		for _, server := range s.Servers {
			wg.Add(1)
			go func(a telegraf.Accumulator, st string) {
				defer wg.Done()
				err := s.parse(a, st)
				if err != nil {
					a.AddError(err)
				}
			}(acc, server)
		}
		wg.Wait()
	} else {
		err := s.parse(acc, "")
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Smcipmi) parse(acc telegraf.Accumulator, server string) error {
	opts := make([]string, 0)
	hostname := ""
	if server != "" {
		conn := NewConnection(server)
		hostname = conn.Hostname
		opts = conn.options()
	}

	ipmiSensorOpts := append(opts, "ipmi", "sensor")
	pmbOpts := append(opts, "pminfo")

	smcIpmiToolPath := s.Path

	ipmiSensorCmd := execCommand(smcIpmiToolPath, ipmiSensorOpts...)
	pmbCmd := execCommandPminfo(smcIpmiToolPath, pmbOpts...)

	ipmiSensorOut, ipmiSensorErr := internal.CombinedOutputTimeout(ipmiSensorCmd, s.Timeout.Duration)
	if ipmiSensorErr != nil {
		return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(ipmiSensorCmd.Args, " "), ipmiSensorErr, string(ipmiSensorOut))
	}

	pmbOut, pmbErr := internal.CombinedOutputTimeout(pmbCmd, s.Timeout.Duration)
	if pmbErr != nil {
		return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(pmbCmd.Args, " "), pmbErr, string(pmbOut))
	}

	return parseSmcIpmi(acc, hostname, s.TempUnit, ipmiSensorOut, pmbOut, time.Now())
}

func parseSmcIpmi(acc telegraf.Accumulator, hostname string, tempUnit string, ipmiSensorOut []byte, pmbOut []byte, measured_at time.Time) error {
	scanner := bufio.NewScanner(bytes.NewReader(ipmiSensorOut))
	for scanner.Scan() {
		ipmiFields := extractFieldsFromRegex(re_ipmi, scanner.Text())
		if len(ipmiFields) != 5 {
			continue
		}

		tags := map[string]string{
			"name": transform(re_sensor_num.ReplaceAllString(ipmiFields["sensor"], "")),
		}

		// tag the server is we have one
		if hostname != "" {
			tags["server"] = hostname
		}

		fields := make(map[string]interface{})

		if strings.EqualFold("OK", trim(ipmiFields["status"])) {
			fields["status"] = 1
		} else {
			fields["status"] = 0
		}

		reading := ipmiFields["reading"]

		// handle temp
		isTemp := re_temp_reading.MatchString(reading)
		if isTemp {
			fields["value"] = toTemp(reading, tempUnit)
			tags["unit"] = transform(tempUnit)
		} else if strings.Index(reading, " ") > 0 {
			// split middle column into value and unit
			valunit := strings.SplitN(reading, " ", 2)
			var err error
			fields["value"], err = aToFloat(valunit[0])
			if err != nil {
				continue
			}
			if len(valunit) > 1 {
				tags["unit"] = transform(valunit[1])
			}
		} else {
			fields["value"] = 0.0
		}

		acc.AddFields("smc_ipmi", fields, tags, measured_at)
	}

	// PMBus info
	pmbScanner := bufio.NewScanner(bytes.NewReader(pmbOut))
	for pmbScanner.Scan() {
		pmbFields := extractFieldsFromRegex(re_pmb, pmbScanner.Text())
		item := transform(pmbFields["item"])

		// Skip invalid, header row, or rows we don't care about
		if len(pmbFields) != 2 ||
			item == "item" || item == "pmbus_revision" ||
			strings.HasPrefix(item, "pws") {
			continue
		}

		tags := map[string]string{
			"name": fmt.Sprintf("pmbus_%s", item),
		}

		fields := make(map[string]interface{})

		value := pmbFields["value"]
		if item == "status" {
			if strings.Contains(value, "STATUS OK") {
				fields["status"] = 1
			} else {
				fields["status"] = 0
			}
		} else {
			isTemp := re_temp_reading.MatchString(value)
			if isTemp {
				fields["value"] = toTemp(value, tempUnit)
				tags["unit"] = transform(tempUnit)
			} else if strings.Index(value, " ") > 0 {
				// split middle column into value and unit
				valunit := strings.SplitN(value, " ", 2)
				var err error
				fields["value"], err = aToFloat(valunit[0])
				if err != nil {
					continue
				}
				if len(valunit) > 1 {
					tags["unit"] = transform(valunit[1])
				}
			} else {
				fields["value"] = 0.0
			}
		}

		// tag the server is we have one
		if hostname != "" {
			tags["server"] = hostname
		}

		acc.AddFields("smc_ipmi", fields, tags, measured_at)
	}

	return scanner.Err()
}

func toTemp(value string, tempUnit string) float64 {
	var tempVal string
	temps := strings.SplitN(value, "/", 2)

	if tempUnit == "F" {
		tempVal = temps[1]
	} else {
		tempVal = temps[0]
	}

	tempInt, err := strconv.ParseInt(strings.TrimRight(tempVal, tempUnit), 0, 64)
	if err != nil {
		log.Printf("E! Error parsing to Int: '%s'", tempVal)
	}

	return float64(tempInt)
}

// extractFieldsFromRegex consumes a regex with named capture groups and returns a kvp map of strings with the results
func extractFieldsFromRegex(re *regexp.Regexp, input string) map[string]string {
	submatches := re.FindStringSubmatch(input)
	results := make(map[string]string)
	subexpNames := re.SubexpNames()
	//log.Printf("D! submatches: %s", submatches)
	//log.Printf("D! subexNames: %s", subexpNames)
	if len(subexpNames) > len(submatches) {
		log.Printf("D! No matches found in '%s'", input)
		return results
	}
	for i, name := range subexpNames {
		if name != input && name != "" && input != "" {
			results[name] = trim(submatches[i])
		}
	}
	return results
}

// aToFloat converts string representations of numbers to float64 values
func aToFloat(val string) (float64, error) {
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0.0, err
	}
	return f, nil
}

func trim(s string) string {
	return strings.TrimSpace(s)
}

func transform(s string) string {
	s = trim(s)
	s = strings.ToLower(s)
	return strings.Replace(s, " ", "_", -1)
}

func init() {
	s := Smcipmi{}
	path, _ := exec.LookPath("SMCIPMITool")
	if len(path) > 0 {
		s.Path = path
	}
	s.Timeout = internal.Duration{Duration: time.Second * 20}
	inputs.Add("smc_ipmi", func() telegraf.Input {
		s := s
		return &s
	})
}

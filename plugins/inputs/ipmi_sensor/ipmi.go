package ipmi_sensor

import (
	"fmt"
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
	execCommand = exec.Command // execCommand is used to mock commands in tests.
)

// Ipmi stores the configuration values for the ipmi_sensor input plugin
type Ipmi struct {
	Path          string
	Privilege     string
	Servers       []string
	Timeout       internal.Duration
	SchemaVersion int
}

var sampleConfig = `
  ## optionally specify the path to the ipmitool executable
  # path = "/usr/bin/ipmitool"
  ##
  ## optionally force session privilege level. Can be CALLBACK, USER, OPERATOR, ADMINISTRATOR
  # privilege = "ADMINISTRATOR"
  ##
  ## optionally specify one or more servers via a url matching
  ##  [username[:password]@][protocol[(address)]]
  ##  e.g.
  ##    root:passwd@lan(127.0.0.1)
  ##
  ## if no servers are specified, local machine sensor stats will be queried
  ##
  # servers = ["USERID:PASSW0RD@lan(192.168.1.1)"]

  ## Recommended: use metric 'interval' that is a multiple of 'timeout' to avoid
  ## gaps or overlap in pulled data
  interval = "30s"

  ## Timeout for the ipmitool command to complete
  timeout = "20s"

  ## Schema Version: (Optional, defaults to version 1)
  schemaVersion = 2
`

// SampleConfig returns the documentation about the sample configuration
func (m *Ipmi) SampleConfig() string {
	return sampleConfig
}

// Description returns a basic description for the plugin functions
func (m *Ipmi) Description() string {
	return "Read metrics from the bare metal servers via IPMI"
}

// Gather is the main execution function for the plugin
func (m *Ipmi) Gather(acc telegraf.Accumulator) error {
	if len(m.Path) == 0 {
		return fmt.Errorf("ipmitool not found: verify that ipmitool is installed and that ipmitool is in your PATH")
	}

	if len(m.Servers) > 0 {
		wg := sync.WaitGroup{}
		for _, server := range m.Servers {
			wg.Add(1)
			go func(a telegraf.Accumulator, s string) {
				defer wg.Done()
				err := m.parse(a, s)
				if err != nil {
					a.AddError(err)
				}
			}(acc, server)
		}
		wg.Wait()
	} else {
		err := m.parse(acc, "")
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Ipmi) parse(acc telegraf.Accumulator, server string) error {
	opts := make([]string, 0)
	hostname := ""
	if server != "" {
		conn := NewConnection(server, m.Privilege)
		hostname = conn.Hostname
		opts = conn.options()
	}
	opts = append(opts, "sdr")
	if m.SchemaVersion == 2 {
		opts = append(opts, "elist")
	}
	cmd := execCommand(m.Path, opts...)
	out, err := internal.CombinedOutputTimeout(cmd, m.Timeout.Duration)
	if err != nil {
		return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(cmd.Args, " "), err, string(out))
	}
	if m.SchemaVersion == 2 {
		return parseV2(acc, hostname, string(out))
	}
	return parseV1(acc, hostname, string(out))
}

func parseV1(acc telegraf.Accumulator, hostname string, cmdOut string) error {
	// each line will look something like
	// Planar VBAT      | 3.05 Volts        | ok
	lines := strings.Split(cmdOut, "\n")
	for i := 0; i < len(lines); i++ {
		ipmiFields := strings.Split(lines[i], "|")
		if len(ipmiFields) != 3 {
			continue
		}

		tags := map[string]string{
			"name": transform(ipmiFields[0]),
		}

		// tag the server is we have one
		if hostname != "" {
			tags["server"] = hostname
		}

		fields := make(map[string]interface{})
		if strings.EqualFold("ok", trim(ipmiFields[2])) {
			fields["status"] = 1
		} else {
			fields["status"] = 0
		}

		val1 := trim(ipmiFields[1])

		if strings.Index(val1, " ") > 0 {
			// split middle column into value and unit
			valunit := strings.SplitN(val1, " ", 2)
			fields["value"] = aToFloat(valunit[0])
			if len(valunit) > 1 {
				tags["unit"] = transform(valunit[1])
			}
		} else {
			fields["value"] = 0.0
		}

		acc.AddFields("ipmi_sensor", fields, tags, time.Now())
	}

	return nil
}

func parseV2(acc telegraf.Accumulator, hostname string, cmdOut string) error {
	// each line will look something like
	// CMOS Battery     | 65h | ok  |  7.1 |
	// Temp             | 0Eh | ok  |  3.1 | 55 degrees C
	// Drive 0          | A0h | ok  |  7.1 | Drive Present
	lines := strings.Split(cmdOut, "\n")
	for i := 0; i < len(lines); i++ {
		ipmiFields := strings.Split(lines[i], "|")
		if len(ipmiFields) != 5 {
			continue
		}

		tags := map[string]string{
			"name": transform(ipmiFields[0]),
		}

		// tag the server is we have one
		if hostname != "" {
			tags["server"] = hostname
		}
		tags["entity_id"] = transform(ipmiFields[3])
		tags["status_code"] = trim(ipmiFields[2])

		fields := make(map[string]interface{})
		result := extractFieldsFromRegex(`^(?P<analogValue>[0-9.]+)\s(?P<analogUnit>.*)|(?P<status>.+)|^$`, trim(ipmiFields[4]))
		// This is an analog value with a unit
		if result["analogValue"] != "" && len(result["analogUnit"]) >= 1 {
			fields["value"] = aToFloat(result["analogValue"])
			// Some implementations add an extra status to their analog units
			unitResults := extractFieldsFromRegex(`^(?P<realAnalogUnit>[^,]+)(?:,\s*(?P<statusDesc>.*))?`, result["analogUnit"])
			tags["unit"] = transform(unitResults["realAnalogUnit"])
			if unitResults["statusDesc"] != "" {
				tags["status_desc"] = transform(unitResults["statusDesc"])
			}
		} else {
			// This is a status value
			fields["value"] = 0.0
			// Extended status descriptions aren't required, in which case for consistency re-use the status code
			if result["status"] != "" {
				tags["status_desc"] = transform(result["status"])
			} else {
				tags["status_desc"] = transform(ipmiFields[2])
			}
		}

		acc.AddFields("ipmi_sensor", fields, tags, time.Now())
	}

	return nil
}

// extractFieldsFromRegex consumes a regex with named capture groups and returns a kvp map of strings with the results
func extractFieldsFromRegex(regex string, input string) map[string]string {
	re := regexp.MustCompile(regex)
	submatches := re.FindStringSubmatch(input)
	results := make(map[string]string)
	for i, name := range re.SubexpNames() {
		results[name] = submatches[i]
	}
	return results
}

// aToFloat converts string representations of numbers to float64 values
func aToFloat(val string) float64 {
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0.0
	}
	return f
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
	m := Ipmi{}
	path, _ := exec.LookPath("ipmitool")
	if len(path) > 0 {
		m.Path = path
	}
	m.Timeout = internal.Duration{Duration: time.Second * 20}
	inputs.Add("ipmi_sensor", func() telegraf.Input {
		m := m
		return &m
	})
}

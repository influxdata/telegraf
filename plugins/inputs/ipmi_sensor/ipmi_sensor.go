package ipmi_sensor

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	execCommand          = exec.Command // execCommand is used to mock commands in tests.
	reV1ParseLine        = regexp.MustCompile(`^(?P<name>[^|]*)\|(?P<description>[^|]*)\|(?P<status_code>.*)`)
	reV2ParseLine        = regexp.MustCompile(`^(?P<name>[^|]*)\|[^|]+\|(?P<status_code>[^|]*)\|(?P<entity_id>[^|]*)\|(?:(?P<description>[^|]+))?`)
	reV2ParseDescription = regexp.MustCompile(`^(?P<analogValue>-?[0-9.]+)\s(?P<analogUnit>.*)|(?P<status>.+)|^$`)
	reV2ParseUnit        = regexp.MustCompile(`^(?P<realAnalogUnit>[^,]+)(?:,\s*(?P<statusDesc>.*))?`)
)

// Ipmi stores the configuration values for the ipmi_sensor input plugin
type Ipmi struct {
	Path          string
	Privilege     string
	HexKey        string `toml:"hex_key"`
	Servers       []string
	Timeout       config.Duration
	MetricVersion int
	UseSudo       bool
	UseCache      bool
	CachePath     string

	Log telegraf.Logger `toml:"-"`
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
		conn := NewConnection(server, m.Privilege, m.HexKey)
		hostname = conn.Hostname
		opts = conn.options()
	}
	opts = append(opts, "sdr")
	if m.UseCache {
		cacheFile := filepath.Join(m.CachePath, server+"_ipmi_cache")
		_, err := os.Stat(cacheFile)
		if os.IsNotExist(err) {
			dumpOpts := opts
			// init cache file
			dumpOpts = append(dumpOpts, "dump")
			dumpOpts = append(dumpOpts, cacheFile)
			name := m.Path
			if m.UseSudo {
				// -n - avoid prompting the user for input of any kind
				dumpOpts = append([]string{"-n", name}, dumpOpts...)
				name = "sudo"
			}
			cmd := execCommand(name, dumpOpts...)
			out, err := internal.CombinedOutputTimeout(cmd, time.Duration(m.Timeout))
			if err != nil {
				return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(sanitizeIPMICmd(cmd.Args), " "), err, string(out))
			}
		}
		opts = append(opts, "-S")
		opts = append(opts, cacheFile)
	}
	if m.MetricVersion == 2 {
		opts = append(opts, "elist")
	}
	name := m.Path
	if m.UseSudo {
		// -n - avoid prompting the user for input of any kind
		opts = append([]string{"-n", name}, opts...)
		name = "sudo"
	}
	cmd := execCommand(name, opts...)
	out, err := internal.CombinedOutputTimeout(cmd, time.Duration(m.Timeout))
	timestamp := time.Now()
	if err != nil {
		return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(sanitizeIPMICmd(cmd.Args), " "), err, string(out))
	}
	if m.MetricVersion == 2 {
		return m.parseV2(acc, hostname, out, timestamp)
	}
	return m.parseV1(acc, hostname, out, timestamp)
}

func (m *Ipmi) parseV1(acc telegraf.Accumulator, hostname string, cmdOut []byte, measuredAt time.Time) error {
	// each line will look something like
	// Planar VBAT      | 3.05 Volts        | ok
	scanner := bufio.NewScanner(bytes.NewReader(cmdOut))
	for scanner.Scan() {
		ipmiFields := m.extractFieldsFromRegex(reV1ParseLine, scanner.Text())
		if len(ipmiFields) != 3 {
			continue
		}

		tags := map[string]string{
			"name": transform(ipmiFields["name"]),
		}

		// tag the server is we have one
		if hostname != "" {
			tags["server"] = hostname
		}

		fields := make(map[string]interface{})
		if strings.EqualFold("ok", trim(ipmiFields["status_code"])) {
			fields["status"] = 1
		} else {
			fields["status"] = 0
		}

		description := ipmiFields["description"]

		// handle hex description field
		if strings.HasPrefix(description, "0x") {
			descriptionInt, err := strconv.ParseInt(description, 0, 64)
			if err != nil {
				continue
			}

			fields["value"] = float64(descriptionInt)
		} else if strings.Index(description, " ") > 0 {
			// split middle column into value and unit
			valunit := strings.SplitN(description, " ", 2)
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

		acc.AddFields("ipmi_sensor", fields, tags, measuredAt)
	}

	return scanner.Err()
}

func (m *Ipmi) parseV2(acc telegraf.Accumulator, hostname string, cmdOut []byte, measuredAt time.Time) error {
	// each line will look something like
	// CMOS Battery     | 65h | ok  |  7.1 |
	// Temp             | 0Eh | ok  |  3.1 | 55 degrees C
	// Drive 0          | A0h | ok  |  7.1 | Drive Present
	scanner := bufio.NewScanner(bytes.NewReader(cmdOut))
	for scanner.Scan() {
		ipmiFields := m.extractFieldsFromRegex(reV2ParseLine, scanner.Text())
		if len(ipmiFields) < 3 || len(ipmiFields) > 4 {
			continue
		}

		tags := map[string]string{
			"name": transform(ipmiFields["name"]),
		}

		// tag the server is we have one
		if hostname != "" {
			tags["server"] = hostname
		}
		tags["entity_id"] = transform(ipmiFields["entity_id"])
		tags["status_code"] = trim(ipmiFields["status_code"])
		fields := make(map[string]interface{})
		descriptionResults := m.extractFieldsFromRegex(reV2ParseDescription, trim(ipmiFields["description"]))
		// This is an analog value with a unit
		if descriptionResults["analogValue"] != "" && len(descriptionResults["analogUnit"]) >= 1 {
			var err error
			fields["value"], err = aToFloat(descriptionResults["analogValue"])
			if err != nil {
				continue
			}
			// Some implementations add an extra status to their analog units
			unitResults := m.extractFieldsFromRegex(reV2ParseUnit, descriptionResults["analogUnit"])
			tags["unit"] = transform(unitResults["realAnalogUnit"])
			if unitResults["statusDesc"] != "" {
				tags["status_desc"] = transform(unitResults["statusDesc"])
			}
		} else {
			// This is a status value
			fields["value"] = 0.0
			// Extended status descriptions aren't required, in which case for consistency re-use the status code
			if descriptionResults["status"] != "" {
				tags["status_desc"] = transform(descriptionResults["status"])
			} else {
				tags["status_desc"] = transform(ipmiFields["status_code"])
			}
		}

		acc.AddFields("ipmi_sensor", fields, tags, measuredAt)
	}

	return scanner.Err()
}

// extractFieldsFromRegex consumes a regex with named capture groups and returns a kvp map of strings with the results
func (m *Ipmi) extractFieldsFromRegex(re *regexp.Regexp, input string) map[string]string {
	submatches := re.FindStringSubmatch(input)
	results := make(map[string]string)
	subexpNames := re.SubexpNames()
	if len(subexpNames) > len(submatches) {
		m.Log.Debugf("No matches found in '%s'", input)
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

func sanitizeIPMICmd(args []string) []string {
	for i, v := range args {
		if v == "-P" {
			args[i+1] = "REDACTED"
		}
	}

	return args
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
	m.Timeout = config.Duration(time.Second * 20)
	m.UseCache = false
	m.CachePath = os.TempDir()
	inputs.Add("ipmi_sensor", func() telegraf.Input {
		m := m
		return &m
	})
}

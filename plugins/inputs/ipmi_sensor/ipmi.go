package ipmi_sensor

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	execCommand = exec.Command // execCommand is used to mock commands in tests.
)

type Ipmi struct {
	Path    string
	Servers []string
	Timeout internal.Duration
}

var sampleConfig = `
  ## optionally specify the path to the ipmitool executable
  # path = "/usr/bin/ipmitool"
  #
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
`

func (m *Ipmi) SampleConfig() string {
	return sampleConfig
}

func (m *Ipmi) Description() string {
	return "Read metrics from the bare metal servers via IPMI"
}

func (m *Ipmi) Gather(acc telegraf.Accumulator) error {
	if len(m.Path) == 0 {
		return fmt.Errorf("ipmitool not found: verify that ipmitool is installed and that ipmitool is in your PATH")
	}

	if len(m.Servers) > 0 {
		for _, server := range m.Servers {
			err := m.parse(acc, server)
			if err != nil {
				acc.AddError(err)
				continue
			}
		}
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
		conn := NewConnection(server)
		hostname = conn.Hostname
		opts = conn.options()
	}

	opts = append(opts, "sdr")
	cmd := execCommand(m.Path, opts...)
	out, err := internal.CombinedOutputTimeout(cmd, m.Timeout.Duration)
	if err != nil {
		return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(cmd.Args, " "), err, string(out))
	}

	// each line will look something like
	// Planar VBAT      | 3.05 Volts        | ok
	lines := strings.Split(string(out), "\n")
	for i := 0; i < len(lines); i++ {
		vals := strings.Split(lines[i], "|")
		if len(vals) != 3 {
			continue
		}

		tags := map[string]string{
			"name": transform(vals[0]),
		}

		// tag the server is we have one
		if hostname != "" {
			tags["server"] = hostname
		}

		fields := make(map[string]interface{})
		if strings.EqualFold("ok", trim(vals[2])) {
			fields["status"] = 1
		} else {
			fields["status"] = 0
		}

		val1 := trim(vals[1])

		if strings.Index(val1, " ") > 0 {
			// split middle column into value and unit
			valunit := strings.SplitN(val1, " ", 2)
			fields["value"] = Atofloat(valunit[0])
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

func Atofloat(val string) float64 {
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0.0
	} else {
		return f
	}
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

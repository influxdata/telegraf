package ipmi_sensor

import (
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Ipmi struct {
	Servers []string
	runner  Runner
}

var sampleConfig = `
  ## specify servers via a url matching:
  ##  [username[:password]@][protocol[(address)]]
  ##  e.g.
  ##    root:passwd@lan(127.0.0.1)
  ##
  servers = ["USERID:PASSW0RD@lan(192.168.1.1)"]
`

func NewIpmi() *Ipmi {
	return &Ipmi{
		runner: CommandRunner{},
	}
}

func (m *Ipmi) SampleConfig() string {
	return sampleConfig
}

func (m *Ipmi) Description() string {
	return "Read metrics from one or many bare metal servers"
}

func (m *Ipmi) Gather(acc telegraf.Accumulator) error {
	if m.runner == nil {
		m.runner = CommandRunner{}
	}
	for _, serv := range m.Servers {
		err := m.gatherServer(serv, acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Ipmi) gatherServer(serv string, acc telegraf.Accumulator) error {
	conn := NewConnection(serv)

	res, err := m.runner.Run(conn, "sdr")
	if err != nil {
		return err
	}

	// each line will look something like
	// Planar VBAT      | 3.05 Volts        | ok
	lines := strings.Split(res, "\n")
	for i := 0; i < len(lines); i++ {
		vals := strings.Split(lines[i], "|")
		if len(vals) != 3 {
			continue
		}

		tags := map[string]string{
			"server": conn.Hostname,
			"name":   transform(vals[0]),
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

type Runner interface {
	Run(conn *Connection, args ...string) (string, error)
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
	inputs.Add("ipmi_sensor", func() telegraf.Input {
		return &Ipmi{}
	})
}

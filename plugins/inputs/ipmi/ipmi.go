// ipmi
package ipmi

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

	lines := strings.Split(res, "\n")

	for i := 0; i < len(lines); i++ {
		vals := strings.Split(lines[i], "|")
		if len(vals) == 3 {
			tags := map[string]string{"server": conn.Hostname, "name": trim(vals[0])}
			fields := make(map[string]interface{})
			if strings.EqualFold("ok", trim(vals[2])) {
				fields["status"] = 1
			} else {
				fields["status"] = 0
			}

			val1 := trim(vals[1])

			if strings.Index(val1, " ") > 0 {
				val := strings.Split(val1, " ")[0]
				fields["value"] = Atofloat(val)
			} else {
				fields["value"] = 0.0
			}

			acc.AddFields("ipmi_sensor", fields, tags, time.Now())
		}
	}

	return nil
}

type Runner interface {
	Run(conn *Connection, args ...string) (string, error)
}

func Atofloat(val string) float64 {
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return float64(0)
	} else {
		return float64(f)
	}
}

func trim(s string) string {
	return strings.TrimSpace(s)
}

func init() {
	inputs.Add("ipmi", func() telegraf.Input {
		return &Ipmi{}
	})
}

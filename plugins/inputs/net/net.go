//go:generate ../../../tools/readme_config_includer/generator
package net

import (
	_ "embed"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/psutil"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Net struct {
	Interfaces          []string `toml:"interfaces"`
	IgnoreProtocolStats bool     `toml:"ignore_protocol_stats" deprecated:"1.37.0;1.45.0;option is ignored"`

	filter     filter.Filter
	ps         psutil.PS
	skipChecks bool
}

func (*Net) SampleConfig() string {
	return sampleConfig
}

func (n *Net) Init() error {
	// So not use the interface list of the system if the HOST_PROC variable is
	// set as the interfaces are determined by a syscall and therefore might
	// differ especially in container environments.
	n.skipChecks = os.Getenv("HOST_PROC") != ""

	return nil
}

func (n *Net) Gather(acc telegraf.Accumulator) error {
	netio, err := n.ps.NetIO()
	if err != nil {
		return fmt.Errorf("error getting net io info: %w", err)
	}

	if n.filter == nil {
		if n.filter, err = filter.Compile(n.Interfaces); err != nil {
			return fmt.Errorf("error compiling filter: %w", err)
		}
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("error getting list of interfaces: %w", err)
	}
	interfacesByName := make(map[string]net.Interface, len(interfaces))
	for _, iface := range interfaces {
		interfacesByName[iface.Name] = iface
	}

	for _, io := range netio {
		if len(n.Interfaces) != 0 {
			var found bool

			if n.filter.Match(io.Name) {
				found = true
			}

			if !found {
				continue
			}
		} else if !n.skipChecks {
			iface, ok := interfacesByName[io.Name]
			if !ok {
				continue
			}

			if iface.Flags&net.FlagLoopback == net.FlagLoopback {
				continue
			}

			if iface.Flags&net.FlagUp == 0 {
				continue
			}
		}

		tags := map[string]string{
			"interface": io.Name,
		}

		fields := map[string]interface{}{
			"bytes_sent":   io.BytesSent,
			"bytes_recv":   io.BytesRecv,
			"packets_sent": io.PacketsSent,
			"packets_recv": io.PacketsRecv,
			"err_in":       io.Errin,
			"err_out":      io.Errout,
			"drop_in":      io.Dropin,
			"drop_out":     io.Dropout,
			"speed":        getInterfaceSpeed(io.Name),
		}
		acc.AddCounter("net", fields, tags)
	}

	return nil
}

// Get the interface speed from /sys/class/net/*/speed file. returns -1 if unsupported
func getInterfaceSpeed(ioName string) int64 {
	sysPath := internal.GetSysPath()

	raw, err := os.ReadFile(filepath.Join(sysPath, "class", "net", ioName, "speed"))
	if err != nil {
		return -1
	}

	speed, err := strconv.ParseInt(strings.TrimSuffix(string(raw), "\n"), 10, 64)
	if err != nil {
		return -1
	}
	return speed
}

func init() {
	inputs.Add("net", func() telegraf.Input {
		return &Net{ps: psutil.NewSystemPS()}
	})
}

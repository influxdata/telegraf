//go:generate ../../../tools/readme_config_includer/generator
package net

import (
	_ "embed"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

//go:embed sample.conf
var sampleConfig string

type NetIOStats struct {
	filter filter.Filter
	ps     system.PS

	skipChecks          bool
	IgnoreProtocolStats bool
	Interfaces          []string
}

func (*NetIOStats) SampleConfig() string {
	return sampleConfig
}

func (n *NetIOStats) Init() error {
	if !n.IgnoreProtocolStats {
		models.PrintOptionValueDeprecationNotice(telegraf.Warn, "inputs.net", "ignore_protocol_stats", "false",
			telegraf.DeprecationInfo{
				Since:     "1.27.3",
				RemovalIn: "1.36.0",
				Notice:    "use the 'inputs.nstat' plugin instead",
			},
		)
	}

	return nil
}

func (n *NetIOStats) Gather(acc telegraf.Accumulator) error {
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
	interfacesByName := map[string]net.Interface{}
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

	// Get system wide stats for different network protocols
	// (ignore these stats if the call fails)
	if !n.IgnoreProtocolStats {
		netprotos, _ := n.ps.NetProto()
		fields := make(map[string]interface{})
		for _, proto := range netprotos {
			for stat, value := range proto.Stats {
				name := fmt.Sprintf("%s_%s", strings.ToLower(proto.Protocol),
					strings.ToLower(stat))
				fields[name] = value
			}
		}
		tags := map[string]string{
			"interface": "all",
		}
		acc.AddFields("net", fields, tags)
	}

	return nil
}

func getInterfaceSpeed(ioName string) uint64 {
	sysPath := os.Getenv("HOST_SYS")
	if sysPath == "" {
		sysPath = "/sys"
	}

	raw, err := os.ReadFile(fmt.Sprintf("%s/class/net/%s/speed", sysPath, ioName))
	if err != nil {
		return 0
	}
	speedStr := string(raw)

	if speedStr == "-1" {
		return 0
	}

	speed, err := strconv.ParseUint(speedStr, 10, 64)
	if err != nil {
		return 0
	}
	return speed
}

func init() {
	inputs.Add("net", func() telegraf.Input {
		return &NetIOStats{ps: system.NewSystemPS()}
	})
}

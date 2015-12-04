package system

import (
	"fmt"
	"net"
	"strings"

	"github.com/influxdb/telegraf/plugins"
)

type NetIOStats struct {
	ps PS

	skipChecks bool
	Interfaces []string
}

func (_ *NetIOStats) Description() string {
	return "Read metrics about network interface usage"
}

var netSampleConfig = `
  # By default, telegraf gathers stats from any up interface (excluding loopback)
  # Setting interfaces will tell it to gather these explicit interfaces,
  # regardless of status.
  #
  # interfaces = ["eth0", ... ]
`

func (_ *NetIOStats) SampleConfig() string {
	return netSampleConfig
}

func (s *NetIOStats) Gather(acc plugins.Accumulator) error {
	netio, err := s.ps.NetIO()
	if err != nil {
		return fmt.Errorf("error getting net io info: %s", err)
	}

	for _, io := range netio {
		if len(s.Interfaces) != 0 {
			var found bool

			for _, name := range s.Interfaces {
				if name == io.Name {
					found = true
					break
				}
			}

			if !found {
				continue
			}
		} else if !s.skipChecks {
			iface, err := net.InterfaceByName(io.Name)
			if err != nil {
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

		acc.Add("bytes_sent", io.BytesSent, tags)
		acc.Add("bytes_recv", io.BytesRecv, tags)
		acc.Add("packets_sent", io.PacketsSent, tags)
		acc.Add("packets_recv", io.PacketsRecv, tags)
		acc.Add("err_in", io.Errin, tags)
		acc.Add("err_out", io.Errout, tags)
		acc.Add("drop_in", io.Dropin, tags)
		acc.Add("drop_out", io.Dropout, tags)
	}

	// Get system wide stats for different network protocols
	// (ignore these stats if the call fails)
	netprotos, _ := s.ps.NetProto()
	for _, proto := range netprotos {
		for stat, value := range proto.Stats {
			name := fmt.Sprintf("%s_%s", strings.ToLower(proto.Protocol),
				strings.ToLower(stat))
			acc.Add(name, value, nil)
		}
	}

	return nil
}

func init() {
	plugins.Add("net", func() plugins.Plugin {
		return &NetIOStats{ps: &systemPS{}}
	})
}

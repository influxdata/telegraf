package net

import (
	"fmt"
	"net"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

type NetIOStats struct {
	filter filter.Filter
	ps     system.PS

	skipChecks          bool
	IgnoreProtocolStats bool
	Interfaces          []string
}

func (_ *NetIOStats) Description() string {
	return "Read metrics about network interface usage"
}

var netSampleConfig = `
  ## By default, telegraf gathers stats from any up interface (excluding loopback)
  ## Setting interfaces will tell it to gather these explicit interfaces,
  ## regardless of status.
  ##
  # interfaces = ["eth0"]
  ##
  ## On linux systems telegraf also collects protocol stats.
  ## Setting ignore_protocol_stats to true will skip reporting of protocol metrics.
  ##
  # ignore_protocol_stats = false
  ##
`

func (_ *NetIOStats) SampleConfig() string {
	return netSampleConfig
}

func GetHostSysFS() string {
	procPath := "/proc"
	if os.Getenv("HOST_PROC") != "" {
		procPath = os.Getenv("HOST_PROC")
	}
	return path.Join(procPath, "/sys")
}

func (s *NetIOStats) Gather(acc telegraf.Accumulator) error {
	netio, err := s.ps.NetIO()
	if err != nil {
		return fmt.Errorf("error getting net io info: %s", err)
	}

	if s.filter == nil {
		if s.filter, err = filter.Compile(s.Interfaces); err != nil {
			return fmt.Errorf("error compiling filter: %s", err)
		}
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("error getting list of interfaces: %s", err)
	}
	interfacesByName := map[string]net.Interface{}
	for _, iface := range interfaces {
		interfacesByName[iface.Name] = iface
	}

	for _, io := range netio {
		var iface net.Interface
		if len(s.Interfaces) != 0 {
			var found bool

			if s.filter.Match(io.Name) {
				found = true
			}

			if !found {
				continue
			}
		} else if !s.skipChecks {
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

		var state string
		if iface.Flags&net.FlagUp > 0 {
			state = "up"
		} else {
			state = "down"
		}

		tags := map[string]string{
			"interface": io.Name,
			"state":     state,
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
			"mtu":          uint64(iface.MTU),
		}

		// linux specific metrics from /sys
		if runtime.GOOS == "linux" {
			var net_sysfs string
			var iface_path string
			net_sysfs = path.Join(GetHostSysFS(), "/class/net/")
			iface_path = net_sysfs + io.Name

			carrier, err := internal.ReadLines(iface_path + "/carrier")
			if err == nil {
				carrier := strings.TrimSpace(carrier[0])
				if carrier == "1" {
					tags["carrier"] = "up"
				} else {
					tags["carrier"] = "down"
				}
			}

			speed, err := internal.ReadLines(iface_path + "/speed")
			if err == nil {
				speed_uint64, err := strconv.ParseUint(strings.TrimSpace(speed[0]), 10, 64)
				if err == nil {
					fields["speed"] = speed_uint64
				}
			}

			duplex, err := internal.ReadLines(iface_path + "/duplex")
			if err == nil {
				tags["duplex"] = strings.TrimSpace(duplex[0])
			}
		}

		acc.AddCounter("net", fields, tags)

	}

	// Get system wide stats for different network protocols
	// (ignore these stats if the call fails)
	if !s.IgnoreProtocolStats {
		netprotos, _ := s.ps.NetProto()
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

func init() {
	inputs.Add("net", func() telegraf.Input {
		return &NetIOStats{ps: system.NewSystemPS()}
	})
}

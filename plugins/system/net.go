package system

import (
	"fmt"
	"math"
	"net"
	"time"

	"github.com/influxdb/telegraf/plugins"
)

type NetIOStats struct {
	ps PS

	skipChecks bool
	Interfaces []string

	prevTime        time.Time
	prevBytesSent   uint64
	prevBytesRecv   uint64
	prevPacketsSent uint64
	prevPacketsRecv uint64
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

func calcSpeed(current uint64, prev uint64, duration time.Duration) uint64 {
	return uint64(math.Floor(float64(current-prev)/duration.Seconds() + 0.5))
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

		now := time.Now()
		if !s.prevTime.IsZero() {
			delta := now.Sub(s.prevTime)

			if delta.Seconds() > 0 {
				acc.Add("bits_per_second_sent", calcSpeed(io.BytesSent, s.prevBytesSent, delta)*8, tags)
				acc.Add("bits_per_second_recv", calcSpeed(io.BytesRecv, s.prevBytesRecv, delta)*8, tags)
				acc.Add("packets_per_second_sent", calcSpeed(io.PacketsSent, s.prevPacketsSent, delta), tags)
				acc.Add("packets_per_second_recv", calcSpeed(io.PacketsRecv, s.prevPacketsRecv, delta), tags)
			}
		}

		s.prevTime = time.Now()
		s.prevBytesSent = io.BytesSent
		s.prevBytesRecv = io.BytesRecv
		s.prevPacketsSent = io.PacketsSent
		s.prevPacketsRecv = io.PacketsRecv
	}

	return nil
}

func init() {
	plugins.Add("net", func() plugins.Plugin {
		return &NetIOStats{ps: &systemPS{}}
	})
}

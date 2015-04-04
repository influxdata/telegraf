package net

import (
	"encoding/json"
	"net"
)

type NetIOCountersStat struct {
	Name        string `json:"name"`         // interface name
	BytesSent   uint64 `json:"bytes_sent"`   // number of bytes sent
	BytesRecv   uint64 `json:"bytes_recv"`   // number of bytes received
	PacketsSent uint64 `json:"packets_sent"` // number of packets sent
	PacketsRecv uint64 `json:"packets_recv"` // number of packets received
	Errin       uint64 `json:"errin"`        // total number of errors while receiving
	Errout      uint64 `json:"errout"`       // total number of errors while sending
	Dropin      uint64 `json:"dropin"`       // total number of incoming packets which were dropped
	Dropout     uint64 `json:"dropout"`      // total number of outgoing packets which were dropped (always 0 on OSX and BSD)
}

// Addr is implemented compatibility to psutil
type Addr struct {
	IP   string `json:"ip"`
	Port uint32 `json:"port"`
}

type NetConnectionStat struct {
	Fd     uint32 `json:"fd"`
	Family uint32 `json:"family"`
	Type   uint32 `json:"type"`
	Laddr  Addr   `json:"localaddr"`
	Raddr  Addr   `json:"remoteaddr"`
	Status string `json:"status"`
	Pid    int32  `json:"pid"`
}

// NetInterfaceAddr is designed for represent interface addresses
type NetInterfaceAddr struct {
	Addr string `json:"addr"`
}

type NetInterfaceStat struct {
	MTU          int                `json:"mtu"`          // maximum transmission unit
	Name         string             `json:"name"`         // e.g., "en0", "lo0", "eth0.100"
	HardwareAddr string             `json:"hardwareaddr"` // IEEE MAC-48, EUI-48 and EUI-64 form
	Flags        []string           `json:"flags"`        // e.g., FlagUp, FlagLoopback, FlagMulticast
	Addrs        []NetInterfaceAddr `json:"addrs"`
}

func (n NetIOCountersStat) String() string {
	s, _ := json.Marshal(n)
	return string(s)
}

func (n NetConnectionStat) String() string {
	s, _ := json.Marshal(n)
	return string(s)
}

func (a Addr) String() string {
	s, _ := json.Marshal(a)
	return string(s)
}

func (n NetInterfaceStat) String() string {
	s, _ := json.Marshal(n)
	return string(s)
}

func (n NetInterfaceAddr) String() string {
	s, _ := json.Marshal(n)
	return string(s)
}

func NetInterfaces() ([]NetInterfaceStat, error) {
	is, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	ret := make([]NetInterfaceStat, 0, len(is))
	for _, ifi := range is {

		var flags []string
		if ifi.Flags&net.FlagUp != 0 {
			flags = append(flags, "up")
		}
		if ifi.Flags&net.FlagBroadcast != 0 {
			flags = append(flags, "broadcast")
		}
		if ifi.Flags&net.FlagLoopback != 0 {
			flags = append(flags, "loopback")
		}
		if ifi.Flags&net.FlagPointToPoint != 0 {
			flags = append(flags, "pointtopoint")
		}
		if ifi.Flags&net.FlagMulticast != 0 {
			flags = append(flags, "multicast")
		}

		r := NetInterfaceStat{
			Name:         ifi.Name,
			MTU:          ifi.MTU,
			HardwareAddr: ifi.HardwareAddr.String(),
			Flags:        flags,
		}
		addrs, err := ifi.Addrs()
		if err == nil {
			r.Addrs = make([]NetInterfaceAddr, 0, len(addrs))
			for _, addr := range addrs {
				r.Addrs = append(r.Addrs, NetInterfaceAddr{
					Addr: addr.String(),
				})
			}

		}
		ret = append(ret, r)
	}

	return ret, nil
}

func getNetIOCountersAll(n []NetIOCountersStat) ([]NetIOCountersStat, error) {
	r := NetIOCountersStat{
		Name: "all",
	}
	for _, nic := range n {
		r.BytesRecv += nic.BytesRecv
		r.PacketsRecv += nic.PacketsRecv
		r.Errin += nic.Errin
		r.Dropin += nic.Dropin
		r.BytesSent += nic.BytesSent
		r.PacketsSent += nic.PacketsSent
		r.Errout += nic.Errout
		r.Dropout += nic.Dropout
	}

	return []NetIOCountersStat{r}, nil
}

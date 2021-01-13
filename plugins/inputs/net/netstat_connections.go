package net

import (
	"fmt"
	//"syscall"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
	"github.com/shirou/gopsutil/net"
)

type NetStatsConnections struct {
	ps                system.PS
	RemoteConnections bool `toml:"remote_connections"`
}

type PortData struct {
	Pid         int32
	Local_addr  string
	Local_port  uint32
	Remote_addr string
	Remote_port uint32
	Status      string
	Established int
	SynSent     int
	SynRecv     int
	FinWait1    int
	FinWait2    int
	TimeWait    int
	Close       int
	CloseWait   int
	LastAck     int
	Listen      int
	Closing     int
	None        int
	UDP         int
	Count       int
}

func (_ *NetStatsConnections) Description() string {
	return "Read TCP metrics such as established, time wait and sockets counts."
}

var tcpstatProcessSampleConfig = ""

func (_ *NetStatsConnections) SampleConfig() string {
	return tcpstatProcessSampleConfig
}

func (s *NetStatsConnections) Gather(acc telegraf.Accumulator) error {
	netconns, err := s.ps.NetConnections()
	if err != nil {
		return fmt.Errorf("error getting net connections info: %s", err)
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("error getting list of interfaces: %s", err)
	}
	interfacesByAddr := make(map[string]string)
	for _, iface := range interfaces {
		addrs := iface.Addrs
		for _, addr := range addrs {
			interfacesByAddr[addr.Addr] = iface.Name
		}
	}

	listen_ports := make(map[string]*PortData)
	connected_ports := make(map[string]*PortData)

	for _, netcon := range netconns {
		var port = ""
		switch netcon.Status {
		case "LISTEN":
			port = strconv.Itoa(int(netcon.Laddr.Port))
			c, ok := listen_ports[port]
			if !ok {
				c = &PortData{
					Pid:         netcon.Pid,
					Local_addr:  netcon.Laddr.IP,
					Local_port:  netcon.Laddr.Port,
					Remote_addr: netcon.Raddr.IP,
					Remote_port: netcon.Raddr.Port,
					Status:      netcon.Status,
					Listen:      1,
				}
				listen_ports[port] = c
			} else {
				c.Listen += 1
			}
		}
	}
	for _, netcon := range netconns {
		var port = ""
		if netcon.Status != "LISTEN" {
			//Add count status by listen port
			port = strconv.Itoa(int(netcon.Laddr.Port))
			c, ok := listen_ports[port]
			if ok {
				switch netcon.Status {
				case "ESTABLISHED":
					c.Established += 1
				case "SYN_SENT":
					c.SynSent += 1
				case "SYN_RECV":
					c.SynRecv += 1
				case "FIN_WAIT1":
					c.FinWait1 += 1
				case "FIN_WAIT2":
					c.FinWait2 += 1
				case "TIME_WAIT":
					c.TimeWait += 1
				case "CLOSE":
					c.Close += 1
				case "CLOSE_WAIT":
					c.CloseWait += 1
				case "LAST_ACK":
					c.LastAck += 1
				case "CLOSING":
					c.Closing += 1
				case "NONE":
					c.None += 1
				}
			} else if s.RemoteConnections && netcon.Raddr.Port > 0 { //Only generate remote connectios by parameter
				port = netcon.Raddr.IP + "_" + strconv.Itoa(int(netcon.Raddr.Port))
				c, ok := connected_ports[port]
				if !ok {
					c = &PortData{
						Pid:         netcon.Pid,
						Local_addr:  netcon.Laddr.IP,
						Local_port:  netcon.Laddr.Port,
						Remote_addr: netcon.Raddr.IP,
						Remote_port: netcon.Raddr.Port,
						Status:      netcon.Status,
					}
					connected_ports[port] = c
				}
				switch netcon.Status {
				case "ESTABLISHED":
					c.Established += 1
				case "SYN_SENT":
					c.SynSent += 1
				case "SYN_RECV":
					c.SynRecv += 1
				case "FIN_WAIT1":
					c.FinWait1 += 1
				case "FIN_WAIT2":
					c.FinWait2 += 1
				case "TIME_WAIT":
					c.TimeWait += 1
				case "CLOSE":
					c.Close += 1
				case "CLOSE_WAIT":
					c.CloseWait += 1
				case "LAST_ACK":
					c.LastAck += 1
				case "CLOSING":
					c.Closing += 1
				case "NONE":
					c.None += 1
				}
			}
		}
	}

	for _, value := range listen_ports {
		acc.AddFields("netstat_incoming",
			map[string]interface{}{
				"tcp_established": value.Established,
				"tcp_syn_send":    value.SynSent,
				"tcp_syn_recv":    value.SynRecv,
				"tcp_fin_wait1":   value.FinWait1,
				"tcp_fin_wait2":   value.FinWait2,
				"tcp_time_wait":   value.TimeWait,
				"tcp_close":       value.Close,
				"tcp_close_wait":  value.CloseWait,
				"tcp_last_ack":    value.LastAck,
				"tcp_listen":      value.Listen,
				"tcp_closing":     value.Closing,
				"tcp_none":        value.None,
			},
			map[string]string{
				//"pid": strconv.Itoa(int(value.Pid)),
				//"local_addr": value.Local_addr,
				"port": strconv.Itoa(int(value.Local_port)),
				//"remote_addr": value.Remote_addr,
				//"remote_port": strconv.Itoa(int(value.Remote_port)),
				//"status": value.Status,
			})
	}
	for _, value := range connected_ports {
		acc.AddFields("netstat_outgoing",
			map[string]interface{}{
				"tcp_established": value.Established,
				"tcp_syn_send":    value.SynSent,
				"tcp_syn_recv":    value.SynRecv,
				"tcp_fin_wait1":   value.FinWait1,
				"tcp_fin_wait2":   value.FinWait2,
				"tcp_time_wait":   value.TimeWait,
				"tcp_close":       value.Close,
				"tcp_close_wait":  value.CloseWait,
				"tcp_last_ack":    value.LastAck,
				"tcp_listen":      value.Listen,
				"tcp_closing":     value.Closing,
				"tcp_none":        value.None,
			},
			map[string]string{
				//"pid": strconv.Itoa(int(value.Pid)),
				//"local_addr": value.Local_addr,
				//"local_port": strconv.Itoa(int(value.Local_port)),
				"addr": value.Remote_addr,
				"port": strconv.Itoa(int(value.Remote_port)),
				//"status": value.Status,
			})
	}

	return nil
}

func init() {
	inputs.Add("netstat_connections", func() telegraf.Input {
		return &NetStatsConnections{ps: system.NewSystemPS(), RemoteConnections: false}
	})
}

// +build linux

package procstat

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/elastic/gosigar/sys/linux"
	"github.com/influxdata/telegraf"
)

// addConnectionStats count the number of connections in each TCP state and add those values to the metric
func addConnectionStats(pidConnections []ConnInfo, fields map[string]interface{}, prefix string) {
	counts := make(map[linux.TCPState]int)
	for _, netcon := range pidConnections {
		counts[netcon.state]++
	}

	fields[prefix+"tcp_established"] = counts[linux.TCP_ESTABLISHED]
	fields[prefix+"tcp_syn_sent"] = counts[linux.TCP_SYN_SENT]
	fields[prefix+"tcp_syn_recv"] = counts[linux.TCP_SYN_RECV]
	fields[prefix+"tcp_fin_wait1"] = counts[linux.TCP_FIN_WAIT1]
	fields[prefix+"tcp_fin_wait2"] = counts[linux.TCP_FIN_WAIT2]
	//fields[prefix+"tcp_time_wait"] = counts[linux.TCP_TIME_WAIT]  // TIME-WAIT connections does not have a pid associated
	fields[prefix+"tcp_close"] = counts[linux.TCP_CLOSE]
	fields[prefix+"tcp_close_wait"] = counts[linux.TCP_CLOSE_WAIT]
	fields[prefix+"tcp_last_ack"] = counts[linux.TCP_LAST_ACK]
	fields[prefix+"tcp_listen"] = counts[linux.TCP_LISTEN]
	fields[prefix+"tcp_closing"] = counts[linux.TCP_CLOSING]
}

// addConnectionEnpoints add listen and connection endpoints to the procstat_tcp metric.
// If listen is 0.0.0.0 or ::, it will be added one value for each of the IP addresses of the host.
// Listeners in private IPs are ignored (maybe a flag could be added, but now the reasoning is matching connections between hosts).
// Connections made to this server are ignored (the local port is one of the listening ports).
func addConnectionEnpoints(acc telegraf.Accumulator, proc Process, netInfo NetworkInfo) error {
	TCPListen := map[string]interface{}{}
	TCPConn := map[string]interface{}{}

	pidConnections, err := netInfo.GetConnectionsByPid(uint32(proc.PID()))
	if err != nil {
		if errors.Is(err, ErrorPIDNotFound) {
			return nil
		}

		return fmt.Errorf("W! [inputs.procstat] not able to get connections for pid=%v: %v", proc.PID(), err)
	}

	// In case of error, ppid=0 and will be ignored in IsPidListeningInPort
	ppid, _ := proc.Ppid()

	for _, c := range pidConnections {
		// Ignore listeners or connections in/to localhost or private IPs
		if c.srcIP.IsLoopback() || containsIP(netInfo.GetPrivateIPs(), c.srcIP) {
			continue
		}

		if c.state == linux.TCP_LISTEN {
			if netInfo.IsPidListeningInAddr(uint32(ppid), c.srcIP, c.srcPort) {
				continue
			}

			if c.srcIP.IsUnspecified() {
				// 0.0.0.0 listen in all IPv4 addresses
				// :: listen in all IPv4 + IPv6 addresses
				for _, ip := range netInfo.GetPublicIPs() {
					if isIPV4(ip) || isIPV6(c.srcIP) {
						TCPListen[endpointString(ip, c.srcPort)] = nil
					}
				}
			} else {
				TCPListen[endpointString(c.srcIP, c.srcPort)] = nil
			}
		} else if c.state != linux.TCP_SYN_SENT { // All TCP states except LISTEN (already processed) and SYN_SENT imply a connection between the hosts
			// Ignore connections from outside hosts to listeners in this host (status != LISTEN and localPort in listenPorts)
			if !netInfo.IsAListenPort(c.srcPort) {
				TCPConn[endpointString(c.dstIP, c.dstPort)] = nil
			}
		}
	}

	// Only add metrics if we have data
	if len(TCPConn) > 0 || len(TCPListen) > 0 {
		tcpConnections := []string{}
		tcpListeners := []string{}

		for k := range TCPConn {
			tcpConnections = append(tcpConnections, k)
		}
		sort.Strings(tcpConnections) // sort to make testing simplier

		for k := range TCPListen {
			tcpListeners = append(tcpListeners, k)
		}
		sort.Strings(tcpListeners)

		fields := map[string]interface{}{
			TCPConnectionKey: strings.Join(tcpConnections, ","),
			TCPListenKey:     strings.Join(tcpListeners, ","),
		}

		acc.AddFields(MetricNameTCPConnections, fields, proc.Tags())
	}

	return nil
}

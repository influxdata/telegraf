// +build linux

package net

import (
	"errors"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/internal/common"
)

// NetIOCounters returnes network I/O statistics for every network
// interface installed on the system.  If pernic argument is false,
// return only sum of all information (which name is 'all'). If true,
// every network interface installed on the system is returned
// separately.
func NetIOCounters(pernic bool) ([]NetIOCountersStat, error) {
	filename := common.HostProc("net/dev")
	lines, err := common.ReadLines(filename)
	if err != nil {
		return nil, err
	}

	statlen := len(lines) - 1

	ret := make([]NetIOCountersStat, 0, statlen)

	for _, line := range lines[2:] {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		interfaceName := strings.TrimSpace(parts[0])
		if interfaceName == "" {
			continue
		}

		fields := strings.Fields(strings.TrimSpace(parts[1]))
		bytesRecv, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			return ret, err
		}
		packetsRecv, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return ret, err
		}
		errIn, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			return ret, err
		}
		dropIn, err := strconv.ParseUint(fields[3], 10, 64)
		if err != nil {
			return ret, err
		}
		bytesSent, err := strconv.ParseUint(fields[8], 10, 64)
		if err != nil {
			return ret, err
		}
		packetsSent, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			return ret, err
		}
		errOut, err := strconv.ParseUint(fields[10], 10, 64)
		if err != nil {
			return ret, err
		}
		dropOut, err := strconv.ParseUint(fields[13], 10, 64)
		if err != nil {
			return ret, err
		}

		nic := NetIOCountersStat{
			Name:        interfaceName,
			BytesRecv:   bytesRecv,
			PacketsRecv: packetsRecv,
			Errin:       errIn,
			Dropin:      dropIn,
			BytesSent:   bytesSent,
			PacketsSent: packetsSent,
			Errout:      errOut,
			Dropout:     dropOut,
		}
		ret = append(ret, nic)
	}

	if pernic == false {
		return getNetIOCountersAll(ret)
	}

	return ret, nil
}

var netProtocols = []string{
	"ip",
	"icmp",
	"icmpmsg",
	"tcp",
	"udp",
	"udplite",
}

// NetProtoCounters returns network statistics for the entire system
// If protocols is empty then all protocols are returned, otherwise
// just the protocols in the list are returned.
// Available protocols:
//   ip,icmp,icmpmsg,tcp,udp,udplite
func NetProtoCounters(protocols []string) ([]NetProtoCountersStat, error) {
	if len(protocols) == 0 {
		protocols = netProtocols
	}

	stats := make([]NetProtoCountersStat, 0, len(protocols))
	protos := make(map[string]bool, len(protocols))
	for _, p := range protocols {
		protos[p] = true
	}

	filename := "/proc/net/snmp"
	lines, err := common.ReadLines(filename)
	if err != nil {
		return nil, err
	}

	linecount := len(lines)
	for i := 0; i < linecount; i++ {
		line := lines[i]
		r := strings.IndexRune(line, ':')
		if r == -1 {
			return nil, errors.New(filename + " is not fomatted correctly, expected ':'.")
		}
		proto := strings.ToLower(line[:r])
		if !protos[proto] {
			// skip protocol and data line
			i++
			continue
		}

		// Read header line
		statNames := strings.Split(line[r+2:], " ")

		// Read data line
		i++
		statValues := strings.Split(lines[i][r+2:], " ")
		if len(statNames) != len(statValues) {
			return nil, errors.New(filename + " is not fomatted correctly, expected same number of columns.")
		}
		stat := NetProtoCountersStat{
			Protocol: proto,
			Stats:    make(map[string]int64, len(statNames)),
		}
		for j := range statNames {
			value, err := strconv.ParseInt(statValues[j], 10, 64)
			if err != nil {
				return nil, err
			}
			stat.Stats[statNames[j]] = value
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

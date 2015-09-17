// +build windows

package net

import (
	"net"
	"os"
	"syscall"
	"unsafe"

	common "github.com/shirou/gopsutil/common"
)

var (
	modiphlpapi             = syscall.NewLazyDLL("iphlpapi.dll")
	procGetExtendedTcpTable = modiphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable = modiphlpapi.NewProc("GetExtendedUdpTable")
)

const (
	TCPTableBasicListener = iota
	TCPTableBasicConnections
	TCPTableBasicAll
	TCPTableOwnerPIDListener
	TCPTableOwnerPIDConnections
	TCPTableOwnerPIDAll
	TCPTableOwnerModuleListener
	TCPTableOwnerModuleConnections
	TCPTableOwnerModuleAll
)

func NetIOCounters(pernic bool) ([]NetIOCountersStat, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	ai, err := getAdapterList()
	if err != nil {
		return nil, err
	}
	var ret []NetIOCountersStat

	for _, ifi := range ifs {
		name := ifi.Name
		for ; ai != nil; ai = ai.Next {
			name = common.BytePtrToString(&ai.Description[0])
			c := NetIOCountersStat{
				Name: name,
			}

			row := syscall.MibIfRow{Index: ai.Index}
			e := syscall.GetIfEntry(&row)
			if e != nil {
				return nil, os.NewSyscallError("GetIfEntry", e)
			}
			c.BytesSent = uint64(row.OutOctets)
			c.BytesRecv = uint64(row.InOctets)
			c.PacketsSent = uint64(row.OutUcastPkts)
			c.PacketsRecv = uint64(row.InUcastPkts)
			c.Errin = uint64(row.InErrors)
			c.Errout = uint64(row.OutErrors)
			c.Dropin = uint64(row.InDiscards)
			c.Dropout = uint64(row.OutDiscards)

			ret = append(ret, c)
		}
	}

	if pernic == false {
		return getNetIOCountersAll(ret)
	}
	return ret, nil
}

// Return a list of network connections opened by a process
func NetConnections(kind string) ([]NetConnectionStat, error) {
	var ret []NetConnectionStat

	return ret, common.NotImplementedError
}

// borrowed from src/pkg/net/interface_windows.go
func getAdapterList() (*syscall.IpAdapterInfo, error) {
	b := make([]byte, 1000)
	l := uint32(len(b))
	a := (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
	err := syscall.GetAdaptersInfo(a, &l)
	if err == syscall.ERROR_BUFFER_OVERFLOW {
		b = make([]byte, l)
		a = (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
		err = syscall.GetAdaptersInfo(a, &l)
	}
	if err != nil {
		return nil, os.NewSyscallError("GetAdaptersInfo", err)
	}
	return a, nil
}

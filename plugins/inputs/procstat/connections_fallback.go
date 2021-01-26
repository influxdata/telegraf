// +build !linux

package procstat

import (
	"fmt"
	"net"
)

type ConnInfo struct {
}

func (n *NetworkInfo) IsAListenPort(port uint32) bool {
	return false
}

func (n *NetworkInfo) Fetch() error {
	return nil
}

func (n *NetworkInfo) GetConnectionsByPid(pid uint32) (conn []ConnInfo, err error) {
	return conn, fmt.Errorf("platform not supported")
}

func (n *NetworkInfo) GetPublicIPs() []net.IP {
	return []net.IP{}
}

func (n *NetworkInfo) GetPrivateIPs() []net.IP {
	return []net.IP{}
}

func (n *NetworkInfo) IsPidListeningInAddr(pid uint32, ip net.IP, port uint32) bool {
	return false
}

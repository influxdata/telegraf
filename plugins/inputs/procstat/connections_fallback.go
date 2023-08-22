//go:build !linux
// +build !linux

package procstat

import (
	"fmt"
	"net"
)

type connInfo struct{}

func (n *networkInfo) IsAListenPort(_ uint32) bool {
	return false
}

func (n *networkInfo) Fetch() error {
	return nil
}

func (n *networkInfo) GetConnectionsByPid(_ uint32) (conn []connInfo, err error) {
	// Avoid "unused" errors
	_ = dockerMACPrefix
	_ = virtualBoxMACPrefix
	_ = hardwareAddrLength
	i := inodeInfo{}
	_ = i.pid
	_ = n.tcp
	_ = n.listenPorts
	_ = n.publicIPs
	_ = n.privateIPs

	return conn, fmt.Errorf("platform not supported")
}

func (n *networkInfo) GetPublicIPs() []net.IP {
	return []net.IP{}
}

func (n *networkInfo) GetPrivateIPs() []net.IP {
	return []net.IP{}
}

func (n *networkInfo) IsPidListeningInAddr(_ uint32, _ net.IP, _ uint32) bool {
	return false
}

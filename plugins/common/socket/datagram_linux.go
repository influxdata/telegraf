package socket

import (
	"net"

	"golang.org/x/sys/unix"
)

// Go binds multicast listeners to the wildcard address, so by default Linux
// delivers every group joined on the port to each of them.
func ownGroupsOnly(conn *net.UDPConn, ipv4 bool) error {
	level, opt := unix.IPPROTO_IPV6, unix.IPV6_MULTICAST_ALL
	if ipv4 {
		level, opt = unix.IPPROTO_IP, unix.IP_MULTICAST_ALL
	}

	rc, err := conn.SyscallConn()
	if err != nil {
		return err
	}
	var serr error
	if err := rc.Control(func(fd uintptr) {
		serr = unix.SetsockoptInt(int(fd), level, opt, 0)
	}); err != nil {
		return err
	}
	return serr
}

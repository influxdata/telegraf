//go:build !linux

package socket

import "net"

func ownGroupsOnly(*net.UDPConn, bool) error { return nil }

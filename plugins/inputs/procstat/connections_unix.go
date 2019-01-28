// +build linux freebsd darwin

package procstat

import "github.com/shirou/gopsutil/net"

func getConnectionsByPid(kind string, pid int32) ([]net.ConnectionStat, error) {
	return net.ConnectionsPid(kind, pid)
}

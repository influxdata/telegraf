// +build !linux,!freebsd,!darwin

package procstat

import (
	"fmt"

	"github.com/shirou/gopsutil/net"
)

func getConnectionsByPid(kind string, pid int32) ([]net.ConnectionStat, error) {
	return []net.ConnectionStat{}, fmt.Errorf("platform not supported")
}

//go:build !linux
// +build !linux

package procstat

import (
	"fmt"
	"net"

	"github.com/influxdata/telegraf"
)

func addConnectionStats(_ []connInfo, _ map[string]interface{}, _ string) {
}

func addConnectionEndpoints(_ telegraf.Accumulator, _ Process, _ networkInfo) error {
	// Avoid "unused" errors
	_ = metricNameTCPConnections
	_ = tcpConnectionKey
	_ = tcpListenKey
	_, _ = extractIPs([]net.Addr{})
	_ = containsIP([]net.IP{}, net.IP{})
	_ = isIPV4(net.IP{})
	_ = isIPV6(net.IP{})
	_ = endpointString(net.IP{}, uint32(0))

	return fmt.Errorf("platform not supported")
}

//go:build !linux && !windows

package procstat

import (
	"errors"

	"github.com/shirou/gopsutil/v3/net"
	gopsnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

const supportsSocketStat = false

func processName(p *process.Process) (string, error) {
	return p.Exe()
}

func queryPidWithWinServiceName(string) (uint32, error) {
	return 0, errors.New("os not supporting win_service option")
}

func collectMemmap(Process, string, map[string]any) {}

func findBySystemdUnits([]string) ([]processGroup, error) {
	return nil, nil
}

func findByWindowsServices([]string) ([]processGroup, error) {
	return nil, nil
}

func collectTotalReadWrite(Process) (r, w uint64, err error) {
	return 0, 0, errors.ErrUnsupported
}

func unixConnectionsPid(int32) ([]gopsnet.ConnectionStat, error) {
	return nil, errors.ErrUnsupported
}

func statsTCP([]net.ConnectionStat, uint8) ([]map[string]interface{}, error) {
	return nil, errors.ErrUnsupported
}

func statsUDP([]net.ConnectionStat, uint8) ([]map[string]interface{}, error) {
	return nil, errors.ErrUnsupported
}
func statsUnix([]net.ConnectionStat) ([]map[string]interface{}, error) {
	return nil, errors.ErrUnsupported
}

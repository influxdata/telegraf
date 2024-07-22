//go:build !linux && !windows

package procstat

import (
	"errors"
	"syscall"

	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

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

func statsTCP(conns []net.ConnectionStat, _ uint8) ([]map[string]interface{}, error) {
	if len(conns) == 0 {
		return nil, nil
	}

	// Filter the responses via the inodes belonging to the process
	fieldslist := make([]map[string]interface{}, 0, len(conns))
	for _, c := range conns {
		var proto string
		switch c.Family {
		case syscall.AF_INET:
			proto = "tcp4"
		case syscall.AF_INET6:
			proto = "tcp6"
		default:
			continue
		}

		fields := map[string]interface{}{
			"protocol":  proto,
			"state":     c.Status,
			"pid":       c.Pid,
			"src":       c.Laddr.IP,
			"src_port":  c.Laddr.Port,
			"dest":      c.Raddr.IP,
			"dest_port": c.Raddr.Port,
		}
		fieldslist = append(fieldslist, fields)
	}

	return fieldslist, nil
}

func statsUDP(conns []net.ConnectionStat, _ uint8) ([]map[string]interface{}, error) {
	if len(conns) == 0 {
		return nil, nil
	}

	// Filter the responses via the inodes belonging to the process
	fieldslist := make([]map[string]interface{}, 0, len(conns))
	for _, c := range conns {
		var proto string
		switch c.Family {
		case syscall.AF_INET:
			proto = "udp4"
		case syscall.AF_INET6:
			proto = "udp6"
		default:
			continue
		}

		fields := map[string]interface{}{
			"protocol":  proto,
			"state":     c.Status,
			"pid":       c.Pid,
			"src":       c.Laddr.IP,
			"src_port":  c.Laddr.Port,
			"dest":      c.Raddr.IP,
			"dest_port": c.Raddr.Port,
		}
		fieldslist = append(fieldslist, fields)
	}

	return fieldslist, nil
}

func statsUnix([]net.ConnectionStat) ([]map[string]interface{}, error) {
	return nil, errors.ErrUnsupported
}

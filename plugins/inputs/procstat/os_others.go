//go:build !linux && !windows

package procstat

import (
	"errors"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	gopsnet "github.com/shirou/gopsutil/v4/net"
	gopsprocess "github.com/shirou/gopsutil/v4/process"
)

func processName(p *gopsprocess.Process) (string, error) {
	return p.Exe()
}

func username(p *gopsprocess.Process) string {
	// Use the local lookup
	n, err := p.Username()
	if err == nil {
		return n
	}

	// Exit on errors other than unknown user-ID
	var uerr user.UnknownUserIdError
	if !errors.As(err, &uerr) {
		return ""
	}

	// Try to run the `id` command on the UID of the process to resolve remote
	// users such as LDAP or NIS.
	uid := strconv.Itoa(int(uerr))
	buf, err := exec.Command("id", "-nu", uid).Output()
	if n := strings.TrimSpace(string(buf)); err == nil && n != "" {
		return n
	}

	// We were either not able to run the command or the user cannot be
	// resolved so just return the user ID instead.
	return uid
}

func queryPidWithWinServiceName(string) (uint32, error) {
	return 0, errors.New("os not supporting win_service option")
}

func collectMemmap(process, string, map[string]any) {}

func findBySystemdUnits([]string) ([]processGroup, error) {
	return nil, nil
}

func findByWindowsServices([]string) ([]processGroup, error) {
	return nil, nil
}

func collectTotalReadWrite(process) (r, w uint64, err error) {
	return 0, 0, errors.ErrUnsupported
}

func statsTCP(conns []gopsnet.ConnectionStat, _ uint8) ([]map[string]interface{}, error) {
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

func statsUDP(conns []gopsnet.ConnectionStat, _ uint8) ([]map[string]interface{}, error) {
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

func statsUnix([]gopsnet.ConnectionStat) ([]map[string]interface{}, error) {
	return nil, errors.ErrUnsupported
}

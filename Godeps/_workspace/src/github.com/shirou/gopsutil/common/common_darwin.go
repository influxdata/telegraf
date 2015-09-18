// +build darwin

package common

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func DoSysctrl(mib string) ([]string, error) {
	out, err := exec.Command("/usr/sbin/sysctl", "-n", mib).Output()
	if err != nil {
		return []string{}, err
	}
	v := strings.Replace(string(out), "{ ", "", 1)
	v = strings.Replace(string(v), " }", "", 1)
	values := strings.Fields(string(v))

	return values, nil
}

func CallSyscall(mib []int32) ([]byte, uint64, error) {
	miblen := uint64(len(mib))

	// get required buffer size
	length := uint64(0)
	_, _, err := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(miblen),
		0,
		uintptr(unsafe.Pointer(&length)),
		0,
		0)
	if err != 0 {
		var b []byte
		return b, length, err
	}
	if length == 0 {
		var b []byte
		return b, length, err
	}
	// get proc info itself
	buf := make([]byte, length)
	_, _, err = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(miblen),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&length)),
		0,
		0)
	if err != 0 {
		return buf, length, err
	}

	return buf, length, nil
}

func CallLsof(invoke Invoker, pid int32, args ...string) ([]string, error) {
	var cmd []string
	if pid == 0 { // will get from all processes.
		cmd = []string{"-a", "-n", "-P"}
	} else {
		cmd = []string{"-a", "-n", "-P", "-p", strconv.Itoa(int(pid))}
	}
	cmd = append(cmd, args...)
	lsof, err := exec.LookPath("lsof")
	if err != nil {
		return []string{}, err
	}
	out, err := invoke.Command(lsof, cmd...)
	if err != nil {
		// if not pid found, lsof returnes code 1
		if err.Error() == "exit status 1" && len(out) == 0 {
			return []string{}, nil
		}
		return []string{}, err
	}
	lines := strings.Split(string(out), "\n")

	var ret []string
	for _, l := range lines[1:] {
		if len(l) == 0 {
			continue
		}
		ret = append(ret, l)
	}
	return ret, nil
}

// +build linux freebsd darwin

package common

import (
	"os/exec"
	"strconv"
	"strings"
)

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
		// if no pid found, lsof returnes code 1.
		if err.Error() == "exit status 1" && len(out) == 0 {
			return []string{}, nil
		}
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

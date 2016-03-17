// command
package ipmi

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type IpmiCommand struct {
	*Connection
}

func (t *IpmiCommand) options() []string {
	intf := t.Interface
	if intf == "" {
		intf = "lanplus"
	}

	options := []string{
		"-H", t.Hostname,
		"-U", t.Username,
		"-P", t.Password,
		"-I", intf,
	}

	if t.Port != 0 {
		options = append(options, "-p", strconv.Itoa(t.Port))
	}

	return options
}

func (t *IpmiCommand) cmd(args ...string) *exec.Cmd {
	path := t.Path
	opts := append(t.options(), args...)

	if path == "" {
		path = "ipmitool"
	}

	return exec.Command(path, opts...)

}

func (t *IpmiCommand) Run(args ...string) (string, error) {
	cmd := t.cmd(args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("run %s %s: %s (%s)",
			cmd.Path, strings.Join(cmd.Args, " "), stderr.String(), err)
	}

	return stdout.String(), err
}

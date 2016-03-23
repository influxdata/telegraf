package ipmi_sensor

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type CommandRunner struct{}

func (t CommandRunner) cmd(conn *Connection, args ...string) *exec.Cmd {
	path := conn.Path
	opts := append(conn.options(), args...)

	if path == "" {
		path = "ipmitool"
	}

	return exec.Command(path, opts...)

}

func (t CommandRunner) Run(conn *Connection, args ...string) (string, error) {
	cmd := t.cmd(conn, args...)
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

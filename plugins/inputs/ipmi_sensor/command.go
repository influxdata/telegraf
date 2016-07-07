package ipmi_sensor

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/influxdata/telegraf/internal"
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

	output, err := internal.CombinedOutputTimeout(cmd, time.Second*5)
	if err != nil {
		return "", fmt.Errorf("run %s %s: %s (%s)",
			cmd.Path, strings.Join(cmd.Args, " "), string(output), err)
	}

	return string(output), err
}

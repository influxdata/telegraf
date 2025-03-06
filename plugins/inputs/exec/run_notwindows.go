//go:build !windows

package exec

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/kballard/go-shellquote"

	"github.com/influxdata/telegraf/internal"
)

func (c *commandRunner) run(command string) (out, errout []byte, err error) {
	splitCmd, err := shellquote.Split(command)
	if err != nil || len(splitCmd) == 0 {
		return nil, nil, fmt.Errorf("exec: unable to parse command %q: %w", command, err)
	}

	cmd := exec.Command(splitCmd[0], splitCmd[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if len(c.environment) > 0 {
		cmd.Env = append(os.Environ(), c.environment...)
	}

	var outbuf, stderr bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &stderr

	runErr := internal.RunTimeout(cmd, c.timeout)

	if stderr.Len() > 0 && !c.debug {
		truncate(&stderr)
	}

	return outbuf.Bytes(), stderr.Bytes(), runErr
}

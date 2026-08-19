//go:build !windows

package exec

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"

	"github.com/influxdata/telegraf/internal"
)

func (c *commandRunner) run(splitCmd []string) (out, errout []byte, err error) {
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

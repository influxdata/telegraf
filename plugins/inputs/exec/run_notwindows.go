//go:build !windows

package exec

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/kballard/go-shellquote"

	"github.com/influxdata/telegraf/internal"
)

func (c commandRunner) run(
	command string,
	environments []string,
	timeout time.Duration,
) (out, errout []byte, err error) {
	splitCmd, err := shellquote.Split(command)
	if err != nil || len(splitCmd) == 0 {
		return nil, nil, fmt.Errorf("exec: unable to parse command: %w", err)
	}

	cmd := exec.Command(splitCmd[0], splitCmd[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if len(environments) > 0 {
		cmd.Env = append(os.Environ(), environments...)
	}

	var (
		outbuf bytes.Buffer
		stderr bytes.Buffer
	)
	cmd.Stdout = &outbuf
	cmd.Stderr = &stderr

	runErr := internal.RunTimeout(cmd, timeout)

	outbuf = removeWindowsCarriageReturns(outbuf)
	if stderr.Len() > 0 && !c.debug {
		stderr = removeWindowsCarriageReturns(stderr)
		stderr = truncate(stderr)
	}

	return outbuf.Bytes(), stderr.Bytes(), runErr
}

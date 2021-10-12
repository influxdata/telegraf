package exec

import (
	"bytes"
	"io"
	osExec "os/exec"
	"runtime"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

const MaxStderrBytes int = 512

type Runner interface {
	Run(command string, args []string, env []string, input []byte, timeout time.Duration) ([]byte, []byte, error)
}

type CommandRunner struct{}

var commandRunner = CommandRunner{}

func NewRunner() Runner {
	return commandRunner
}

func (c CommandRunner) Run(
	command string,
	args []string,
	env []string,
	input []byte,
	timeout time.Duration,
) ([]byte, []byte, error) {
	cmd := osExec.Command(command, args...)

	var (
		out    bytes.Buffer
		stderr bytes.Buffer
	)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Env = env
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}

	runErr := internal.RunTimeout(cmd, timeout)

	out = removeWindowsCarriageReturns(out)
	if stderr.Len() > 0 && !telegraf.Debug {
		stderr = removeWindowsCarriageReturns(stderr)
		stderr = c.truncate(stderr)
	}

	return out.Bytes(), stderr.Bytes(), runErr
}

func (c CommandRunner) truncate(buf bytes.Buffer) bytes.Buffer {
	// Limit the number of bytes.
	didTruncate := false
	if buf.Len() > MaxStderrBytes {
		buf.Truncate(MaxStderrBytes)
		didTruncate = true
	}
	if i := bytes.IndexByte(buf.Bytes(), '\n'); i > 0 {
		// Only show truncation if the newline wasn't the last character.
		if i < buf.Len()-1 {
			didTruncate = true
		}
		buf.Truncate(i)
	}
	if didTruncate {
		//nolint:errcheck,revive // Will always return nil or panic
		buf.WriteString("...")
	}
	return buf
}

// removeWindowsCarriageReturns removes all carriage returns from the input if the
// OS is Windows. It does not return any errors.
func removeWindowsCarriageReturns(b bytes.Buffer) bytes.Buffer {
	if runtime.GOOS == "windows" {
		var buf bytes.Buffer
		for {
			byt, err := b.ReadBytes(0x0D)
			byt = bytes.TrimRight(byt, "\x0d")
			if len(byt) > 0 {
				_, _ = buf.Write(byt)
			}
			if err == io.EOF {
				return buf
			}
		}
	}
	return b
}

package exec

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

const maxStderrBytes = 512

// Exec defines the exec output plugin.
type Exec struct {
	Command []string        `toml:"command"`
	Timeout config.Duration `toml:"timeout"`
	Log     telegraf.Logger `toml:"-"`

	runner     Runner
	serializer serializers.Serializer
}

func (e *Exec) Init() error {
	e.runner = &CommandRunner{log: e.Log}

	return nil
}

// SetSerializer sets the serializer for the output.
func (e *Exec) SetSerializer(serializer serializers.Serializer) {
	e.serializer = serializer
}

// Connect satisfies the Output interface.
func (e *Exec) Connect() error {
	return nil
}

// Close satisfies the Output interface.
func (e *Exec) Close() error {
	return nil
}

// Write writes the metrics to the configured command.
func (e *Exec) Write(metrics []telegraf.Metric) error {
	var buffer bytes.Buffer
	serializedMetrics, err := e.serializer.SerializeBatch(metrics)
	if err != nil {
		return err
	}
	buffer.Write(serializedMetrics) //nolint:revive // from buffer.go: "err is always nil"

	if buffer.Len() <= 0 {
		return nil
	}

	return e.runner.Run(time.Duration(e.Timeout), e.Command, &buffer)
}

// Runner provides an interface for running exec.Cmd.
type Runner interface {
	Run(time.Duration, []string, io.Reader) error
}

// CommandRunner runs a command with the ability to kill the process before the timeout.
type CommandRunner struct {
	cmd *exec.Cmd
	log telegraf.Logger
}

// Run runs the command.
func (c *CommandRunner) Run(timeout time.Duration, command []string, buffer io.Reader) error {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = buffer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := internal.RunTimeout(cmd, timeout)
	s := stderr

	if err != nil {
		if err == internal.ErrTimeout {
			return fmt.Errorf("%q timed out and was killed", command)
		}

		s = removeWindowsCarriageReturns(s)
		if s.Len() > 0 {
			if !telegraf.Debug {
				c.log.Errorf("Command error: %q", c.truncate(s))
			} else {
				c.log.Debugf("Command error: %q", s)
			}
		}

		if status, ok := internal.ExitStatus(err); ok {
			return fmt.Errorf("%q exited %d with %s", command, status, err.Error())
		}

		return fmt.Errorf("%q failed with %s", command, err.Error())
	}

	c.cmd = cmd

	return nil
}

func (c *CommandRunner) truncate(buf bytes.Buffer) string {
	// Limit the number of bytes.
	didTruncate := false
	if buf.Len() > maxStderrBytes {
		buf.Truncate(maxStderrBytes)
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
		buf.WriteString("...") //nolint:revive // from buffer.go: "err is always nil"
	}
	return buf.String()
}

func init() {
	outputs.Add("exec", func() telegraf.Output {
		return &Exec{
			Timeout: config.Duration(time.Second * 5),
		}
	})
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

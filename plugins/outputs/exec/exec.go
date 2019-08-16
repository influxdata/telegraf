package exec

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

// Exec defines the exec output plugin.
type Exec struct {
	Command []string          `toml:"command"`
	Timeout internal.Duration `toml:"timeout"`

	runner     RunCloser
	serializer serializers.Serializer
}

var sampleConfig = `
  ## Command to injest metrics via stdin.
  command = ["tee", "-a", "/dev/null"]

  ## Timeout for command to complete.
  # timeout = "5s"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"
`

// SetSerializer sets the serializer for the output.
func (e *Exec) SetSerializer(serializer serializers.Serializer) {
	e.serializer = serializer
}

// Connect satisfies the Ouput interface.
func (e *Exec) Connect() error {
	return nil
}

// Close kills the running process if any.
func (e *Exec) Close() error {
	if e.runner == nil {
		return nil
	}
	return e.runner.Close()
}

// Description describes the plugin.
func (e *Exec) Description() string {
	return "Send metrics to command as input over stdin"
}

// SampleConfig returns a sample configuration.
func (e *Exec) SampleConfig() string {
	return sampleConfig
}

// Write writes the metrics to the configured command.
func (e *Exec) Write(metrics []telegraf.Metric) error {
	var buffer bytes.Buffer
	for _, metric := range metrics {
		value, err := e.serializer.Serialize(metric)
		if err != nil {
			return err
		}
		buffer.Write(value)
	}

	if buffer.Len() <= 0 {
		return nil
	}

	return e.runner.Run(e.Timeout.Duration, e.Command, &buffer)
}

// RunCloser provides an interface for running and ending an exec.Cmd before a
// timeout is reached.
type RunCloser interface {
	Run(time.Duration, []string, io.Reader) error
	Close() error
}

// CommandRunner runs a command with the ability to kill the process before the timeout.
type CommandRunner struct {
	cmd *exec.Cmd
}

// Run runs the command.
func (c *CommandRunner) Run(timeout time.Duration, command []string, buffer io.Reader) error {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = buffer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := internal.RunTimeout(cmd, timeout); err != nil {
		if err == internal.TimeoutErr {
			return fmt.Errorf("%q timed out and was killed", command)
		}

		s := stderr.String()
		if s != "" {
			log.Printf("E! [outputs.exec] Command error: %q", s)
		}

		status, _ := internal.ExitStatus(err)
		return fmt.Errorf("%q exited %d with %s", command, status, err.Error())
	}
	c.cmd = cmd

	return nil
}

// Close kills the process if it is still running.
func (c *CommandRunner) Close() error {
	if c.cmd == nil {
		return nil
	}

	return c.cmd.Process.Kill()
}

func init() {
	outputs.Add("exec", func() telegraf.Output {
		return &Exec{
			runner:  &CommandRunner{},
			Timeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}

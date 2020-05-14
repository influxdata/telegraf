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

const maxStderrBytes = 512

// Exec defines the exec output plugin.
type Exec struct {
	Command []string          `toml:"command"`
	Timeout internal.Duration `toml:"timeout"`

	runner     Runner
	serializer serializers.Serializer
}

var sampleConfig = `
  ## Command to ingest metrics via stdin.
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

// Connect satisfies the Output interface.
func (e *Exec) Connect() error {
	return nil
}

// Close satisfies the Output interface.
func (e *Exec) Close() error {
	return nil
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
	serializedMetrics, err := e.serializer.SerializeBatch(metrics)
	if err != nil {
		return err
	}
	buffer.Write(serializedMetrics)

	if buffer.Len() <= 0 {
		return nil
	}

	return e.runner.Run(e.Timeout.Duration, e.Command, &buffer)
}

// Runner provides an interface for running exec.Cmd.
type Runner interface {
	Run(time.Duration, []string, io.Reader) error
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

	err := internal.RunTimeout(cmd, timeout)
	s := stderr

	if err != nil {
		if err == internal.TimeoutErr {
			return fmt.Errorf("%q timed out and was killed", command)
		}

		if s.Len() > 0 {
			log.Printf("E! [outputs.exec] Command error: %q", truncate(s))
		}

		if status, ok := internal.ExitStatus(err); ok {
			return fmt.Errorf("%q exited %d with %s", command, status, err.Error())
		}

		return fmt.Errorf("%q failed with %s", command, err.Error())
	}

	c.cmd = cmd

	return nil
}

func truncate(buf bytes.Buffer) string {
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
		buf.WriteString("...")
	}
	return buf.String()
}

func init() {
	outputs.Add("exec", func() telegraf.Output {
		return &Exec{
			runner:  &CommandRunner{},
			Timeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}

package exec

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gonuts/go-shellquote"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/errchan"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

var sampleConfig = `
  ## Commands array
  commands = [
    "/tmp/test.sh",
    "/usr/bin/mycollector --foo=bar",
    "/tmp/collect_*.sh"
  ]

  ## Timeout for each command to complete.
  timeout = "5s"

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

type Exec struct {
	Commands []string
	Timeout  internal.Duration

	serializer serializers.Serializer

	runner  Runner
	errChan chan error
}

func NewExec() *Exec {
	return &Exec{
		runner:  CommandRunner{},
		Timeout: internal.Duration{Duration: time.Second * 5},
	}
}

type Runner interface {
	Run(*Exec, string, bytes.Buffer) error
}

type CommandRunner struct{}

func (c CommandRunner) Run(e *Exec, command string, buffer bytes.Buffer) error {
	split_cmd, err := shellquote.Split(command)
	if err != nil || len(split_cmd) == 0 {
		return fmt.Errorf("exec: unable to parse command, %s", err)
	}

	cmd := exec.Command(split_cmd[0], split_cmd[1:]...)
	cmd.Stdin = &buffer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := internal.RunTimeout(cmd, e.Timeout.Duration); err != nil {
		s := stderr.String()
		if s != "" {
			log.Printf("D! Command error: %s\n", s)
		}

		return fmt.Errorf("exec: %s for command '%s'", err, command)
	}

	return nil
}

func (e *Exec) Description() string {
	return "Send Telegraf metrics to one or more commands that can input from stdin"
}

func (e *Exec) SampleConfig() string {
	return sampleConfig
}

func (e *Exec) SetSerializer(serializer serializers.Serializer) {
	e.serializer = serializer
}

func (e *Exec) Connect() error {
	return nil
}

func (e *Exec) Close() error {
	return nil
}

func (e *Exec) ProcessCommand(command string, buffer bytes.Buffer, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := e.runner.Run(e, command, buffer); err != nil {
		e.errChan <- err
		return
	}
}

func (e *Exec) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	var buffer bytes.Buffer
	for _, metric := range metrics {
		values, err := e.serializer.Serialize(metric)
		if err != nil {
			return err
		}
		for _, value := range values {
			buffer.WriteString(value)
			buffer.WriteString("\n")
		}
	}

	// Lifted from 'plugins/inputs/exec/exec.go:Gather'
	commands := make([]string, 0, len(e.Commands))
	for _, pattern := range e.Commands {
		cmdAndArgs := strings.SplitN(pattern, " ", 2)
		if len(cmdAndArgs) == 0 {
			continue
		}

		matches, err := filepath.Glob(cmdAndArgs[0])
		if err != nil {
			return err
		}

		if len(matches) == 0 {
			// There were no matches with the glob pattern, so let's assume
			// that the command is in PATH and just run it as it is
			commands = append(commands, pattern)
		} else {
			// There were matches, so we'll append each match together with
			// the arguments to the commands slice
			for _, match := range matches {
				if len(cmdAndArgs) == 1 {
					commands = append(commands, match)
				} else {
					commands = append(commands,
						strings.Join([]string{match, cmdAndArgs[1]}, " "))
				}
			}
		}
	}

	var wg sync.WaitGroup
	errChan := errchan.New(len(commands))
	e.errChan = errChan.C

	wg.Add(len(commands))
	for _, command := range commands {
		go e.ProcessCommand(command, buffer, &wg)
	}
	wg.Wait()
	return errChan.Error()
}

func init() {
	outputs.Add("exec", func() telegraf.Output { return NewExec() })
}

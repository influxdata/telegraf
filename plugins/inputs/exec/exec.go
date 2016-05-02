package exec

import (
	"bytes"
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/gonuts/go-shellquote"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/nagios"
)

const sampleConfig = `
  ## Commands array
  commands = ["/tmp/test.sh", "/usr/bin/mycollector --foo=bar"]

  ## Timeout for each command to complete.
  timeout = "5s"

  ## measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

type Exec struct {
	Commands []string
	Command  string
	Timeout  internal.Duration

	parser parsers.Parser

	wg sync.WaitGroup

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
	Run(*Exec, string, telegraf.Accumulator) ([]byte, error)
}

type CommandRunner struct{}

func AddNagiosState(exitCode error, acc telegraf.Accumulator) error {
	nagiosState := 0
	if exitCode != nil {
		exiterr, ok := exitCode.(*exec.ExitError)
		if ok {
			status, ok := exiterr.Sys().(syscall.WaitStatus)
			if ok {
				nagiosState = status.ExitStatus()
			} else {
				return fmt.Errorf("exec: unable to get nagios plugin exit code")
			}
		} else {
			return fmt.Errorf("exec: unable to get nagios plugin exit code")
		}
	}
	fields := map[string]interface{}{"state": nagiosState}
	acc.AddFields("nagios_state", fields, nil)
	return nil
}

func (c CommandRunner) Run(
	e *Exec,
	command string,
	acc telegraf.Accumulator,
) ([]byte, error) {
	split_cmd, err := shellquote.Split(command)
	if err != nil || len(split_cmd) == 0 {
		return nil, fmt.Errorf("exec: unable to parse command, %s", err)
	}

	cmd := exec.Command(split_cmd[0], split_cmd[1:]...)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := internal.RunTimeout(cmd, e.Timeout.Duration); err != nil {
		switch e.parser.(type) {
		case *nagios.NagiosParser:
			AddNagiosState(err, acc)
		default:
			return nil, fmt.Errorf("exec: %s for command '%s'", err, command)
		}
	} else {
		switch e.parser.(type) {
		case *nagios.NagiosParser:
			AddNagiosState(nil, acc)
		}
	}

	return out.Bytes(), nil
}

func (e *Exec) ProcessCommand(command string, acc telegraf.Accumulator) {
	defer e.wg.Done()

	out, err := e.runner.Run(e, command, acc)
	if err != nil {
		e.errChan <- err
		return
	}

	metrics, err := e.parser.Parse(out)
	if err != nil {
		e.errChan <- err
	} else {
		for _, metric := range metrics {
			acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
		}
	}
}

func (e *Exec) SampleConfig() string {
	return sampleConfig
}

func (e *Exec) Description() string {
	return "Read metrics from one or more commands that can output to stdout"
}

func (e *Exec) SetParser(parser parsers.Parser) {
	e.parser = parser
}

func (e *Exec) Gather(acc telegraf.Accumulator) error {
	// Legacy single command support
	if e.Command != "" {
		e.Commands = append(e.Commands, e.Command)
		e.Command = ""
	}

	e.errChan = make(chan error, len(e.Commands))

	e.wg.Add(len(e.Commands))
	for _, command := range e.Commands {
		go e.ProcessCommand(command, acc)
	}
	e.wg.Wait()

	select {
	default:
		close(e.errChan)
		return nil
	case err := <-e.errChan:
		close(e.errChan)
		return err
	}

}

func init() {
	inputs.Add("exec", func() telegraf.Input {
		return NewExec()
	})
}

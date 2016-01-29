package exec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/gonuts/go-shellquote"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  # NOTE This plugin only reads numerical measurements, strings and booleans
  # will be ignored.

  # the command to run
  command = "/usr/bin/mycollector --foo=bar"

  # measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"
`

type Exec struct {
	Command string

	runner Runner
}

type Runner interface {
	Run(*Exec) ([]byte, error)
}

type CommandRunner struct{}

func (c CommandRunner) Run(e *Exec) ([]byte, error) {
	split_cmd, err := shellquote.Split(e.Command)
	if err != nil || len(split_cmd) == 0 {
		return nil, fmt.Errorf("exec: unable to parse command, %s", err)
	}

	cmd := exec.Command(split_cmd[0], split_cmd[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("exec: %s for command '%s'", err, e.Command)
	}

	return out.Bytes(), nil
}

func NewExec() *Exec {
	return &Exec{runner: CommandRunner{}}
}

func (e *Exec) SampleConfig() string {
	return sampleConfig
}

func (e *Exec) Description() string {
	return "Read flattened metrics from one or more commands that output JSON to stdout"
}

func (e *Exec) Gather(acc telegraf.Accumulator) error {
	out, err := e.runner.Run(e)
	if err != nil {
		return err
	}

	var jsonOut interface{}
	err = json.Unmarshal(out, &jsonOut)
	if err != nil {
		return fmt.Errorf("exec: unable to parse output of '%s' as JSON, %s",
			e.Command, err)
	}

	f := internal.JSONFlattener{}
	err = f.FlattenJSON("", jsonOut)
	if err != nil {
		return err
	}

	acc.AddFields("exec", f.Fields, nil)
	return nil
}

func init() {
	inputs.Add("exec", func() telegraf.Input {
		return NewExec()
	})
}

package exec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gonuts/go-shellquote"
	"github.com/influxdb/telegraf/plugins"
	"os/exec"
	"sync"
)

const sampleConfig = `
	# specify commands via an array of tables
	[[exec.commands]]
	# the command to run
	command = "/usr/bin/mycollector --foo=bar"

	# name of the command (used as a prefix for measurements)
	name = "mycollector"
`

type Command struct {
	Command string
	Name    string
}

type Exec struct {
	Commands []*Command
	runner   Runner
}

type Runner interface {
	Run(string, ...string) ([]byte, error)
}

type CommandRunner struct {
}

func NewExec() *Exec {
	return &Exec{runner: CommandRunner{}}
}

func (c CommandRunner) Run(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("exec: %s for command '%s'", err, command)
	}

	return out.Bytes(), nil
}

func (e *Exec) SampleConfig() string {
	return sampleConfig
}

func (e *Exec) Description() string {
	return "Read flattened metrics from one or more commands that output JSON to stdout"
}

func (e *Exec) Gather(acc plugins.Accumulator) error {
	var wg sync.WaitGroup

	var outerr error

	for _, c := range e.Commands {
		wg.Add(1)
		go func(c *Command, acc plugins.Accumulator) {
			defer wg.Done()
			outerr = e.gatherCommand(c, acc)
		}(c, acc)
	}

	wg.Wait()

	return outerr
}

func (e *Exec) gatherCommand(c *Command, acc plugins.Accumulator) error {
	words, err := shellquote.Split(c.Command)
	if err != nil || len(words) == 0 {
		return fmt.Errorf("exec: unable to parse command, %s", err)
	}

	out, err := e.runner.Run(words[0], words[1:]...)
	if err != nil {
		return err
	}

	var jsonOut interface{}
	err = json.Unmarshal(out, &jsonOut)
	if err != nil {
		return fmt.Errorf("exec: unable to parse output of '%s' as JSON, %s", c.Command, err)
	}

	processResponse(acc, c.Name, map[string]string{}, jsonOut)
	return nil
}

func processResponse(acc plugins.Accumulator, prefix string, tags map[string]string, v interface{}) {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, v := range t {
			processResponse(acc, prefix+"_"+k, tags, v)
		}
	case float64:
		acc.Add(prefix, v, tags)
	}
}

func init() {
	plugins.Add("exec", func() plugins.Plugin {
		return NewExec()
	})
}

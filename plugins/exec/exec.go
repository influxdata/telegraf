package exec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gonuts/go-shellquote"

	"github.com/influxdb/telegraf/internal"
	"github.com/influxdb/telegraf/plugins"
)

const sampleConfig = `
  # specify commands via an array of tables
  [[plugins.exec.commands]]
  # the command to run
  command = "/usr/bin/mycollector --foo=bar"

  # name of the command (used as a prefix for measurements)
  name = "mycollector"

  # Only run this command if it has been at least this many
  # seconds since it last ran
  interval = 10
`

type Exec struct {
	Commands []*Command
	runner   Runner
	clock    Clock
}

type Command struct {
	Command   string
	Name      string
	Interval  int
	lastRunAt time.Time
}

type Runner interface {
	Run(*Command) ([]byte, error)
}

type Clock interface {
	Now() time.Time
}

type CommandRunner struct{}

type RealClock struct{}

func (c CommandRunner) Run(command *Command) ([]byte, error) {
	command.lastRunAt = time.Now()
	split_cmd, err := shellquote.Split(command.Command)
	if err != nil || len(split_cmd) == 0 {
		return nil, fmt.Errorf("exec: unable to parse command, %s", err)
	}

	cmd := exec.Command(split_cmd[0], split_cmd[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("exec: %s for command '%s'", err, command.Command)
	}

	return out.Bytes(), nil
}

func (c RealClock) Now() time.Time {
	return time.Now()
}

func NewExec() *Exec {
	return &Exec{runner: CommandRunner{}, clock: RealClock{}}
}

func (e *Exec) SampleConfig() string {
	return sampleConfig
}

func (e *Exec) Description() string {
	return "Read flattened metrics from one or more commands that output JSON to stdout"
}

func (e *Exec) Gather(acc plugins.Accumulator) error {
	var wg sync.WaitGroup

	errorChannel := make(chan error, len(e.Commands))

	for _, c := range e.Commands {
		wg.Add(1)
		go func(c *Command, acc plugins.Accumulator) {
			defer wg.Done()
			err := e.gatherCommand(c, acc)
			if err != nil {
				errorChannel <- err
			}
		}(c, acc)
	}

	wg.Wait()
	close(errorChannel)

	// Get all errors and return them as one giant error
	errorStrings := []string{}
	for err := range errorChannel {
		errorStrings = append(errorStrings, err.Error())
	}

	if len(errorStrings) == 0 {
		return nil
	}
	return errors.New(strings.Join(errorStrings, "\n"))
}

func (e *Exec) gatherCommand(c *Command, acc plugins.Accumulator) error {
	secondsSinceLastRun := 0.0

	if c.lastRunAt.Unix() == 0 { // means time is uninitialized
		secondsSinceLastRun = math.Inf(1)
	} else {
		secondsSinceLastRun = (e.clock.Now().Sub(c.lastRunAt)).Seconds()
	}

	if secondsSinceLastRun >= float64(c.Interval) {
		out, err := e.runner.Run(c)
		if err != nil {
			return err
		}

		var jsonOut interface{}
		err = json.Unmarshal(out, &jsonOut)
		if err != nil {
			return fmt.Errorf("exec: unable to parse output of '%s' as JSON, %s",
				c.Command, err)
		}

		f := internal.JSONFlattener{}
		err = f.FlattenJSON("", jsonOut)
		if err != nil {
			return err
		}

		var msrmnt_name string
		if c.Name == "" {
			msrmnt_name = "exec"
		} else {
			msrmnt_name = "exec_" + c.Name
		}
		acc.AddFields(msrmnt_name, f.Fields, nil)
	}
	return nil
}

func init() {
	plugins.Add("exec", func() plugins.Plugin {
		return NewExec()
	})
}

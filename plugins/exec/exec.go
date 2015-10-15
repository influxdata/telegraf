package exec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gonuts/go-shellquote"
	"github.com/influxdb/telegraf/plugins"
	"math"
	"os/exec"
	"sync"
	"time"
)

const sampleConfig = `
  # specify commands via an array of tables
  [[exec.commands]]
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
			return fmt.Errorf("exec: unable to parse output of '%s' as JSON, %s", c.Command, err)
		}

		processResponse(acc, c.Name, map[string]string{}, jsonOut)
	}
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

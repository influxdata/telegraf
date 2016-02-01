package exec

import (
	"bytes"
	"fmt"
	"os/exec"
	"sync"

	"github.com/gonuts/go-shellquote"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/encoding"
	"github.com/influxdata/telegraf/plugins/inputs"

	_ "github.com/influxdata/telegraf/internal/encoding/graphite"
	_ "github.com/influxdata/telegraf/internal/encoding/influx"
	_ "github.com/influxdata/telegraf/internal/encoding/json"
)

const sampleConfig = `
  # Shell/commands array
  # compatible with old version
  # we can still use the old command configuration
  # command = "/usr/bin/mycollector --foo=bar"
  commands = ["/tmp/test.sh","/tmp/test2.sh"]

  # Data format to consume. This can be "json", "influx" or "graphite" (line-protocol)
  # NOTE json only reads numerical measurements, strings and booleans are ignored.
  data_format = "json"

  # measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ### Below configuration will be used for data_format = "graphite", can be ignored for other data_format
  ### If matching multiple measurement files, this string will be used to join the matched values.
  separator = "."

  ### Each template line requires a template pattern.  It can have an optional
  ### filter before the template and separated by spaces.  It can also have optional extra
  ### tags following the template.  Multiple tags should be separated by commas and no spaces
  ### similar to the line protocol format.  The can be only one default template.
  ### Templates support below format:
  ### 1. filter + template
  ### 2. filter + template + extra tag
  ### 3. filter + template with field key
  ### 4. default template
  templates = [
    "*.app env.service.resource.measurement",
    "stats.* .host.measurement* region=us-west,agent=sensu",
    "stats2.* .host.measurement.field",
    "measurement*"
 ]
`

type Exec struct {
	Commands   []string
	Command    string
	DataFormat string

	Separator string
	Templates []string

	encodingParser encoding.Parser

	initedConfig bool

	wg sync.WaitGroup
	sync.Mutex

	runner Runner
	errc   chan error
}

type Runner interface {
	Run(*Exec, string) ([]byte, error)
}

type CommandRunner struct{}

func (c CommandRunner) Run(e *Exec, command string) ([]byte, error) {
	split_cmd, err := shellquote.Split(command)
	if err != nil || len(split_cmd) == 0 {
		return nil, fmt.Errorf("exec: unable to parse command, %s", err)
	}

	cmd := exec.Command(split_cmd[0], split_cmd[1:]...)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("exec: %s for command '%s'", err, command)
	}

	return out.Bytes(), nil
}

func NewExec() *Exec {
	return &Exec{runner: CommandRunner{}}
}

func (e *Exec) ProcessCommand(command string, acc telegraf.Accumulator) {
	defer e.wg.Done()

	out, err := e.runner.Run(e, command)
	if err != nil {
		e.errc <- err
		return
	}

	metrics, err := e.encodingParser.Parse(out)
	if err != nil {
		e.errc <- err
	} else {
		for _, metric := range metrics {
			acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
		}
	}
}

func (e *Exec) initConfig() error {
	e.Lock()
	defer e.Unlock()

	if e.Command != "" && len(e.Commands) < 1 {
		e.Commands = []string{e.Command}
	}

	if e.DataFormat == "" {
		e.DataFormat = "json"
	}

	var err error

	configs := make(map[string]interface{})
	configs["Separator"] = e.Separator
	configs["Templates"] = e.Templates

	e.encodingParser, err = encoding.NewParser(e.DataFormat, configs)

	if err != nil {
		return fmt.Errorf("exec configuration is error: %s ", err.Error())
	}

	return nil
}

func (e *Exec) SampleConfig() string {
	return sampleConfig
}

func (e *Exec) Description() string {
	return "Read metrics from one or more commands that can output JSON, influx or graphite line protocol to stdout"
}

func (e *Exec) Gather(acc telegraf.Accumulator) error {

	if !e.initedConfig {
		if err := e.initConfig(); err != nil {
			return err
		}
		e.initedConfig = true
	}

	e.Lock()
	e.errc = make(chan error, 10)
	e.Unlock()

	for _, command := range e.Commands {
		e.wg.Add(1)
		go e.ProcessCommand(command, acc)
	}
	e.wg.Wait()

	select {
	default:
		close(e.errc)
		return nil
	case err := <-e.errc:
		close(e.errc)
		return err
	}

}

func init() {
	inputs.Add("exec", func() telegraf.Input {
		return NewExec()
	})
}

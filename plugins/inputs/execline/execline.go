package execline

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gonuts/go-shellquote"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  # NOTE This execline plugin only reads numerical measurements output by commands, 
  # strings and booleans ill be ignored.
  commands = ["/tmp/test.sh","/tmp/test2.sh"] # the bind address

  ### If matching multiple measurement files, this string will be used to join the matched values.
  separator = "."
  
  ### Default tags that will be added to all metrics.  These can be overridden at the template level
  ### or by tags extracted from metric
  tags = ["region=north-china", "zone=1c"]
  
  ### Each template line requires a template pattern.  It can have an optional
  ### filter before the template and separated by spaces.  It can also have optional extra
  ### tags following the template.  Multiple tags should be separated by commas and no spaces
  ### similar to the line protocol format.  The can be only one default template.
  ### Templates support below format:
  ### filter + template
  ### filter + template + extra tag
  ### filter + template with field key
  ### default template. Ignore the first graphite component "servers"
  templates = [
    "*.app env.service.resource.measurement",
    "stats.* .host.measurement* region=us-west,agent=sensu",
    "stats2.* .host.measurement.field",
    "measurement*"
 ]
`

type ExecLine struct {
	Commands  []string
	Separator string
	Tags      []string
	Templates []string

	parser *Parser
	config *Config

	initedConfig bool

	wg sync.WaitGroup
	sync.Mutex
}

func (e *ExecLine) Run(command string, acc telegraf.Accumulator) error {
	defer e.wg.Done()

	split_cmd, err := shellquote.Split(command)
	if err != nil || len(split_cmd) == 0 {
		return fmt.Errorf("execline: unable to parse command, %s", err)
	}

	cmd := exec.Command(split_cmd[0], split_cmd[1:]...)
	name := strings.Replace(filepath.Base(cmd.Path), "/", "_", -1)
	name = strings.Replace(name, ".", "_", -1)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execline: %s for command '%s'", err, command)
	}

	reader := bufio.NewReader(&out)
	for {
		// Read up to the next newline.
		buf, err := reader.ReadBytes('\n')
		if err != nil {
			return err
		}

		// Trim the buffer, even though there should be no padding
		line := strings.TrimSpace(string(buf))
		e.handleLine(name, acc, line)
	}

	return nil
}

func (e *ExecLine) handleLine(name string, acc telegraf.Accumulator, line string) {
	if line == "" {
		return
	}

	// Parse it.
	metric, err := e.parser.Parse(line)
	if err != nil {
		switch err := err.(type) {
		case *UnsupposedValueError:
			// Graphite ignores NaN values with no error.
			if math.IsNaN(err.Value) {
				return
			}
		}
		fmt.Errorf("unable to parse line: %s: %s", line, err)
		return
	}

	acc.AddFields(name+"."+metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
}

func (e *ExecLine) initConfig() error {
	e.Lock()
	defer e.Unlock()

	c := &Config{
		Commands:  e.Commands,
		Separator: e.Separator,
		Tags:      e.Tags,
		Templates: e.Templates,
	}
	c.WithDefaults()
	if err := c.Validate(); err != nil {
		return fmt.Errorf("ExecLine configuration is error! ", err.Error())

	}
	e.config = c

	parser, err := NewParserWithOptions(Options{
		Templates:   e.config.Templates,
		DefaultTags: e.config.DefaultTags(),
		Separator:   e.config.Separator})
	if err != nil {
		return fmt.Errorf("ExecLine input parser config is error! ", err.Error())
	}
	e.parser = parser

	return nil
}

func (e *ExecLine) SampleConfig() string {
	return sampleConfig
}

func (e *ExecLine) Description() string {
	return "Read metrics from one or more commands that output graphite line protocol to stdout"
}

func (e *ExecLine) Gather(acc telegraf.Accumulator) error {

	if !e.initedConfig {
		if err := e.initConfig(); err != nil {
			return err
		}
		e.initedConfig = true
	}

	for _, command := range e.Commands {
		e.wg.Add(1)
		go e.Run(command, acc)
	}

	e.wg.Wait()
	return nil
}

func init() {
	inputs.Add("execline", func() telegraf.Input {
		return &ExecLine{}
	})
}

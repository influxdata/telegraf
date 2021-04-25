package exec

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/nagios"
	"github.com/kballard/go-shellquote"
)

const sampleConfig = `
  ## Commands array
  commands = [
    "/tmp/test.sh",
    "/usr/bin/mycollector --foo=bar",
    "/tmp/collect_*.sh"
  ]

  # Extended commands array with support for custom tags per command
  commands_extended = [
    { command = "/tmp/test.sh",tags = [ ["tag_1", "metricA"], ["tag_2", "custom"] ] },
    { command = "/usr/bin/mycollector --foo=bar", tags = [ ["tag_1", "metricB"], ["tag_2", "custom"] ] }
  ]

  ## Timeout for each command to complete.
  timeout = "5s"

  ## measurement name suffix (for separating different commands)
  name_suffix = "_mycollector"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

const MaxStderrBytes int = 512

// CommandExtended defines a command object with custom data
type CommandExtended struct {
	Command string
	Tags    [][]string
}

type Exec struct {
	Commands         []string          `toml:"commands"`
	CommandsExtended []CommandExtended `toml:"commands_extended"`
	Command          string            `toml:"command"`
	Timeout          config.Duration   `toml:"timeout"`

	parser parsers.Parser

	runner Runner
	Log    telegraf.Logger `toml:"-"`
}

func NewExec() *Exec {
	return &Exec{
		runner:  CommandRunner{},
		Timeout: config.Duration(time.Second * 5),
	}
}

type Runner interface {
	Run(string, time.Duration) ([]byte, []byte, error)
}

type CommandRunner struct{}

func (c CommandRunner) Run(
	command string,
	timeout time.Duration,
) ([]byte, []byte, error) {
	splitCmd, err := shellquote.Split(command)
	if err != nil || len(splitCmd) == 0 {
		return nil, nil, fmt.Errorf("exec: unable to parse command, %s", err)
	}

	cmd := exec.Command(splitCmd[0], splitCmd[1:]...)

	var (
		out    bytes.Buffer
		stderr bytes.Buffer
	)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	runErr := internal.RunTimeout(cmd, timeout)

	out = removeWindowsCarriageReturns(out)
	if stderr.Len() > 0 && !telegraf.Debug {
		stderr = removeWindowsCarriageReturns(stderr)
		stderr = c.truncate(stderr)
	}

	return out.Bytes(), stderr.Bytes(), runErr
}

func (c CommandRunner) truncate(buf bytes.Buffer) bytes.Buffer {
	// Limit the number of bytes.
	didTruncate := false
	if buf.Len() > MaxStderrBytes {
		buf.Truncate(MaxStderrBytes)
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
		//nolint:errcheck,revive // Will always return nil or panic
		buf.WriteString("...")
	}
	return buf
}

// removeWindowsCarriageReturns removes all carriage returns from the input if the
// OS is Windows. It does not return any errors.
func removeWindowsCarriageReturns(b bytes.Buffer) bytes.Buffer {
	if runtime.GOOS == "windows" {
		var buf bytes.Buffer
		for {
			byt, err := b.ReadBytes(0x0D)
			byt = bytes.TrimRight(byt, "\x0d")
			if len(byt) > 0 {
				_, _ = buf.Write(byt)
			}
			if err == io.EOF {
				return buf
			}
		}
	}
	return b
}

func (e *Exec) ProcessCommand(command CommandExtended, acc telegraf.Accumulator, wg *sync.WaitGroup) {
	defer wg.Done()
	_, isNagios := e.parser.(*nagios.NagiosParser)

	out, errbuf, runErr := e.runner.Run(command.Command, time.Duration(e.Timeout))
	if !isNagios && runErr != nil {
		err := fmt.Errorf("exec: %s for command '%s': %s", runErr, command, string(errbuf))
		acc.AddError(err)
		return
	}

	metrics, err := e.parser.Parse(out)
	if err != nil {
		acc.AddError(err)
		return
	}

	if isNagios {
		metrics, err = nagios.TryAddState(runErr, metrics)
		if err != nil {
			e.Log.Errorf("Failed to add nagios state: %s", err)
		}
	}

	for _, m := range metrics {
		for _, tag := range command.Tags {
			m.AddTag(tag[0], tag[1])
		}
		acc.AddMetric(m)
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
	var wg sync.WaitGroup
	//transform all commands into extended ones to unify them
	for _, cmd := range e.Commands {
		e.CommandsExtended = append(e.CommandsExtended, CommandExtended{Command: cmd})
	}

	// Legacy single command support
	if e.Command != "" {
		e.CommandsExtended = append(e.CommandsExtended, CommandExtended{Command: e.Command})
		e.Command = ""
	}

	commands := make([]CommandExtended, 0, len(e.CommandsExtended))
	for _, cmd := range e.CommandsExtended {
		cmdAndArgs := strings.SplitN(cmd.Command, " ", 2)
		if len(cmdAndArgs) == 0 {
			continue
		}

		matches, err := filepath.Glob(cmdAndArgs[0])
		if err != nil {
			acc.AddError(err)
			continue
		}

		if len(matches) == 0 {
			// There were no matches with the glob cmd, so let's assume
			// that the command is in PATH and just run it as it is
			commands = append(commands, cmd)
		} else {
			// There were matches, so we'll append each match together with
			// the arguments to the commands slice
			for _, match := range matches {
				if len(cmdAndArgs) == 1 {
					commands = append(commands, CommandExtended{
						Command: match,
						Tags:    cmd.Tags,
					})
				} else {
					commands = append(commands, CommandExtended{
						Command: strings.Join([]string{match, cmdAndArgs[1]}, " "),
						Tags:    cmd.Tags,
					})
				}
			}
		}
	}

	wg.Add(len(commands))
	for _, command := range commands {
		go e.ProcessCommand(command, acc, &wg)
	}
	wg.Wait()
	return nil
}

func (e *Exec) Init() error {
	return nil
}

func init() {
	inputs.Add("exec", func() telegraf.Input {
		return NewExec()
	})
}

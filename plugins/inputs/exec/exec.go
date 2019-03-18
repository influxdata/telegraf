package exec

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/kballard/go-shellquote"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/nagios"
)

const sampleConfig = `
  ## Commands array
  commands = [
    "/tmp/test.sh",
    "/usr/bin/mycollector --foo=bar",
    "/tmp/collect_*.sh"
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

const MaxStderrBytes = 512

type Exec struct {
	Commands []string
	Command  string
	Timeout  internal.Duration

	parser parsers.Parser

	runner Runner
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

	var (
		out    bytes.Buffer
		stderr bytes.Buffer
	)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	_, isNagiosParser := e.parser.(*nagios.NagiosParser)

	err = internal.RunTimeout(cmd, e.Timeout.Duration)
	nagiosErr := err
	if err != nil && !isNagiosParser {
		var errMessage = ""
		if stderr.Len() > 0 {
			stderr = removeCarriageReturns(stderr)
			// Limit the number of bytes.
			didTruncate := false
			if stderr.Len() > MaxStderrBytes {
				stderr.Truncate(MaxStderrBytes)
				didTruncate = true
			}
			if i := bytes.IndexByte(stderr.Bytes(), '\n'); i > 0 {
				// Only show truncation if the newline wasn't the last character.
				if i < stderr.Len()-1 {
					didTruncate = true
				}
				stderr.Truncate(i)
			}
			if didTruncate {
				stderr.WriteString("...")
			}

			errMessage = fmt.Sprintf(": %s", stderr.String())
		}
		return nil, fmt.Errorf("exec: %s for command '%s'%s", err, command, errMessage)
	}

	out = removeCarriageReturns(out)

	if isNagiosParser {
		exitCode, err := nagios.GetExitCode(nagiosErr)
		if err != nil {
			return nil, fmt.Errorf("exec: get exit code: %s", err)
		}
		nagios.AppendExitCode(&out, exitCode)
	}

	return out.Bytes(), nil
}

// removeCarriageReturns removes all carriage returns from the input if the
// OS is Windows. It does not return any errors.
func removeCarriageReturns(b bytes.Buffer) bytes.Buffer {
	if runtime.GOOS == "windows" {
		var buf bytes.Buffer
		for {
			byt, er := b.ReadBytes(0x0D)
			end := len(byt)
			if nil == er {
				end -= 1
			}
			if nil != byt {
				buf.Write(byt[:end])
			} else {
				break
			}
			if nil != er {
				break
			}
		}
		b = buf
	}
	return b

}

func (e *Exec) ProcessCommand(command string, acc telegraf.Accumulator, wg *sync.WaitGroup) {
	defer wg.Done()

	out, err := e.runner.Run(e, command, acc)
	if err != nil {
		acc.AddError(err)
		return
	}

	metrics, err := e.parser.Parse(out)
	if err != nil {
		acc.AddError(err)
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
	var wg sync.WaitGroup
	// Legacy single command support
	if e.Command != "" {
		e.Commands = append(e.Commands, e.Command)
		e.Command = ""
	}

	commands := make([]string, 0, len(e.Commands))
	for _, pattern := range e.Commands {
		cmdAndArgs := strings.SplitN(pattern, " ", 2)
		if len(cmdAndArgs) == 0 {
			continue
		}

		matches, err := filepath.Glob(cmdAndArgs[0])
		if err != nil {
			acc.AddError(err)
			continue
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

	wg.Add(len(commands))
	for _, command := range commands {
		go e.ProcessCommand(command, acc, &wg)
	}
	wg.Wait()
	return nil
}

func init() {
	inputs.Add("exec", func() telegraf.Input {
		return NewExec()
	})
}

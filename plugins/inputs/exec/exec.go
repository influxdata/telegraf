//go:generate ../../../tools/readme_config_includer/generator
package exec

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kballard/go-shellquote"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/nagios"
)

//go:embed sample.conf
var sampleConfig string

var once sync.Once

const maxStderrBytes int = 512

type Exec struct {
	Commands    []interface{}   `toml:"commands"`
	Command     string          `toml:"command"`
	Environment []string        `toml:"environment"`
	IgnoreError bool            `toml:"ignore_error"`
	LogStdErr   bool            `toml:"log_stderr"`
	Timeout     config.Duration `toml:"timeout"`
	Log         telegraf.Logger `toml:"-"`

	parser telegraf.Parser

	runner runner
	cmds   [][]string

	// Allow post-processing of command exit codes
	exitCodeHandler   exitCodeHandlerFunc
	parseDespiteError bool
}

type exitCodeHandlerFunc func([]telegraf.Metric, error, []byte) []telegraf.Metric

type runner interface {
	run([]string) ([]byte, []byte, error)
}

type commandRunner struct {
	environment []string
	timeout     time.Duration
	debug       bool
}

func (*Exec) SampleConfig() string {
	return sampleConfig
}

func (e *Exec) Init() error {
	// Legacy single command support
	if e.Command != "" {
		e.Commands = append(e.Commands, e.Command)
	}

	if len(e.Commands) == 0 {
		return errors.New("no command specified")
	}

	e.cmds = make([][]string, 0, len(e.Commands))
	for _, raw := range e.Commands {
		switch c := raw.(type) {
		case string:
			// Legacy single string command setting
			if c == "" {
				return errors.New("command string cannot be empty")
			}

			// Convert the legacy command string to a string list and output
			// deprecation notice
			cmd, err := shellquote.Split(c)
			if err != nil {
				return fmt.Errorf("unable to parse command %q: %w", c, err)
			}
			if len(cmd) == 0 {
				return errors.New("command cannot be empty")
			}
			// Create the corresponding command in the new syntax to ease migration
			suggestion := make([]string, 0, len(cmd))
			for _, a := range cmd {
				suggestion = append(suggestion, strconv.Quote(a))
			}
			config.PrintOptionValueDeprecationNotice("inputs.exec", "command", c, telegraf.DeprecationInfo{
				Since:     "1.39.0",
				RemovalIn: "1.45.0",
				Notice:    fmt.Sprintf("Use array syntax instead: [%s]", strings.Join(suggestion, ",")),
			})
			e.cmds = append(e.cmds, cmd)
		case []string:
			if len(c) == 0 {
				return errors.New("command cannot be empty")
			}
			e.cmds = append(e.cmds, c)
		case []interface{}:
			if len(c) == 0 {
				return errors.New("command cannot be empty")
			}

			// Convert the entries to a string list
			converted := make([]string, 0, len(c))
			for _, r := range c {
				v, ok := r.(string)
				if !ok {
					return fmt.Errorf("command %v has invalid entry %v of type %T, expected string", raw, r, r)
				}
				converted = append(converted, v)
			}
			e.cmds = append(e.cmds, converted)
		default:
			return fmt.Errorf("command %v has invalid type %T, expected string list", raw, raw)
		}
	}

	e.runner = &commandRunner{
		environment: e.Environment,
		timeout:     time.Duration(e.Timeout),
		debug:       e.Log.Level().Includes(telegraf.Debug),
	}

	return nil
}

func (e *Exec) SetParser(parser telegraf.Parser) {
	e.parser = parser
	unwrapped, ok := parser.(*models.RunningParser)
	if ok {
		if _, ok := unwrapped.Parser.(*nagios.Parser); ok {
			e.exitCodeHandler = func(metrics []telegraf.Metric, err error, msg []byte) []telegraf.Metric {
				return nagios.AddState(err, msg, metrics)
			}
			e.parseDespiteError = true
		}
	}
}

func (e *Exec) Gather(acc telegraf.Accumulator) error {
	commands := e.updateRunners()

	var wg sync.WaitGroup
	for _, item := range commands {
		wg.Add(1)
		go func(c []string) {
			defer wg.Done()
			acc.AddError(e.processCommand(acc, c))
		}(item)
	}
	wg.Wait()
	return nil
}

func (e *Exec) updateRunners() [][]string {
	commands := make([][]string, 0, len(e.cmds))
	for _, cmd := range e.cmds {
		// Try to expand globbing expressions
		matches, err := filepath.Glob(cmd[0])
		if err != nil {
			e.Log.Errorf("Matching command %q failed: %v", cmd[0], err)
			continue
		}

		if len(matches) == 0 {
			// There were no matches with the glob pattern, so let's assume
			// the command is in PATH and just run it as it is
			commands = append(commands, cmd)
		} else {
			// There were matches, so we'll append each match together with
			// the arguments to the commands slice
			for _, match := range matches {
				expanded := make([]string, 0, len(cmd))
				expanded = append(expanded, match)
				expanded = append(expanded, cmd[1:]...)
				commands = append(commands, expanded)
			}
		}
	}

	return commands
}

func (e *Exec) processCommand(acc telegraf.Accumulator, cmd []string) error {
	out, errBuf, runErr := e.runner.run(cmd)
	if !e.IgnoreError && !e.parseDespiteError && runErr != nil {
		return fmt.Errorf("exec: %w for command %q: %s", runErr, strings.Join(cmd, " "), string(errBuf))
	}

	// Log output in stderr
	if e.LogStdErr && len(errBuf) > 0 {
		scanner := bufio.NewScanner(bytes.NewBuffer(errBuf))

		for scanner.Scan() {
			msg := scanner.Text()
			switch {
			case strings.TrimSpace(msg) == "":
				continue
			case strings.HasPrefix(msg, "E! "):
				e.Log.Error(msg[3:])
			case strings.HasPrefix(msg, "W! "):
				e.Log.Warn(msg[3:])
			case strings.HasPrefix(msg, "I! "):
				e.Log.Info(msg[3:])
			case strings.HasPrefix(msg, "D! "):
				e.Log.Debug(msg[3:])
			case strings.HasPrefix(msg, "T! "):
				e.Log.Trace(msg[3:])
			default:
				e.Log.Error(msg)
			}
		}

		if err := scanner.Err(); err != nil {
			acc.AddError(fmt.Errorf("error reading stderr: %w", err))
		}
	}

	metrics, err := e.parser.Parse(out)
	if err != nil {
		return err
	}

	if len(metrics) == 0 {
		once.Do(func() {
			e.Log.Debug(internal.NoMetricsCreatedMsg)
		})
	}

	if e.exitCodeHandler != nil {
		metrics = e.exitCodeHandler(metrics, runErr, errBuf)
	}

	for _, m := range metrics {
		acc.AddMetric(m)
	}

	return nil
}

func truncate(buf *bytes.Buffer) {
	// Limit the number of bytes.
	didTruncate := false
	if buf.Len() > maxStderrBytes {
		buf.Truncate(maxStderrBytes)
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
		buf.WriteString("...")
	}
}

func init() {
	inputs.Add("exec", func() telegraf.Input {
		return &Exec{
			Timeout: config.Duration(5 * time.Second),
		}
	})
}

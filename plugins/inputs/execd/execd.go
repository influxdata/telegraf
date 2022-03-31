package execd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/process"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/parsers/prometheus"
)

type Execd struct {
	Command      []string        `toml:"command"`
	Signal       string          `toml:"signal"`
	RestartDelay config.Duration `toml:"restart_delay"`
	Log          telegraf.Logger `toml:"-"`

	process *process.Process
	acc     telegraf.Accumulator
	parser  parsers.Parser
}

func (e *Execd) SetParser(parser parsers.Parser) {
	e.parser = parser
}

func (e *Execd) Start(acc telegraf.Accumulator) error {
	e.acc = acc
	var err error
	e.process, err = process.New(e.Command)
	if err != nil {
		return fmt.Errorf("error creating new process: %w", err)
	}
	e.process.Log = e.Log
	e.process.RestartDelay = time.Duration(e.RestartDelay)
	e.process.ReadStdoutFn = e.cmdReadOut
	e.process.ReadStderrFn = e.cmdReadErr

	if err = e.process.Start(); err != nil {
		// if there was only one argument, and it contained spaces, warn the user
		// that they may have configured it wrong.
		if len(e.Command) == 1 && strings.Contains(e.Command[0], " ") {
			e.Log.Warn("The inputs.execd Command contained spaces but no arguments. " +
				"This setting expects the program and arguments as an array of strings, " +
				"not as a space-delimited string. See the plugin readme for an example.")
		}
		return fmt.Errorf("failed to start process %s: %w", e.Command, err)
	}

	return nil
}

func (e *Execd) Stop() {
	e.process.Stop()
}

func (e *Execd) cmdReadOut(out io.Reader) {
	if _, isInfluxParser := e.parser.(*influx.Parser); isInfluxParser {
		// work around the lack of built-in streaming parser. :(
		e.cmdReadOutStream(out)
		return
	}

	_, isPrometheus := e.parser.(*prometheus.Parser)

	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		data := scanner.Bytes()
		if isPrometheus {
			data = append(data, []byte("\n")...)
		}

		metrics, err := e.parser.Parse(data)
		if err != nil {
			e.acc.AddError(fmt.Errorf("parse error: %w", err))
		}

		for _, metric := range metrics {
			e.acc.AddMetric(metric)
		}
	}

	if err := scanner.Err(); err != nil {
		e.acc.AddError(fmt.Errorf("error reading stdout: %w", err))
	}
}

func (e *Execd) cmdReadOutStream(out io.Reader) {
	parser := influx.NewStreamParser(out)

	for {
		metric, err := parser.Next()
		if err != nil {
			if err == influx.EOF {
				break // stream ended
			}
			if parseErr, isParseError := err.(*influx.ParseError); isParseError {
				// parse error.
				e.acc.AddError(parseErr)
				continue
			}
			// some non-recoverable error?
			e.acc.AddError(err)
			return
		}

		e.acc.AddMetric(metric)
	}
}

func (e *Execd) cmdReadErr(out io.Reader) {
	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		e.Log.Errorf("stderr: %q", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		e.acc.AddError(fmt.Errorf("error reading stderr: %w", err))
	}
}

func (e *Execd) Init() error {
	if len(e.Command) == 0 {
		return errors.New("no command specified")
	}
	return nil
}

func init() {
	inputs.Add("execd", func() telegraf.Input {
		return &Execd{
			Signal:       "none",
			RestartDelay: config.Duration(10 * time.Second),
		}
	})
}

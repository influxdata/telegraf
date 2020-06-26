package execd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/process"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
)

const sampleConfig = `
  ## Program to run as daemon
  command = ["telegraf-smartctl", "-d", "/dev/sda"]

  ## Define how the process is signaled on each collection interval.
  ## Valid values are:
  ##   "none"   : Do not signal anything.
  ##              The process must output metrics by itself.
  ##   "STDIN"   : Send a newline on STDIN.
  ##   "SIGHUP"  : Send a HUP signal. Not available on Windows.
  ##   "SIGUSR1" : Send a USR1 signal. Not available on Windows.
  ##   "SIGUSR2" : Send a USR2 signal. Not available on Windows.
  signal = "none"

  ## Delay before the process is restarted after an unexpected termination
  restart_delay = "10s"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

type Execd struct {
	Command      []string
	Signal       string
	RestartDelay config.Duration
	Log          telegraf.Logger

	process *process.Process
	acc     telegraf.Accumulator
	parser  parsers.Parser
}

func (e *Execd) SampleConfig() string {
	return sampleConfig
}

func (e *Execd) Description() string {
	return "Run executable as long-running input plugin"
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

	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		metrics, err := e.parser.Parse(scanner.Bytes())
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

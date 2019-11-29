package execd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/kballard/go-shellquote"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

const sampleConfig = `
  ## Program to run as daemon
  command = "telegraf-smartctl"

  ## Define how the process is signaled on each collection interval.

  ## Valid values are:
  ##   "none"   : Do not signal anything.
  ##              The process must output metrics by itself.
  ##   "STDIN"  : Send a newline on STDIN.
  ##   "SIGHUP" : Send a HUP signal. Not available on Windows.
  signal = "none"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

type Execd struct {
	Command string
	Signal  string

	acc     telegraf.Accumulator
	args    []string
	cmd     *exec.Cmd
	parser  parsers.Parser
	stdin   io.WriteCloser
	stopped bool
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

	args, err := shellquote.Split(e.Command)
	if err != nil || len(args) == 0 {
		return fmt.Errorf("unable to parse command: %s", err)
	}

	e.args = args
	go e.cmdRun()

	return nil
}

func (e *Execd) Stop() {
	e.stopped = true

	if e.cmd == nil || e.cmd.Process == nil {
		return
	}

	if err := e.cmd.Process.Kill(); err != nil {
		log.Printf("E! [inputs.execd] FATAL error killing process: %s", err)
	}
}

func (e *Execd) cmdRun() error {
	var wg sync.WaitGroup

	if len(e.args) > 1 {
		e.cmd = exec.Command(e.args[0], e.args[1:]...)
	} else {
		e.cmd = exec.Command(e.args[0])
	}

	stdin, err := e.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error opening stdin pipe: %s", err)
	}

	e.stdin = stdin

	stdout, err := e.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error opening stdout pipe: %s", err)
	}

	stderr, err := e.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error opening stderr pipe: %s", err)
	}

	log.Printf("D! [inputs.execd] Start program: %s", e.Command)

	err = e.cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting program: %s", err)
	}

	wg.Add(2)

	go func() {
		e.cmdReadOut(stdout)
		wg.Done()
	}()

	go func() {
		e.cmdReadErr(stderr)
		wg.Done()
	}()

	wg.Wait()
	e.cmd.Wait()

	if e.stopped {
		return nil
	}

	log.Printf("E! [inputs.execd] %s terminated. Restart in one second...", e.Command)

	go func() {
		<-time.After(time.Second)
		e.cmdRun()
	}()

	return nil
}

func (e *Execd) cmdReadOut(out io.Reader) {
	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		metrics, err := e.parser.Parse(scanner.Bytes())
		if err != nil {
			e.acc.AddError(fmt.Errorf("parse error: %s", err.Error()))
		}

		for _, metric := range metrics {
			e.acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
		}
	}

	if err := scanner.Err(); err != nil {
		e.acc.AddError(fmt.Errorf("Error reading stdout: %s", err.Error()))
	}
}

func (e *Execd) cmdReadErr(out io.Reader) {
	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		log.Printf("E! [inputs.execd] stderr output: %q", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		e.acc.AddError(fmt.Errorf("Error reading stderr: %s", err.Error()))
	}
}

func init() {
	inputs.Add("execd", func() telegraf.Input {
		return &Execd{
			Signal: "none",
		}
	})
}

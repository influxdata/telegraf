package execd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
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
	RestartDelay internal.Duration

	acc    telegraf.Accumulator
	cmd    *exec.Cmd
	parser parsers.Parser
	stdin  io.WriteCloser
	cancel context.CancelFunc
	wg     sync.WaitGroup
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

	if len(e.Command) == 0 {
		return fmt.Errorf("E! [inputs.execd] FATAL no command specified")
	}

	e.wg.Add(1)

	var ctx context.Context
	ctx, e.cancel = context.WithCancel(context.Background())

	go func() {
		e.cmdLoop(ctx)
		e.wg.Done()
	}()

	return nil
}

func (e *Execd) Stop() {
	e.cancel()
	e.wg.Wait()
}

func (e *Execd) cmdLoop(ctx context.Context) {
	for {
		// Use a buffered channel to ensure goroutine below can exit
		// if `ctx.Done` is selected and nothing reads on `done` anymore
		done := make(chan error, 1)
		go func() {
			done <- e.cmdRun()
		}()

		select {
		case <-ctx.Done():
			e.stdin.Close()
			// Immediately exit process but with a graceful shutdown
			// period before killing
			internal.WaitTimeout(e.cmd, 200*time.Millisecond)
			return
		case err := <-done:
			log.Printf("E! [inputs.execd] Process %s terminated: %s", e.Command, err)
		}

		log.Printf("E! [inputs.execd] Restarting in %s...", e.RestartDelay.Duration)

		select {
		case <-ctx.Done():
			return
		case <-time.After(e.RestartDelay.Duration):
			// Continue the loop and restart the process
		}
	}
}

func (e *Execd) cmdRun() error {
	var wg sync.WaitGroup

	if len(e.Command) > 1 {
		e.cmd = exec.Command(e.Command[0], e.Command[1:]...)
	} else {
		e.cmd = exec.Command(e.Command[0])
	}

	stdin, err := e.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("E! [inputs.execd] Error opening stdin pipe: %s", err)
	}

	e.stdin = stdin

	stdout, err := e.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("E! [inputs.execd] Error opening stdout pipe: %s", err)
	}

	stderr, err := e.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("E! [inputs.execd] Error opening stderr pipe: %s", err)
	}

	log.Printf("D! [inputs.execd] Starting process: %s", e.Command)

	err = e.cmd.Start()
	if err != nil {
		return fmt.Errorf("E! [inputs.execd] Error starting process: %s", err)
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
	return e.cmd.Wait()
}

func (e *Execd) cmdReadOut(out io.Reader) {
	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		metrics, err := e.parser.Parse(scanner.Bytes())
		if err != nil {
			e.acc.AddError(fmt.Errorf("E! [inputs.execd] Parse error: %s", err))
		}

		for _, metric := range metrics {
			e.acc.AddMetric(metric)
		}
	}

	if err := scanner.Err(); err != nil {
		e.acc.AddError(fmt.Errorf("E! [inputs.execd] Error reading stdout: %s", err))
	}
}

func (e *Execd) cmdReadErr(out io.Reader) {
	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		log.Printf("E! [inputs.execd] stderr: %q", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		e.acc.AddError(fmt.Errorf("E! [inputs.execd] Error reading stderr: %s", err))
	}
}

func init() {
	inputs.Add("execd", func() telegraf.Input {
		return &Execd{
			Signal:       "none",
			RestartDelay: internal.Duration{Duration: 10 * time.Second},
		}
	})
}
